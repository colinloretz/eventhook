package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/eventhook/eventhook/assets"
	"github.com/eventhook/eventhook/internal/api"
	"github.com/eventhook/eventhook/internal/config"
	"github.com/eventhook/eventhook/internal/store"
	"github.com/eventhook/eventhook/internal/worker"
	"github.com/eventhook/eventhook/migrations"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start EventHook in development mode with live event log",
	RunE:  runDev,
}

func runDev(_ *cobra.Command, _ []string) error {
	// Quiet structured logs — we print our own pretty output
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: false})
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	cfg := config.Load()
	// Force dev defaults
	if cfg.APIKey == "" {
		cfg.APIKey = "dev-api-key"
	}

	printBanner(cfg)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Ensure postgres + redis are reachable, start via docker if not
	if err := ensurePostgres(ctx, cfg); err != nil {
		return err
	}
	if err := ensureRedis(ctx, cfg); err != nil {
		return err
	}

	devPrint("✓", colorGreen, "postgres and redis ready")

	if err := store.RunMigrations(cfg.DatabaseURL, migrations.FS); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	devPrint("✓", colorGreen, "migrations up to date")

	st, err := store.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("postgres connect: %w", err)
	}
	defer st.Close()

	pool := worker.NewPool(st, cfg.WorkerCount)
	pool.Start(ctx)
	defer pool.Stop()

	srv := api.NewServer(cfg, st)
	srv.ServeDashboard(assets.Dashboard)

	fmt.Printf("\n%s  Dashboard → %shttp://localhost:%d/dashboard%s\n\n",
		colorDim+"─────────────────────────────────────────"+colorReset,
		colorCyan+colorBold, cfg.Port, colorReset)
	fmt.Printf("%sWaiting for events…%s\n\n", colorDim, colorReset)

	// Start live event tailer
	go tailEvents(ctx, cfg)

	return srv.Run()
}

// ensurePostgres verifies a real connection to postgres; starts docker on port
// 5433 if the configured URL isn't reachable (avoids conflicts with system postgres).
func ensurePostgres(ctx context.Context, cfg *config.Config) error {
	devPrint("→", colorDim, "checking postgres…")

	if pgConnectOK(ctx, cfg.DatabaseURL) {
		return nil
	}

	// Use 5433 to avoid colliding with a local system postgres on 5432.
	devURL := "postgres://eventhook:eventhook@localhost:5433/eventhook?sslmode=disable"
	devPrint("!", colorYellow, "postgres not reachable — starting via docker on :5433…")
	if err := dockerStart(ctx, "eventhook-dev-postgres", "postgres:16",
		[]string{
			"-e", "POSTGRES_DB=eventhook",
			"-e", "POSTGRES_USER=eventhook",
			"-e", "POSTGRES_PASSWORD=eventhook",
			"-p", "5433:5432",
		},
		func() bool { return pgConnectOK(ctx, devURL) },
	); err != nil {
		return err
	}
	cfg.DatabaseURL = devURL
	return nil
}

func pgConnectOK(ctx context.Context, databaseURL string) bool {
	tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	st, err := store.NewPostgresStore(tctx, databaseURL)
	if err != nil {
		return false
	}
	st.Close()
	return true
}

// ensureRedis checks if redis is reachable; starts docker if not.
func ensureRedis(ctx context.Context, cfg *config.Config) error {
	devPrint("→", colorDim, "checking redis…")

	host, port := extractHostPort(cfg.RedisURL, "6379")
	addr := net.JoinHostPort(host, port)

	if redisPingOK(addr) {
		return nil
	}

	devPrint("!", colorYellow, "redis not found — starting via docker…")
	return dockerStart(ctx, "eventhook-redis-1", "redis:7-alpine",
		[]string{"-p", "6379:6379"},
		func() bool { return redisPingOK("localhost:6379") },
	)
}

// redisPingOK sends a raw PING and checks for +PONG.
func redisPingOK(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(time.Second)) //nolint
	fmt.Fprintf(conn, "*1\r\n$4\r\nPING\r\n")
	buf := make([]byte, 7)
	n, err := conn.Read(buf)
	return err == nil && n > 0 && strings.HasPrefix(string(buf[:n]), "+PONG")
}

func dockerStart(ctx context.Context, name, image string, args []string, ready func() bool) error {
	// Check docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found — please start %s manually on the default port", image)
	}

	// Remove existing stopped container with same name
	exec.CommandContext(ctx, "docker", "rm", "-f", name).Run() //nolint

	cmdArgs := append([]string{"run", "-d", "--name", name}, args...)
	cmdArgs = append(cmdArgs, image)

	out, err := exec.CommandContext(ctx, "docker", cmdArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker run %s: %s", image, strings.TrimSpace(string(out)))
	}

	// Wait up to 30s for the service to become reachable
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(time.Second)
		if ready() {
			return nil
		}
	}
	return fmt.Errorf("timed out waiting for %s to be ready", image)
}

// tailEvents polls the events API and prints new events as they arrive.
func tailEvents(ctx context.Context, cfg *config.Config) {
	client := &http.Client{Timeout: 5 * time.Second}
	seen := map[string]bool{}
	url := fmt.Sprintf("http://localhost:%d/api/v1/events", cfg.Port)

	// Wait for server to be ready
	time.Sleep(500 * time.Millisecond)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			events := fetchEvents(client, url, cfg.APIKey)
			for _, ev := range events {
				id, _ := ev["id"].(string)
				if id == "" || seen[id] {
					continue
				}
				seen[id] = true
				printEvent(ev)
			}
		}
	}
}

func fetchEvents(client *http.Client, url, apiKey string) []map[string]any {
	req, err := http.NewRequest("GET", url+"?limit=50", nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}
	// Reverse so oldest prints first
	events := result.Data
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}
	return events
}

func printEvent(ev map[string]any) {
	id, _        := ev["id"].(string)
	eventType, _ := ev["event_type"].(string)
	status, _    := ev["status"].(string)
	receivedAt, _ := ev["received_at"].(string)

	ts := ""
	if t, err := time.Parse(time.RFC3339Nano, receivedAt); err == nil {
		ts = t.Local().Format("15:04:05")
	}

	statusColor := statusBadgeColor(status)
	shortID := id
	if len(id) > 8 {
		shortID = id[:8]
	}

	payload, _ := ev["payload"].(map[string]any)
	preview := payloadPreview(payload)

	fmt.Printf("  %s%s%s  %s%-10s%s  %s%-36s%s  %s%s%s\n",
		colorDim, ts, colorReset,
		statusColor, status, colorReset,
		colorCyan, eventType, colorReset,
		colorDim, preview+colorDim+"  "+shortID+"…"+colorReset, colorReset,
	)
}

func payloadPreview(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	b, _ := json.Marshal(payload)
	s := string(b)
	if len(s) > 60 {
		s = s[:60] + "…"
	}
	return colorDim + s + colorReset
}

func statusBadgeColor(status string) string {
	switch status {
	case "delivered":
		return colorGreen
	case "pending":
		return colorDim
	case "retrying":
		return colorYellow
	case "failed":
		return colorRed
	default:
		return colorDim
	}
}

func printBanner(cfg *config.Config) {
	fmt.Printf("\n%s%s EventHook v0.1.0 %s\n", colorBold, colorGreen, colorReset)
	fmt.Printf("%s─────────────────────────────────────────%s\n", colorDim, colorReset)
	fmt.Printf("  Runtime   %shttp://localhost:%d%s\n", colorCyan, cfg.Port, colorReset)
	fmt.Printf("  Dashboard %shttp://localhost:%d/dashboard%s\n", colorCyan, cfg.Port, colorReset)
	fmt.Printf("  API Key   %s%s%s\n", colorDim, cfg.APIKey, colorReset)
	fmt.Printf("%s─────────────────────────────────────────%s\n\n", colorDim, colorReset)
}

func devPrint(icon, color, msg string) {
	fmt.Printf("  %s%s%s  %s\n", color, icon, colorReset, msg)
}

func portOpen(host, port string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func extractHostPort(rawURL, defaultPort string) (string, string) {
	// Strip scheme (postgres://, redis://)
	s := rawURL
	if idx := strings.Index(s, "://"); idx != -1 {
		s = s[idx+3:]
	}
	// Strip userinfo
	if idx := strings.Index(s, "@"); idx != -1 {
		s = s[idx+1:]
	}
	// Strip path/query
	if idx := strings.IndexAny(s, "/?"); idx != -1 {
		s = s[:idx]
	}
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return s, defaultPort
	}
	return host, port
}

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/eventhook/eventhook/assets"
	"github.com/eventhook/eventhook/internal/api"
	"github.com/eventhook/eventhook/internal/config"
	"github.com/eventhook/eventhook/internal/store"
	"github.com/eventhook/eventhook/internal/worker"
	"github.com/eventhook/eventhook/migrations"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := config.Load()

	if cfg.Env == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Info().Msg("running migrations")
	if err := store.RunMigrations(cfg.DatabaseURL, migrations.FS); err != nil {
		log.Fatal().Err(err).Msg("migrations failed")
	}

	st, err := store.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect postgres")
	}
	defer st.Close()

	pool := worker.NewPool(st, cfg.WorkerCount)
	pool.Start(ctx)
	defer pool.Stop()

	srv := api.NewServer(cfg, st)
	srv.ServeDashboard(assets.Dashboard)

	fmt.Printf("\nEventHook v0.1.0\n")
	fmt.Printf("─────────────────────────────────\n")
	fmt.Printf("Runtime:   http://localhost:%d\n", cfg.Port)
	fmt.Printf("Dashboard: http://localhost:%d/dashboard\n", cfg.Port)
	fmt.Printf("API Key:   %s\n", cfg.APIKey)
	fmt.Printf("─────────────────────────────────\n\n")

	go func() {
		<-ctx.Done()
		log.Info().Msg("shutting down")
	}()

	if err := srv.Run(); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}

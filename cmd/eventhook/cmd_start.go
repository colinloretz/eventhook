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
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the EventHook runtime (production)",
	RunE:  runStart,
}

func runStart(_ *cobra.Command, _ []string) error {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	cfg := config.Load()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Info().Msg("running migrations")
	if err := store.RunMigrations(cfg.DatabaseURL, migrations.FS); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	st, err := store.NewPostgresStore(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	defer st.Close()

	pool := worker.NewPool(st, cfg.WorkerCount)
	pool.Start(ctx)
	defer pool.Stop()

	srv := api.NewServer(cfg, st)
	srv.ServeDashboard(assets.Dashboard)

	log.Info().Msgf("listening on :%d", cfg.Port)
	return srv.Run()
}

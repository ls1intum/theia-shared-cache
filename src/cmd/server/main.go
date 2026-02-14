package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/kevingruber/gradle-cache/internal/telemetry"

	"github.com/kevingruber/gradle-cache/internal/config"
	"github.com/kevingruber/gradle-cache/internal/server"
	"github.com/kevingruber/gradle-cache/internal/storage"
	"github.com/rs/zerolog"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}

	// Setup logger
	logger := setupLogger(cfg.Logging)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Fatal().Err(err).Msg("invalid configuration")
	}

	cleanup, err := telemetry.SetupTelemetry(cfg.Sentry.Enabled, cfg.Sentry.Dsn)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to setup telemetry")
	}
	defer cleanup()

	store, err := storage.NewRedisStorage(storage.RedisConfig{
		Addr:     cfg.Storage.Addr,
		Password: cfg.Storage.Password,
		DB:       cfg.Storage.DB,
	})

	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create storage")
	}

	// Create and run server
	srv := server.New(cfg, store, logger)

	ctx := context.Background()
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.Info().Str("signal", sig.String()).Msg("received shutdown signal")
		cancel()
	}()

	// Run server
	if err := srv.Run(ctx); err != nil {
		logger.Fatal().Err(err).Msg("server error")
	}

	logger.Info().Msg("server stopped")
}

func setupLogger(cfg config.LoggingConfig) zerolog.Logger {
	// Set log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output format
	var logger zerolog.Logger
	if cfg.Format == "json" {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
			With().Timestamp().Logger()
	}

	return logger
}

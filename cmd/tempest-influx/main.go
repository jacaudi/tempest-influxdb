package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/samber/lo"

	"tempest_influx/internal/config"
	"tempest_influx/internal/logger"
	"tempest_influx/internal/processor"
)

func main() {
	log.SetPrefix("tempest_influx: ")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Check for config path override using Lo library patterns
	configDir := lo.CoalesceOrEmpty(os.Getenv("TEMPEST_INFLUX_CONFIG_DIR"), "/config")

	cfg := config.Load(configDir, "tempest_influx")

	// Initialize structured logger
	appLogger := logger.New(cfg)

	go func() {
		<-sigCh
		appLogger.Info("Received shutdown signal")
		cancel()
	}()

	appLogger.Info("Starting tempest_influx",
		slog.String("config_dir", configDir),
		slog.String("version", "2.0.0"))

	if cfg.Debug {
		appLogger.Debug("Configuration loaded",
			slog.String("listen_address", cfg.Listen_Address),
			slog.String("influx_url", cfg.Influx_URL),
			slog.String("influx_bucket", cfg.Influx_Bucket),
			slog.Bool("rapid_wind", cfg.Rapid_Wind))
	}

	appLogger.Info("Service configuration loaded",
		slog.Bool("verbose", cfg.Verbose),
		slog.Bool("debug", cfg.Debug),
		slog.String("listen_address", cfg.Listen_Address),
		slog.String("influx_url", cfg.Influx_URL),
		slog.String("bucket", cfg.Influx_Bucket),
		slog.Bool("rapid_wind", cfg.Rapid_Wind),
		slog.String("rapid_wind_bucket", cfg.Influx_Bucket_Rapid_Wind))

	// Use the service-oriented approach
	service, err := processor.NewWeatherService(cfg, appLogger)
	if err != nil {
		appLogger.Error("Failed to create weather service", slog.String("error", err.Error()))
		return
	}

	if err := service.Start(ctx); err != nil && err != context.Canceled {
		appLogger.Error("Weather service error", slog.String("error", err.Error()))
	}
}

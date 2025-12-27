package main

import (
	"log/slog"
	"os"
	"runtime/debug"

	"bros_kiosk/internal/config"
	"bros_kiosk/internal/logger"
	"bros_kiosk/internal/server"
)

func main() {
	// 0. Setup structured logging
	logger.Setup()

	// Set GC percent to be more aggressive for Pi Zero if not already set
	if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(50)
	}

	// 1. Load config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Create a default config if it doesn't exist for manual verification
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := `
server:
  port: 8080
  host: "0.0.0.0"
  update_interval: "5s"
`
		if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
			slog.Error("Failed to create default config", "error", err)
			os.Exit(1)
		}
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		slog.Error("Invalid configuration", "error", err)
		os.Exit(1)
	}

	// 2. Initialize server
	srv := server.New(cfg)

	// 3. Start server
	slog.Info("Bros Kiosk Server starting", "host", cfg.Server.Host, "port", cfg.Server.Port)
	if err := srv.Start(); err != nil {
		slog.Error("Server stopped", "error", err)
		os.Exit(1)
	}

	slog.Info("Server shut down gracefully")
}

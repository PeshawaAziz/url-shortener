package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/PeshawaAziz/url-shortener/internal/config"
	"github.com/PeshawaAziz/url-shortener/pkg/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.Server.Environment)

	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	logger.Info("database connected")

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("failed to get underlying sql.DB", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	logger.Info("starting server",
		"environment", cfg.Server.Environment,
		"port", cfg.Server.Port,
		"db_host", cfg.Database.Host,
	)
}

func setupLogger(env string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if env == "local" || env == "development" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

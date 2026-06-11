package main

func main(){
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup structured logger based on environment
	logger := setupLogger(cfg.Server.Environment)

	logger.Info("starting server",
		"environment", cfg.Server.Environment,
		"port", cfg.Server.Port,
		"db_host", cfg.Database.Host,
	)

	// Your server setup continues here...
	_ = cfg // use cfg
}

func setupLogger(env string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	// Development: human-readable logs
	// Production: JSON logs for log aggregators
	if env == "local" || env == "development" {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	// For staging/production, use JSON format
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/punnch/ankiwords/internal/app"
	"github.com/punnch/ankiwords/internal/config"
	"github.com/punnch/ankiwords/internal/logger"
	"github.com/punnch/ankiwords/internal/repository"

	"go.uber.org/zap"
)

// main loads configuration, initializes dependencies, and runs the terminal UI
// until a shutdown signal cancels the root context.
func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		panic("config validation failed " + err.Error())
	}

	logger, err := logger.NewLogger(cfg)
	if err != nil {
		panic("logger initialization failed " + err.Error())
	}
	defer logger.Close()

	store, err := repository.NewFileRepository(cfg.SettingsFile)
	if err != nil {
		logger.Fatal("store init failed", zap.Error(err))
	}

	application, err := app.New(
		cfg,
		store,
		logger,
	)
	if err != nil {
		logger.Fatal("app init failed", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	if err := application.Run(ctx); err != nil && ctx.Err() == nil {
		logger.Fatal("run failed", zap.Error(err))
	}
}

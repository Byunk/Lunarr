// Package main is the entry point for the agent-broker service.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/lunarr-ai/lunarr/agent-broker/internal/config"
	"github.com/lunarr-ai/lunarr/agent-broker/internal/handler"
	"github.com/lunarr-ai/lunarr/agent-broker/internal/server"
)

func main() {
	cfg := config.Load()
	logger := setupLogger(cfg.LogLevel)

	logger.Info("starting agent-broker",
		"port", cfg.Port,
		"log_level", cfg.LogLevel.String(),
	)

	mux := http.NewServeMux()
	handler.NewHealthHandler(nil).RegisterRoutes(mux)

	srv := server.New(mux,
		server.WithPort(cfg.Port),
		server.WithLogger(logger),
	)

	if err := srv.Run(context.Background()); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}

func setupLogger(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

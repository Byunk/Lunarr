package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/lunarr-ai/lunarr/agent-broker/internal/config"
	"github.com/lunarr-ai/lunarr/agent-broker/internal/handler"
	"github.com/lunarr-ai/lunarr/agent-broker/internal/registry"
	"github.com/lunarr-ai/lunarr/agent-broker/internal/server"
	"github.com/lunarr-ai/lunarr/agent-broker/internal/store"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	logger := setupLogger(cfg.LogLevel)

	logger.Info("starting agent-broker",
		"port", cfg.Port,
		"log_level", cfg.LogLevel.String(),
		"qdrant_host", cfg.QdrantHost,
		"qdrant_port", cfg.QdrantPort,
	)

	ctx := context.Background()

	qdrantStore, err := store.NewQdrantStore(ctx,
		store.WithHost(cfg.QdrantHost),
		store.WithPort(cfg.QdrantPort),
		store.WithAPIKey(cfg.QdrantAPIKey),
		store.WithTLS(cfg.QdrantUseTLS),
	)
	if err != nil {
		logger.Error("failed to connect to qdrant", "error", err)
		return err
	}
	defer func() {
		if err := qdrantStore.Close(); err != nil {
			logger.Error("failed to close qdrant connection", "error", err)
		}
	}()

	logger.Info("connected to qdrant")

	agentStore, err := store.NewQdrantStore(ctx)
	if err != nil {
		logger.Error("failed to create qdrant store", "error", err)
		return err
	}
	defer func() {
		if err := agentStore.Close(); err != nil {
			logger.Error("failed to close agent store", "error", err)
		}
	}()

	registryService := registry.NewRegistryService(agentStore)

	mux := http.NewServeMux()
	handler.NewHealthHandler(qdrantStore).RegisterRoutes(mux)
	handler.NewAdminHandler(registryService).RegisterRoutes(mux)
	handler.NewAgentsHandler(registryService).RegisterRoutes(mux)

	srv := server.New(mux,
		server.WithPort(cfg.Port),
		server.WithLogger(logger),
	)

	if err := srv.Run(ctx); err != nil {
		logger.Error("server error", "error", err)
		return err
	}

	return nil
}

func setupLogger(level slog.Level) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

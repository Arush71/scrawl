// Package app is the main app
package app

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/Arush71/scrawl/internal/api"
	"github.com/Arush71/scrawl/internal/server"
	"github.com/Arush71/scrawl/internal/ws"
)

type App struct {
	logger   *slog.Logger
	server   *server.Server
	registry *ws.Registry
}

func StartApp() (*App, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	logger.Info("logger initialized")
	mux := http.NewServeMux()
	registry := ws.NewRegistry(logger)
	handler := &api.Handler{
		Logger:   logger,
		Registry: registry,
	}
	api.AddRoutes(mux, handler)
	server := server.New(logger, mux)
	if err := server.Run(); err != nil {
		return nil, err
	}
	return &App{
		logger:   logger,
		server:   server,
		registry: registry,
	}, nil
}

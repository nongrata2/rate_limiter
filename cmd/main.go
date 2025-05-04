package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"ratelimiter/internal/config"
	"ratelimiter/internal/handlers"
	"ratelimiter/internal/models"
	"ratelimiter/internal/rate_limiter"
	"syscall"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	defaultLimit := models.Limit{
		Capacity:   cfg.DefaultLimit.Capacity,
		RefillRate: cfg.DefaultLimit.RefillRate,
	}

	store := rate_limiter.NewBucketStore(defaultLimit, cfg.ClientRateLimits)

	mux := http.NewServeMux()
	mux.Handle("POST /clients", handlers.CreateClient(store))
	mux.Handle("GET /clients", handlers.ListClients(store))
	mux.Handle("GET /clients/{clientID}", handlers.GetClient(store))
	mux.Handle("DELETE /clients/", handlers.DeleteClient(store))

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	server := http.Server{
		Addr:        cfg.Address,
		Handler:     mux,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	log.Info("Starting server", "address", cfg.Address)
	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	log.Info("Running HTTP server", "address", cfg.Address)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Error("server closed unexpectedly", "error", err)
			return
		}
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level, AddSource: true})
	return slog.New(handler)
}

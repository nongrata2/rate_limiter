package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"ratelimiter/internal/config"
	"ratelimiter/internal/handlers"
	"ratelimiter/internal/rate_limiter"
	"ratelimiter/internal/repositories"
	"time"

	"syscall"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	storage, err := repositories.New(log, dsn)
	if err != nil {
		log.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	if err := storage.Migrate(); err != nil {
		log.Error("failed to migrate db", "error", err)
		os.Exit(1)
	}

	log.Info("successfully connected to database")

	store := rate_limiter.NewBucketStore()
	go store.StartBackgroundRefill(100 * time.Millisecond)

	clientsFromDB, err := storage.ListClients(context.Background())
	if err != nil {
		log.Error("failed to list clients", "error", err)
	} else {
		for _, cl := range clientsFromDB {
			log.Info("Loading client from DB into BucketStore",
				"key", cl.Key,
				"capacity", cl.Capacity,
				"refill_rate", cl.RefillRate.String(),
				"unlimited", cl.Unlimited,
			)

			store.Set(cl.Key, rate_limiter.NewTokenBucket(
				cl.Capacity,
				cl.RefillRate,
				cl.Unlimited,
			))
		}
	}

	mux := http.NewServeMux()
	mux.Handle("POST /clients", handlers.AddClientHandler(log, storage, store))
	mux.Handle("PUT /clients/{clientID}", handlers.EditClientHandler(log, storage))
	mux.Handle("GET /clients", handlers.ListClientsHandler(log, storage))
	mux.Handle("GET /clients/{clientID}", handlers.GetClientHandler(log, storage))
	mux.Handle("DELETE /clients/{clientID}", handlers.DeleteClientHandler(log, storage, store))
	mux.Handle("/api", rate_limiter.RateLimitMiddleware(store, cfg.DefaultLimit, storage)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Request allowed\n")
	})))

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

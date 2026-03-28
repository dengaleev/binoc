package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dengaleev/binoc/service/internal/config"
	"github.com/dengaleev/binoc/service/internal/instrument"
	"github.com/dengaleev/binoc/service/internal/server"
)

func main() {
	cfg := config.Load()
	logger := instrument.SetupLogger(cfg.LogFormat, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	opts := []server.Option{
		server.WithLogger(logger),
	}

	if cfg.MetricsEnabled {
		m := instrument.NewMetrics()
		opts = append(opts, server.WithMetrics(m))
		logger.Info("metrics enabled")
	}

	if cfg.TracingEnabled {
		_, shutdown, err := instrument.SetupTracing(ctx, cfg.OTELExporterEndpoint, cfg.ServiceName)
		if err != nil {
			logger.Error("failed to setup tracing", "error", err)
			os.Exit(1)
		}
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdown(shutdownCtx); err != nil {
				logger.Error("tracing shutdown error", "error", err)
			}
		}()
		opts = append(opts, server.WithTracer(cfg.ServiceName))
		logger.Info("tracing enabled", "endpoint", cfg.OTELExporterEndpoint)
	}

	srv := server.New(opts...)
	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("starting server", "addr", cfg.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}

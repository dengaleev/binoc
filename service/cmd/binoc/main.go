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
	"github.com/dengaleev/binoc/service/internal/store"
)

func main() {
	cfg := config.Load()
	logger := instrument.SetupLogger(cfg.LogFormat, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var opts []server.Option

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
		opts = append(opts, server.WithTracing())
		logger.Info("tracing enabled", "endpoint", cfg.OTELExporterEndpoint)

		otelHandler, logShutdown, err := instrument.SetupOTELLogging(ctx, cfg.OTELExporterEndpoint, cfg.ServiceName)
		if err != nil {
			logger.Error("failed to setup OTLP logging", "error", err)
		} else {
			defer func() {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := logShutdown(shutdownCtx); err != nil {
					logger.Error("log provider shutdown error", "error", err)
				}
			}()
			logger = slog.New(instrument.NewMultiHandler(logger.Handler(), otelHandler))
			slog.SetDefault(logger)
			logger.Info("OTLP logging enabled")
		}
	}

	if cfg.SelfURL != "" {
		opts = append(opts, server.WithSelfURL(cfg.SelfURL))
		logger.Info("chain endpoint enabled", "self_url", cfg.SelfURL)
	}

	if cfg.DBPath != "" {
		db, err := store.New(ctx, cfg.DBPath)
		if err != nil {
			logger.Error("failed to open database", "error", err)
			os.Exit(1)
		}
		defer db.Close()
		opts = append(opts, server.WithStore(db))
		logger.Info("notes API enabled", "db", cfg.DBPath)
	}

	opts = append(opts, server.WithLogger(logger))
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

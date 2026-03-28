package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/dengaleev/binoc/service/internal/instrument"
)

// Option configures the server.
type Option func(*Server)

// Server wraps an http.ServeMux with instrumentation.
type Server struct {
	mux     *http.ServeMux
	logger  *slog.Logger
	metrics *instrument.Metrics
	tracer  string // tracer name for OTEL spans
}

// WithLogger sets the server logger.
func WithLogger(l *slog.Logger) Option {
	return func(s *Server) { s.logger = l }
}

// WithMetrics enables Prometheus metrics.
func WithMetrics(m *instrument.Metrics) Option {
	return func(s *Server) { s.metrics = m }
}

// WithTracer sets the OTEL tracer name.
func WithTracer(name string) Option {
	return func(s *Server) { s.tracer = name }
}

// New creates a configured Server with all routes registered.
func New(opts ...Option) *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		logger: slog.Default(),
	}
	for _, o := range opts {
		o(s)
	}

	s.mux.HandleFunc("GET /echo", s.handleGetEcho)
	s.mux.HandleFunc("POST /echo", s.handlePostEcho)
	s.mux.HandleFunc("GET /healthz", handleHealthz)
	s.mux.HandleFunc("GET /readyz", handleReadyz)
	s.mux.Handle("GET /metrics", promhttp.Handler())

	return s
}

// Handler returns the fully wrapped handler with middleware applied.
// Middleware chain order (outermost first): tracing → logging → metrics → mux.
func (s *Server) Handler() http.Handler {
	var h http.Handler = s.mux

	if s.metrics != nil {
		h = metricsMiddleware(s.metrics, h)
	}
	h = loggingMiddleware(s.logger, h)
	if s.tracer != "" {
		h = tracingMiddleware(s.tracer, h)
	}

	return h
}

func (s *Server) handleGetEcho(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	resp := map[string]any{
		"message":   msg,
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		s.logger.Error("encoding response", "error", err)
	}
}

func (s *Server) handlePostEcho(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	if _, err := io.Copy(w, r.Body); err != nil {
		s.logger.Error("copying request body", "error", err)
	}
}

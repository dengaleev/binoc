package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/dengaleev/binoc/service/internal/instrument"
)

// traceExemplar extracts a trace_id exemplar label from the request context.
// Returns nil if no active trace span exists.
func traceExemplar(r *http.Request) prometheus.Labels {
	spanCtx := trace.SpanContextFromContext(r.Context())
	if !spanCtx.HasTraceID() {
		return nil
	}
	return prometheus.Labels{"trace_id": spanCtx.TraceID().String()}
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	bytes       int
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// Unwrap returns the underlying ResponseWriter, allowing middleware like
// otelhttp to access the original writer's optional interfaces.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

func otelMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewMiddleware("binoc",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			if r.Pattern != "" {
				return r.Pattern
			}
			return r.Method + " " + r.URL.Path
		}),
		otelhttp.WithFilter(func(r *http.Request) bool {
			return !isInternalPath(r.URL.Path)
		}),
	)(next)
}

// isInternalPath reports whether a path should be excluded from
// tracing and logging (metrics scrapes).
func isInternalPath(path string) bool {
	return path == "/metrics"
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isInternalPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"bytes", rw.bytes,
		}

		spanCtx := trace.SpanContextFromContext(r.Context())
		if spanCtx.HasTraceID() {
			attrs = append(attrs, "trace_id", spanCtx.TraceID().String())
			attrs = append(attrs, "span_id", spanCtx.SpanID().String())
		}

		logger.InfoContext(r.Context(), "request", attrs...)
	})
}

func metricsMiddleware(m *instrument.Metrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.InFlightRequests.Inc()
		defer m.InFlightRequests.Dec()

		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		route := r.Pattern
		method := r.Method
		code := fmt.Sprintf("%d", rw.status)

		exemplar := traceExemplar(r)
		m.RequestsTotal.WithLabelValues(method, route, code).(prometheus.ExemplarAdder).AddWithExemplar(1, exemplar)
		m.RequestDuration.WithLabelValues(method, route).(prometheus.ExemplarObserver).ObserveWithExemplar(time.Since(start).Seconds(), exemplar)
		m.ResponseSize.WithLabelValues(method, route).(prometheus.ExemplarObserver).ObserveWithExemplar(float64(rw.bytes), exemplar)
	})
}

package instrument

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds Prometheus metrics for the echo service.
type Metrics struct {
	RequestsTotal    *prometheus.CounterVec
	RequestDuration  *prometheus.HistogramVec
	ResponseSize     *prometheus.HistogramVec
	InFlightRequests prometheus.Gauge
}

// NewMetrics registers and returns Prometheus metrics.
func NewMetrics() *Metrics {
	return &Metrics{
		RequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "echo",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		}, []string{"method", "route", "code"}),

		RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "echo",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "route"}),

		ResponseSize: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "echo",
			Name:      "response_size_bytes",
			Help:      "HTTP response size in bytes.",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 6),
		}, []string{"method", "route"}),

		InFlightRequests: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "echo",
			Name:      "in_flight_requests",
			Help:      "Number of in-flight HTTP requests.",
		}),
	}
}

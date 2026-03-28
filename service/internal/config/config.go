package config

import (
	"os"
	"strings"
)

// Config holds all service configuration, sourced from environment variables.
type Config struct {
	Addr                   string
	LogFormat              string
	LogLevel               string
	MetricsEnabled         bool
	TracingEnabled         bool
	OTELExporterEndpoint   string
	ServiceName            string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		Addr:                   envOr("ADDR", ":8080"),
		LogFormat:              envOr("LOG_FORMAT", "json"),
		LogLevel:               envOr("LOG_LEVEL", "info"),
		MetricsEnabled:         envBool("METRICS_ENABLED", true),
		TracingEnabled:         envBool("TRACING_ENABLED", true),
		OTELExporterEndpoint:   envOr("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		ServiceName:            envOr("SERVICE_NAME", "echo"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		if fallback {
			return true
		}
		return false
	}
	return strings.EqualFold(v, "true") || v == "1"
}

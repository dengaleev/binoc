package config

import "github.com/caarlos0/env/v11"

// Config holds application configuration sourced from environment variables.
// OTEL-specific settings (endpoint, service name) are handled by the OTEL SDK
// via standard env vars (OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_SERVICE_NAME).
type Config struct {
	Addr           string `env:"ADDR"            envDefault:":8080"`
	LogFormat      string `env:"LOG_FORMAT"       envDefault:"json"`
	LogLevel       string `env:"LOG_LEVEL"        envDefault:"info"`
	MetricsEnabled bool   `env:"METRICS_ENABLED"  envDefault:"true"`
	TracingEnabled bool   `env:"TRACING_ENABLED"  envDefault:"true"`
	SelfURL        string `env:"SELF_URL"         envDefault:""`
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return env.Must(env.ParseAs[Config]())
}

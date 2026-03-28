package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

// Config holds all service configuration, sourced from environment variables.
type Config struct {
	Addr                 string `env:"ADDR"                          envDefault:":8080"`
	LogFormat            string `env:"LOG_FORMAT"                    envDefault:"json"`
	LogLevel             string `env:"LOG_LEVEL"                     envDefault:"info"`
	MetricsEnabled       bool   `env:"METRICS_ENABLED"               envDefault:"true"`
	TracingEnabled       bool   `env:"TRACING_ENABLED"               envDefault:"true"`
	OTELExporterEndpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"   envDefault:"localhost:4317"`
	ServiceName          string `env:"SERVICE_NAME"                  envDefault:"echo"`
	DBPath               string `env:"DB_PATH"                       envDefault:""`
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("parsing config: %v", err)
	}
	return cfg
}

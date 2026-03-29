package server

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("binoc/ticker")

// startTicker runs a background loop that pokes /api/random through Caddy
// every second, producing continuous traces, logs, and error-rate data.
func (s *Server) startTicker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				s.pokeRandom(ctx)
			}
		}
	}()
}

func (s *Server) pokeRandom(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "ticker.poke_random")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.selfURL+"/api/random", nil)
	if err != nil {
		return
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("ticker: random request failed", "error", err)
		return
	}
	resp.Body.Close()
}

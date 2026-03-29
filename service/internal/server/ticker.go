package server

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("binoc/ticker")

// startTicker runs a background loop that fetches /time and /api/random from
// Caddy every second, producing continuous traces, logs, and error rate.
func (s *Server) startTicker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				s.fetchTime(ctx)
				s.pokeRandom(ctx)
			}
		}
	}()
}

func (s *Server) fetchTime(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "ticker.fetch_time")
	defer span.End()

	target := s.selfURL + "/time"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		s.logger.Error("ticker: creating request", "error", err)
		return
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("ticker: request failed", "error", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Time string `json:"time"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		s.logger.Warn("ticker: bad response", "body", string(body))
		return
	}

	s.logger.Info("tick", slog.String("caddy_time", result.Time))
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

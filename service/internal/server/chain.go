package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// handleChain calls the echo endpoint through Caddy, producing a distributed
// trace: client → Caddy → app(/chain) → Caddy → app(/echo).
func (s *Server) handleChain(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	if msg == "" {
		msg = "chain"
	}

	target := fmt.Sprintf("%s/api/echo?msg=%s", s.selfURL, msg)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		s.logger.Error("creating chain request", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("chain call failed", "error", err, "target", target)
		http.Error(w, `{"error":"upstream error"}`, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := map[string]any{
		"chain_timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"upstream_status": resp.StatusCode,
		"upstream_body":   json.RawMessage(body),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func newTracedHTTPClient() *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
		Timeout:   5 * time.Second,
	}
}

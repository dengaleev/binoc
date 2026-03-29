package server

import (
	"math/rand/v2"
	"net/http"
	"time"
)

// handleRandom sleeps for a random duration (0-500ms) and returns an error
// ~10% of the time. Useful for generating realistic latency distributions
// and non-zero error rates on dashboards.
func (s *Server) handleRandom(w http.ResponseWriter, _ *http.Request) {
	delay := time.Duration(rand.IntN(500)) * time.Millisecond
	time.Sleep(delay)

	if rand.IntN(10) == 0 {
		jsonError(w, "random failure", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"delay_ms":  delay.Milliseconds(),
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
	})
}

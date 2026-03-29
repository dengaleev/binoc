package server

import (
	"encoding/json"
	"net/http"

	"github.com/dengaleev/binoc/service/internal/store"
)

func (s *Server) handleListNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := s.store.List(r.Context())
	if err != nil {
		s.logger.Error("listing notes", "error", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	if notes == nil {
		notes = make([]store.Note, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

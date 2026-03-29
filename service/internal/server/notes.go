package server

import (
	"net/http"

	"github.com/dengaleev/binoc/service/internal/store"
)

func (s *Server) handleListNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := s.store.List(r.Context())
	if err != nil {
		s.logger.Error("listing notes", "error", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if notes == nil {
		notes = make([]store.Note, 0)
	}

	writeJSON(w, http.StatusOK, notes)
}

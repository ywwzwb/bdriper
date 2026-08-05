package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/zwb/bdriper/internal/db"
)

func (s *Server) handleListSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := db.ListSettings(s.DB)
	if err != nil {
		slog.Error("failed to list settings", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var input map[string]string
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	for k, v := range input {
		if err := db.SetSetting(s.DB, k, v); err != nil {
			slog.Error("failed to set setting", "key", k, "error", err)
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

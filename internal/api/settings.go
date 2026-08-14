package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

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
	var input map[string]any
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	for k, v := range input {
		if err := db.SetSetting(s.DB, k, fmt.Sprint(v)); err != nil {
			slog.Error("failed to set setting", "key", k, "error", err)
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if k == "max_concurrent" && s.Runner != nil {
			if n, err := strconv.Atoi(fmt.Sprint(v)); err == nil {
				s.Runner.SetMaxConcurrent(n)
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

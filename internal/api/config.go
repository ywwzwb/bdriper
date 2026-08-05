package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/zwb/bdriper/internal/config"
	"github.com/zwb/bdriper/internal/db"
)

func (s *Server) handleListConfigs(w http.ResponseWriter, r *http.Request) {
	cfgs, err := db.ListConfigs(s.DB)
	if err != nil {
		slog.Error("failed to list configs", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cfgs)
}

func (s *Server) handleCreateConfig(w http.ResponseWriter, r *http.Request) {
	var c db.TranscodeConfig
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	id, err := db.CreateConfig(s.DB, &c)
	if err != nil {
		slog.Error("failed to create config", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	c.ID = id
	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	c, err := db.GetConfig(s.DB, id)
	if err != nil {
		slog.Warn("config not found", "id", id, "error", err)
		writeError(w, http.StatusNotFound, "config not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var input map[string]any
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := db.UpdateConfig(s.DB, id, input); err != nil {
		slog.Error("failed to update config", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleDeleteConfig(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err := db.DeleteConfig(s.DB, id); err != nil {
		slog.Error("failed to delete config", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleListPresets(w http.ResponseWriter, r *http.Request) {
	presets, err := config.ListPresets(s.DB)
	if err != nil {
		slog.Error("failed to list presets", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, presets)
}

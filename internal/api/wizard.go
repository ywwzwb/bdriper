package api

import (
	"encoding/json"
	"net/http"

	"github.com/zwb/bdriper/internal/wizard"
)

func (s *Server) handleParseBDMV(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SourcePath string `json:"source_path"`
		Path       string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	srcPath := input.SourcePath
	if srcPath == "" {
		srcPath = input.Path
	}

	info, err := wizard.ParseBDMV(srcPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *Server) handleFileStreams(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	streams, err := wizard.GetFileStreams(filePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, streams)
}

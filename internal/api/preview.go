package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/zwb/bdriper/internal/db"
	"github.com/zwb/bdriper/internal/preview"
)

func (s *Server) handleCreatePreview(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TaskID     int64  `json:"task_id"`
		SourceFile string `json:"source_file"`
		StartTime  string `json:"start_time"`
		Duration   int    `json:"duration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	outputFile := filepath.Join(s.DataDir, "previews", strconv.FormatInt(time.Now().UnixNano(), 10)+".mp4")
	os.MkdirAll(filepath.Dir(outputFile), 0755)

	pj := &db.PreviewJob{
		TaskID:     input.TaskID,
		SourceFile: input.SourceFile,
		StartTime:  input.StartTime,
		Duration:   input.Duration,
		OutputFile: outputFile,
		Status:     "running",
		ExpiresAt:  time.Now().Add(30 * time.Minute),
	}

	id, err := db.CreatePreview(s.DB, pj)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	go func() {
		preview.RunPreview(pj.SourceFile, pj.StartTime, pj.Duration, pj.OutputFile)
		db.UpdatePreview(s.DB, id, map[string]any{"status": "completed"})
	}()

	pj.ID = id
	writeJSON(w, http.StatusCreated, pj)
}

func (s *Server) handlePreviewStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	p, err := db.GetPreview(s.DB, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "preview not found")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handlePreviewDownload(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	p, err := db.GetPreview(s.DB, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "preview not found")
		return
	}
	http.ServeFile(w, r, p.OutputFile)
}

func (s *Server) handleDeletePreview(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	p, err := db.GetPreview(s.DB, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "preview not found")
		return
	}
	os.Remove(p.OutputFile)
	db.DeletePreview(s.DB, id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

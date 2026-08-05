package api

import (
	"encoding/json"
	"log/slog"
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
		TaskID      int64          `json:"task_id"`
		SourceFile  string         `json:"source_file"`
		StartTime   string         `json:"start_time"`
		Duration    int            `json:"duration"`
		Encoder     string         `json:"encoder"`
		VideoParams map[string]any `json:"video_params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if input.Encoder == "" {
		input.Encoder = "x264"
	}
	if input.VideoParams == nil {
		input.VideoParams = map[string]any{}
	}

	outputFile := filepath.Join(s.DataDir, "previews", strconv.FormatInt(time.Now().UnixNano(), 10)+".mkv")
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
		slog.Error("failed to create preview", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	go func() {
		cmd, progressCh, err := preview.RunPreview(pj.SourceFile, pj.StartTime, input.Encoder, input.VideoParams, pj.Duration, pj.OutputFile)
		if err != nil {
			slog.Error("preview run failed", "preview_id", id, "error", err)
			db.UpdatePreview(s.DB, id, map[string]any{"status": "failed"})
			return
		}
		for pct := range progressCh {
			db.UpdatePreview(s.DB, id, map[string]any{"progress": pct})
		}
		cmd.Wait()
		db.UpdatePreview(s.DB, id, map[string]any{"status": "completed", "progress": 1.0})
		slog.Info("preview completed", "preview_id", id)
	}()

	pj.ID = id
	writeJSON(w, http.StatusCreated, pj)
}

func (s *Server) handlePreviewStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	p, err := db.GetPreview(s.DB, id)
	if err != nil {
		slog.Warn("preview not found", "id", id, "error", err)
		writeError(w, http.StatusNotFound, "preview not found")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handlePreviewDownload(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	p, err := db.GetPreview(s.DB, id)
	if err != nil {
		slog.Warn("preview not found", "id", id, "error", err)
		writeError(w, http.StatusNotFound, "preview not found")
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=preview.mkv")
	w.Header().Set("Content-Type", "video/x-matroska")
	http.ServeFile(w, r, p.OutputFile)
}

func (s *Server) handleDeletePreview(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	p, err := db.GetPreview(s.DB, id)
	if err != nil {
		slog.Warn("preview not found", "id", id, "error", err)
		writeError(w, http.StatusNotFound, "preview not found")
		return
	}
	if err := os.Remove(p.OutputFile); err != nil {
		slog.Warn("failed to remove preview file", "file", p.OutputFile, "error", err)
	}
	db.DeletePreview(s.DB, id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

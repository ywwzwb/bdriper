package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/zwb/bdriper/internal/log"
	"github.com/zwb/bdriper/internal/task"
)

type Server struct {
	DB         *sql.DB
	Logger     *slog.Logger
	LogHandler *log.MultiHandler
	TaskHub    *Hub
	LogHub     *Hub
	Runner     *task.Runner
	DataDir    string
	SPAFS      http.FileSystem
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("GET /api/overview/status", s.handleOverview)

	mux.HandleFunc("GET /api/tasks", s.handleListTasks)
	mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	mux.HandleFunc("GET /api/tasks/{id}", s.handleGetTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", s.handleUpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.handleDeleteTask)
	mux.HandleFunc("DELETE /api/tasks/completed", s.handleDeleteCompleted)
	mux.HandleFunc("POST /api/tasks/{id}/retry", s.handleRetryTask)
	mux.HandleFunc("POST /api/tasks/batch", s.handleBatchTasks)

	mux.HandleFunc("POST /api/wizard/parse", s.handleParseBDMV)
	mux.HandleFunc("GET /api/wizard/file/{path}/streams", s.handleFileStreams)

	mux.HandleFunc("GET /api/configs", s.handleListConfigs)
	mux.HandleFunc("POST /api/configs", s.handleCreateConfig)
	mux.HandleFunc("GET /api/configs/{id}", s.handleGetConfig)
	mux.HandleFunc("PUT /api/configs/{id}", s.handleUpdateConfig)
	mux.HandleFunc("DELETE /api/configs/{id}", s.handleDeleteConfig)
	mux.HandleFunc("GET /api/presets", s.handleListPresets)

	mux.HandleFunc("GET /api/settings", s.handleListSettings)
	mux.HandleFunc("PATCH /api/settings", s.handleUpdateSettings)
	mux.HandleFunc("GET /api/settings/gpu-info", s.handleGPUInfo)

	mux.HandleFunc("POST /api/preview", s.handleCreatePreview)
	mux.HandleFunc("GET /api/preview/{id}/status", s.handlePreviewStatus)
	mux.HandleFunc("GET /api/preview/{id}/download", s.handlePreviewDownload)
	mux.HandleFunc("DELETE /api/preview/{id}", s.handleDeletePreview)

	mux.HandleFunc("GET /api/logs", s.handleGetLogs)
	mux.HandleFunc("GET /api/logs/download", s.handleDownloadLogs)

	mux.HandleFunc("GET /ws/events", s.TaskHub.HandleWS)
	mux.HandleFunc("GET /ws/logs", s.handleWSLogs)

	if s.SPAFS != nil {
		mux.Handle("/", http.FileServer(s.SPAFS))
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

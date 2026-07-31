package api

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/zwb/bdriper/internal/log"
)

type Server struct {
	DB         *sql.DB
	Logger     *slog.Logger
	LogHandler *log.BroadcastHandler
	TaskHub    *Hub
	LogHub     *Hub
	DataDir    string
	SPAFS      http.FileSystem
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// Health
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Overview
	mux.HandleFunc("GET /api/overview/status", s.handleOverview)

	// Tasks
	mux.HandleFunc("GET /api/tasks", s.handleListTasks)
	mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	mux.HandleFunc("GET /api/tasks/{id}", s.handleGetTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", s.handleUpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.handleDeleteTask)
	mux.HandleFunc("DELETE /api/tasks/completed", s.handleDeleteCompleted)
	mux.HandleFunc("POST /api/tasks/{id}/retry", s.handleRetryTask)
	mux.HandleFunc("POST /api/tasks/batch", s.handleBatchTasks)

	// Wizard
	mux.HandleFunc("POST /api/wizard/parse", s.handleParseBDMV)
	mux.HandleFunc("GET /api/wizard/file/{path}/streams", s.handleFileStreams)

	// Configs
	mux.HandleFunc("GET /api/configs", s.handleListConfigs)
	mux.HandleFunc("POST /api/configs", s.handleCreateConfig)
	mux.HandleFunc("GET /api/configs/{id}", s.handleGetConfig)
	mux.HandleFunc("PUT /api/configs/{id}", s.handleUpdateConfig)
	mux.HandleFunc("DELETE /api/configs/{id}", s.handleDeleteConfig)
	mux.HandleFunc("GET /api/presets", s.handleListPresets)

	// Settings
	mux.HandleFunc("GET /api/settings", s.handleListSettings)
	mux.HandleFunc("PATCH /api/settings", s.handleUpdateSettings)
	mux.HandleFunc("GET /api/settings/gpu-info", s.handleGPUInfo)

	// Preview
	mux.HandleFunc("POST /api/preview", s.handleCreatePreview)
	mux.HandleFunc("GET /api/preview/{id}/status", s.handlePreviewStatus)
	mux.HandleFunc("GET /api/preview/{id}/download", s.handlePreviewDownload)
	mux.HandleFunc("DELETE /api/preview/{id}", s.handleDeletePreview)

	// Logs
	mux.HandleFunc("GET /api/logs", s.handleGetLogs)
	mux.HandleFunc("GET /api/logs/download", s.handleDownloadLogs)

	// WebSocket
	mux.HandleFunc("GET /ws/events", s.TaskHub.HandleWS)

	// SPA static files (from embedded FS)
	if s.SPAFS != nil {
		mux.Handle("/", http.FileServer(s.SPAFS))
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func notImplemented(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "not implemented"})
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request)         { notImplemented(w) }
func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request)        { notImplemented(w) }
func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request)       { notImplemented(w) }
func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request)          { notImplemented(w) }
func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request)       { notImplemented(w) }
func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request)       { notImplemented(w) }
func (s *Server) handleDeleteCompleted(w http.ResponseWriter, r *http.Request)  { notImplemented(w) }
func (s *Server) handleRetryTask(w http.ResponseWriter, r *http.Request)        { notImplemented(w) }
func (s *Server) handleBatchTasks(w http.ResponseWriter, r *http.Request)       { notImplemented(w) }
func (s *Server) handleParseBDMV(w http.ResponseWriter, r *http.Request)        { notImplemented(w) }
func (s *Server) handleFileStreams(w http.ResponseWriter, r *http.Request)      { notImplemented(w) }
func (s *Server) handleListConfigs(w http.ResponseWriter, r *http.Request)      { notImplemented(w) }
func (s *Server) handleCreateConfig(w http.ResponseWriter, r *http.Request)     { notImplemented(w) }
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request)        { notImplemented(w) }
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request)     { notImplemented(w) }
func (s *Server) handleDeleteConfig(w http.ResponseWriter, r *http.Request)     { notImplemented(w) }
func (s *Server) handleListPresets(w http.ResponseWriter, r *http.Request)      { notImplemented(w) }
func (s *Server) handleListSettings(w http.ResponseWriter, r *http.Request)     { notImplemented(w) }
func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request)   { notImplemented(w) }
func (s *Server) handleGPUInfo(w http.ResponseWriter, r *http.Request)          { notImplemented(w) }
func (s *Server) handleCreatePreview(w http.ResponseWriter, r *http.Request)    { notImplemented(w) }
func (s *Server) handlePreviewStatus(w http.ResponseWriter, r *http.Request)    { notImplemented(w) }
func (s *Server) handlePreviewDownload(w http.ResponseWriter, r *http.Request)  { notImplemented(w) }
func (s *Server) handleDeletePreview(w http.ResponseWriter, r *http.Request)    { notImplemented(w) }
func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request)          { notImplemented(w) }
func (s *Server) handleDownloadLogs(w http.ResponseWriter, r *http.Request)     { notImplemented(w) }

package api

import (
	"archive/tar"
	"compress/gzip"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/zwb/bdriper/internal/log"
)

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	var allLines []string

	logDir := filepath.Join(s.DataDir, "logs")
	files, _ := filepath.Glob(filepath.Join(logDir, "*.log"))
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			slog.Warn("failed to read log file", "file", f, "error", err)
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) != "" {
				allLines = append(allLines, line)
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"lines": allLines,
		"total": len(allLines),
	})
}

func (s *Server) handleDownloadLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", "attachment; filename=bdriper-logs.tar.gz")

	gw := gzip.NewWriter(w)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	logDir := filepath.Join(s.DataDir, "logs")
	filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			slog.Warn("failed to read log file for download", "file", path, "error", err)
			return nil
		}
		hdr := &tar.Header{
			Name:    filepath.Base(path),
			Size:    int64(len(data)),
			Mode:    0644,
			ModTime: info.ModTime(),
		}
		tw.WriteHeader(hdr)
		tw.Write(data)
		return nil
	})
}

func (s *Server) handleWSLogs(w http.ResponseWriter, r *http.Request) {
	if s.LogHandler == nil {
		writeError(w, http.StatusInternalServerError, "log handler not configured")
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("log websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ch := s.LogHandler.Subscribe()
	defer s.LogHandler.Unsubscribe(ch)

	for entry := range ch {
		var e log.LogEntry = entry
		if err := conn.WriteJSON(e); err != nil {
			slog.Warn("log websocket write failed", "error", err)
			return
		}
	}
}


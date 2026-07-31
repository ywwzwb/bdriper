package api

import (
	"archive/tar"
	"compress/gzip"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/zwb/bdriper/internal/log"
)

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	entries := make([]log.LogEntry, 0)

	logDir := filepath.Join(s.DataDir, "logs")
	files, _ := filepath.Glob(filepath.Join(logDir, "*.log"))
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			entries = append(entries, log.LogEntry{
				Raw: line + "\n",
				Msg: line,
			})
		}
	}

	writeJSON(w, http.StatusOK, entries)
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
		return
	}
	defer conn.Close()

	ch := s.LogHandler.Subscribe()
	defer s.LogHandler.Unsubscribe(ch)

	for entry := range ch {
		if err := conn.WriteJSON(entry); err != nil {
			return
		}
	}
}


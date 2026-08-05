package api

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DirEntry struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

func (s *Server) handleFSList(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	path = filepath.Clean(path)

	allowedRoots := []string{s.DataDir, "/mnt", "/media", "/tmp"}
	if !isPathAllowed(path, allowedRoots) {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		slog.Warn("cannot read directory", "path", path, "error", err)
		writeError(w, http.StatusNotFound, "cannot read directory: "+err.Error())
		return
	}

	parent := filepath.Dir(path)
	if parent == path || !isPathAllowed(parent, allowedRoots) {
		parent = ""
	}

	var result []DirEntry
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		result = append(result, DirEntry{
			Name:  e.Name(),
			IsDir: e.IsDir(),
			Size:  info.Size(),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir
		}
		return result[i].Name < result[j].Name
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"path":    path,
		"parent":  parent,
		"entries": result,
	})
}

func isPathAllowed(path string, allowedRoots []string) bool {
	for _, root := range allowedRoots {
		rel, err := filepath.Rel(root, path)
		if err == nil && !strings.HasPrefix(rel, "..") {
			return true
		}
	}
	return false
}

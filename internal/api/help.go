package api

import (
	"net/http"
	"path/filepath"

	"github.com/zwb/bdriper/internal/help"
)

func (s *Server) handleHelpDoc(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("doc")
	if name == "" {
		name = "x264-params"
	}
	name = filepath.Base(name) // prevent path traversal

	path := filepath.Join(s.HelpDir, name+".md")
	html, err := help.RenderFile(path)
	if err != nil {
		path = filepath.Join(s.PresetsDir, "..", "docs", "help", name+".md")
		html, err = help.RenderFile(path)
		if err != nil {
			writeError(w, http.StatusNotFound, "help doc not found: "+name)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(html)
}

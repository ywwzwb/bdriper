package api

import (
	"net/http"

	"github.com/zwb/bdriper/internal/gpu"
)

func (s *Server) handleGPUInfo(w http.ResponseWriter, r *http.Request) {
	gpus := gpu.DetectGPUs()
	writeJSON(w, http.StatusOK, gpus)
}

package api

import (
	"net/http"
	"runtime"

	"github.com/zwb/bdriper/internal/db"
)

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	tasks, _ := db.ListTasks(s.DB, "")
	var running, total int
	for _, t := range tasks {
		total++
		if t.Status == "running" {
			running++
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"cpu_usage":  getCPUUsage(),
		"gpu_usage":  getGPUUsage(),
		"running":    running,
		"total":      total,
		"goroutines": runtime.NumGoroutine(),
		"mem_mb":     mem.Alloc / 1024 / 1024,
	})
}

func getCPUUsage() float64 { return 0 }
func getGPUUsage() float64 { return 0 }

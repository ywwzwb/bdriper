package api

import (
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/zwb/bdriper/internal/db"
	"github.com/zwb/bdriper/internal/gpu"
)

var (
	lastCPUPercent  float64
	cpuPollMutex    sync.Mutex
	cpuPolling      bool
)

func pollCPU() {
	cpuPollMutex.Lock()
	if cpuPolling {
		cpuPollMutex.Unlock()
		return
	}
	cpuPolling = true
	cpuPollMutex.Unlock()

	p, err := cpu.Percent(1*time.Second, false)
	if err == nil && len(p) > 0 {
		cpuPollMutex.Lock()
		lastCPUPercent = p[0]
		cpuPollMutex.Unlock()
	}

	cpuPollMutex.Lock()
	cpuPolling = false
	cpuPollMutex.Unlock()
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	go pollCPU()

	tasks, _ := db.ListTasks(s.DB, "")
	var running, total int
	for _, t := range tasks {
		total++
		if t.Status == "running" {
			running++
		}
	}

	gpuAvail := false
	gpuVendor := ""
	gpuUsage := 0.0

	gpus := gpu.DetectGPUs()
	if len(gpus) > 0 {
		gpuAvail = true
		gpuVendor = gpus[0].Vendor
		if gpuVendor == "NVIDIA" {
			if out, err := execCmdOutput("nvidia-smi", "--query-gpu=utilization.gpu", "--format=csv,noheader,nounits"); err == nil {
				u, parseErr := strconv.ParseFloat(strings.TrimSpace(out), 64)
				if parseErr == nil {
					gpuUsage = u
				}
			}
		}
	}

	cpuPollMutex.Lock()
	cp := lastCPUPercent
	cpuPollMutex.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"cpu_usage":     cp,
		"gpu_available": gpuAvail,
		"gpu_vendor":    gpuVendor,
		"gpu_usage":     gpuUsage,
		"running":       running,
		"total":         total,
		"goroutines":    runtime.NumGoroutine(),
		"mem_mb":        mem.Alloc / 1024 / 1024,
	})
}

func execCmdOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	return string(out), err
}

var _ = os.Stat

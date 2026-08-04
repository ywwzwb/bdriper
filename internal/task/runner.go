package task

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/zwb/bdriper/internal/db"
)

type ProgressBroadcaster interface {
	BroadcastProgress(taskID int64, progress float64)
}

type runningTask struct {
	cmd    *exec.Cmd
	stderr *bytes.Buffer
}

type Runner struct {
	DB            *sql.DB
	Logger        *slog.Logger
	Hub           ProgressBroadcaster
	MaxConcurrent int
	mu            sync.Mutex
	running       map[int64]*runningTask
	sem           chan struct{}
}

func NewRunner(database *sql.DB, logger *slog.Logger, hub ProgressBroadcaster, maxConcurrent int) *Runner {
	return &Runner{
		DB:            database,
		Logger:        logger,
		Hub:           hub,
		MaxConcurrent: maxConcurrent,
		running:       make(map[int64]*runningTask),
		sem:           make(chan struct{}, maxConcurrent),
	}
}

func (r *Runner) Start(task *db.Task, files []db.FileEntry, cfg *db.TranscodeConfig) error {
	r.mu.Lock()
	if _, already := r.running[task.ID]; already {
		r.mu.Unlock()
		return fmt.Errorf("task %d already running", task.ID)
	}
	r.running[task.ID] = &runningTask{} // placeholder, prevents duplicate Start
	r.mu.Unlock()
	go r.run(task, files, cfg)
	return nil
}

func (r *Runner) run(task *db.Task, files []db.FileEntry, cfg *db.TranscodeConfig) {
	r.sem <- struct{}{} // acquire slot first, may block goroutine but not HTTP handler
	defer func() { <-r.sem }()
	defer r.cleanup(task)
	defer func() {
		if rec := recover(); rec != nil {
			r.Logger.Error("task panic", "id", task.ID, "panic", rec)
			r.failTask(task.ID, fmt.Sprintf("internal error: %v", rec))
		}
	}()

	r.Logger.Info("task started", "id", task.ID, "name", task.Name, "encoder", cfg.VideoEncoder)

	for _, file := range files {
		if !file.Selected {
			continue
		}

		pipelineCfg := PipelineConfig{
			SourceFile:      file.SourceFile,
			OutputDir:       task.OutputPath,
			TaskID:          task.ID,
			VideoEncoder:    cfg.VideoEncoder,
			VideoParams:     cfg.VideoParams,
			ChaptersEnabled: cfg.ChaptersEnabled,
		}
		json.Unmarshal([]byte(cfg.AudioTracks), &pipelineCfg.AudioTracks)
		json.Unmarshal([]byte(cfg.SubtitleTracks), &pipelineCfg.SubtitleTracks)

		extractCmds, err := extractStreams(pipelineCfg)
		if err != nil {
			r.failTask(task.ID, err.Error())
			return
		}
		for _, cmd := range extractCmds {
			cmd.Run()
			cmd.Wait()
		}
		r.Logger.Info("streams extracted", "task", task.ID, "file", filepath.Base(file.SourceFile))

	getTotalFrames := func() int64 {
			out, err := exec.Command("/usr/bin/ffprobe",
				"-v", "error",
				"-select_streams", "v:0",
				"-show_entries", "stream=r_frame_rate,duration,nb_frames",
				"-of", "default=nokey=1:noprint_wrappers=1",
				file.SourceFile).Output()
			if err != nil {
				return 0
			}
			parts := strings.Fields(strings.TrimSpace(string(out)))
			if len(parts) >= 3 {
				dur, _ := strconv.ParseFloat(parts[1], 64)
				if dur > 0 {
					if n, err := strconv.ParseInt(parts[2], 10, 64); err == nil && n > 0 {
						return n
					}
					if fr := parts[0]; strings.Contains(fr, "/") {
						frac := strings.SplitN(fr, "/", 2)
						if num, err := strconv.ParseFloat(frac[0], 64); err == nil {
							if den, err := strconv.ParseFloat(frac[1], 64); err == nil && den > 0 {
								return int64(dur * num / den)
							}
						}
					}
				}
			}
			return 0
		}
		totalFrames := getTotalFrames()
		r.Logger.Info("total frames", "file", filepath.Base(file.SourceFile), "frames", totalFrames)

		ffmpegCmd, videoCmd, err := encodeVideo(pipelineCfg)
		if err != nil {
			r.failTask(task.ID, err.Error())
			return
		}
		r.Logger.Info("encode command", "cmd", strings.Join(videoCmd.Args, " "))

		stderrPipe, err := videoCmd.StderrPipe()
		if err != nil {
			r.failTask(task.ID, "stderr pipe: "+err.Error())
			return
		}
		var ffmpegStderr bytes.Buffer
		if ffmpegCmd != nil {
			ffmpegCmd.Stderr = &ffmpegStderr
			if err := ffmpegCmd.Start(); err != nil {
				r.failTask(task.ID, "ffmpeg start failed: "+err.Error())
				return
			}
		}
		if err := videoCmd.Start(); err != nil {
			r.failTask(task.ID, "encoder start failed: "+err.Error())
			return
		}
		// Close parent's pipe fds — children hold their own copies
		if pw, ok := ffmpegCmd.Stdout.(*os.File); ok && pw != nil {
			pw.Close()
		}
		if pr, ok := videoCmd.Stdin.(*os.File); ok && pr != nil {
			pr.Close()
		}

		// Register the running task AFTER Start succeeds
		var stderrBuf bytes.Buffer
		rt := &runningTask{cmd: videoCmd, stderr: &stderrBuf}
		r.mu.Lock()
		r.running[task.ID] = rt
		r.mu.Unlock()

		r.updateTaskPID(task.ID, videoCmd.Process.Pid)
		r.Logger.Info("video encode started", "pid", videoCmd.Process.Pid, "file", filepath.Base(file.SourceFile))

		encodeStart := time.Now()

		// Read stderr in a goroutine to a buffer + parse progress
		go func() {
			buf := make([]byte, 4096)
			for {
				n, err := stderrPipe.Read(buf)
				if n > 0 {
					stderrBuf.Write(buf[:n])
					updateProgress(r, task.ID, string(buf[:n]), encodeStart, totalFrames)
				}
				if err != nil {
					return
				}
			}
		}()

	if err := videoCmd.Wait(); err != nil {
			r.Logger.Error("encoder exited with error", "task", task.ID, "err", err)
		}
		var ffmpegExitErr error
		if ffmpegCmd != nil {
			waitCh := make(chan error, 1)
			go func() { waitCh <- ffmpegCmd.Wait() }()
			select {
			case ffmpegExitErr = <-waitCh:
			case <-time.After(10 * time.Second):
				r.Logger.Error("ffmpeg wait timeout, killing", "task", task.ID)
				ffmpegCmd.Process.Kill()
				ffmpegExitErr = <-waitCh
			}
			if ffmpegExitErr != nil {
				r.Logger.Error("ffmpeg exited with error", "task", task.ID, "err", ffmpegExitErr)
			}
			if ffmpegStderr.Len() > 0 {
				r.Logger.Error("ffmpeg stderr", "task", task.ID, "stderr", ffmpegStderr.String())
			}
		}

		updateProgress(r, task.ID, stderrBuf.String(), encodeStart, totalFrames)

		r.mu.Lock()
		delete(r.running, task.ID)
		r.mu.Unlock()

		baseName := filepath.Base(file.SourceFile)
		ext := filepath.Ext(baseName)
		videoFile := filepath.Join(task.OutputPath, fmt.Sprintf("%s_t%d%s_video.265", baseName[:len(baseName)-len(ext)], task.ID, ext))
		if stderrBuf.Len() > 0 {
			s := stderrBuf.String()
			if len(s) > 500 {
				s = s[:500]
			}
			r.Logger.Info("encoder stderr", "task", task.ID, "stderr", s)
		}
		if _, err := os.Stat(videoFile); os.IsNotExist(err) {
			msg := "video encode produced no output: " + videoFile
			msg += " stderr: " + stderrBuf.String()
			if ffmpegCmd != nil && ffmpegStderr.Len() > 0 {
				msg += " ffmpeg: " + ffmpegStderr.String()
			}
			if ffmpegExitErr != nil {
				msg += " ffmpeg_exit: " + ffmpegExitErr.Error()
			}
			r.failTask(task.ID, msg)
			return
		}

		muxCmd := muxMKV(videoFile, task.OutputPath, file.SourceFile, task.ID)
		if err := muxCmd.Run(); err != nil {
			r.Logger.Error("mux failed", "task", task.ID, "error", err)
			r.failTask(task.ID, "muxing failed: "+err.Error())
			return
		}
		os.Remove(videoFile)
		r.Logger.Info("file completed", "task", task.ID, "file", filepath.Base(file.SourceFile), "output", videoFile)
	}

	r.updateTaskStatus(task.ID, "completed")
	r.Logger.Info("task completed", "id", task.ID, "name", task.Name)
}

func updateProgress(r *Runner, taskID int64, text string, encodeStart time.Time, totalFrames int64) {
	var pct float64
	lines := strings.Split(text, "\r")
	for _, line := range lines {
		if m := x265PercentRe.FindStringSubmatch(line); m != nil {
			pct, _ = strconv.ParseFloat(m[1], 64)
		} else if m := x264FrameRe.FindStringSubmatch(line); m != nil {
			frames, _ := strconv.ParseInt(m[1], 10, 64)
			if totalFrames > 0 {
				pct = float64(frames) / float64(totalFrames) * 100
			}
		}
	}
	if pct > 0 {
		db.UpdateTask(r.DB, taskID, map[string]any{"progress": pct / 100, "status": "running"})
		if r.Hub != nil {
			r.Hub.BroadcastProgress(taskID, pct/100)
		}

		eta := calcETA(pct, encodeStart)
		if eta >= 0 {
			db.UpdateTask(r.DB, taskID, map[string]any{"estimated_eta": strconv.Itoa(int(eta))})
		}
	}
}

func calcETA(pct float64, encodeStart time.Time) int {
	if pct <= 0 || pct > 100 {
		return -1
	}
	elapsed := time.Since(encodeStart).Seconds()
	if elapsed < 2 {
		return -1
	}
	remaining := elapsed * (100 - pct) / pct
	return int(remaining)
}

func (r *Runner) Pause(taskID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rt, ok := r.running[taskID]; ok {
		return pauseProcess(rt.cmd)
	}
	return nil
}

func (r *Runner) Cancel(taskID int64) error {
	r.mu.Lock()
	rt, ok := r.running[taskID]
	if ok {
		delete(r.running, taskID)
	}
	r.mu.Unlock()
	if ok {
		return cancelProcess(rt.cmd)
	}
	return nil
}

func (r *Runner) updateTaskStatus(taskID int64, status string) {
	db.UpdateTask(r.DB, taskID, map[string]any{"status": status})
}

func (r *Runner) updateTaskPID(taskID int64, pid int) {
	db.UpdateTask(r.DB, taskID, map[string]any{"pid": pid})
}

func (r *Runner) failTask(taskID int64, errMsg string) {
	r.mu.Lock()
	rt := r.running[taskID]
	r.mu.Unlock()

	stderrInfo := ""
	if rt != nil && rt.stderr != nil && rt.stderr.Len() > 0 {
		stderrInfo = ": " + rt.stderr.String()
	}

	fullMsg := errMsg + stderrInfo
	if len(fullMsg) > 2000 {
		fullMsg = fullMsg[:2000]
	}

	db.UpdateTask(r.DB, taskID, map[string]any{"status": "failed", "error_msg": fullMsg})
	r.Logger.Error("task failed", "id", taskID, "error", errMsg)
}

func (r *Runner) cleanup(task *db.Task) {
	if task.OutputPath != "" {
		idSuffix := fmt.Sprintf("_t%d", task.ID)
		entries, err := os.ReadDir(task.OutputPath)
		if err == nil {
			for _, e := range entries {
				if strings.Contains(e.Name(), idSuffix) && strings.HasSuffix(e.Name(), "_video.265") {
					os.Remove(filepath.Join(task.OutputPath, e.Name()))
				}
			}
		}
	}
}

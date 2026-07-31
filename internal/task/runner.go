package task

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

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
	r.sem <- struct{}{}
	go r.run(task, files, cfg)
	return nil
}

func (r *Runner) run(task *db.Task, files []db.FileEntry, cfg *db.TranscodeConfig) {
	defer func() { <-r.sem }()
	defer r.cleanup(task)

	r.updateTaskStatus(task.ID, "running")

	for _, file := range files {
		if !file.Selected {
			continue
		}

		pipelineCfg := PipelineConfig{
			SourceFile:   file.SourceFile,
			OutputDir:    task.OutputPath,
			VideoEncoder: cfg.VideoEncoder,
			VideoParams:  cfg.VideoParams,
		}

		extractCmds, err := extractStreams(pipelineCfg)
		if err != nil {
			r.failTask(task.ID, err.Error())
			return
		}
		for _, cmd := range extractCmds {
			cmd.Run()
			cmd.Wait()
		}

		videoCmd, err := encodeVideo(pipelineCfg)
		if err != nil {
			r.failTask(task.ID, err.Error())
			return
		}

		var stderrBuf bytes.Buffer
		videoCmd.Stderr = &stderrBuf
		stderrPipe, _ := videoCmd.StderrPipe()

		rt := &runningTask{cmd: videoCmd, stderr: &stderrBuf}
		r.mu.Lock()
		r.running[task.ID] = rt
		r.mu.Unlock()

		videoCmd.Start()

		r.updateTaskPID(task.ID, videoCmd.Process.Pid)

		progressCh := make(chan ProgressInfo, 100)
		go parseProgress(stderrPipe, 0, progressCh)

		for pi := range progressCh {
			db.UpdateTask(r.DB, task.ID, map[string]any{"progress": pi.Percent / 100})
			if r.Hub != nil {
				r.Hub.BroadcastProgress(task.ID, pi.Percent/100)
			}
		}

		videoCmd.Wait()

		r.mu.Lock()
		delete(r.running, task.ID)
		r.mu.Unlock()

		videoFile := filepath.Join(task.OutputPath, filepath.Base(file.SourceFile)) + "_video.265"
		muxCmd := muxMKV(videoFile, task.OutputPath, file.SourceFile)
		if err := muxCmd.Run(); err != nil {
			r.failTask(task.ID, "muxing failed: "+err.Error())
			return
		}
	}

	r.updateTaskStatus(task.ID, "completed")
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
	defer r.mu.Unlock()
	if rt, ok := r.running[taskID]; ok {
		delete(r.running, taskID)
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
	if len(fullMsg) > 1000 {
		fullMsg = fullMsg[:1000]
	}

	db.UpdateTask(r.DB, taskID, map[string]any{"status": "failed", "error_msg": fullMsg})
	r.Logger.Error("task failed", "id", taskID, "error", errMsg)
}

func (r *Runner) cleanup(task *db.Task) {
	if task.Status == "failed" {
		os.RemoveAll(task.OutputPath)
	}
}

var _ = context.Background

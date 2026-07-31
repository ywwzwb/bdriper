package task

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/exec"
	"sync"

	"github.com/zwb/bdriper/internal/db"
)

type Runner struct {
	DB           *sql.DB
	Logger       *slog.Logger
	MaxConcurrent int
	mu           sync.Mutex
	running      map[int64]*exec.Cmd
	sem          chan struct{}
}

func NewRunner(database *sql.DB, logger *slog.Logger, maxConcurrent int) *Runner {
	return &Runner{
		DB:            database,
		Logger:        logger,
		MaxConcurrent: maxConcurrent,
		running:       make(map[int64]*exec.Cmd),
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

		r.mu.Lock()
		r.running[task.ID] = videoCmd
		r.mu.Unlock()

		stderr, _ := videoCmd.StderrPipe()
		videoCmd.Start()

		r.updateTaskPID(task.ID, videoCmd.Process.Pid)

		progressCh := make(chan ProgressInfo, 100)
		go parseProgress(stderr, 0, progressCh)

		for pi := range progressCh {
			db.UpdateTask(r.DB, task.ID, map[string]any{"progress": pi.Percent / 100})
		}

		videoCmd.Wait()

		r.mu.Lock()
		delete(r.running, task.ID)
		r.mu.Unlock()
	}

	r.updateTaskStatus(task.ID, "completed")
}

func (r *Runner) Pause(taskID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cmd, ok := r.running[taskID]; ok {
		return pauseProcess(cmd)
	}
	return nil
}

func (r *Runner) Cancel(taskID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cmd, ok := r.running[taskID]; ok {
		delete(r.running, taskID)
		return cancelProcess(cmd)
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
	db.UpdateTask(r.DB, taskID, map[string]any{"status": "failed", "error_msg": errMsg})
}

func (r *Runner) cleanup(task *db.Task) {
	if task.Status == "failed" {
		os.RemoveAll(task.OutputPath)
	}
}

var _ = context.Background

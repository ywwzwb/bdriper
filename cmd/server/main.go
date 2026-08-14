package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/zwb/bdriper/internal/api"
	"github.com/zwb/bdriper/internal/config"
	"github.com/zwb/bdriper/internal/db"
	"github.com/zwb/bdriper/internal/log"
	"github.com/zwb/bdriper/internal/preview"
	"github.com/zwb/bdriper/internal/task"
)

func main() {
	dataDir := os.Getenv("BDRI_PER_DATA_DIR")
	if dataDir == "" {
		dataDir = getEnv("DATA_DIR", "")
	}
	if dataDir == "" {
		dataDir = filepath.Join(mustUserHomeDir(), ".bdriper")
	}

	database, err := db.Open(dataDir)
	if err != nil {
		panic("failed to open database: " + err.Error())
	}

	logHandler, err := log.NewLogger(filepath.Join(dataDir, "logs"), "debug", 5, 10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	logger := slog.New(logHandler)
	slog.SetDefault(logger)
	logger.Info("data dir", "path", dataDir)

	taskHub := api.NewHub()

	maxConcurrent := getEnvInt("MAX_CONCURRENT", 2)
	// A saved max_concurrent setting overrides the environment variable.
	if v, err := db.GetSetting(database, "max_concurrent"); err == nil && v != "" {
		if n, aerr := strconv.Atoi(v); aerr == nil && n > 0 {
			maxConcurrent = n
		}
	}
	hubAdapter := &taskHubAdapter{hub: taskHub}
	runner := task.NewRunner(database, logger, hubAdapter, maxConcurrent)

	// Mark orphaned running tasks as failed on startup
	db.RecoverOrphanedTasks(database, logger)

	presetsDir := os.Getenv("BDRI_PER_PRESETS_DIR")
	if presetsDir == "" {
		presetsDir = getEnv("PRESETS_DIR", "presets")
	}
	if err := config.LoadPresets(database, presetsDir); err != nil {
		logger.Warn("failed to load presets", "dir", presetsDir, "error", err)
	}

	preview.StartCleanup(database, 30*time.Minute)
	api.StartCPUPoller(5 * time.Second)

	spaFS := http.Dir(getEnv("WEB_DIST", "web/dist"))

	helpDir := os.Getenv("BDRI_PER_HELP_DIR")
	if helpDir == "" {
		helpDir = getEnv("HELP_DIR", filepath.Join(presetsDir, "..", "docs", "help"))
	}

	srv := &api.Server{
		DB:         database,
		Logger:     logger,
		LogHandler: logHandler,
		TaskHub:    taskHub,
		LogHub:     api.NewHub(),
		Runner:     runner,
		DataDir:    dataDir,
		HelpDir:    helpDir,
		PresetsDir: presetsDir,
		SPAFS:      spaFS,
	}

	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: withMiddleware(mux),
	}

	go func() {
		logger.Info("server starting", "addr", ":8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	logger.Info("shutting down")
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Error("server shutdown error", "err", err)
	}
	if err := database.Close(); err != nil {
		logger.Error("database close error", "err", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func mustUserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	return home
}

type taskHubAdapter struct {
	hub *api.Hub
}

func (a *taskHubAdapter) BroadcastProgress(taskID int64, progress float64) {
	data, _ := json.Marshal(map[string]any{
		"task_id":  taskID,
		"progress": progress,
	})
	a.hub.Broadcast(api.Event{Type: "progress", Data: data})
}

func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		slog.Debug("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

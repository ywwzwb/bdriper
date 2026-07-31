package main

import (
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

	bw := log.NewBroadcastHandler(os.Stdout, slog.LevelInfo)
	logger := slog.New(bw)

	maxConcurrent := getEnvInt("MAX_CONCURRENT", 2)
	runner := task.NewRunner(database, logger, maxConcurrent)

	presetsDir := os.Getenv("BDRI_PER_PRESETS_DIR")
	if presetsDir == "" {
		presetsDir = getEnv("PRESETS_DIR", "presets")
	}
	config.LoadPresets(database, presetsDir)

	preview.StartCleanup(database, 30*time.Minute)

	spaFS := http.Dir(getEnv("WEB_DIST", "web/dist"))

	srv := &api.Server{
		DB:         database,
		Logger:     logger,
		LogHandler: bw,
		TaskHub:    api.NewHub(),
		LogHub:     api.NewHub(),
		Runner:     runner,
		DataDir:    dataDir,
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

func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		slog.Info("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

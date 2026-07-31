package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/zwb/bdriper/internal/db"
)

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	tasks, err := db.ListTasks(s.DB, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name       string `json:"name"`
		SourcePath string `json:"source_path"`
		OutputPath string `json:"output_path"`
		ConfigID   int64  `json:"config_id"`
		Files      []struct {
			SourceFile string `json:"source_file"`
			Streams    string `json:"streams"`
			Selected   bool   `json:"selected"`
		} `json:"files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task := &db.Task{
		Name:       input.Name,
		Status:     "pending",
		SourcePath: input.SourcePath,
		OutputPath: input.OutputPath,
		ConfigID:   input.ConfigID,
	}
	taskID, err := db.CreateTask(s.DB, task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, f := range input.Files {
		fe := &db.FileEntry{
			TaskID:     taskID,
			SourceFile: f.SourceFile,
			Streams:    f.Streams,
			Selected:   f.Selected,
			OutputFile: input.OutputPath + "/" + filepath.Base(f.SourceFile) + ".mkv",
		}
		db.CreateFileEntry(s.DB, fe)
	}

	if s.Runner != nil {
		task.ID = taskID
		files, _ := db.ListFileEntries(s.DB, taskID)
		cfg, _ := db.GetConfig(s.DB, input.ConfigID)
		s.Runner.Start(task, files, cfg)
	}

	task.ID = taskID
	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	task, err := db.GetTask(s.DB, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	files, _ := db.ListFileEntries(s.DB, id)
	writeJSON(w, http.StatusOK, map[string]any{"task": task, "files": files})
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var input struct {
		Status string `json:"status"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	if s.Runner != nil {
		if input.Status == "paused" {
			s.Runner.Pause(id)
		} else if input.Status == "cancelled" {
			s.Runner.Cancel(id)
		}
	}

	db.UpdateTask(s.DB, id, map[string]any{"status": input.Status})
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if s.Runner != nil {
		s.Runner.Cancel(id)
	}
	db.DeleteTask(s.DB, id)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleDeleteCompleted(w http.ResponseWriter, r *http.Request) {
	db.DeleteCompleted(s.DB)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleRetryTask(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	db.UpdateTask(s.DB, id, map[string]any{"status": "pending", "error_msg": ""})
	task, _ := db.GetTask(s.DB, id)
	files, _ := db.ListFileEntries(s.DB, id)
	cfg, _ := db.GetConfig(s.DB, task.ConfigID)
	if s.Runner != nil {
		s.Runner.Start(task, files, cfg)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "retrying"})
}

func (s *Server) handleBatchTasks(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDs    []int64 `json:"ids"`
		Action string  `json:"action"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	switch input.Action {
	case "delete":
		db.BatchDelete(s.DB, input.IDs)
	case "pause":
		for _, id := range input.IDs {
			if s.Runner != nil {
				s.Runner.Pause(id)
			}
			db.UpdateTask(s.DB, id, map[string]any{"status": "paused"})
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

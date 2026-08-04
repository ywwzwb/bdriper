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

	if s.Runner != nil {
		// Create one task per selected file
		var created []db.Task
		for _, f := range input.Files {
			if !f.Selected {
				continue
			}
			fileName := filepath.Base(f.SourceFile)
			fileTask := &db.Task{
				Name:       input.Name + " - " + fileName,
				Status:     "pending",
				SourcePath: input.SourcePath,
				OutputPath: input.OutputPath,
				ConfigID:   input.ConfigID,
			}
			ftID, err := db.CreateTask(s.DB, fileTask)
			if err != nil {
				continue
			}
			fe := &db.FileEntry{
				TaskID:     ftID,
				SourceFile: f.SourceFile,
				Streams:    f.Streams,
				Selected:   true,
				OutputFile: input.OutputPath + "/" + fileName + ".mkv",
			}
			db.CreateFileEntry(s.DB, fe)

			cfg, err := db.GetConfig(s.DB, input.ConfigID)
			if err != nil || cfg == nil || cfg.ID == 0 {
				cfg = &db.TranscodeConfig{
					ID:           0,
					VideoEncoder: "x264",
					VideoParams:  `{"crf": 23, "preset": "fast"}`,
				}
			}
			fileTask.ID = ftID
			s.Runner.Start(fileTask, []db.FileEntry{*fe}, cfg)
			created = append(created, *fileTask)
		}
		if len(created) > 0 {
			writeJSON(w, http.StatusCreated, created)
			return
		}
		writeError(w, http.StatusBadRequest, "no files selected")
		return
	}
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
	task, err := db.GetTask(s.DB, id)
	if err != nil || task == nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	files, _ := db.ListFileEntries(s.DB, id)
	cfg, err := db.GetConfig(s.DB, task.ConfigID)
	if err != nil || cfg == nil {
		writeError(w, http.StatusNotFound, "config not found")
		return
	}
	if s.Runner != nil {
		if err := s.Runner.Start(task, files, cfg); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
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

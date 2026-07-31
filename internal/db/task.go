package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Task struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	SourcePath   string    `json:"source_path"`
	OutputPath   string    `json:"output_path"`
	Progress     float64   `json:"progress"`
	EstimatedETA string    `json:"estimated_eta"`
	PID          int       `json:"pid"`
	ConfigID     int64     `json:"config_id"`
	ErrorMsg     string    `json:"error_msg"`
	Deleted      bool      `json:"deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type FileEntry struct {
	ID         int64   `json:"id"`
	TaskID     int64   `json:"task_id"`
	SourceFile string  `json:"source_file"`
	OutputFile string  `json:"output_file"`
	Streams    string  `json:"streams"`
	Selected   bool    `json:"selected"`
	Progress   float64 `json:"progress"`
}

func CreateTask(db *sql.DB, t *Task) (int64, error) {
	res, err := db.Exec(`INSERT INTO tasks (name, status, source_path, output_path, config_id) VALUES (?,?,?,?,?)`,
		t.Name, t.Status, t.SourcePath, t.OutputPath, t.ConfigID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ListTasks(db *sql.DB, status string) ([]Task, error) {
	query := `SELECT id,name,status,source_path,output_path,progress,estimated_eta,pid,config_id,error_msg,created_at,updated_at FROM tasks WHERE deleted=0`
	args := []any{}
	if status != "" && status != "all" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var t Task
		rows.Scan(&t.ID, &t.Name, &t.Status, &t.SourcePath, &t.OutputPath, &t.Progress, &t.EstimatedETA, &t.PID, &t.ConfigID, &t.ErrorMsg, &t.CreatedAt, &t.UpdatedAt)
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func GetTask(db *sql.DB, id int64) (*Task, error) {
	t := &Task{}
	err := db.QueryRow(`SELECT id,name,status,source_path,output_path,progress,estimated_eta,pid,config_id,error_msg,created_at,updated_at FROM tasks WHERE id=? AND deleted=0`, id).
		Scan(&t.ID, &t.Name, &t.Status, &t.SourcePath, &t.OutputPath, &t.Progress, &t.EstimatedETA, &t.PID, &t.ConfigID, &t.ErrorMsg, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func UpdateTask(db *sql.DB, id int64, updates map[string]any) error {
	setClauses := ""
	args := []any{}
	for k, v := range updates {
		if setClauses != "" {
			setClauses += ", "
		}
		setClauses += k + " = ?"
		args = append(args, v)
	}
	args = append(args, id)
	_, err := db.Exec("UPDATE tasks SET "+setClauses+", updated_at=CURRENT_TIMESTAMP WHERE id=?", args...)
	return err
}

func DeleteTask(db *sql.DB, id int64) error {
	_, err := db.Exec("UPDATE tasks SET deleted=1, updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
	return err
}

func BatchDelete(db *sql.DB, ids []int64) error {
	tx, _ := db.Begin()
	for _, id := range ids {
		tx.Exec("UPDATE tasks SET deleted=1, updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
	}
	return tx.Commit()
}

func DeleteCompleted(db *sql.DB) error {
	_, err := db.Exec("UPDATE tasks SET deleted=1 WHERE status='completed'")
	return err
}

func CreateFileEntry(db *sql.DB, fe *FileEntry) (int64, error) {
	res, err := db.Exec(`INSERT INTO file_entries (task_id, source_file, output_file, streams, selected) VALUES (?,?,?,?,?)`,
		fe.TaskID, fe.SourceFile, fe.OutputFile, fe.Streams, fe.Selected)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ListFileEntries(db *sql.DB, taskID int64) ([]FileEntry, error) {
	rows, err := db.Query(`SELECT id,task_id,source_file,output_file,streams,selected,progress FROM file_entries WHERE task_id=?`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []FileEntry
	for rows.Next() {
		var fe FileEntry
		rows.Scan(&fe.ID, &fe.TaskID, &fe.SourceFile, &fe.OutputFile, &fe.Streams, &fe.Selected, &fe.Progress)
		entries = append(entries, fe)
	}
	return entries, nil
}

func UpdateFileEntry(db *sql.DB, id int64, updates map[string]any) error {
	setClauses := ""
	args := []any{}
	for k, v := range updates {
		if setClauses != "" {
			setClauses += ", "
		}
		setClauses += k + " = ?"
		args = append(args, v)
	}
	args = append(args, id)
	_, err := db.Exec("UPDATE file_entries SET "+setClauses+" WHERE id=?", args...)
	return err
}

var _ = json.Marshal

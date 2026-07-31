package db

import (
	"database/sql"
	"time"
)

type PreviewJob struct {
	ID         int64     `json:"id"`
	TaskID     int64     `json:"task_id"`
	SourceFile string    `json:"source_file"`
	StartTime  string    `json:"start_time"`
	Duration   int       `json:"duration"`
	OutputFile string    `json:"output_file"`
	Status     string    `json:"status"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

func CreatePreview(db *sql.DB, p *PreviewJob) (int64, error) {
	res, err := db.Exec(`INSERT INTO preview_jobs (task_id, source_file, start_time, duration, output_file, status, expires_at) VALUES (?,?,?,?,?,?,?)`,
		p.TaskID, p.SourceFile, p.StartTime, p.Duration, p.OutputFile, p.Status, p.ExpiresAt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetPreview(db *sql.DB, id int64) (*PreviewJob, error) {
	p := &PreviewJob{}
	err := db.QueryRow(`SELECT id,task_id,source_file,start_time,duration,output_file,status,expires_at,created_at FROM preview_jobs WHERE id=?`, id).
		Scan(&p.ID, &p.TaskID, &p.SourceFile, &p.StartTime, &p.Duration, &p.OutputFile, &p.Status, &p.ExpiresAt, &p.CreatedAt)
	return p, err
}

func UpdatePreview(db *sql.DB, id int64, updates map[string]any) error {
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
	_, err := db.Exec("UPDATE preview_jobs SET "+setClauses+" WHERE id=?", args...)
	return err
}

func ListExpiredPreviews(db *sql.DB) ([]PreviewJob, error) {
	rows, err := db.Query(`SELECT id,output_file FROM preview_jobs WHERE expires_at < CURRENT_TIMESTAMP`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []PreviewJob
	for rows.Next() {
		var j PreviewJob
		rows.Scan(&j.ID, &j.OutputFile)
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func DeletePreview(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM preview_jobs WHERE id=?", id)
	return err
}

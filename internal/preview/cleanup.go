package preview

import (
	"database/sql"
	"log/slog"
	"os"
	"time"
)

func StartCleanup(db *sql.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			rows, err := db.Query(`SELECT id, output_file FROM preview_jobs WHERE expires_at < CURRENT_TIMESTAMP`)
			if err != nil {
				slog.Warn("preview cleanup query failed", "error", err)
				continue
			}
			for rows.Next() {
				var id int64
				var outputFile string
				rows.Scan(&id, &outputFile)
				if err := os.Remove(outputFile); err != nil {
					slog.Warn("failed to remove expired preview file", "file", outputFile, "error", err)
				}
				db.Exec("DELETE FROM preview_jobs WHERE id=?", id)
			}
			rows.Close()
		}
	}()
}

package preview

import (
	"database/sql"
	"os"
	"time"
)

func StartCleanup(db *sql.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			rows, err := db.Query(`SELECT id, output_file FROM preview_jobs WHERE expires_at < CURRENT_TIMESTAMP`)
			if err != nil {
				continue
			}
			for rows.Next() {
				var id int64
				var outputFile string
				rows.Scan(&id, &outputFile)
				os.Remove(outputFile)
				db.Exec("DELETE FROM preview_jobs WHERE id=?", id)
			}
			rows.Close()
		}
	}()
}

package config

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/zwb/bdriper/internal/db"
)

func LoadPresets(database *sql.DB, presetsDir string) error {
	entries, err := os.ReadDir(presetsDir)
	if err != nil {
		slog.Warn("failed to read presets directory", "dir", presetsDir, "error", err)
		return err
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(presetsDir, e.Name()))
		if err != nil {
			slog.Warn("failed to read preset file", "file", e.Name(), "error", err)
			continue
		}
		var p db.PresetTemplate
		json.Unmarshal(data, &p)
		database.Exec(`INSERT OR IGNORE INTO preset_templates (name,encoder,mode,params,description,builtin) VALUES (?,?,?,?,?,1)`,
			p.Name, p.Encoder, p.Mode, string(data), p.Description)
	}
	slog.Info("presets loaded", "dir", presetsDir, "count", len(entries))
	return nil
}

func ListPresets(database *sql.DB) ([]db.PresetTemplate, error) {
	rows, err := database.Query(`SELECT name,encoder,mode,params,description,builtin FROM preset_templates ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var presets []db.PresetTemplate
	for rows.Next() {
		var p db.PresetTemplate
		var data string
		rows.Scan(&p.Name, &p.Encoder, &p.Mode, &data, &p.Description, &p.Builtin)
		// Extract inner params from preset JSON wrapper
		var wrapper struct {
			Params map[string]any `json:"params"`
		}
		if err := json.Unmarshal([]byte(data), &wrapper); err == nil && wrapper.Params != nil {
			p.Params = wrapper.Params
		} else {
			json.Unmarshal([]byte(data), &p.Params)
		}
		presets = append(presets, p)
	}
	return presets, nil
}

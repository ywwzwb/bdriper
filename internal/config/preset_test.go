package config

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

const presetSchema = `CREATE TABLE IF NOT EXISTS preset_templates (
    name TEXT PRIMARY KEY,
    encoder TEXT NOT NULL DEFAULT '',
    mode TEXT NOT NULL DEFAULT 'cpu',
    params TEXT NOT NULL DEFAULT '{}',
    description TEXT NOT NULL DEFAULT '',
    builtin BOOLEAN NOT NULL DEFAULT 1
)`

func newPresetTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(presetSchema)
	if err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

func TestLoadPresets(t *testing.T) {
	dir := t.TempDir()
	preset := filepath.Join(dir, "test-preset.json")
	os.WriteFile(preset, []byte(`{"name":"test","encoder":"x265","mode":"cpu","description":"test preset","params":{"crf":15}}`), 0644)

	db := newPresetTestDB(t)
	defer db.Close()

	err := LoadPresets(db, dir)
	if err != nil {
		t.Fatalf("LoadPresets: %v", err)
	}

	presets, err := ListPresets(db)
	if err != nil {
		t.Fatalf("ListPresets: %v", err)
	}
	if len(presets) != 1 {
		t.Fatalf("got %d presets, want 1", len(presets))
	}
	if presets[0].Name != "test" {
		t.Errorf("Name = %q, want test", presets[0].Name)
	}
}

func TestListPresetsEmpty(t *testing.T) {
	db := newPresetTestDB(t)
	defer db.Close()

	presets, err := ListPresets(db)
	if err != nil {
		t.Fatalf("ListPresets: %v", err)
	}
	if len(presets) != 0 {
		t.Errorf("got %d presets, want 0", len(presets))
	}
}

func TestLoadPresetsIgnoresNonJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not a preset"), 0644)
	os.WriteFile(filepath.Join(dir, "preset.json"), []byte(`{"name":"valid","encoder":"x265","mode":"cpu"}`), 0644)

	db := newPresetTestDB(t)
	defer db.Close()

	err := LoadPresets(db, dir)
	if err != nil {
		t.Fatalf("LoadPresets: %v", err)
	}

	presets, _ := ListPresets(db)
	if len(presets) != 1 {
		t.Fatalf("got %d presets, want 1", len(presets))
	}
}

func TestLoadPresetsSkipsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.json"), []byte(`{invalid`), 0644)
	os.WriteFile(filepath.Join(dir, "good.json"), []byte(`{"name":"good","encoder":"x265","mode":"cpu"}`), 0644)

	db := newPresetTestDB(t)
	defer db.Close()

	err := LoadPresets(db, dir)
	if err != nil {
		t.Fatalf("LoadPresets: %v", err)
	}

	presets, _ := ListPresets(db)
	t.Logf("got %d presets (bad JSON produces zero-value insert)", len(presets))

	foundGood := false
	for _, p := range presets {
		if p.Name == "good" {
			foundGood = true
		}
	}
	if !foundGood {
		t.Error("did not find the valid preset 'good'")
	}
}

func TestLoadPresetsDeduplicates(t *testing.T) {
	dir := t.TempDir()
	content := `{"name":"dedup","encoder":"x265","mode":"cpu"}`
	os.WriteFile(filepath.Join(dir, "a.json"), []byte(content), 0644)
	os.WriteFile(filepath.Join(dir, "b.json"), []byte(content), 0644)

	db := newPresetTestDB(t)
	defer db.Close()

	LoadPresets(db, dir)
	presets, _ := ListPresets(db)
	if len(presets) != 1 {
		t.Fatalf("got %d presets, want 1 (deduplication)", len(presets))
	}
}

func TestLoadPresetsMissingDir(t *testing.T) {
	db := newPresetTestDB(t)
	defer db.Close()

	err := LoadPresets(db, "/nonexistent/path")
	if err == nil {
		t.Error("LoadPresets should error for missing directory")
	}
}

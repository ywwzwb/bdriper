package db

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_journal=WAL&_foreign_keys=on")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.SetMaxOpenConns(1)
	_, err = db.Exec(schemaSQL)
	if err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

func TestCreateAndGetTask(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	task := &Task{Name: "test-task", Status: "pending", SourcePath: "/input/test", OutputPath: "/output/test"}
	id, err := CreateTask(db, task)
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	got, err := GetTask(db, id)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got.Name != "test-task" {
		t.Errorf("Name = %q, want %q", got.Name, "test-task")
	}
	if got.Status != "pending" {
		t.Errorf("Status = %q, want pending", got.Status)
	}
}

func TestListTasks(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	CreateTask(db, &Task{Name: "running", Status: "running", SourcePath: "/a", OutputPath: "/a"})
	CreateTask(db, &Task{Name: "done", Status: "completed", SourcePath: "/b", OutputPath: "/b"})
	CreateTask(db, &Task{Name: "pending", Status: "pending", SourcePath: "/c", OutputPath: "/c"})

	tasks, err := ListTasks(db, "running")
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	if tasks[0].Name != "running" {
		t.Errorf("Name = %q", tasks[0].Name)
	}
}

func TestListTasksAll(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	CreateTask(db, &Task{Name: "a", Status: "running", SourcePath: "/a", OutputPath: "/a"})
	CreateTask(db, &Task{Name: "b", Status: "completed", SourcePath: "/b", OutputPath: "/b"})

	tasks, err := ListTasks(db, "all")
	if err != nil {
		t.Fatalf("ListTasks all: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}
}

func TestUpdateTask(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	id, _ := CreateTask(db, &Task{Name: "test", Status: "pending"})

	err := UpdateTask(db, id, map[string]any{"status": "running", "progress": 0.5})
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got, _ := GetTask(db, id)
	if got.Status != "running" {
		t.Errorf("Status = %q", got.Status)
	}
}

func TestDeleteTask(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	id, _ := CreateTask(db, &Task{Name: "test", Status: "running"})
	DeleteTask(db, id)

	_, err := GetTask(db, id)
	if err == nil {
		t.Error("GetTask should fail after soft delete")
	}
}

func TestFileEntryCRUD(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	tid, _ := CreateTask(db, &Task{Name: "t", Status: "pending"})

	fe := &FileEntry{TaskID: tid, SourceFile: "/input/a.m2ts", OutputFile: "/output/a.mkv", Selected: true}
	fid, err := CreateFileEntry(db, fe)
	if err != nil {
		t.Fatalf("CreateFileEntry: %v", err)
	}
	if fid == 0 {
		t.Error("expected non-zero id")
	}

	entries, _ := ListFileEntries(db, tid)
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
}

func TestConfigCRUD(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	cfg := &TranscodeConfig{Name: "test-cfg", EncoderType: "cpu", VideoEncoder: "x265"}
	id, err := CreateConfig(db, cfg)
	if err != nil {
		t.Fatalf("CreateConfig: %v", err)
	}

	got, _ := GetConfig(db, id)
	if got.Name != "test-cfg" {
		t.Errorf("Name = %q", got.Name)
	}

	cfgs, _ := ListConfigs(db)
	if len(cfgs) != 1 {
		t.Errorf("got %d configs, want 1", len(cfgs))
	}
}

func TestConfigDelete(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	cfg := &TranscodeConfig{Name: "del-me", EncoderType: "cpu", VideoEncoder: "x265"}
	id, _ := CreateConfig(db, cfg)

	err := DeleteConfig(db, id)
	if err != nil {
		t.Fatalf("DeleteConfig: %v", err)
	}

	_, err = GetConfig(db, id)
	if err == nil {
		t.Error("GetConfig should fail after delete")
	}
}

func TestConfigUpdate(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	cfg := &TranscodeConfig{Name: "old-name", EncoderType: "cpu", VideoEncoder: "x265"}
	id, _ := CreateConfig(db, cfg)

	err := UpdateConfig(db, id, map[string]any{"name": "new-name"})
	if err != nil {
		t.Fatalf("UpdateConfig: %v", err)
	}

	got, _ := GetConfig(db, id)
	if got.Name != "new-name" {
		t.Errorf("Name = %q, want new-name", got.Name)
	}
}

func TestSettingsCRUD(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	SetSetting(db, "max_concurrent", "3")
	val, err := GetSetting(db, "max_concurrent")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if val != "3" {
		t.Errorf("value = %q, want 3", val)
	}

	settings, _ := ListSettings(db)
	if settings["max_concurrent"] != "3" {
		t.Error("ListSettings missing key")
	}
}

func TestSettingsUpdate(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	SetSetting(db, "key1", "val1")
	SetSetting(db, "key1", "val2")

	val, _ := GetSetting(db, "key1")
	if val != "val2" {
		t.Errorf("value = %q, want val2", val)
	}
}

func TestGetSettingNotFound(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	_, err := GetSetting(db, "nonexistent")
	if err == nil {
		t.Error("GetSetting should error for nonexistent key")
	}
}

func TestPreviewCRUD(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	p := &PreviewJob{TaskID: 1, SourceFile: "/a.m2ts", OutputFile: "/tmp/p.mkv", Status: "completed"}
	id, err := CreatePreview(db, p)
	if err != nil {
		t.Fatalf("CreatePreview: %v", err)
	}

	got, _ := GetPreview(db, id)
	if got.SourceFile != "/a.m2ts" {
		t.Errorf("SourceFile = %q", got.SourceFile)
	}
}

func TestPreviewUpdate(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	p := &PreviewJob{TaskID: 2, SourceFile: "/b.m2ts", Status: "pending"}
	id, _ := CreatePreview(db, p)

	err := UpdatePreview(db, id, map[string]any{"status": "processing"})
	if err != nil {
		t.Fatalf("UpdatePreview: %v", err)
	}

	got, _ := GetPreview(db, id)
	if got.Status != "processing" {
		t.Errorf("Status = %q, want processing", got.Status)
	}
}

func TestPreviewDelete(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	p := &PreviewJob{TaskID: 3, SourceFile: "/c.m2ts", Status: "done"}
	id, _ := CreatePreview(db, p)

	DeletePreview(db, id)
	_, err := GetPreview(db, id)
	if err == nil {
		t.Error("GetPreview should fail after delete")
	}
}

func TestListExpiredPreviews(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	p := &PreviewJob{TaskID: 4, SourceFile: "/d.m2ts", Status: "completed"}
	_, err := CreatePreview(db, p)
	if err != nil {
		t.Fatalf("CreatePreview: %v", err)
	}

	jobs, err := ListExpiredPreviews(db)
	if err != nil {
		t.Fatalf("ListExpiredPreviews: %v", err)
	}
	t.Logf("found %d expired previews (zero ExpiresAt defaults to epoch, considered expired)", len(jobs))
}

func TestDeleteCompleted(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	CreateTask(db, &Task{Name: "done1", Status: "completed"})
	CreateTask(db, &Task{Name: "done2", Status: "completed"})
	CreateTask(db, &Task{Name: "running", Status: "running"})

	DeleteCompleted(db)

	all, _ := ListTasks(db, "all")
	if len(all) != 1 {
		t.Errorf("got %d tasks after DeleteCompleted, want 1", len(all))
	}
}

func TestBatchDelete(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	id1, _ := CreateTask(db, &Task{Name: "del1", Status: "pending"})
	id2, _ := CreateTask(db, &Task{Name: "del2", Status: "pending"})
	id3, _ := CreateTask(db, &Task{Name: "keep", Status: "pending"})

	BatchDelete(db, []int64{id1, id2})

	_, err := GetTask(db, id1)
	if err == nil {
		t.Error("GetTask id1 should fail after batch delete")
	}
	_, err = GetTask(db, id2)
	if err == nil {
		t.Error("GetTask id2 should fail after batch delete")
	}
	_, err = GetTask(db, id3)
	if err != nil {
		t.Errorf("GetTask id3 should succeed: %v", err)
	}
}

func TestUpdateFileEntry(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	tid, _ := CreateTask(db, &Task{Name: "t", Status: "pending"})
	fid, _ := CreateFileEntry(db, &FileEntry{TaskID: tid, SourceFile: "/s.m2ts", OutputFile: "/o.mkv", Selected: false})

	err := UpdateFileEntry(db, fid, map[string]any{"selected": true, "progress": 0.75})
	if err != nil {
		t.Fatalf("UpdateFileEntry: %v", err)
	}

	entries, _ := ListFileEntries(db, tid)
	if len(entries) != 1 {
		t.Fatalf("got %d entries", len(entries))
	}
	if !entries[0].Selected {
		t.Error("Selected should be true")
	}
	if entries[0].Progress != 0.75 {
		t.Errorf("Progress = %f, want 0.75", entries[0].Progress)
	}
}

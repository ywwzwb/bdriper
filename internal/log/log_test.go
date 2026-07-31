package log

import (
	"log/slog"
	"testing"
	"time"
)

func TestBroadcastHandlerSubscribe(t *testing.T) {
	h := NewBroadcastHandler(nilWriter{}, slog.LevelInfo)

	ch := h.Subscribe()
	if ch == nil {
		t.Fatal("Subscribe returned nil channel")
	}

	h.Unsubscribe(ch)
}

func TestBroadcastHandlerHandle(t *testing.T) {
	h := NewBroadcastHandler(nilWriter{}, slog.LevelInfo)

	ch := h.Subscribe()
	defer h.Unsubscribe(ch)

	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "test message", 0)
	h.Handle(nilContext{}, r)

	select {
	case entry := <-ch:
		if entry.Msg != "test message" {
			t.Errorf("Msg = %q", entry.Msg)
		}
		if entry.Level != "INFO" {
			t.Errorf("Level = %q", entry.Level)
		}
	default:
		t.Error("no entry received on channel")
	}
}

func TestBroadcastHandlerLevel(t *testing.T) {
	h := NewBroadcastHandler(nilWriter{}, slog.LevelWarn)

	if !h.Enabled(nilContext{}, slog.LevelError) {
		t.Error("Error should be enabled for Warn level handler")
	}
	if h.Enabled(nilContext{}, slog.LevelInfo) {
		t.Error("Info should not be enabled for Warn level handler")
	}
}

func TestBroadcastHandlerSetLevel(t *testing.T) {
	h := NewBroadcastHandler(nilWriter{}, slog.LevelWarn)

	h.SetLevel(slog.LevelDebug)
	if !h.Enabled(nilContext{}, slog.LevelInfo) {
		t.Error("Info should be enabled after SetLevel to Debug")
	}
}

func TestNewLogger(t *testing.T) {
	dir := t.TempDir()
	h, err := NewLogger(dir, "debug", 3, 1)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	if h == nil {
		t.Fatal("NewLogger returned nil")
	}

	ch := h.Subscribe()
	defer h.Unsubscribe(ch)

	r := slog.NewRecord(time.Time{}, slog.LevelInfo, "hello world", 0)
	h.Handle(nilContext{}, r)

	select {
	case entry := <-ch:
		if entry.Msg != "hello world" {
			t.Errorf("Msg = %q", entry.Msg)
		}
	default:
		t.Error("no entry received")
	}
}

func TestNewLoggerDefaultLevel(t *testing.T) {
	dir := t.TempDir()
	h, err := NewLogger(dir, "unknown", 3, 1)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	if !h.Enabled(nilContext{}, slog.LevelInfo) {
		t.Error("should default to Info level")
	}
	if h.Enabled(nilContext{}, slog.LevelDebug) {
		t.Error("Debug should not be enabled at default Info level")
	}
}

func TestRotateWriter(t *testing.T) {
	dir := t.TempDir()
	w, err := NewRotateWriter(dir, "test.log", 3, 1)
	if err != nil {
		t.Fatalf("NewRotateWriter: %v", err)
	}

	data := []byte("hello world\n")
	n, err := w.Write(data)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(data) {
		t.Errorf("wrote %d bytes, want %d", n, len(data))
	}
}

func TestRotateWriterCurrentPath(t *testing.T) {
	dir := t.TempDir()
	w, err := NewRotateWriter(dir, "app.log", 3, 1)
	if err != nil {
		t.Fatalf("NewRotateWriter: %v", err)
	}

	path := w.CurrentPath()
	if path == "" {
		t.Error("CurrentPath returned empty string")
	}
}

func TestMultiHandlerRotate(t *testing.T) {
	dir := t.TempDir()
	h, err := NewLogger(dir, "info", 3, 1)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	h.Rotate()
}

func TestBroadcastHandlerWithAttrs(t *testing.T) {
	h := NewBroadcastHandler(nilWriter{}, slog.LevelInfo)
	h2 := h.WithAttrs([]slog.Attr{slog.String("key", "val")})
	if h2 != h {
		t.Error("WithAttrs should return same handler")
	}
}

func TestBroadcastHandlerWithGroup(t *testing.T) {
	h := NewBroadcastHandler(nilWriter{}, slog.LevelInfo)
	h2 := h.WithGroup("test-group")
	if h2 != h {
		t.Error("WithGroup should return same handler")
	}
}

type nilWriter struct{}

func (nilWriter) Write(p []byte) (int, error) { return len(p), nil }

type nilContext struct{}

func (nilContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (nilContext) Done() <-chan struct{}       { return nil }
func (nilContext) Err() error                  { return nil }
func (nilContext) Value(any) any               { return nil }

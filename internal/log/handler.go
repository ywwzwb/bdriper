package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
)

type LogEntry struct {
	Level string `json:"level"`
	Time  string `json:"time"`
	Msg   string `json:"msg"`
	Raw   string `json:"raw"`
}

type BroadcastHandler struct {
	mu     sync.RWMutex
	subs   []chan LogEntry
	writer io.Writer
	level  slog.Level
}

func NewBroadcastHandler(w io.Writer, level slog.Level) *BroadcastHandler {
	return &BroadcastHandler{
		writer: w,
		level:  level,
	}
}

func (h *BroadcastHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *BroadcastHandler) Handle(_ context.Context, r slog.Record) error {
	var attrs []string
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value))
		return true
	})
	attrStr := ""
	if len(attrs) > 0 {
		attrStr = " " + strings.Join(attrs, " ")
	}
	raw := r.Time.Format("2006-01-02T15:04:05") + " [" + r.Level.String() + "] " + r.Message + attrStr + "\n"

	entry := LogEntry{
		Level: r.Level.String(),
		Time:  r.Time.Format("2006-01-02T15:04:05"),
		Msg:   r.Message + attrStr,
		Raw:   raw,
	}
	h.writer.Write([]byte(entry.Raw))
	h.mu.RLock()
	for _, ch := range h.subs {
		select {
		case ch <- entry:
		default:
		}
	}
	h.mu.RUnlock()
	return nil
}

func (h *BroadcastHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *BroadcastHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *BroadcastHandler) Subscribe() <-chan LogEntry {
	ch := make(chan LogEntry, 256)
	h.mu.Lock()
	h.subs = append(h.subs, ch)
	h.mu.Unlock()
	return ch
}

func (h *BroadcastHandler) Unsubscribe(ch <-chan LogEntry) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, sub := range h.subs {
		if sub == ch {
			h.subs = append(h.subs[:i], h.subs[i+1:]...)
			close(sub)
			return
		}
	}
}

func (h *BroadcastHandler) SetLevel(level slog.Level) {
	h.mu.Lock()
	h.level = level
	h.mu.Unlock()
}

type MultiHandler struct {
	*BroadcastHandler
	writer *RotateWriter
}

func NewLogger(dataDir, level string, maxFiles int, maxSizeMB int64) (*MultiHandler, error) {
	w, err := NewRotateWriter(dataDir, "bdriper.log", maxFiles, maxSizeMB)
	if err != nil {
		return nil, err
	}
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	bh := NewBroadcastHandler(io.MultiWriter(w, os.Stdout), l)
	return &MultiHandler{BroadcastHandler: bh, writer: w}, nil
}

func (m *MultiHandler) Rotate() {
	m.writer.Rotate()
}

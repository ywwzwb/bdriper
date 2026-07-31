package log

import (
	"context"
	"io"
	"log/slog"
	"sync"
)

type LogEntry struct {
	Level string `json:"level"`
	Time  string `json:"time"`
	Msg   string `json:"msg"`
	Raw   string `json:"raw"`
}

type BroadcastHandler struct {
	mu      sync.RWMutex
	subs    []chan LogEntry
	writer  io.Writer
	level   slog.Level
	formatter func(slog.Record) LogEntry
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
	entry := LogEntry{
		Level: r.Level.String(),
		Time:  r.Time.Format("2006-01-02T15:04:05"),
		Msg:   r.Message,
		Raw:   r.Time.Format("2006-01-02T15:04:05") + " [" + r.Level.String() + "] " + r.Message + "\n",
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

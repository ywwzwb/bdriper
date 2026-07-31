package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type RotateWriter struct {
	mu       sync.Mutex
	dir      string
	baseName string
	maxFiles int
	maxSize  int64
	current  *os.File
	curSize  int64
}

func NewRotateWriter(dir, baseName string, maxFiles int, maxSizeMB int64) (*RotateWriter, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	rw := &RotateWriter{
		dir:      dir,
		baseName: baseName,
		maxFiles: maxFiles,
		maxSize:  maxSizeMB * 1024 * 1024,
	}
	if err := rw.openCurrent(); err != nil {
		return nil, err
	}
	return rw, nil
}

func (rw *RotateWriter) openCurrent() error {
	path := filepath.Join(rw.dir, rw.baseName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if rw.current != nil {
		rw.current.Close()
	}
	rw.current = f
	stat, _ := f.Stat()
	rw.curSize = stat.Size()
	return nil
}

func (rw *RotateWriter) Write(p []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	if rw.curSize+int64(len(p)) >= rw.maxSize {
		rw.rotate()
	}
	n, err := rw.current.Write(p)
	rw.curSize += int64(n)
	return n, err
}

func (rw *RotateWriter) Rotate() {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.rotate()
}

func (rw *RotateWriter) rotate() {
	ts := time.Now().Format("2006-01-02-150405")
	oldPath := filepath.Join(rw.dir, rw.baseName)
	newPath := filepath.Join(rw.dir, fmt.Sprintf("bdriper-%s.log", ts))
	rw.current.Close()
	os.Rename(oldPath, newPath)
	rw.openCurrent()
	rw.cleanup()
}

func (rw *RotateWriter) cleanup() {
	files, _ := filepath.Glob(filepath.Join(rw.dir, "bdriper-*.log"))
	if len(files) <= rw.maxFiles {
		return
	}
	sort.Strings(files)
	for i := 0; i < len(files)-rw.maxFiles; i++ {
		os.Remove(files[i])
	}
}

func (rw *RotateWriter) CurrentPath() string {
	return filepath.Join(rw.dir, rw.baseName)
}

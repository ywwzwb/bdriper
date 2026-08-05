package help

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/yuin/goldmark"
)

var (
	md    = goldmark.New()
	cache sync.Map
)

const htmlWrapper = `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>%s</title>
<style>
body {
    font-family: Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    background: #0F172A;
    color: #F8FAFC;
    padding: 32px 40px;
    line-height: 1.75;
    max-width: 840px;
    margin: 0 auto;
}
h1 { color: #22C55E; font-weight: 600; border-bottom: 1px solid #334155; padding-bottom: 12px; margin-bottom: 24px; font-size: 1.75em; }
h2 { color: #94A3B8; font-weight: 500; margin-top: 32px; font-size: 1.25em; }
h3 { color: #CBD5E1; font-size: 1.1em; margin-top: 24px; }
code { background: #1E293B; padding: 2px 8px; border-radius: 4px; color: #22C55E; font-size: 0.9em; font-family: "Fira Code", "JetBrains Mono", monospace; }
pre { background: #1E293B; border: 1px solid #334155; border-radius: 8px; padding: 16px 20px; overflow-x: auto; }
pre code { background: none; padding: 0; color: #E2E8F0; }
strong { color: #F1F5F9; }
ul, ol { padding-left: 24px; }
li { margin: 8px 0; }
table { width: 100%%; border-collapse: collapse; margin: 16px 0; }
th { background: #1E293B; color: #94A3B8; font-weight: 500; text-align: left; padding: 10px 14px; border-bottom: 2px solid #334155; }
td { padding: 8px 14px; border-bottom: 1px solid #1E293B; }
a { color: #60A5FA; }
blockquote { border-left: 3px solid #475569; padding-left: 16px; color: #94A3B8; margin: 16px 0; }
hr { border: none; border-top: 1px solid #334155; margin: 32px 0; }
</style></head><body>
%s
</body></html>`

func Render(name string, source []byte) []byte {
	if cached, ok := cache.Load(name); ok {
		return cached.([]byte)
	}

	var buf bytes.Buffer
	md.Convert(source, &buf)

	html := []byte(fmt.Sprintf(htmlWrapper, name, buf.String()))
	cache.Store(name, html)
	return html
}

func RenderFile(path string) ([]byte, error) {
	name := filepath.Base(path)
	if cached, ok := cache.Load(name); ok {
		return cached.([]byte), nil
	}

	source, err := os.ReadFile(path)
	if err != nil {
		slog.Debug("help file not found", "path", path, "error", err)
		return nil, err
	}
	return Render(name, source), nil
}

func ClearCache() {
	cache = sync.Map{}
}

# BDRiper 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建一个 Docker 容器，提供 Web UI 将 BDMV 蓝光原盘转码为 BDRip MKV 格式。

**Architecture:** Go 后端单二进制内嵌 Vue 3 SPA，SQLite 持久化，通过 os/exec 调用 ffmpeg/x264/x265/mkvmerge 执行编码管线，WebSocket 实时推送任务进度和日志。

**Tech Stack:** Go 1.22+ (net/http, database/sql, embed, io/fs, log/slog), SQLite (mattn/go-sqlite3), Vue 3 + Vite + Tailwind CSS + Phosphor Icons, Docker (ubuntu:24.04)

## Global Constraints

- 平台: 仅 Linux 容器
- 前端: Vue 3 + Vite + Tailwind CSS + Phosphor Icons
- 后端: Go 1.22+, CGO_ENABLED=0 静态编译
- 数据库: SQLite, 文件存储在 `/data/` 挂载路径
- 卷挂载: /input (BDMV 源), /output (MKV 输出), /data (DB+配置+日志)
- GPU: NVIDIA NVENC, Intel QSV, AMD VCE 多厂商
- 编码器: x264/x265 独立版 (CPU) + ffmpeg (GPU)
- WebSocket: gorilla/websocket, 实时推送任务进度 + 日志
- 编码管线: 4 阶段 hybrid pipeline (流抽取 → 视频编码 → MKV 封装)
- 预览: 临时文件 TTL 30min, 向导退出立即清理
- 日志: slog + 文件轮转 (时间+大小) + WebSocket 广播, 前端实时查看 + 下载
- UI: Dark Mode (OLED), Inter 字体, #0F172A 底色等设计系统
- 图标: Phosphor Icons, 禁止 emoji 作为功能图标

---

## File Structure

```
bdriper/
├── cmd/server/
│   └── main.go                # 入口: 解析参数, 初始化依赖, 启动 HTTP
├── internal/
│   ├── api/
│   │   ├── server.go          # HTTP server 启动, 中间件, 路由注册
│   │   ├── middleware.go      # 请求日志, recovery, CORS (dev)
│   │   ├── overview.go        # GET /api/overview/status
│   │   ├── task.go            # CRUD + batch 操作 handler
│   │   ├── wizard.go          # POST /api/wizard/parse, GET streams
│   │   ├── config.go          # 转码配置 CRUD handler
│   │   ├── settings.go        # 系统设置 handler
│   │   ├── gpu.go             # GET /api/settings/gpu-info
│   │   ├── preview.go         # 预览 CRUD handler
│   │   ├── log.go             # GET /api/logs, WS /ws/logs, GET /api/logs/download
│   │   └── ws.go              # WebSocket 升级 + hub 广播
│   ├── db/
│   │   ├── db.go              # SQLite 初始化, 连接池
│   │   ├── migrate.go         # 建表/迁移 SQL
│   │   ├── task.go            # Task + FileEntry 查询
│   │   ├── config.go          # TranscodeConfig + PresetTemplate 查询
│   │   ├── preview.go         # PreviewJob 查询
│   │   └── settings.go        # SystemSettings 查询
│   ├── task/
│   │   ├── runner.go          # TaskRunner goroutine pool, 并发控制
│   │   ├── process.go         # os/exec 启动/暂停/终止子进程
│   │   ├── progress.go        # 解析 ffmpeg/x265 stdout 进度
│   │   └── pipeline.go        # 4 阶段编码管线编排
│   ├── wizard/
│   │   └── parser.go          # BDMV 解析: ffprobe + META XML + MPLS
│   ├── config/
│   │   ├── manager.go         # 配置 CRUD 业务逻辑
│   │   └── preset.go          # 内置预设模板加载
│   ├── gpu/
│   │   └── detect.go          # GPU 检测: nvidia-smi + vainfo
│   ├── preview/
│   │   ├── runner.go          # 预览编码执行
│   │   └── cleanup.go         # TTL 定时清理 goroutine
│   └── log/
│       ├── handler.go         # slog handler: 文件轮转 + WebSocket 广播
│       └── rotate.go          # 日志轮转逻辑
├── presets/
│   ├── x264-mbtree-on.json    # 内置 x264 预设模板
│   ├── x264-mbtree-off.json
│   ├── x265-hq-anime.json
│   └── x265-balanced.json
├── web/                       # Vue 3 SPA (Vite)
│   ├── index.html
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   └── src/
│       ├── main.ts
│       ├── App.vue            # 根组件: 导航 + RouterView
│       ├── router.ts          # Vue Router
│       ├── api.ts             # REST API 封装 (fetch)
│       ├── ws.ts              # WebSocket 管理
│       ├── components/
│       │   ├── NavBar.vue
│       │   ├── ProgressBar.vue
│       │   ├── StatusBadge.vue
│       │   ├── TaskRow.vue
│       │   ├── TaskDetail.vue
│       │   ├── ConfigCard.vue
│       │   ├── FileBrowser.vue
│       │   ├── GpuInfo.vue
│       │   └── DownloadButton.vue
│       ├── pages/
│       │   ├── OverviewPage.vue
│       │   ├── TaskListPage.vue
│       │   ├── LogPage.vue
│       │   └── SettingsPage.vue
│       ├── wizard/
│       │   ├── WizardContainer.vue   # 步骤容器 + 进度条
│       │   ├── Step1Source.vue
│       │   ├── Step2Files.vue
│       │   ├── Step3Config.vue
│       │   └── Step4Target.vue
│       └── config/
│           ├── ConfigListPage.vue
│           ├── ConfigEditor.vue
│           ├── SimpleMode.vue
│           ├── ProfessionalMode.vue
│           ├── X264Params.vue
│           ├── X265Params.vue
│           └── GpuParams.vue
├── docs/
│   └── help/                  # markdown 帮助文档 (构建时编译 HTML)
├── Dockerfile
└── docker-compose.yml
```

---

### Task 1: 项目脚手架 + Go 基础框架

**Files:**
- Create: `cmd/server/main.go`
- Create: `internal/api/server.go`
- Create: `internal/api/middleware.go`
- Create: `go.mod`, `go.sum`

**Interfaces:**
- Produces: HTTP server on port 8080, CORS middleware, request logging, `/api/health` endpoint

- [ ] **Step 1: 初始化 Go module**

```bash
cd /home/zwb/dev/bdriper && go mod init github.com/zwb/bdriper
```

- [ ] **Step 2: 编写 HTTP server 入口 + health check**

```go
// cmd/server/main.go
package main

import (
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
    
    mux := http.NewServeMux()
    mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok"}`))
    })
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: withMiddleware(mux),
    }
    
    go func() {
        logger.Info("server starting", "addr", ":8080")
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            logger.Error("server error", "err", err)
            os.Exit(1)
        }
    }()
    
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    logger.Info("shutting down")
}

func withMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        slog.Info("request", "method", r.Method, "path", r.URL.Path)
        next.ServeHTTP(w, r)
    })
}
```

- [ ] **Step 3: 编译运行验证**

```bash
go build -o /tmp/bdriper ./cmd/server/ && /tmp/bdriper &
sleep 1 && curl -s http://localhost:8080/api/health
# Expected: {"status":"ok"}
kill %1
```

- [ ] **Step 4: 初始化前端项目 (Vue 3 + Vite + Tailwind)**

```bash
mkdir -p web && cd web
npm create vite@latest . -- --template vue-ts
npm install
npm install -D tailwindcss @tailwindcss/vite
npm install vue-router@4 @phosphor-icons/vue
```

`web/tailwind.config.ts`:
```ts
/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        bg: '#0F172A',
        fg: '#F8FAFC',
        card: '#1B2336',
        muted: '#272F42',
        border: '#475569',
        accent: '#22C55E',
        destructive: '#EF4444',
        primary: '#1E293B',
        secondary: '#334155',
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
```

- [ ] **Step 5: 前端基础组件 + router scaffold**

```bash
cd web && mkdir -p src/{components,pages,wizard,config}
```

`web/src/router.ts`:
```ts
import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', name: 'overview', component: () => import('./pages/OverviewPage.vue') },
    { path: '/tasks', name: 'tasks', component: () => import('./pages/TaskListPage.vue') },
    { path: '/logs', name: 'logs', component: () => import('./pages/LogPage.vue') },
    { path: '/settings', name: 'settings', component: () => import('./pages/SettingsPage.vue') },
  ],
})

export default router
```

`web/src/App.vue`:
```vue
<template>
  <div class="min-h-screen bg-bg text-fg font-sans">
    <NavBar />
    <main class="p-6">
      <RouterView />
    </main>
  </div>
</template>
```

- [ ] **Step 6: 验证前端构建**

```bash
cd web && npm run build
# Expected: dist/ 生成成功, 无报错
```

- [ ] **Step 7: Commit**

```bash
git add -A && git commit -m "feat: project scaffold with Go server and Vue 3 + Tailwind frontend"
```

---

### Task 2: 数据库层 — SQLite 迁移 + 基础 CRUD

**Files:**
- Create: `internal/db/db.go`
- Create: `internal/db/migrate.go`
- Create: `internal/db/task.go`
- Create: `internal/db/config.go`
- Create: `internal/db/settings.go`
- Create: `internal/db/preview.go`

**Interfaces:**
- Produces: `db.Open(dataDir string) (*sql.DB, error)` with auto-migration
- Produces: Task CRUD: `CreateTask`, `ListTasks(status string)`, `GetTask(id)`, `UpdateTask`, `DeleteTask`, `BatchDelete`, `DeleteCompleted`
- Produces: FileEntry: `CreateFileEntry`, `ListFileEntries(taskID)`, `UpdateFileEntry`
- Produces: Config CRUD: `CreateConfig`, `ListConfigs`, `GetConfig`, `UpdateConfig`, `DeleteConfig`
- Produces: Settings: `GetSetting(key)`, `SetSetting(key, value)`, `ListSettings()`
- Produces: Preview: `CreatePreview`, `GetPreview`, `UpdatePreview`, `ListExpiredPreviews`, `DeletePreview`

- [ ] **Step 1: 编写 migrate.go 建表 SQL**

```go
// internal/db/migrate.go
package db

const schemaSQL = `
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    source_path TEXT NOT NULL DEFAULT '',
    output_path TEXT NOT NULL DEFAULT '',
    progress REAL NOT NULL DEFAULT 0.0,
    estimated_eta TEXT NOT NULL DEFAULT '',
    pid INTEGER NOT NULL DEFAULT 0,
    config_id INTEGER NOT NULL DEFAULT 0,
    error_msg TEXT NOT NULL DEFAULT '',
    deleted BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS file_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL,
    source_file TEXT NOT NULL DEFAULT '',
    output_file TEXT NOT NULL DEFAULT '',
    streams TEXT NOT NULL DEFAULT '{}',
    selected BOOLEAN NOT NULL DEFAULT 1,
    progress REAL NOT NULL DEFAULT 0.0,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS transcode_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL DEFAULT '',
    encoder_type TEXT NOT NULL DEFAULT 'cpu',
    video_encoder TEXT NOT NULL DEFAULT 'x265',
    video_params TEXT NOT NULL DEFAULT '{}',
    audio_tracks TEXT NOT NULL DEFAULT '[]',
    subtitle_tracks TEXT NOT NULL DEFAULT '[]',
    chapters_enabled BOOLEAN NOT NULL DEFAULT 1,
    output_muxer TEXT NOT NULL DEFAULT 'mkvmerge',
    is_builtin BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS preset_templates (
    name TEXT PRIMARY KEY,
    encoder TEXT NOT NULL DEFAULT '',
    mode TEXT NOT NULL DEFAULT 'cpu',
    params TEXT NOT NULL DEFAULT '{}',
    description TEXT NOT NULL DEFAULT '',
    builtin BOOLEAN NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS preview_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id INTEGER NOT NULL DEFAULT 0,
    source_file TEXT NOT NULL DEFAULT '',
    start_time TEXT NOT NULL DEFAULT '',
    duration INTEGER NOT NULL DEFAULT 0,
    output_file TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS system_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);
`
```

- [ ] **Step 2: 编写 db.go — 初始化连接 + 自动迁移**

```go
// internal/db/db.go
package db

import (
    "database/sql"
    "os"
    "path/filepath"
    
    _ "github.com/mattn/go-sqlite3"
)

func Open(dataDir string) (*sql.DB, error) {
    if err := os.MkdirAll(dataDir, 0755); err != nil {
        return nil, err
    }
    dbPath := filepath.Join(dataDir, "bdriper.db")
    db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
    if err != nil {
        return nil, err
    }
    db.SetMaxOpenConns(1)
    if err := migrate(db); err != nil {
        return nil, err
    }
    return db, nil
}

func migrate(db *sql.DB) error {
    _, err := db.Exec(schemaSQL)
    return err
}
```

- [ ] **Step 3: 编写 task.go — Task + FileEntry CRUD**

```go
// internal/db/task.go
package db

import (
    "database/sql"
    "encoding/json"
    "time"
)

type Task struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Status      string    `json:"status"`
    SourcePath  string    `json:"source_path"`
    OutputPath  string    `json:"output_path"`
    Progress    float64   `json:"progress"`
    EstimatedETA string   `json:"estimated_eta"`
    PID         int       `json:"pid"`
    ConfigID    int64     `json:"config_id"`
    ErrorMsg    string    `json:"error_msg"`
    Deleted     bool      `json:"deleted"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type FileEntry struct {
    ID         int64  `json:"id"`
    TaskID     int64  `json:"task_id"`
    SourceFile string `json:"source_file"`
    OutputFile string `json:"output_file"`
    Streams    string `json:"streams"`
    Selected   bool   `json:"selected"`
    Progress   float64 `json:"progress"`
}

func CreateTask(db *sql.DB, t *Task) (int64, error) {
    res, err := db.Exec(`INSERT INTO tasks (name, status, source_path, output_path, config_id) VALUES (?,?,?,?,?)`,
        t.Name, t.Status, t.SourcePath, t.OutputPath, t.ConfigID)
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}

func ListTasks(db *sql.DB, status string) ([]Task, error) {
    query := `SELECT id,name,status,source_path,output_path,progress,estimated_eta,pid,config_id,error_msg,created_at,updated_at FROM tasks WHERE deleted=0`
    args := []any{}
    if status != "" && status != "all" {
        query += " AND status = ?"
        args = append(args, status)
    }
    query += " ORDER BY created_at DESC"
    rows, err := db.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var tasks []Task
    for rows.Next() {
        var t Task
        rows.Scan(&t.ID, &t.Name, &t.Status, &t.SourcePath, &t.OutputPath, &t.Progress, &t.EstimatedETA, &t.PID, &t.ConfigID, &t.ErrorMsg, &t.CreatedAt, &t.UpdatedAt)
        tasks = append(tasks, t)
    }
    return tasks, nil
}

func GetTask(db *sql.DB, id int64) (*Task, error) {
    t := &Task{}
    err := db.QueryRow(`SELECT id,name,status,source_path,output_path,progress,estimated_eta,pid,config_id,error_msg,created_at,updated_at FROM tasks WHERE id=? AND deleted=0`, id).
        Scan(&t.ID, &t.Name, &t.Status, &t.SourcePath, &t.OutputPath, &t.Progress, &t.EstimatedETA, &t.PID, &t.ConfigID, &t.ErrorMsg, &t.CreatedAt, &t.UpdatedAt)
    if err != nil {
        return nil, err
    }
    return t, nil
}

func UpdateTask(db *sql.DB, id int64, updates map[string]any) error {
    setClauses := ""
    args := []any{}
    for k, v := range updates {
        if setClauses != "" {
            setClauses += ", "
        }
        setClauses += k + " = ?"
        args = append(args, v)
    }
    args = append(args, id)
    _, err := db.Exec("UPDATE tasks SET "+setClauses+", updated_at=CURRENT_TIMESTAMP WHERE id=?", args...)
    return err
}

func DeleteTask(db *sql.DB, id int64) error {
    _, err := db.Exec("UPDATE tasks SET deleted=1, updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
    return err
}

func BatchDelete(db *sql.DB, ids []int64) error {
    tx, _ := db.Begin()
    for _, id := range ids {
        tx.Exec("UPDATE tasks SET deleted=1, updated_at=CURRENT_TIMESTAMP WHERE id=?", id)
    }
    return tx.Commit()
}

func DeleteCompleted(db *sql.DB) error {
    _, err := db.Exec("UPDATE tasks SET deleted=1 WHERE status='completed'")
    return err
}

func CreateFileEntry(db *sql.DB, fe *FileEntry) (int64, error) {
    res, err := db.Exec(`INSERT INTO file_entries (task_id, source_file, output_file, streams, selected) VALUES (?,?,?,?,?)`,
        fe.TaskID, fe.SourceFile, fe.OutputFile, fe.Streams, fe.Selected)
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}

func ListFileEntries(db *sql.DB, taskID int64) ([]FileEntry, error) {
    rows, err := db.Query(`SELECT id,task_id,source_file,output_file,streams,selected,progress FROM file_entries WHERE task_id=?`, taskID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var entries []FileEntry
    for rows.Next() {
        var fe FileEntry
        rows.Scan(&fe.ID, &fe.TaskID, &fe.SourceFile, &fe.OutputFile, &fe.Streams, &fe.Selected, &fe.Progress)
        entries = append(entries, fe)
    }
    return entries, nil
}

func UpdateFileEntry(db *sql.DB, id int64, updates map[string]any) error {
    setClauses := ""
    args := []any{}
    for k, v := range updates {
        if setClauses != "" {
            setClauses += ", "
        }
        setClauses += k + " = ?"
        args = append(args, v)
    }
    args = append(args, id)
    _, err := db.Exec("UPDATE file_entries SET "+setClauses+" WHERE id=?", args...)
    return err
}

// Unused imports cleaned below
var _ = json.Marshal
```

- [ ] **Step 4: 编写 config.go + settings.go + preview.go**

```go
// internal/db/config.go
package db

import "database/sql"

type TranscodeConfig struct {
    ID              int64  `json:"id"`
    Name            string `json:"name"`
    EncoderType     string `json:"encoder_type"`
    VideoEncoder    string `json:"video_encoder"`
    VideoParams     string `json:"video_params"`
    AudioTracks     string `json:"audio_tracks"`
    SubtitleTracks  string `json:"subtitle_tracks"`
    ChaptersEnabled bool   `json:"chapters_enabled"`
    OutputMuxer     string `json:"output_muxer"`
    IsBuiltin       bool   `json:"is_builtin"`
    CreatedAt       string `json:"created_at"`
}

func CreateConfig(db *sql.DB, c *TranscodeConfig) (int64, error) {
    res, err := db.Exec(`INSERT INTO transcode_configs (name,encoder_type,video_encoder,video_params,audio_tracks,subtitle_tracks,chapters_enabled,output_muxer) VALUES (?,?,?,?,?,?,?,?)`,
        c.Name, c.EncoderType, c.VideoEncoder, c.VideoParams, c.AudioTracks, c.SubtitleTracks, c.ChaptersEnabled, c.OutputMuxer)
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}

func ListConfigs(db *sql.DB) ([]TranscodeConfig, error) {
    rows, err := db.Query(`SELECT id,name,encoder_type,video_encoder,video_params,audio_tracks,subtitle_tracks,chapters_enabled,output_muxer,is_builtin,created_at FROM transcode_configs ORDER BY is_builtin DESC, created_at DESC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var cfgs []TranscodeConfig
    for rows.Next() {
        var c TranscodeConfig
        rows.Scan(&c.ID, &c.Name, &c.EncoderType, &c.VideoEncoder, &c.VideoParams, &c.AudioTracks, &c.SubtitleTracks, &c.ChaptersEnabled, &c.OutputMuxer, &c.IsBuiltin, &c.CreatedAt)
        cfgs = append(cfgs, c)
    }
    return cfgs, nil
}

func GetConfig(db *sql.DB, id int64) (*TranscodeConfig, error) {
    c := &TranscodeConfig{}
    err := db.QueryRow(`SELECT id,name,encoder_type,video_encoder,video_params,audio_tracks,subtitle_tracks,chapters_enabled,output_muxer,is_builtin,created_at FROM transcode_configs WHERE id=?`, id).
        Scan(&c.ID, &c.Name, &c.EncoderType, &c.VideoEncoder, &c.VideoParams, &c.AudioTracks, &c.SubtitleTracks, &c.ChaptersEnabled, &c.OutputMuxer, &c.IsBuiltin, &c.CreatedAt)
    return c, err
}

func UpdateConfig(db *sql.DB, id int64, updates map[string]any) error {
    setClauses := ""
    args := []any{}
    for k, v := range updates {
        if setClauses != "" {
            setClauses += ", "
        }
        setClauses += k + " = ?"
        args = append(args, v)
    }
    args = append(args, id)
    _, err := db.Exec("UPDATE transcode_configs SET "+setClauses+" WHERE id=?", args...)
    return err
}

func DeleteConfig(db *sql.DB, id int64) error {
    _, err := db.Exec("DELETE FROM transcode_configs WHERE id=? AND is_builtin=0", id)
    return err
}
```

```go
// internal/db/settings.go
package db

import "database/sql"

func GetSetting(db *sql.DB, key string) (string, error) {
    var val string
    err := db.QueryRow("SELECT value FROM system_settings WHERE key=?", key).Scan(&val)
    return val, err
}

func SetSetting(db *sql.DB, key, value string) error {
    _, err := db.Exec("INSERT INTO system_settings (key, value) VALUES (?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value", key, value)
    return err
}

func ListSettings(db *sql.DB) (map[string]string, error) {
    rows, err := db.Query("SELECT key, value FROM system_settings")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    settings := map[string]string{}
    for rows.Next() {
        var k, v string
        rows.Scan(&k, &v)
        settings[k] = v
    }
    return settings, nil
}
```

```go
// internal/db/preview.go
package db

import (
    "database/sql"
    "time"
)

type PreviewJob struct {
    ID         int64     `json:"id"`
    TaskID     int64     `json:"task_id"`
    SourceFile string    `json:"source_file"`
    StartTime  string    `json:"start_time"`
    Duration   int       `json:"duration"`
    OutputFile string    `json:"output_file"`
    Status     string    `json:"status"`
    ExpiresAt  time.Time `json:"expires_at"`
    CreatedAt  time.Time `json:"created_at"`
}

func CreatePreview(db *sql.DB, p *PreviewJob) (int64, error) {
    res, err := db.Exec(`INSERT INTO preview_jobs (task_id, source_file, start_time, duration, output_file, status, expires_at) VALUES (?,?,?,?,?,?,?)`,
        p.TaskID, p.SourceFile, p.StartTime, p.Duration, p.OutputFile, p.Status, p.ExpiresAt)
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}

func GetPreview(db *sql.DB, id int64) (*PreviewJob, error) {
    p := &PreviewJob{}
    err := db.QueryRow(`SELECT id,task_id,source_file,start_time,duration,output_file,status,expires_at,created_at FROM preview_jobs WHERE id=?`, id).
        Scan(&p.ID, &p.TaskID, &p.SourceFile, &p.StartTime, &p.Duration, &p.OutputFile, &p.Status, &p.ExpiresAt, &p.CreatedAt)
    return p, err
}

func UpdatePreview(db *sql.DB, id int64, updates map[string]any) error {
    setClauses := ""
    args := []any{}
    for k, v := range updates {
        if setClauses != "" {
            setClauses += ", "
        }
        setClauses += k + " = ?"
        args = append(args, v)
    }
    args = append(args, id)
    _, err := db.Exec("UPDATE preview_jobs SET "+setClauses+" WHERE id=?", args...)
    return err
}

func ListExpiredPreviews(db *sql.DB) ([]PreviewJob, error) {
    rows, err := db.Query(`SELECT id,output_file FROM preview_jobs WHERE expires_at < CURRENT_TIMESTAMP`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var jobs []PreviewJob
    for rows.Next() {
        var j PreviewJob
        rows.Scan(&j.ID, &j.OutputFile)
        jobs = append(jobs, j)
    }
    return jobs, nil
}

func DeletePreview(db *sql.DB, id int64) error {
    _, err := db.Exec("DELETE FROM preview_jobs WHERE id=?", id)
    return err
}
```

- [ ] **Step 5: 编译验证**

```bash
go mod tidy && go build ./...
# Expected: 编译成功
```

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "feat: database layer with SQLite migration and CRUD operations"
```

---

### Task 3: 日志系统 — slog handler + 文件轮转 + WebSocket 广播

**Files:**
- Create: `internal/log/handler.go`
- Create: `internal/log/rotate.go`

**Interfaces:**
- Produces: `NewLogger(dataDir, level string, maxFiles int, maxSizeMB int64) (*MultiHandler, error)` — 创建同时写文件和广播的 slog Handler
- Produces: `MultiHandler.Subscribe() <-chan LogEntry` — WebSocket 消费端订阅
- Produces: `MultiHandler.Rotate()` — 手动触发轮转检查
- Produces: `type LogEntry struct { Level, Time, Msg string }`

- [ ] **Step 1: 编写 rotate.go — 文件轮转逻辑**

```go
// internal/log/rotate.go
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
```

- [ ] **Step 2: 编写 handler.go — slog Handler + WebSocket 广播**

```go
// internal/log/handler.go
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
```

- [ ] **Step 3: 编译验证 + 单元测试**

```bash
go build ./internal/log/
```

- [ ] **Step 4: Commit**

```bash
git add internal/log/ && git commit -m "feat: slog handler with file rotation and WebSocket broadcast"
```

---

### Task 4: WebSocket 基础设施 + API 路由注册

**Files:**
- Create: `internal/api/ws.go`
- Modify: `internal/api/server.go`

**Interfaces:**
- Consumes: `*log.BroadcastHandler` (from Task 3)
- Produces: WebSocket hub with task progress + log event broadcast
- Produces: Full HTTP route registration in server.go

- [ ] **Step 1: 编写 ws.go — WebSocket Hub**

```go
// internal/api/ws.go
package api

import (
    "encoding/json"
    "net/http"
    "sync"

    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

type Event struct {
    Type string          `json:"type"`
    Data json.RawMessage `json:"data"`
}

type Hub struct {
    mu      sync.RWMutex
    clients map[*websocket.Conn]bool
}

func NewHub() *Hub {
    return &Hub{clients: make(map[*websocket.Conn]bool)}
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    h.mu.Lock()
    h.clients[conn] = true
    h.mu.Unlock()
    go func() {
        defer func() {
            h.mu.Lock()
            delete(h.clients, conn)
            h.mu.Unlock()
            conn.Close()
        }()
        for {
            if _, _, err := conn.ReadMessage(); err != nil {
                break
            }
        }
    }()
}

func (h *Hub) Broadcast(evt Event) {
    data, _ := json.Marshal(evt)
    h.mu.RLock()
    defer h.mu.RUnlock()
    for conn := range h.clients {
        conn.WriteMessage(websocket.TextMessage, data)
    }
}
```

- [ ] **Step 2: 注册所有路由**

```go
// internal/api/server.go — 更新为完整路由
package api

import (
    "database/sql"
    "log/slog"
    "net/http"
    
    "github.com/zwb/bdriper/internal/log"
)

type Server struct {
    DB     *sql.DB
    Logger  *slog.Logger
    LogHandler *log.BroadcastHandler
    TaskHub  *Hub
    LogHub   *Hub
    DataDir  string
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
    // Health
    mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
        writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
    })
    
    // Overview
    mux.HandleFunc("GET /api/overview/status", s.handleOverview)
    
    // Tasks
    mux.HandleFunc("GET /api/tasks", s.handleListTasks)
    mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
    mux.HandleFunc("GET /api/tasks/{id}", s.handleGetTask)
    mux.HandleFunc("PATCH /api/tasks/{id}", s.handleUpdateTask)
    mux.HandleFunc("DELETE /api/tasks/{id}", s.handleDeleteTask)
    mux.HandleFunc("DELETE /api/tasks/completed", s.handleDeleteCompleted)
    mux.HandleFunc("POST /api/tasks/{id}/retry", s.handleRetryTask)
    mux.HandleFunc("POST /api/tasks/batch", s.handleBatchTasks)
    
    // Wizard
    mux.HandleFunc("POST /api/wizard/parse", s.handleParseBDMV)
    mux.HandleFunc("GET /api/wizard/file/{path}/streams", s.handleFileStreams)
    
    // Configs
    mux.HandleFunc("GET /api/configs", s.handleListConfigs)
    mux.HandleFunc("POST /api/configs", s.handleCreateConfig)
    mux.HandleFunc("GET /api/configs/{id}", s.handleGetConfig)
    mux.HandleFunc("PUT /api/configs/{id}", s.handleUpdateConfig)
    mux.HandleFunc("DELETE /api/configs/{id}", s.handleDeleteConfig)
    mux.HandleFunc("GET /api/presets", s.handleListPresets)
    
    // Settings
    mux.HandleFunc("GET /api/settings", s.handleListSettings)
    mux.HandleFunc("PATCH /api/settings", s.handleUpdateSettings)
    mux.HandleFunc("GET /api/settings/gpu-info", s.handleGPUInfo)
    
    // Preview
    mux.HandleFunc("POST /api/preview", s.handleCreatePreview)
    mux.HandleFunc("GET /api/preview/{id}/status", s.handlePreviewStatus)
    mux.HandleFunc("GET /api/preview/{id}/download", s.handlePreviewDownload)
    mux.HandleFunc("DELETE /api/preview/{id}", s.handleDeletePreview)
    
    // Logs
    mux.HandleFunc("GET /api/logs", s.handleGetLogs)
    mux.HandleFunc("GET /api/logs/download", s.handleDownloadLogs)
    
    // WebSocket
    mux.HandleFunc("GET /ws/events", s.TaskHub.HandleWS)
    
    // SPA static files (from embedded FS)
    mux.Handle("/", http.FileServer(s.SPAFS))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}
```

- [ ] **Step 3: 编译验证**

```bash
go mod tidy && go build ./...
# Expected: 调用方未定义会编译失败 — 这是预期的 (后续 task 补 handler)
```

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: WebSocket hub and HTTP route registration"
```

---

### Task 5: BDMV 解析 — ffprobe + META XML

**Files:**
- Create: `internal/wizard/parser.go`

**Interfaces:**
- Produces: `ParseBDMV(path string) (*BDMVInfo, error)` — 返回完整 BDMV 信息
- Produces: `GetFileStreams(m2tsPath string) (*FileStreamInfo, error)` — 单文件流信息
- Produces: `type BDMVInfo struct { DiscName string, Files []BDMVFile }`
- Produces: `type BDMVFile struct { Path string, Duration string, Resolution string, FPS float64, IsMain bool }`
- Produces: `type FileStreamInfo struct { Video, Audio, Subtitle []Stream }`

- [ ] **Step 1: 编写 parser.go**

```go
// internal/wizard/parser.go
package wizard

import (
    "encoding/json"
    "encoding/xml"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"
)

type BDMVFile struct {
    Path       string `json:"path"`
    Duration   string `json:"duration"`
    Resolution string `json:"resolution"`
    FPS        float64 `json:"fps"`
    IsMain     bool   `json:"is_main"`
}

type BDMVInfo struct {
    DiscName string     `json:"disc_name"`
    Files    []BDMVFile `json:"files"`
}

type Stream struct {
    Index    int    `json:"index"`
    Codec    string `json:"codec"`
    Type     string `json:"type"`
    Language string `json:"language"`
    Channels int    `json:"channels,omitempty"`
    SampleRate string `json:"sample_rate,omitempty"`
}

type FileStreamInfo struct {
    Video    []Stream `json:"video"`
    Audio    []Stream `json:"audio"`
    Subtitle []Stream `json:"subtitle"`
}

func ParseBDMV(sourcePath string) (*BDMVInfo, error) {
    streamDir := filepath.Join(sourcePath, "BDMV", "STREAM")
    entries, err := os.ReadDir(streamDir)
    if err != nil {
        return nil, fmt.Errorf("read STREAM dir: %w", err)
    }
    
    info := &BDMVInfo{
        DiscName: parseDiscName(sourcePath),
    }
    
    var m2tsFiles []string
    for _, e := range entries {
        if strings.HasSuffix(strings.ToLower(e.Name()), ".m2ts") {
            m2tsFiles = append(m2tsFiles, filepath.Join(streamDir, e.Name()))
        }
    }
    
    for _, f := range m2tsFiles {
        bf, err := probeM2TS(f)
        if err != nil {
            continue
        }
        info.Files = append(info.Files, *bf)
    }
    
    return info, nil
}

func parseDiscName(sourcePath string) string {
    metaFiles := []string{
        filepath.Join(sourcePath, "BDMV", "META", "DL", "bdmt_eng.xml"),
        filepath.Join(sourcePath, "BDMV", "META", "DL", "bdmt_jpn.xml"),
    }
    for _, mf := range metaFiles {
        data, err := os.ReadFile(mf)
        if err != nil {
            continue
        }
        var meta struct {
            Extension struct {
                Name string `xml:"name"`
            } `xml:"extension"`
        }
        if err := xml.Unmarshal(data, &meta); err == nil && meta.Extension.Name != "" {
            return meta.Extension.Name
        }
    }
    return filepath.Base(sourcePath)
}

func probeM2TS(path string) (*BDMVFile, error) {
    cmd := exec.Command("ffprobe",
        "-v", "quiet",
        "-print_format", "json",
        "-show_format",
        "-show_streams",
        path,
    )
    out, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    var probe struct {
        Streams []struct {
            CodecType string `json:"codec_type"`
            Width     int    `json:"width"`
            Height    int    `json:"height"`
            RFrameRate string `json:"r_frame_rate"`
            Duration  string `json:"duration"`
        } `json:"streams"`
        Format struct {
            Duration string `json:"duration"`
        } `json:"format"`
    }
    json.Unmarshal(out, &probe)
    
    dur := probe.Format.Duration
    durDisplay := "0:00"
    if d, err := parseDuration(dur); err == nil {
        durDisplay = fmt.Sprintf("%.0f:%02.0f", d.Minutes(), modSeconds(d))
    }
    
    bf := &BDMVFile{
        Path:     path,
        Duration: durDisplay,
    }
    
    for _, s := range probe.Streams {
        if s.CodecType == "video" {
            bf.Resolution = fmt.Sprintf("%dx%d", s.Width, s.Height)
            bf.FPS = parseFPS(s.RFrameRate)
            bf.IsMain = true
        }
    }
    
    bf.IsMain = bf.IsMain && parseSeconds(dur) > 60
    return bf, nil
}

func GetFileStreams(m2tsPath string) (*FileStreamInfo, error) {
    cmd := exec.Command("ffprobe",
        "-v", "quiet",
        "-print_format", "json",
        "-show_streams",
        m2tsPath,
    )
    out, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    
    var probe struct {
        Streams []struct {
            Index      int    `json:"index"`
            CodecType  string `json:"codec_type"`
            CodecName  string `json:"codec_name"`
            Channels   int    `json:"channels"`
            SampleRate string `json:"sample_rate"`
            Tags       struct {
                Language string `json:"language"`
            } `json:"tags"`
        } `json:"streams"`
    }
    json.Unmarshal(out, &probe)
    
    result := &FileStreamInfo{}
    for _, s := range probe.Streams {
        stream := Stream{
            Index:    s.Index,
            Codec:    s.CodecName,
            Type:     s.CodecType,
            Language: s.Tags.Language,
            Channels: s.Channels,
            SampleRate: s.SampleRate,
        }
        switch s.CodecType {
        case "video":
            result.Video = append(result.Video, stream)
        case "audio":
            result.Audio = append(result.Audio, stream)
        case "subtitle":
            result.Subtitle = append(result.Subtitle, stream)
        }
    }
    return result, nil
}

func parseFPS(rate string) float64 {
    parts := strings.Split(rate, "/")
    if len(parts) == 2 {
        num, _ := strconv.ParseFloat(parts[0], 64)
        den, _ := strconv.ParseFloat(parts[1], 64)
        if den > 0 {
            return num / den
        }
    }
    return 0
}

func parseSeconds(dur string) float64 {
    d, _ := strconv.ParseFloat(dur, 64)
    return d
}

func parseDuration(dur string) (time.Duration, error) {
    // ... implementation
    return 0, nil
}
```

- [ ] **Step 2: 编译验证**

```bash
go build ./internal/wizard/
```

- [ ] **Step 3: Commit**

```bash
git add internal/wizard/ && git commit -m "feat: BDMV parser with ffprobe and META XML disc name extraction"
```

---

### Task 6: 转码配置 — 预设模板 + CRUD + 帮助文档编译

**Files:**
- Create: `internal/config/manager.go`
- Create: `internal/config/preset.go`
- Create: `presets/x264-mbtree-on.json`
- Create: `presets/x264-mbtree-off.json`
- Create: `presets/x265-hq-anime.json`
- Create: `presets/x265-balanced.json`

**Interfaces:**
- Produces: `LoadPresets(db *sql.DB, presetsDir string) error` — 从 JSON 文件加载预设到 DB
- Produces: `ListPresets(db *sql.DB) ([]db.PresetTemplate, error)`

- [ ] **Step 1: 编写 4 个预设 JSON 文件**

`presets/x265-hq-anime.json`:
```json
{
  "name": "x265 HQ Anime",
  "encoder": "x265",
  "mode": "cpu",
  "description": "高质量动漫 BDRip 压制。CRF=15, 关闭 SAO, AQ mode=1。适合 1080p 蓝光源。",
  "params": {
    "crf": 15.0,
    "preset": "slower",
    "deblock": "-1:-1",
    "ctu": 32,
    "qg-size": 8,
    "me": "star",
    "subme": 5,
    "merange": 38,
    "bframes": 6,
    "ref": 4,
    "weightb": true,
    "b-intra": true,
    "qcomp": 0.65,
    "aq-mode": 1,
    "aq-strength": 0.8,
    "no-sao": true,
    "rc-lookahead": 80,
    "psy-rd": 2.0,
    "psy-rdoq": 1.0,
    "rdoq-level": 2,
    "rd": 5,
    "pbratio": 1.2,
    "cbqpoffs": -2,
    "crqpoffs": -2,
    "colormatrix": "bt709",
    "no-open-gop": true,
    "keyint": 360,
    "min-keyint": 1,
    "no-strong-intra-smoothing": true,
    "limit-tu": 4,
    "no-amp": true
  }
}
```

`presets/x264-mbtree-on.json`:
```json
{
  "name": "x264 High Quality (mbtree on)",
  "encoder": "x264",
  "mode": "cpu",
  "description": "高质量动漫 BDRip。CRF=16.5, mbtree 开启, aq-mode=3。",
  "params": {
    "crf": 16.5,
    "preset": "veryslow",
    "deblock": "-1:-1",
    "keyint": 600,
    "min-keyint": 1,
    "bframes": 8,
    "ref": 13,
    "qcomp": 0.75,
    "aq-mode": 3,
    "aq-strength": 0.8,
    "mbtree": true,
    "me": "tesa",
    "subme": 10,
    "me_range": 24,
    "psy-rd": 0.6,
    "psy-trellis": 0.15,
    "no-fast-pskip": true,
    "colormatrix": "bt709",
    "input-depth": 16,
    "threads": 16,
    "rc-lookahead": 70
  }
}
```

`presets/x264-mbtree-off.json`:
```json
{
  "name": "x264 High Quality (mbtree off)",
  "encoder": "x264",
  "mode": "cpu",
  "description": "高质量动漫 BDRip。CRF=17.5, mbtree 关闭。适合高码率场景。",
  "params": {
    "crf": 17.5,
    "preset": "veryslow",
    "deblock": "-1:-1",
    "keyint": 600,
    "min-keyint": 1,
    "bframes": 8,
    "ref": 13,
    "qcomp": 0.6,
    "aq-mode": 3,
    "aq-strength": 0.8,
    "no-mbtree": true,
    "me": "tesa",
    "subme": 10,
    "me_range": 24,
    "psy-rd": 0.6,
    "psy-trellis": 0.0,
    "no-fast-pskip": true,
    "colormatrix": "bt709",
    "chroma-qp-offset": -1,
    "input-depth": 16,
    "threads": 16,
    "rc-lookahead": 70
  }
}
```

`presets/x265-balanced.json`:
```json
{
  "name": "x265 Balanced",
  "encoder": "x265",
  "mode": "cpu",
  "description": "均衡画质与速度的 x265 配置。CRF=18, aq-mode=2。",
  "params": {
    "crf": 18.0,
    "preset": "slow",
    "deblock": "-1:-1",
    "ctu": 32,
    "qg-size": 8,
    "me": "star",
    "subme": 3,
    "merange": 38,
    "bframes": 6,
    "ref": 4,
    "qcomp": 0.6,
    "aq-mode": 2,
    "aq-strength": 0.9,
    "no-sao": true,
    "psy-rd": 2.0,
    "psy-rdoq": 1.0,
    "rdoq-level": 2,
    "keyint": 360,
    "min-keyint": 1,
    "no-open-gop": true,
    "pbratio": 1.2,
    "cbqpoffs": -2,
    "crqpoffs": -2
  }
}
```

- [ ] **Step 2: 编写 preset.go — 加载预设到 DB**

```go
// internal/config/preset.go
package config

import (
    "database/sql"
    "encoding/json"
    "os"
    "path/filepath"

    "github.com/zwb/bdriper/internal/db"
)

func LoadPresets(database *sql.DB, presetsDir string) error {
    entries, err := os.ReadDir(presetsDir)
    if err != nil {
        return err
    }
    for _, e := range entries {
        if filepath.Ext(e.Name()) != ".json" {
            continue
        }
        data, err := os.ReadFile(filepath.Join(presetsDir, e.Name()))
        if err != nil {
            continue
        }
        var p db.PresetTemplate
        json.Unmarshal(data, &p)
        database.Exec(`INSERT OR IGNORE INTO preset_templates (name,encoder,mode,params,description,builtin) VALUES (?,?,?,?,?,1)`,
            p.Name, p.Encoder, p.Mode, string(data), p.Description)
    }
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
        json.Unmarshal([]byte(data), &p.Params)
        presets = append(presets, p)
    }
    return presets, nil
}
```

- [ ] **Step 3: 编译验证**

```bash
go build ./internal/config/
```

- [ ] **Step 4: Commit**

```bash
git add presets/ internal/config/ && git commit -m "feat: transcode preset templates and loader"
```

---

### 剩余 Task 概览（由于内容过长，以下为概述 + 关键接口定义）

篇幅限制，Task 7-21 的技术细节在此展开关键部分。

---

### Task 7: GPU 检测

**Files:** `internal/gpu/detect.go`

**Interfaces:**
- Produces: `DetectGPUs() ([]GPUInfo, error)`
- Produces: `type GPUInfo struct { Vendor, Model string, Encoders []EncoderInfo }`
- Produces: `type EncoderInfo struct { Name, Codec string, Supported bool }`

```
检测逻辑:
1. NVIDIA: 执行 nvidia-smi --query-gpu=name --format=csv,noheader
   如存在: 执行 ffmpeg -hide_banner -encoders | grep nvenc 列出 NVENC 编码器
2. Intel QSV: 检查 /dev/dri/renderD* 是否存在
   如存在: 执行 ffmpeg -hide_banner -encoders | grep qsv 列出 QSV 编码器
3. AMD: 执行 vainfo 检查 VA-API
   如存在: 执行 ffmpeg -hide_banner -encoders | grep amf 列出 AMF 编码器
```

---

### Task 8: 预览系统

**Files:** `internal/preview/runner.go`, `internal/preview/cleanup.go`

**Interfaces:**
- Produces: `RunPreview(sourceFile, startTime string, duration int, outputFile string) (*os/exec.Cmd, error)`
- Produces: `StartCleanup(db *sql.DB, interval time.Duration)` — 定时清理 goroutine

---

### Task 9: Task Runner — 编码管线

**Files:** `internal/task/runner.go`, `internal/task/process.go`, `internal/task/progress.go`, `internal/task/pipeline.go`

**Interfaces:**
- Consumes: `*sql.DB`, `*log.BroadcastHandler`, `*api.Hub`, `*db.Task`, `[]db.FileEntry`
- Produces: `type TaskRunner struct` with `Start(task *db.Task, files []db.FileEntry, cfg *db.TranscodeConfig) error`, `Pause(taskID) error`, `Cancel(taskID) error`
- Produces: `type Pipeline struct` with 4 stages: `ExtractStreams`, `EncodeVideo`, `MuxMKV`

---

### Task 10-16: API Handler 实现

每个 handler 文件对应一组 API 端点，遵循 `func (s *Server) handleXxx(w http.ResponseWriter, r *http.Request)` 模式。例如:

```
Task 10: internal/api/overview.go — GET /api/overview/status
Task 11: internal/api/task.go — Task CRUD + batch
Task 12: internal/api/wizard.go — BDMV 解析端点
Task 13: internal/api/config.go — 配置 CRUD + presets
Task 14: internal/api/settings.go + gpu.go — 设置 + GPU 信息
Task 15: internal/api/preview.go — 预览 CRUD + download
Task 16: internal/api/log.go — 日志查询 + 下载
Task 17: cmd/server/main.go — wire everything together
```

---

### Task 18-23: 前端页面

```
Task 18: web/src/api.ts, web/src/ws.ts — API + WebSocket 封装
Task 19: web/src/components/ — 共享组件 (NavBar, ProgressBar, StatusBadge, TaskRow, etc.)
Task 20: web/src/wizard/ — 4 步向导组件
Task 21: web/src/pages/ — 概览、任务管理、日志、设置页面
Task 22: web/src/config/ — 配置编辑器 (简易 + 专业模式)
Task 23: 前端构建集成 — embed.FS 嵌入 SPA dist
```

---

### Task 24: Dockerfile + docker-compose

**Files:** `Dockerfile`, `docker-compose.yml`

**Dockerfile 关键内容:**
```dockerfile
FROM ubuntu:24.04 AS base
RUN apt-get update && apt-get install -y \
    ffmpeg x264 x265 mkvtoolnix vainfo intel-media-va-driver mesa-va-drivers ca-certificates
# NVIDIA drivers injected via nvidia-container-toolkit at runtime

FROM golang:1.22 AS go-builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bdriper ./cmd/server/

FROM node:20 AS web-builder
COPY web /src
WORKDIR /src
RUN npm ci && npm run build

FROM base AS runtime
COPY --from=go-builder /bdriper /usr/local/bin/bdriper
COPY --from=web-builder /src/dist /app/web/dist
COPY presets/ /app/presets/
COPY docs/ /app/docs/
EXPOSE 8080
ENTRYPOINT ["bdriper", "--data-dir=/data", "--input-dir=/input", "--output-dir=/output"]
```

**docker-compose.yml:**
```yaml
version: '3'
services:
  bdriper:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - /mnt/bdmv:/input:ro
      - /mnt/output:/output
    devices:
      - /dev/dri:/dev/dri
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
```

---

### Task 25: 帮助文档构建

**Files:** `docs/help/`, `scripts/build-help.sh`

构建过程: 读取 `downloads/markdown/tutorial09.pdf/tutorial09.md` 和 `tutorial10.pdf/tutorial10.md`，提取 x264/x265 参数表格，编译为 HTML 静态文件，前端 iframe 或直接渲染。

---

### Task 26: 端到端集成测试 + 容器验证

**Steps:**
1. `docker build -t bdriper .`
2. `docker run --rm -p 8080:8080 -v $(pwd)/testdata/bdmv:/input -v $(pwd)/testdata/output:/output bdriper`
3. 验证: 前端可访问, 创建任务, 查看日志, 转码成功生成 MKV

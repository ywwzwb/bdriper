# BDRiper 设计文档

## 概述

BDRiper 是一个 Docker 容器化的 Web 应用，提供可视化界面用于将 BDMV（蓝光原盘）压制为 BDRip（MKV 格式）。面向动漫蓝光压制场景，参考 VCB-Studio 教程体系。

### 核心决策

| 决策点 | 选择 |
|--------|------|
| 交付方式 | 一次性全部交付 |
| 前端 | Vue 3 + TypeScript + Element Plus |
| 后端 | Go + chi 路由 + 标准库 |
| 数据库 | SQLite（文件持久化） |
| 通信 | REST + WebSocket |
| 基础镜像 | Ubuntu 24.04 |
| 编码管线 | ffmpeg 内置滤镜 |
| 任务调度 | 内置 Worker Pool |
| 目标场景 | 动漫蓝光 |

---

## 1. 整体架构

```
┌─────────────────────────────────────────────────┐
│                   Docker 容器                     │
│  ┌──────────────┐     ┌──────────────────────┐  │
│  │  Vue 3 前端   │────▶│   Go 后端 (chi)       │  │
│  │  (SPA, 静态)  │◀─WS─│   :8080              │  │
│  │  Element Plus │     │                      │  │
│  └──────────────┘     │  ┌────────────────┐  │  │
│                        │  │  Worker Pool    │  │  │
│  ┌──────────────┐     │  │  (可配并发数)    │──┼──┼──▶ ffmpeg/x265/...
│  │   SQLite      │◀───▶│  └────────────────┘  │  │
│  │   (文件挂载)   │     │                      │  │
│  └──────────────┘     └──────────────────────┘  │
│                                                   │
│  挂载卷: /data (DB+配置), /source (源), /output   │
└─────────────────────────────────────────────────┘
```

- 前端 Vue 3 SPA，生产模式由 Go embed 静态文件 serve，开发模式 Vite 代理
- 后端 Go 单体进程，chi 路由 + 内置 worker pool
- SQLite 单文件放在 `/data` 挂载卷，容器重启数据持久
- Worker pool 通过 `os/exec` 管理 ffmpeg/mkvmerge 子进程，解析 stderr 获取进度
- REST CRUD + 单一 WebSocket 连接推送进度和系统状态

---

## 2. 后端设计

### 2.1 目录结构

```
backend/
├── main.go                  # 入口
├── api/                     # HTTP handler
│   ├── router.go            # chi 路由 + 中间件
│   ├── task_handler.go      # 任务 CRUD + 批量操作
│   ├── source_handler.go    # BDMV 解析、文件浏览
│   ├── config_handler.go    # 转码配置 CRUD + 导入导出
│   ├── system_handler.go    # 系统设置、GPU 检测
│   └── ws_handler.go        # WebSocket
├── models/                  # 数据模型
│   ├── db.go                # SQLite 初始化
│   ├── task.go              # Task
│   ├── config.go            # TranscodeConfig
│   └── settings.go          # 系统设置
├── services/                # 核心业务
│   ├── parser.go            # BDMV 解析
│   ├── encoder.go           # ffmpeg 命令构建 + 执行
│   ├── worker.go            # Worker pool
│   └── monitor.go           # CPU/GPU 监控
├── middleware/
│   └── logger.go
└── config/
    └── config.go
```

### 2.2 数据库表

**tasks**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT (UUID) | 主键 |
| name | TEXT | 任务名称 |
| source_path | TEXT | 源文件路径 |
| output_path | TEXT | 目标路径 |
| status | TEXT | pending/running/paused/done/failed |
| progress | REAL | 0.0 ~ 100.0 |
| eta_seconds | INTEGER | 预估剩余秒数 |
| selected_tracks | TEXT (JSON) | 选中的文件/音轨/字幕 |
| config_id | TEXT | 关联转码配置 |
| pid | INTEGER | 子进程 PID |
| error_message | TEXT | 失败原因 |
| is_preview | INTEGER (bool) | 是否预览任务 |
| temp_dir | TEXT | 临时目录（预览） |
| created_at | TEXT | 创建时间 |
| updated_at | TEXT | 更新时间 |

**transcode_configs**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | TEXT (UUID) | 主键 |
| name | TEXT | 配置名称 |
| is_simple | INTEGER (bool) | 简易/专业模式 |
| encoder | TEXT | x264 / x265 / copy |
| params | TEXT (JSON) | 完整参数 |
| created_at | TEXT | 创建时间 |
| updated_at | TEXT | 更新时间 |

### 2.3 核心服务

**Parser** — 扫描 BDMV `PLAYLIST/*.mpls`，调用 ffprobe 解析 m2ts 流信息（视频/音频/字幕轨道、时长、编码格式）。

**Encoder** — 根据 Task 的轨道选择和 TranscodeConfig 构建 ffmpeg 命令，执行时通过 `-progress pipe:2` 解析进度。

**Worker** — 带缓冲 channel 的 worker pool，从队列取 pending task，创建 `context.WithCancel`，启动 ffmpeg。暂停=cancel ctx + kill 进程。通过 WebSocket hub 推送进度。

**Monitor** — 定时读取 `/proc/stat`（CPU）和 GPU 工具（nvidia-smi / rocm-smi / intel_gpu_top），通过 WebSocket 推送。GPU 类型自动检测。

---

## 3. 前端设计

### 3.1 页面结构

- **概览页**: CPU/GPU 占用实时显示，当前任务数，居中「新增任务」按钮
- **任务管理页**: 任务列表（筛选/批量操作/展开详情），一键清除已完成
- **设置页**: 系统设置（并发数/GPU 开关/GPU 测试）、转码配置管理（CRUD + 导入导出）

### 3.2 新增任务向导

用 Element Plus `el-steps`，7 个步骤：

1. **选择源文件** — 文件浏览器，支持 BDMV 目录和 ISO
2. **解析** — 自动解析，失败则报错退出
3. **选择文件与轨道** — 主文件列表（默认全选）+ 附属列表，每个文件可展开选择音轨/字幕/章节
4. **转码配置** — 已有配置下拉 + 快捷新建入口
5. **目标位置 + 预览** — 输出路径，预览按钮弹出面板（选文件→设时间戳→编码→下载→自动清理）
6. **确认** — 汇总展示，点击完成创建任务

### 3.3 技术选型

- UI 框架: Element Plus
- 状态管理: Pinia
- 构建: Vite
- HTTP: axios
- WebSocket: 原生 API 封装在 Pinia store

---

## 4. 转码配置模型

### 4.1 简易模式

| 参数 | 选项 | 说明 |
|------|------|------|
| 编码器 | x264 / x265 | |
| 位深 | 8bit / 10bit | |
| 质量 | 拖动条：低/中/高/极高 | CRF 映射 |
| 速度预设 | 快/均衡/慢/极慢 | preset 映射 |
| 画面类型 | 动漫/电影/噪点 | tune 映射 |
| 音频编码 | 无损复制/FLAC/AAC 192K/AAC 256K | |
| 字幕处理 | 内封 | 仅内封，不烧录 |

### 4.2 专业模式

三大类参数：

**编码规范**: depth, profile, level, colormatrix, opengop, keyint

**效率取舍**: preset, ref, bframes, me, subme, merange, rect, amp, ctu (x265), qg-size (x265), b-intra, weightb, rc-lookahead, rd

**码率控制**: rate mode (CRF/ABR/2pass), crf, aq-mode, aq-strength, psy-rd, psy-rdoq (x265), rdoq-level (x265), qcomp, pbratio, deblock, sao (x265), cbqpoffs/crqpoffs (x265)

**滤镜链**: 反交错(yadif/bwdif)、去噪(hqdn3d)、缩放、裁剪、色调

**封装**: 容器固定 MKV，字幕仅内封，章节可提取

### 4.3 存储格式

```json
{
  "mode": "simple",
  "encoder": "x265",
  "video": {
    "depth": 10,
    "crf": 16.0,
    "preset": "veryslow",
    "tune": "animation",
    "advanced": {}
  },
  "audio": {"codec": "copy"},
  "subtitle": {"mode": "mux"},
  "chapter": true,
  "filter_chain": []
}
```

专业模式下 `video.advanced` 填入完整参数，`filter_chain` 为用户自定义滤镜列表。

### 4.4 导入导出

JSON 格式完整序列化。导出时从 DB 读取，导入时校验字段合法性后写入。

---

## 5. 错误处理与边界

### 5.1 任务生命周期

```
pending → running ⇄ paused
running → done
running → failed
```

- paused 恢复：重新用同样参数启动 ffmpeg
- 失败重试：复制新 pending 任务，保留原记录
- 删除：done 只删 DB 记录，其余删 DB + 部分输出文件

### 5.2 预览任务

- `is_preview=true`，定时 goroutine 清理超过 4 小时的临时文件
- 容器重启时清空所有残留临时文件

### 5.3 异常处理

| 场景 | 处理 |
|------|------|
| 源路径不存在 | 解析阶段报错 |
| BDMV 结构不完整 | 解析失败 |
| 磁盘空间不足 | 检测 ffmpeg stderr "No space left"，标记 failed |
| Worker pool 满 | 任务进入 pending 队列 |
| 容器重启 | running/paused → failed（子进程丢失），pending 恢复 |
| ffmpeg 僵死 | timeout context + 心跳检测，超时 kill + failed |

### 5.4 WebSocket 重连

- 指数退避重连，上限 30 秒
- 重连后请求任务列表全量快照
- 消息类型: `task_progress`, `task_status`, `system_stats`, `gpu_test_result`

---

## 6. Docker 与部署

### 6.1 多阶段构建

```
Stage 1: go-builder   → Go 二进制
Stage 2: node-builder → Vue SPA
Stage 3: runtime      → Ubuntu + 工具链 + 产物
```

### 6.2 运行方式

```bash
docker run -d \
  --name bdriper \
  --gpus all \
  -p 8080:8080 \
  -v /mnt/data:/data \
  -v /mnt/blurays:/source:ro \
  -v /mnt/output:/output \
  bdriper:latest
```

### 6.3 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `BD_LISTEN` | `:8080` | 监听地址 |
| `BD_DB_PATH` | `/data/bdriper.db` | SQLite 路径 |
| `BD_WEB_DIR` | `/opt/bdriper/web` | 前端静态文件 |
| `BD_TEMP_DIR` | `/tmp/bdriper` | 预览临时文件 |
| `BD_MAX_CONCURRENT` | `2` | 最大并发 |
| `BD_SOURCE_ROOT` | `/source` | 文件浏览器根 |

### 6.4 GPU 支持

- NVIDIA: NVENC/NVDec，nvidia-smi 监控
- AMD: VAAPI/AMF，rocm-smi 监控
- Intel: QSV/VAAPI，intel_gpu_top 监控
- GPU 类型自动检测，编码器列表根据实际硬件动态过滤

---

## 7. API 概览

### REST

```
GET    /api/health
GET    /api/tasks
POST   /api/tasks
GET    /api/tasks/:id
PATCH  /api/tasks/:id
DELETE /api/tasks/:id
POST   /api/tasks/:id/retry
POST   /api/tasks/batch
DELETE /api/tasks/completed
GET    /api/source/browse
POST   /api/source/parse
GET    /api/source/streams/:file
POST   /api/preview
GET    /api/preview/:id/download
DELETE /api/preview/:id
GET    /api/configs
POST   /api/configs
GET    /api/configs/:id
PUT    /api/configs/:id
DELETE /api/configs/:id
POST   /api/configs/export
POST   /api/configs/import
GET    /api/system/settings
PUT    /api/system/settings
POST   /api/system/gpu-test
GET    /api/system/gpu-test/:id
WS     /api/ws
```

### WebSocket 消息

服务端→客户端: `task_progress`, `task_status`, `system_stats`, `gpu_test`

客户端→服务端: `subscribe`

### 认证

第一版不做认证，后续可按需加 HTTP Basic Auth。

---

## 8. 测试策略

| 层级 | 工具 | 覆盖 |
|------|------|------|
| Go 单元 | testing + testify | 模型 CRUD、配置校验、ffmpeg 命令构建、worker pool |
| Go 集成 | testing + SQLite | API 端到端（httptest）、BDMV 解析、短片段编码 |
| 前端单元 | Vitest + Vue Test Utils | 组件渲染、Pinia store、向导状态管理 |
| 前端 E2E | Playwright | 创建任务向导、暂停/删除、配置导入导出 |

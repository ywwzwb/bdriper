# BDRiper 实现计划

> 基于 [2026-07-19-bdriper-design.md](./2026-07-19-bdriper-design.md)

## 阶段总览

| 阶段 | 内容 | 预估工作量 | 依赖 |
|------|------|-----------|------|
| 1 | 项目脚手架 | Go 模块 + Vue 项目 + Dockerfile | 无 |
| 2 | 数据层 | SQLite 初始化、模型、迁移 | 阶段 1 |
| 3 | 后端核心服务 | Parser、Encoder、Worker、Monitor | 阶段 2 |
| 4 | 后端 API | 全部 REST 端点 + WebSocket | 阶段 3 |
| 5 | 前端基础框架 | 路由、布局、Pinia stores | 阶段 1 |
| 6 | 前端页面 | 概览、任务管理、设置 | 阶段 4 + 5 |
| 7 | 前端向导 | 新增任务多步骤向导 | 阶段 6 |
| 8 | Docker 化 | 多阶段构建、docker-compose | 阶段 1-7 |
| 9 | 测试与收尾 | 单元测试、集成测试、E2E | 全部 |

---

## 阶段 1：项目脚手架

### 1.1 Go 项目初始化
- 初始化 `go.mod`（module: `github.com/user/bdriper`）
- 安装依赖：chi、gorilla/websocket、mattn/go-sqlite3、google/uuid
- 创建目录结构：`backend/{api,models,services,middleware,config}`
- `backend/main.go`：启动 HTTP server、初始化 DB、启动 worker pool
- `backend/config/config.go`：环境变量解析

### 1.2 Vue 项目初始化
- `npm create vue@latest frontend`（TypeScript + Router + Pinia）
- 安装 Element Plus、axios
- 配置 Vite 代理（`/api` → `localhost:8080`）
- `frontend/` 目录结构：`components/`、`views/`、`stores/`、`router/`

### 1.3 Dockerfile 骨架
- 三阶段构建：go-builder、node-builder、runtime
- runtime 阶段安装 ffmpeg、mkvtoolnix、GPU 工具
- 验证点：`docker build` 通过，`docker run` 启动后 `/api/health` 返回 200

---

## 阶段 2：数据层

### 2.1 SQLite 初始化
- `models/db.go`：`InitDB(path string) (*sql.DB, error)`
- 自动建表（`CREATE TABLE IF NOT EXISTS`）
- 连接池配置（WAL 模式、busy_timeout）

### 2.2 Task 模型
- `models/task.go`：Task 结构体 + CRUD 方法
  - `Create(task)`, `GetByID(id)`, `List(filter)`, `Update(task)`, `Delete(id)`
  - `BatchUpdateStatus(ids, status)`、`DeleteCompleted()`

### 2.3 TranscodeConfig 模型
- `models/config.go`：TranscodeConfig 结构体 + CRUD
  - `Create`, `GetByID`, `List`, `Update`, `Delete`

### 2.4 Settings 模型
- `models/settings.go`：单例 key-value 存取
  - `Get(key)`, `Set(key, value)`
  - 系统设置用固定 key（`max_concurrent`, `gpu_enabled`）

---

## 阶段 3：后端核心服务

### 3.1 BDMV Parser
- `services/parser.go`
- `ParseSource(path string) (*SourceInfo, error)`
  - 检测路径类型（目录/ISO）
  - 扫描 `PLAYLIST/*.mpls`
  - 调用 `ffprobe` 解析每个 m2ts 的流信息
  - 区分主文件（时长>阈值）和附属文件
- `GetStreams(m2tsPath string) ([]Track, error)`
  - 返回视频轨、音频轨（语言/编码/声道）、字幕轨（语言/格式）、章节

### 3.2 Encoder（ffmpeg 命令构建）
- `services/encoder.go`
- `BuildCommand(task Task, config TranscodeConfig) ([]string, error)`
  - 根据配置生成 ffmpeg 参数数组
  - 视频编码器选择（libx264/libx265/hwaccel）
  - 滤镜链拼接
  - 音频编码
  - 字幕 map
- `Run(ctx context.Context, args []string, progressCh chan<- Progress) error`
  - `os/exec` 启动 ffmpeg，stderr 管道读取
  - 解析 `progress=` 行提取百分比和时间
  - context 取消时 SIGTERM → SIGKILL

### 3.3 Worker Pool
- `services/worker.go`
- `WorkerPool` 结构体：`tasks chan *Task`, `active map[string]context.CancelFunc`
- `Start()`, `Submit(task)`, `Pause(taskID)`, `Resume(taskID)`, `Cancel(taskID)`
- 并发控制：semaphore channel
- 进度回调 → WebSocket hub

### 3.4 Monitor
- `services/monitor.go`
- `Start(ctx, hub)`
  - 每 2 秒采集 CPU（`/proc/stat` 差值计算）
  - GPU 检测：优先 nvidia-smi → rocm-smi → intel_gpu_top → 回退无 GPU
  - 推送 `system_stats` 到 WebSocket hub

---

## 阶段 4：后端 API

### 4.1 Router
- `api/router.go`：chi.NewRouter()
- 中间件链：Recoverer → Logger → CORS
- `/api` 子路由注册
- 静态文件 serve（`embed` 前端 dist）

### 4.2 WebSocket Hub
- `api/ws_handler.go`
- Hub 结构体：连接注册/注销、广播
- 客户端连接管理，心跳检测
- `POST /api/ws` 升级 + 消息路由

### 4.3 Task API
- `api/task_handler.go`
- 全部端点实现（见 spec §7）
- 批量操作逻辑：paused 只允许选中全为 paused 状态
- 删除逻辑：done 只删 DB；其余删 DB + `os.RemoveAll(output_path)`
- 重试逻辑：复制新 pending 任务

### 4.4 Source API
- `api/source_handler.go`
- `GET /source/browse`：限制在 `BD_SOURCE_ROOT` 内，返回目录树
- `POST /source/parse`：调用 Parser，返回解析结果
- `GET /source/streams/:file`：返回指定 m2ts 的轨道列表

### 4.5 Preview API
- `POST /preview`：创建预览子任务（is_preview=true），限定时间戳+长度
- `GET /preview/:id/download`：serve 临时文件
- `DELETE /preview/:id`：取消运行中预览 + 清理文件
- 定时 goroutine：每 30 分钟扫描超过 4 小时的预览临时目录并删除

### 4.6 Config API
- `api/config_handler.go`
- CRUD 端点
- `POST /export`：接受 `[id1, id2, ...]`，返回 JSON 数组
- `POST /import`：接受 JSON 数组，校验后写入

### 4.7 System API
- `api/system_handler.go`
- `GET/PUT /system/settings`
- `POST /system/gpu-test`：启动异步测试（编码一小段测试视频），结果通过 WebSocket 推送

---

## 阶段 5：前端基础框架

### 5.1 路由与布局
- `router/index.ts`：三个路由 `/`, `/tasks`, `/settings`
- `App.vue`：`el-tabs` 顶部导航 + `<router-view>`（keep-alive）
- 全局样式：深色主题基础色板

### 5.2 WebSocket Store
- `stores/ws.ts`
- 连接管理：自动连接、指数退避重连
- 消息分发：根据 `type` 通知对应 store
- 连接状态暴露

### 5.3 Task Store
- `stores/tasks.ts`
- 任务列表、筛选、分页
- 批量选择状态
- WebSocket 更新 `progress`/`status` 实时反映

### 5.4 System Store
- `stores/system.ts`
- CPU/GPU 占用、设置项
- 从 WebSocket 更新，API 读写设置

### 5.5 Config Store
- `stores/configs.ts`
- 配置列表 CRUD、导入导出

---

## 阶段 6：前端页面

### 6.1 概览页
- `views/Dashboard.vue`
- CPU/GPU 进度环（`el-progress` type="circle"）
- 当前任务数徽标
- 居中「新增任务」按钮 → 跳转到任务管理页并打开向导

### 6.2 任务管理页
- `views/Tasks.vue`
- 筛选栏：状态下拉（`el-select`），一键清除按钮
- 任务列表：`el-table` + selection，每行进度条、状态标签、操作按钮
- 展开详情：`el-table` expand，显示源路径/目标路径/配置链接/PID
- 批量操作栏：选中时显示，暂停/删除
- 空状态：引导新增任务

### 6.3 设置页
- `views/Settings.vue`，内含两个 `el-tabs` 子标签

#### 系统设置
- 并发数（`el-input-number`，范围 1-8）
- GPU 加速开关（`el-switch`）+ 测试按钮
- GPU 测试结果面板（编码器支持列表）

#### 转码配置管理
- `views/SettingsConfigs.vue`
- 配置列表（`el-table`），展开显示 JSON 参数
- 新增/编辑弹窗：
  - 简易模式：编码器、位深、质量拖动条、速度预设、画面类型、音频编码
  - 专业模式：全部参数分组表单（`el-collapse` 折叠面板）
- 导入/导出按钮

---

## 阶段 7：前端向导

### 7.1 向导容器
- `components/wizard/WizardDialog.vue`
- `el-dialog` fullscreen，`el-steps` 步骤条
- 步骤组件动态切换，共享 wizard state（Pinia store 或 provide/inject）
- 上一步/下一步/取消 按钮

### 7.2 步骤组件
- `StepSelectSource.vue`：文件浏览器树（`el-tree`），支持目录和 .iso 选择
- `StepParsing.vue`：自动调用解析 API，loading 动画，失败时错误信息+退出按钮
- `StepSelectTracks.vue`：
  - 主文件/附属文件列表，每项 checkbox
  - 展开后音轨/字幕/章节多选
  - 全选按钮
- `StepConfig.vue`：已有配置下拉 + "新建"按钮打开配置弹窗
- `StepTarget.vue`：
  - 目标路径选择器
  - 预览按钮 → 打开预览面板
  - 预览面板：文件选择 → 开始时间+长度输入 → 开始编码 → 进度条 → 下载链接

### 7.3 预览面板
- `components/wizard/PreviewPanel.vue`
- 独立的 `el-dialog`
- 选择文件、设置时间戳（`el-time-picker`）、长度（秒）
- 调用 preview API，接收 WebSocket 进度
- 完成后出现下载按钮
- 关闭时调用 DELETE cleanup

---

## 阶段 8：Docker 化

### 8.1 Dockerfile 完善
- go-builder：`CGO_ENABLED=1 GOOS=linux go build -o bdriper-server`
- node-builder：`npm run build`
- runtime：apt install ffmpeg mkvtoolnix libnvidia-encode-555
- 暴露 8080，ENTRYPOINT bdriper-server

### 8.2 docker-compose.yml
- bdriper 服务定义
- 卷挂载说明
- GPU 配置（`deploy.resources.reservations.devices`）

### 8.3 构建脚本
- `scripts/build.sh`：多阶段构建 + 打标签
- 验证点：完整构建成功，启动后可访问 UI

---

## 阶段 9：测试

### 9.1 Go 单元测试
- 模型 CRUD（内存 SQLite）
- Encoder 命令构建（参数正确性验证）
- Worker pool 调度逻辑（channel 行为）
- 配置校验

### 9.2 Go 集成测试
- 使用 `httptest.Server` + 真实 SQLite 测试所有 API 端点
- Parser 集成（需要包含一个最小 BDMV 测试 fixture）
- 端到端：创建任务 → worker 执行短片段编码 → 验证输出存在

### 9.3 前端测试
- Vitest + Vue Test Utils 测试关键组件
- Pinia store 逻辑测试

### 9.4 E2E
- Playwright 脚本：向导全流程、任务暂停/删除、配置导入导出

---

## 依赖图

```
阶段 1 (脚手架)
  ├──▶ 阶段 2 (数据层)
  │      └──▶ 阶段 3 (核心服务)
  │             └──▶ 阶段 4 (API)
  │                    ├──▶ 阶段 6 (前端页面)
  │                    │      └──▶ 阶段 7 (前端向导)
  │                    └──▶ 阶段 8 (Docker)
  └──▶ 阶段 5 (前端基础)
         └──▶ 阶段 6 (前端页面)

全部 ──▶ 阶段 9 (测试)
```

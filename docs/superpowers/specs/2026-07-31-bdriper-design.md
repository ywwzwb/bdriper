# BDRiper 设计文档

## 1. 概述

BDRiper 是一个 Docker 容器，提供 Web UI 用于将 BDMV（蓝光原盘）转码为 BDRip（MKV 格式）。

- **后端**: Go 单二进制，内嵌 SPA 前端
- **前端**: 单页应用 (SPA)，通过 REST API + WebSocket 与后端通信
- **持久化**: SQLite，文件可挂载到宿主机保留数据
- **平台**: 仅 Linux 容器
- **GPU**: 支持 NVIDIA NVENC、AMD AMF/VCE、Intel QSV 多厂商硬件编码

## 2. 整体架构

```
┌─────────────────────────────────────────────────┐
│                    Docker Container              │
│  ┌──────────────┐  ┌──────────────────────────┐ │
│  │   Go Backend  │  │   Embedded SPA           │ │
│  │               │  │                          │ │
│  │  HTTP Server  │◄─┤  REST API + WebSocket    │ │
│  │  (serve SPA   │  │                          │ │
│  │   + API)      │  │  概览 / 任务管理 / 设置   │ │
│  │               │  └──────────────────────────┘ │
│  │  Task Manager │                              │
│  │  BDMV Parser  │                              │
│  │  GPU Detector │                              │
│  │  Config Mgr   │                              │
│  └──────┬───────┘                              │
│         │                                       │
│  ┌──────▼───────┐  ┌──────────────────────────┐ │
│  │   SQLite DB   │  │   External Tools          │ │
│  │   tasks       │  │   ffmpeg, x264, x265,    │ │
│  │   configs     │  │   mkvmerge, ffprobe,     │ │
│  │   previews    │  │   vainfo, nvidia-smi     │ │
│  │   settings    │  └──────────────────────────┘ │
│  └──────────────┘                              │
│                                                  │
│  Volume mounts: /input (BDMV)  /output (MKV)     │
│                  /data (DB + configs)            │
└─────────────────────────────────────────────────┘
```

## 3. 数据模型

### Task
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | 自增 ID |
| name | TEXT | 任务名称（从碟名自动生成） |
| status | TEXT | pending / running / paused / completed / failed |
| source_path | TEXT | BDMV 源目录或 ISO 路径 |
| output_path | TEXT | 输出目录 |
| progress | REAL | 聚合进度 0.0~1.0 |
| estimated_eta | TEXT | 预估完成时间 |
| pid | INTEGER | 子进程 PID |
| config_id | INTEGER FK | 关联的转码配置 |
| error_msg | TEXT | 失败时的错误信息（stderr 最后 2000 字符） |
| deleted | BOOL | 软删除标记 |
| created_at | DATETIME | |
| updated_at | DATETIME | |

### FileEntry
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| task_id | INTEGER FK | |
| source_file | TEXT | BDMV 中的 m2ts 路径 |
| output_file | TEXT | 输出 mkv 路径 |
| streams | TEXT (JSON) | 音轨/字幕/章节流信息 |
| selected | BOOL | 是否被用户选中 |
| progress | REAL | 单文件进度 |

### TranscodeConfig
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| name | TEXT | 用户自定义名称 |
| encoder_type | TEXT | cpu / gpu |
| video_encoder | TEXT | x264 / x265 / h264_nvenc / hevc_nvenc / h264_amf / hevc_amf / h264_qsv / hevc_qsv |
| video_params | TEXT (JSON) | 完整的编码参数 |
| audio_tracks | TEXT (JSON) | 选中的音轨信息 |
| subtitle_tracks | TEXT (JSON) | 选中的字幕信息 |
| chapters_enabled | BOOL | |
| output_muxer | TEXT | mkvmerge / ffmpeg |
| is_builtin | BOOL | 是否为内置预设模板 |
| created_at | DATETIME | |

### PresetTemplate
| 字段 | 类型 | 说明 |
|------|------|------|
| name | TEXT | 模板名称 |
| encoder | TEXT | x264 / x265 |
| mode | TEXT | cpu / gpu |
| params | TEXT (JSON) | 预设参数 |
| description | TEXT | 参数说明 |
| builtin | BOOL | 固定为 true |

### PreviewJob
| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER PK | |
| task_id | INTEGER FK | |
| source_file | TEXT | 源 m2ts |
| start_time | TEXT | 开始时间戳 |
| duration | INTEGER | 时长（秒） |
| output_file | TEXT | 临时文件路径 |
| status | TEXT | pending / running / completed / failed |
| expires_at | DATETIME | 过期时间（30分钟 TTL） |
| created_at | DATETIME | |

### SystemSettings
| 字段 | 类型 | 说明 |
|------|------|------|
| key | TEXT PK | 设置键 |
| value | TEXT | 设置值 |

预置键: `max_concurrent_tasks`, `preview_ttl_minutes`, `gpu_enabled`, `log_level`, `log_max_files`, `log_max_size_mb`

### 日志系统

日志同时输出到本地文件和 WebSocket 推送，提供实时日志监视功能。

**日志文件配置：**
| 设置项 | 默认值 | 说明 |
|--------|--------|------|
| log_level | info | debug / info / warn / error |
| log_max_files | 5 | 最大保留日志文件数 |
| log_max_size_mb | 10 | 单文件最大大小（MB），超出后轮转 |

**轮转策略：**
- 按时间 + 大小双重轮转
- 当文件达到 `log_max_size_mb` 时自动轮转到新文件
- 超过 `log_max_files` 数量时自动删除最旧文件
- 日志文件存储在 `/data/logs/bdriper.log`, `/data/logs/bdriper-2026-07-31.log`, ...
- 文件名格式: `bdriper-YYYY-MM-DD-HHmmss.log` (创建时间), 当前活跃文件为 `bdriper.log`

**Web 日志查看器：**
- 实时日志通过 WebSocket 推送到前端
- 前端展示为滚动文本区域，新日志自动追加到底部并滚动
- 单日志文件超过 5000 行时，自动隐藏前半内容（显示 "... 已隐藏 2500 行旧日志 ..."）
- 支持日志级别筛选（debug/info/warn/error 复选框）

**Go 日志库:** 使用 `slog` (Go 标准库) + 自定义 handler，同时写入文件和 WebSocket 广播

## 4. API 设计

### 概览
```
GET  /api/overview/status          # CPU/GPU 占用, 当前任务数
```

### 任务管理
```
GET  /api/tasks                    # 任务列表 (?status=xxx)
POST /api/tasks                    # 创建任务
GET  /api/tasks/:id                # 任务详情
PATCH /api/tasks/:id               # 更新状态 (暂停/取消)
DELETE /api/tasks/:id              # 删除任务
DELETE /api/tasks/completed        # 一键删除已完成
POST /api/tasks/:id/retry          # 重试失败任务
POST /api/tasks/batch              # 批量操作 {ids:[], action:"pause"|"delete"}
```

### 设置
```
GET  /api/settings                 # 获取所有系统设置
PATCH /api/settings               # 更新系统设置
GET  /api/settings/gpu-info        # GPU 信息 + 支持的编解码器列表
```

### 转码配置
```
GET  /api/configs                  # 用户配置列表
POST /api/configs                  # 新建配置
GET  /api/configs/:id              # 配置详情
PUT  /api/configs/:id              # 更新配置
DELETE /api/configs/:id            # 删除配置
GET  /api/presets                  # 内置预设模板列表
```

### 向导 (BDMV 解析)
```
POST /api/wizard/parse             # 解析 BDMV 源文件 {path: "/input/BDROM"}
GET  /api/wizard/file/:path/streams # 获取单个文件的流信息
```

### 预览
```
POST /api/preview                  # 创建预览任务
GET  /api/preview/:id/status       # 预览状态
GET  /api/preview/:id/download?token=xxx  # 下载预览文件（一次性 token）
DELETE /api/preview/:id            # 取消预览
```

### 日志
```
GET  /api/logs                     # 获取最近 N 行日志 (?lines=200&level=info)
GET  /api/logs/download            # 下载完整日志文件（打包当前活跃 + 所有轮转文件为 .tar.gz）
WS   /ws/logs                      # 实时推送日志行
```

### WebSocket
```
WS   /ws/tasks                     # 实时推送任务进度更新
```

推送消息格式:

任务进度: `{type: "progress", task_id: 1, progress: 0.45, eta: "12min", status: "running"}`

日志行: `{type: "log", level: "info", time: "2026-07-31T12:00:00Z", msg: "starting encoding..."}`

## 5. 编码管线 (Hybrid Pipeline)

每个任务分 4 个阶段，由 TaskRunner goroutine 编排：

```
阶段1: 流抽取
  m2ts ──ffmpeg──▶ audio.flac/aac + subtitle.sup
  MPLS ──ffprobe──▶ chapters.txt

阶段2: 视频编码
  CPU模式 (专业): m2ts ──ffmpeg pipe (y4m)──▶ x264/x265 独立版 ──▶ video.264/.265
  GPU模式 (快速): m2ts ──ffmpeg (hwaccel nvdec/vaapi/qsv) ──▶ video.mkv

阶段3: 预览 (可选，向导中触发)
  m2ts ──seek + duration──▶ ffmpeg ──▶ /tmp/bdriper/preview_xxx.mkv

阶段4: MKV 封装
  video + audio + subtitle + chapters ──mkvmerge──▶ output.mkv
```

### 实现细节

- **进度解析**: ffmpeg 输出 `frame=12345`, x264/x265 输出 `[5.2%]`。正则匹配计算百分比
- **暂停/恢复**: 发送 SIGSTOP/SIGCONT 给子进程 (Linux)
- **并发控制**: TaskRunner pool，最多 `max_concurrent_tasks` 个 goroutine
- **错误处理**: 子进程非 0 退出码时，捕获 stderr 最后 2000 字符写入 error_msg
- **删除策略**: 已完成任务的 MKV 不删除；其他状态删除时清理输出文件
- **预览 TTL**: 30 分钟，后台 goroutine 每 30 分钟扫描清理。向导退出时立即清理

## 6. 转码配置

### 编码器层级
```
          ┌─ CPU: x264 (AVC) ─ 内置预设 + 专业/简易模式
编码器 ───┤
          ├─ CPU: x265 (HEVC) ─ 内置预设 + 专业/简易模式
          │
          ├─ GPU: NVENC (h264_nvenc / hevc_nvenc)
          ├─ GPU: AMF (h264_amf / hevc_amf)
          └─ GPU: QSV (h264_qsv / hevc_qsv)
```

### 简易模式参数（跨编码器统一）
| 参数 | 选项 |
|------|------|
| 编码器 | x264 / x265 / NVENC / AMF / QSV |
| 质量 | 低 / 中 / 高 / 无损 (映射到 crf/cq) |
| 速度 | 快 / 平衡 / 慢 (映射到 preset) |
| 编码深度 | 8bit / 10bit (x265/GPU 默认 10bit) |

### 专业模式 x264 参数（按教程分类）
| 分类 | 参数 | 默认值 |
|------|------|--------|
| 基础 | crf, preset, tune | crf=17, preset=veryslow, tune=animation |
| 帧类型 | keyint, min-keyint, bframes, ref, b-adapt, open-gop | keyint=600, ref=13, bframes=8, closed |
| 去块 | deblock (alpha:beta) | -1:-1 |
| 码率控制 | qcomp, aq-mode, aq-strength, mbtree, rc-lookahead | qcomp=0.75, aq-mode=3, aq-str=0.8, mbtree=on |
| 运动估计 | me, subme, me_range | me=tesa, subme=10, me_range=24 |
| 心理视觉 | psy-rd, psy-trellis, no-fast-pskip | psy-rd=0.6:0.15, fast-pskip=off |
| 色彩 | colormatrix, input-depth | bt709, 16 |
| 其他 | threads | 16 |

### 专业模式 x265 参数
| 分类 | 参数 | 默认值 |
|------|------|--------|
| 基础 | crf, preset | crf=15, preset=slower |
| 规范 | no-open-gop, deblock, keyint, min-keyint | keyint=360, deblock=-1:-1 |
| CB | ctu, qg-size, me, subme, merange | ctu=32, qg-size=8, me=star, subme=5, merange=38 |
| 帧 | bframes, ref, weightb, b-intra | bframes=6, ref=4 |
| 码率控制 | qcomp, aq-mode, aq-strength, no-sao, rc-lookahead | qcomp=0.65, aq-mode=1, aq-str=0.8, sao=off |
| 心理视觉 | psy-rd, psy-rdoq, rdoq-level, rd | psy-rd=2.0, psy-rdoq=1.0, rdoq-level=2, rd=5 |
| 色彩 | cbqpoffs, crqpoffs, colormatrix | cbqpoffs=-2, crqpoffs=-2, bt709 |
| 其他 | limit-tu, no-amp, no-rect, no-strong-intra-smoothing | |

### 内置预设模板
1. **x264 High Quality (mbtree on)** — crf=16.5, veryslow, aq-mode=3, mbtree=on
2. **x264 High Quality (mbtree off)** — crf=17.5, veryslow, aq-mode=3, mbtree=off
3. **x265 HQ Anime** — crf=15, slower, aq-mode=1, sao=off
4. **x265 Balanced** — crf=18, slow, aq-mode=2

### 帮助文档
- 转码配置界面提供「帮助」按钮
- 参数说明从 markdown 源文件在容器构建时编译为 HTML，支持后期编辑

## 7. BDMV 解析

### 来源
1. BDMV 文件夹路径（容器内挂载点）
2. ISO 镜像文件（自动挂载后解析）

### 解析流程
```
1. 检测路径类型 (目录/ISO)
2. 如果是 ISO: mount 到临时目录
3. 列出 BDMV/STREAM/*.m2ts
4. 读取 BDMV/META/DL/bdmt_eng.xml 获取碟名 (fallback: 文件夹名)
5. 读取 BDMV/PLAYLIST/*.mpls 获取章节信息
6. 对每个 m2ts 运行 ffprobe 获取:
   - 视频流: 编码格式, 分辨率, 帧率, 时长, bitdepth
   - 音频流: 编码格式, 采样率, 声道, 语言
   - 字幕流: 编码格式 (PGS), 语言
7. 返回结构化数据给前端
```

### 智能分类
- **主文件**: 时长 > 阈值且包含视频流（排除菜单/片头）
- **附属文件**: 时长较短或无视频流（菜单、预告等）

## 8. 前端设计

### 设计系统 (来自 UI/UX Pro Max)
| 元素 | 选择 |
|------|------|
| 风格 | Dark Mode (OLED) |
| 字体 | Inter (300/400/500/600/700) |
| 底色 | #0F172A |
| 前景 | #F8FAFC |
| 卡片 | #1B2336 |
| 强调色 | #22C55E (执行/成功) |
| 错误色 | #EF4444 (失败/取消) |
| 边框 | #475569 |
| 图标库 | Phosphor Icons |

### 页面结构
- **顶部导航**: Logo + 四个 Tab (概览 / 任务管理 / 日志 / 设置)
- **概览 Tab**: CPU/GPU 占用率卡片 + 当前任务状态 + 新增任务按钮
- **任务管理 Tab**: 状态筛选栏 + 任务列表（可展开详情）+ 批量操作按钮
- **日志 Tab**: 实时滚动日志查看器 + 级别筛选 + 自动隐藏旧内容 + 下载日志按钮
- **设置 Tab**: 系统设置（含日志配置）+ 转码配置管理

### 新增任务向导 (4 步)
1. **选择源文件**: 浏览器选择容器内目录，支持 BDMV 文件夹 / ISO
2. **选择文件与流**: 显示碟名，主文件/附属文件列表（可全选），音轨/字幕/章节多选
3. **转码配置**: 选择已有配置或新建（简易/专业模式），帮助按钮查看参数说明
4. **目标路径 & 预览**: 设置输出目录，可选预览转码（临时文件 TTL 30min，离开向导立即清理）

### 交互细节
- 任务进度通过 WebSocket 实时推送
- 任务点击展开/收起详情
- 多选后显示批量操作按钮
- 配置页面：简易模式 4 个下拉框，专业模式分类展示所有参数
- 日志页面：实时滚动到底部，支持 debug/info/warn/error 级别筛选，超长自动截断前半内容

## 9. 容器设计

### 基础镜像
```dockerfile
FROM ubuntu:24.04
# 核心依赖: ffmpeg, x264, x265, mkvtoolnix
# GPU 检测: vainfo, nvidia-smi (runtime 注入)
```

### 多阶段构建
```
Stage 1: apt install 系统依赖
Stage 2: Go 编译 (CGO_ENABLED=0, 静态链接)
Stage 3: 前端构建 (npm build)
Stage 4: 运行时 (复制二进制 + 前端 dist + 预设模板 + 帮助文档)
```

### 卷挂载
| 路径 | 用途 | 必需 |
|------|------|------|
| /input | BDMV 源文件 / ISO 挂载 | 是 |
| /output | 转码输出 MKV | 是 |
| /data | SQLite + 用户配置持久化 | 推荐 |
| /tmp/bdriper | 预览临时文件 | 否 |

### GPU 支持
- **NVIDIA**: `--runtime=nvidia`, 通过 nvidia-container-toolkit 注入驱动
- **Intel QSV**: `--device=/dev/dri`, 依赖 vainfo + intel-media-va-driver
- **AMD**: `--device=/dev/dri`, 依赖 mesa-va-drivers

启动时自动检测并汇总可用 GPU 和编解码器。

## 10. 错误处理

| 场景 | 处理方式 |
|------|---------|
| BDMV 解析失败 | 向导步骤 2 报错，用户可退出 |
| 编码进程崩溃 | task status → failed, 写入 error_msg (stderr 末尾 2000 字符) |
| 磁盘空间不足 | 捕获 ffmpeg ENOSPC, task → failed |
| GPU 不可用 | 启动时检测，前端显示 "无可用 GPU"，GPU 编码选项置灰 |
| 用户关闭浏览器 | 任务继续在后台运行，预览超时后 TTL 自动清理 |
| 容器重启 | 启动时扫描 running task, 标记为 failed (PID 失效) |

## 11. 测试策略

- **单元测试**: Go 各包 (task manager, config, wizard parser, GPU detector)
- **集成测试**: API 端点 + SQLite CRUD
- **容器测试**: Docker build + run 验证工具链可用性
- **前端测试**: 组件渲染 + 向导流程 E2E

## 12. 关键文件结构

```
bdriper/
├── cmd/server/          # Go 入口
├── internal/
│   ├── api/             # HTTP handlers + WebSocket
│   ├── task/            # TaskRunner, 进度解析, 子进程管理
│   ├── config/          # 转码配置 CRUD + 预设模板
│   ├── wizard/          # BDMV 解析 (ffprobe + META XML)
│   ├── gpu/             # GPU 检测 (NVML/VA-API/QSV)
│   ├── preview/         # 预览任务 + TTL 清理
│   ├── log/             # slog handler: 文件轮转 + WebSocket 广播
│   └── db/              # SQLite 迁移 + 查询
├── web/                 # SPA 前端 (Vite + Vue/React)
│   └── src/
│       ├── components/  # UI 组件
│       ├── pages/       # 概览/任务/日志/设置
│       ├── wizard/      # 向导组件
│       └── config/      # 配置编辑器
├── presets/             # 内置转码预设 JSON
├── docs/                # 参数帮助 markdown (构建时编译 HTML)
├── Dockerfile
└── docker-compose.yml
```

前端框架建议: **Vue 3 + Vite + Tailwind CSS**。Vue 适合数据驱动的多状态 UI（任务列表、向导步骤、配置表单），Tailwind 快速实现设计系统。也可以选择 React。

## 13. 转码配置（编码参数说明）

以下内容来自 BDRip 制作流程文档，需要在容器构建时编译为 HTML 前端帮助参考。

### 主要编码器与参数说明

#### x264 (AVC) 参数说明

根据文档所述，x264 参数主要分为三大类：

- **编码规范类 (Specification)**: 规定编码格式、Profile/Level 等
- **效率取舍类 (Trade-off)**: 以时间换画质，如 `--ref`、`--bframes`、`--me`、`--subme` 等
- **码率控制类 (Rate Control)**: 决定码率分配策略，如 `--crf`、`--qcomp`、`--aq`、`--psy` 等

关键参数说明：
- **crf (码率控制方法)**: 固定质量模式，值越低质量越高（范围 0-51，默认 23），通常采用 19-21.5
- **preset (效率预设)**: 编码效率与速度的取舍，推荐 slow/slower/veryslow
- **tune (画质预设)**: 针对不同片源类型优化，如 film、animation、grain
- **deblock (去色块)**: 去除编码产生的色块效应，高码率下建议调低
- **keyint/min-keyint (GOP 区间)**: 控制 IDR 帧间距，影响拖动进度条的响应速度
- **bframes (连续 B 帧数)**: 动漫建议 6-12，真人电影 3-8
- **ref (参考帧数)**: 越高压缩率越好，但解码压力也越大
- **qcomp (码率时域分配)**: 决定 QP 值的时间变化灵活度
- **aq-mode/aq-strength (自适应量化)**: 防止纹理和平面处码率过低，动漫推荐 mode=3、strength=0.6-1.0
- **mbtree (宏块树码率控制)**: 高码率 (crf<16) 关闭，中低码率 (crf>18) 开启
- **me/subme/me_range (运动估计)**: 时间换画质的参数组合
- **psy-rd/psy-trellis (心理视觉优化)**: 保留画面纹理和细节锐度

#### x265 (HEVC) 参数说明

关键参数说明：
- **crf**: 默认 28，高质量编码建议至少 18 起
- **preset**: 新手建议使用 preset 一键设置
- **ctu**: 最大编码单元，1080p 建议限制为 32
- **qg-size**: QP 调整最小单位，值越低灵活度越高
- **me (运动估计)**: star search (me=3) 综合优于 umh
- **no-sao (关闭 SAO)**: 非极低码率编码建议关闭，避免暴力涂抹
- **aq-mode/aq-strength**: mode=1 适合高码率，mode=2 适合中低码率
- **psy-rd/psy-rdoq**: 调节锐利度和细节保留
- **rdoq-level**: slow 及以上 preset 自动开启
- **pbratio**: 降低 P/B 帧间画质差距，动漫编码建议 1.2
- **qcomp**: 略高于默认 (0.6)，中高画质优化推荐 0.65

编码器参数研发遵循遗传算法思路：选定代表性片源 → 逐个参数变异测试 → 筛选优胜参数 → 迭代优化，最终达到目标码率下的最佳画质表现。

### 封装与容器说明

- **MKV**: 支持几乎所有音视频编码格式以及 PGS 字幕
- **MP4**: 兼容性更好，但不支持 FLAC 音轨和 PGS 字幕
- **章节**: 从 BDMV/PLAYLIST/*.mpls 提取
- **字幕**: PGS (BD 原盘图形字幕) 格式，.sup 后缀

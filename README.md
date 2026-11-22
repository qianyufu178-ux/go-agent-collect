# Nexus-Edge · 采集端与推流器

`go-agent-collect` 是「摄像头 → 网关 → 转码器 → 浏览器」链路的最上游工程：
本机摄像头采集 + 低延迟 H.264 编码 + MPEG-TS over HTTP 推流，同时承载边缘侧的 C++ 限流 / 指标采集能力（cgo 桥接）。

> **先看现状再跑代码**：本仓库当前**可直接运行的是 `cmd/cam-push`**（有单测、有真实长跑数据）；
> `edge/`、`controller/` 是 Nexus-Edge / Nexus-Controller 的**目录骨架与设计文档**，尚无 Go 源码。

## 现状

| 模块 | 状态 | 说明 |
| --- | --- | --- |
| `cmd/cam-push` | ✅ 可用 | DirectShow 采集 → `libx264` → MPEG-TS → HTTP chunked PUT 网关；含单测 |
| `pkg/config` | ✅ 可用 | `nexus-edge.yaml` 的结构体映射与校验（`EdgeConfig`）；含单测 |
| `pkg/errcode`、`pkg/logger` | ✅ 可用 | 错误码与日志封装 |
| `cpp/` + `pkg/cbridge` | ✅ 源码可用 | 令牌桶 / 滑动窗口 / 最大并发 + `/proc` 指标读取；需先 CMake 构建 `libnexus_core`，**默认不参与编译**（cgo） |
| `proto/nexus.proto` | 📄 契约 | Edge ↔ Controller 的 gRPC 协议定义，未生成代码 |
| `edge/`、`controller/` | 🚧 规划 | 仅目录骨架（无 `.go` 文件）；Makefile 的 `edge` / `controller` 目标为预留，当前**不可构建** |
| `doc/` | 📚 文档 | `cam-push.md`（已实现部分）、`nexus-edge.md` / `controller.md` / `system-requirements.md`（设计） |

## cam-push 特性

- **单请求长连接推流**：`PUT {gateway}/ingest/live/{stream}`，`Content-Type: video/mp2t`，`ContentLength = -1` 强制 chunked；
  网关先读流，推流结束后回统计 JSON。
- **先鉴权后推流**：`X-Ingest-Token` 由网关在读取 body 之前校验，采集端能立刻拿到 `401` / `409`（单发布者冲突），不会白推一帧。
- **异常自愈**：FFmpeg 退出、网关断开、摄像头被拔都会结束本次推流；进程按 `1s → 2s → 4s …`（上限 10s）退避重连，**稳定运行超过 30s 后重置退避**。
- **进程治理**：推流结束（无论正常/异常）都 `Kill` + `Wait` 掉 FFmpeg，避免野进程占着摄像头；Windows 下用 `CREATE_NO_WINDOW` 隐藏控制台。
- **可观测**：每 5 秒打印实时码率与累计流量，FFmpeg 的 `warning` 逐行转发到日志。
- **编码参数**：`-f dshow → libx264 -preset veryfast -tune zerolatency -profile:v main`，
  `-g/-keyint_min = fps`，`-sc_threshold 0`（每个 GOP 起点必是 IDR），CBR（`maxrate = bitrate`，`bufsize = 2×`），
  封装 `mpegts -mpegts_flags +resend_headers -flush_packets 1` 直接从 stdout 管道推走。

## 快速开始

前置：**Windows 10/11**（DirectShow 采集）、FFmpeg（需带 `dshow`、`libx264`）、Go 1.18+。

```powershell
cd go-agent-collect

# 0) 指定 ffmpeg（也可以把 ffmpeg 放进 PATH，或用 -ffmpeg 参数）
$env:FFMPEG_PATH = 'E:\tools\ffmpeg\bin\ffmpeg.exe'

# 1) 列出本机摄像头（DirectShow 视频设备）
go run ./cmd/cam-push -list

# 2) 默认推流：自动选第一个摄像头 → 127.0.0.1:8082 的 cam0
go run ./cmd/cam-push

# 3) 指定设备 / 码率 / 网关
go run ./cmd/cam-push -device "HD 720P webcam" -bitrate 4000 -gateway http://127.0.0.1:8082
```

### 命令行参数

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-list` | `false` | 只列出摄像头后退出 |
| `-device` | `auto` | 摄像头名称；`auto` 取第一个 DirectShow 视频设备 |
| `-gateway` | `http://127.0.0.1:8082` | 网关地址（推流地址 = `{gateway}/ingest/live/{stream}`） |
| `-stream` | `cam0` | 直播流名称 |
| `-token` | `nexus-demo-token` | 推流令牌，作为 `X-Ingest-Token` 发送 |
| `-ffmpeg` | `ffmpeg` | FFmpeg 可执行文件路径 |
| `-width` / `-height` | `1280` / `720` | 采集分辨率 |
| `-fps` | `30` | 采集帧率（同时作为 GOP 长度） |
| `-bitrate` | `6000` | 目标码率 kbps（同时作为 `maxrate`，`bufsize = 2×`） |

### 环境变量

| 变量 | 对应参数 |
| --- | --- |
| `FFMPEG_PATH` | `-ffmpeg` |
| `CAMERA_NAME` | `-device` |
| `GATEWAY_URL` | `-gateway` |
| `STREAM_NAME` | `-stream` |
| `INGEST_TOKEN` | `-token` |
| `CAM_WIDTH` / `CAM_HEIGHT` / `CAM_FPS` / `CAM_BITRATE_KBPS` | 分辨率 / 帧率 / 码率 |

> 命令行参数优先于环境变量。

## 推流协议（与网关对齐）

```
PUT /ingest/live/{name}
Content-Type: video/mp2t
X-Ingest-Token: <token>
Transfer-Encoding: chunked     (Go: ContentLength = -1)

<MPEG-TS 字节流>               ← 持续写入，直到任一端断开
```

网关侧行为（见 `go-apisix-extend/pkg/live`）：先鉴权 → 单发布者校验（冲突 `409`）→ 按块扇出给所有订阅者；
慢订阅者队列满时**整块丢弃**并计数，绝不阻塞推流端；同一路流在订阅空闲 `idle_timeout` 后主动断开，交给采集端/转码端重连。

## C++ 能力（可选，默认不启用）

`cpp/` 里是边缘侧的原生实现，通过 `pkg/cbridge`（cgo）暴露给 Go：

| 目录 | 内容 |
| --- | --- |
| `cpp/ratelimit/` | 令牌桶、滑动窗口、最大并发限制器（C++17） |
| `cpp/proc_reader/` | `/proc` 解析：CPU、内存、磁盘 IO、网络 IO、磁盘使用率 |
| `pkg/cbridge/bridge.cpp` | C ABI 桥接层（`cgo_*` 函数），供 Go 调用 |

构建静态库（产物必须落在 `build/cpp`，与 `bridge.go` 的 `-L${SRCDIR}/../../build/cpp` 对应）：

```powershell
cd go-agent-collect
mkdir build\cpp -Force
cd build\cpp
cmake ..\..\cpp -DCMAKE_BUILD_TYPE=Release
cmake --build . --config Release

# 回到仓库根目录，开启 cgo 验证桥接包
cd ..\..
$env:CGO_ENABLED = 1
go build ./pkg/cbridge     # 需要 g++（MinGW-w64 / MSVC）与 libnexus_core
```

说明：

- 默认 `CGO_ENABLED=0`，`pkg/cbridge` 会被构建约束排除，`go build ./...` / `go test ./...` 照常通过。
- `cpp/CMakeLists.txt` 的 `BUILD_TESTS` 选项（默认 `OFF`）会引用 `cpp/tests/` 下的 GTest 用例，当前目录未提供，**保持默认 OFF** 即可。

## 测试与质量门禁

```powershell
cd go-agent-collect
go vet ./...          # 实测 exit 0
go test ./...         # cmd/cam-push、pkg/config 两个包含用例；errcode/logger 无测试文件
go test ./cmd/cam-push -v
```

覆盖内容：DirectShow 设备列表解析（跳过 `Alternative name` 与音频段）、FFmpeg 关键参数拼装、推流地址拼接、
`nexus-edge.yaml` 配置解析与校验。

## Makefile

```bash
make help        # 查看所有目标
make cpp         # 构建 C++ 静态库到 build/cpp
make cpp-test    # CMake + ctest（需要 GTest，且 cpp/tests/ 需自行补齐）
make proto       # 由 proto/nexus.proto 生成 gRPC 代码（需要 protoc + go plugins）
make test        # cpp-test + go test ./...
make edge        # 🚧 预留：edge/ 尚无 Go 源码，当前会失败
make controller  # 🚧 预留：controller/ 尚无 Go 源码，当前会失败
```

## 配置文件

- `nexus-edge.yaml`：节点 ID、反向代理上游、限流算法（token_bucket / sliding_window / max_concurrency）、
  采集间隔与日志采集、本地缓存（sqlite/file）、gRPC 上报、管理 API 端口。
- `nexus-controller.yaml`：gRPC 监听、心跳超时、SQLite 存储与批量落库、指标/日志保留天数、管理 API 端口。

这两份 YAML **目前是设计契约**（`pkg/config` 已完成解析与测试），对应的 `edge/`、`controller/` 服务端尚未实现。

## 目录结构

```
cmd/cam-push/        摄像头采集与推流（已实现，含 camera.go / sysproc_*.go）
pkg/cbridge/         cgo 桥接：TokenBucket / SlidingWindow / MaxConcurrency / ProcReader
pkg/config/          EdgeConfig 配置解析与校验
pkg/errcode/         统一错误码
pkg/logger/          日志封装
cpp/                 C++17：ratelimit + proc_reader（CMake 构建为 libnexus_core）
proto/               nexus.proto（Edge ↔ Controller gRPC 契约，gen/ 未生成）
edge/                规划：api / cache / cmd / collector / module / proxy / ratelimit / transport
controller/          规划：api / cmd / node / store / strategy / transport
doc/                 cam-push.md、nexus-edge.md、controller.md、system-requirements.md
```

## 已知限制

- **仅 Windows DirectShow**：Linux/macOS 请改用 RTSP 摄像头（把 `rtsp://…` 直接给转码器的 `NEXUS_INPUT_URL`）。
- **一路摄像头一个进程**：多路请用不同 `-stream` 启动多个进程；未做进程内多路复用与统一调度。
- 摄像头被会议软件等占用时 FFmpeg 报 `Could not open video device`，需先释放设备。
- **无断点续传 / 无本地落盘队列**：网关不可达时只能退避重连，期间的画面不缓存（设计文档中的磁盘队列为规划项）。
- `cbridge` 需要 CGO 与已编译的 `libnexus_core`，默认构建链不包含它。

## 相关文档

- [`doc/cam-push.md`](doc/cam-push.md) 推流器详细说明（参数、协议、测试、限制）
- [`doc/nexus-edge.md`](doc/nexus-edge.md) Nexus-Edge 设计（代理 / 限流 / 采集 / 缓存 / 上报）
- [`doc/controller.md`](doc/controller.md) Nexus-Controller 设计（节点管理 / 存储 / 策略下发）
- [`doc/system-requirements.md`](doc/system-requirements.md) 系统需求与指标口径

> 端到端链路：`cmd/cam-push` → `go-apisix-extend`（`/ingest/live/{name}`）→ `go-nexus-stream`（`NEXUS_INPUT_URL`）→ 浏览器 WebRTC。

## 许可证

[MIT](LICENSE) © 2026 qianyufu178-ux


# cam-push · 本机摄像头推流器

`go-agent-collect/cmd/cam-push` 是端到端串联链路里的**采集端**：
调用 FFmpeg 的 Windows DirectShow 采集本机摄像头，编码为低延迟 H.264，
封装成 MPEG-TS 后以 HTTP chunked 方式长连接推给网关的直播接入面。

```
摄像头(DirectShow) → FFmpeg(libx264 main profile, veryfast) → MPEG-TS(stdout) → HTTP PUT 网关 /ingest/live/{name}
```

## 一、快速开始

```powershell
cd go-agent-collect

# 1) 列出本机摄像头（DirectShow 视频设备）
go run ./cmd/cam-push -list

# 2) 推流（默认自动选择第一个摄像头，推到 127.0.0.1:8082 的 cam0）
$env:FFMPEG_PATH = 'E:\project\other-job\.tools\ffmpeg\bin\ffmpeg.exe'
go run ./cmd/cam-push

# 3) 指定设备与码率
go run ./cmd/cam-push -device "HD 720P webcam" -bitrate 3000 -gateway http://127.0.0.1:8082
```

## 二、命令行参数

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `-list` | `false` | 只列出摄像头后退出 |
| `-device` | `auto` | 摄像头名称，`auto` 取第一个 DirectShow 视频设备 |
| `-gateway` | `http://127.0.0.1:8082` | 网关地址（推流地址 = `{gateway}/ingest/live/{stream}`） |
| `-stream` | `cam0` | 直播流名称 |
| `-token` | `nexus-demo-token` | 推流令牌，作为 `X-Ingest-Token` 请求头发送 |
| `-ffmpeg` | `ffmpeg` | FFmpeg 可执行文件路径 |
| `-width` / `-height` | `1280` / `720` | 采集分辨率 |
| `-fps` | `30` | 采集帧率（同时作为 GOP） |
| `-bitrate` | `6000` | 目标码率 kbps（同时作为 maxrate，bufsize = 2×） |

## 三、环境变量

| 变量 | 对应参数 |
| --- | --- |
| `FFMPEG_PATH` | `-ffmpeg` |
| `CAMERA_NAME` | `-device` |
| `GATEWAY_URL` | `-gateway` |
| `STREAM_NAME` | `-stream` |
| `INGEST_TOKEN` | `-token` |
| `CAM_WIDTH` / `CAM_HEIGHT` / `CAM_FPS` / `CAM_BITRATE_KBPS` | 分辨率 / 帧率 / 码率 |

## 四、实现要点

1. **单次请求、长连接推流**：`PUT {gateway}/ingest/live/{stream}`，
   body 是 FFmpeg 的 stdout，`ContentLength=-1` 强制 chunked；网关先读流，
   推流结束后再回统计 JSON。
2. **先鉴权后推流**：网关在读取 body 之前完成 `X-Ingest-Token` 与「单发布者」校验，
   采集端能第一时间拿到 `401` / `409`，不会白推一帧。
3. **异常自愈**：FFmpeg 退出、网关断开、摄像头被拔都会让本次推流结束；
   进程按 1s→2s→4s…（上限 10s）退避重连，稳定运行超过 30s 后重置退避。
4. **进程生命周期**：推流结束（无论正常/异常）都会 `Kill` 掉 FFmpeg 并 `Wait`，
   避免野进程占着摄像头；Windows 下通过 `CREATE_NO_WINDOW` 隐藏控制台窗口。
5. **可观测**：每 5 秒打印一次实时码率与累计流量；FFmpeg 的 warning 逐行转发。

## 五、单元测试

```powershell
cd go-agent-collect
go test ./cmd/cam-push -v
```

覆盖：DirectShow 设备列表解析（跳过 `Alternative name` 与音频段）、
FFmpeg 关键参数拼装、推流地址拼接。

## 六、已知限制

- 仅支持 Windows DirectShow；Linux/macOS 请改用 RTSP 摄像头
  （`NEXUS_INPUT_URL=rtsp://...` 直接给转码器，或让网关接入其它推流协议）。
- 单实例只采集一路摄像头；多路请用不同 `-stream` 启动多个进程。
- 摄像头被会议软件等占用时，FFmpeg 会报 `Could not open video device`，
  需要先释放设备。

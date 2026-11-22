// Command cam-push 把本机摄像头的实时画面推送给网关直播接入面。
//
// 链路：
//
//	本机摄像头(DirectShow) → FFmpeg(H.264 + MPEG-TS) → HTTP chunked PUT → 网关 /ingest/live/{name}
//
// 网关再把同一份流扇出给 go-nexus-stream 转码器，最终由浏览器通过 WebRTC 观看。
//
// 常用命令：
//
//	cam-push -list                                       # 列出本机摄像头
//	cam-push -gateway http://127.0.0.1:8082 -stream cam0  # 开始推流
//
// 支持的环境变量：FFMPEG_PATH / CAMERA_NAME / GATEWAY_URL / STREAM_NAME / INGEST_TOKEN
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

// config 是推流器的运行参数。
type config struct {
	ffmpeg    string
	device    string
	list      bool
	gateway   string
	stream    string
	token     string
	width     int
	height    int
	fps       int
	bitrate   int
	videoSize string
}

// publishURL 拼出网关推流地址。
func (c config) publishURL() string {
	base := strings.TrimRight(c.gateway, "/")
	return fmt.Sprintf("%s/ingest/live/%s", base, c.stream)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	cfg := parseFlags()
	if cfg.list {
		listDevices(cfg.ffmpeg)
		return
	}

	device, err := resolveDevice(cfg.ffmpeg, cfg.device)
	if err != nil {
		log.Fatalf("解析摄像头失败: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("推流目标 %s（设备 %q, %s, %d kbps）", cfg.publishURL(), device, cfg.videoSize, cfg.bitrate)

	backoff := time.Second
	for {
		if ctx.Err() != nil {
			log.Println("收到退出信号，推流结束")
			return
		}

		started := time.Now()
		err := pushOnce(ctx, cfg, device)
		if ctx.Err() != nil {
			log.Println("收到退出信号，推流结束")
			return
		}
		if err != nil {
			if time.Since(started) > 30*time.Second {
				// 稳定跑过一阵再断，退避重置，避免偶发抖动后越等越久
				backoff = time.Second
			}
			log.Printf("推流中断: %v；%s 后重连", err, backoff)
		}

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			log.Println("收到退出信号，推流结束")
			return
		}
		if backoff < 10*time.Second {
			backoff *= 2
		}
	}
}

func parseFlags() config {
	cfg := config{}
	flag.StringVar(&cfg.ffmpeg, "ffmpeg", envOr("FFMPEG_PATH", "ffmpeg"), "ffmpeg 可执行文件路径")
	flag.StringVar(&cfg.device, "device", envOr("CAMERA_NAME", "auto"), "摄像头名称，auto 表示自动选择第一个")
	flag.BoolVar(&cfg.list, "list", false, "仅列出本机摄像头")
	flag.StringVar(&cfg.gateway, "gateway", envOr("GATEWAY_URL", "http://127.0.0.1:8082"), "网关地址")
	flag.StringVar(&cfg.stream, "stream", envOr("STREAM_NAME", "cam0"), "直播流名称")
	flag.StringVar(&cfg.token, "token", envOr("INGEST_TOKEN", "nexus-demo-token"), "推流令牌（X-Ingest-Token）")
	flag.IntVar(&cfg.width, "width", envInt("CAM_WIDTH", 1280), "采集宽度")
	flag.IntVar(&cfg.height, "height", envInt("CAM_HEIGHT", 720), "采集高度")
	flag.IntVar(&cfg.fps, "fps", envInt("CAM_FPS", 30), "采集帧率")
	flag.IntVar(&cfg.bitrate, "bitrate", envInt("CAM_BITRATE_KBPS", 6000), "目标码率 kbps")
	flag.Parse()

	cfg.videoSize = fmt.Sprintf("%dx%d@%dfps", cfg.width, cfg.height, cfg.fps)
	return cfg
}

// resolveDevice 解析摄像头名称：auto 时自动列出并选择第一个。
func resolveDevice(ffmpeg, name string) (string, error) {
	if trimmed := strings.TrimSpace(name); trimmed != "" && !strings.EqualFold(trimmed, "auto") {
		return trimmed, nil
	}

	devices, err := enumerateCameras(ffmpeg)
	if err != nil {
		return "", err
	}
	if len(devices) == 0 {
		return "", errors.New("未发现可用摄像头（可用 -list 查看）")
	}
	return devices[0], nil
}

// pushOnce 启动一次 FFmpeg 并把输出推给网关，直到任一端断开。
func pushOnce(ctx context.Context, cfg config, device string) error {
	cmd := exec.Command(cfg.ffmpeg, buildFFmpegArgs(cfg, device)...)
	cmd.SysProcAttr = hiddenProcAttr()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("接管 ffmpeg 输出失败: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("接管 ffmpeg 日志失败: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动 ffmpeg 失败: %w", err)
	}
	log.Printf("ffmpeg 已启动 pid=%d", cmd.Process.Pid)

	go pumpStderr(stderr)

	counter := &countingReader{r: stdout}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, cfg.publishURL(), counter)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return fmt.Errorf("构造推流请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "video/mp2t")
	req.Header.Set("X-Ingest-Token", cfg.token)
	// 明确使用 chunked：直播流长度未知
	req.ContentLength = -1

	progressDone := make(chan struct{})
	go reportProgress(ctx, counter, progressDone)

	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	resp, err := client.Do(req)
	close(progressDone)

	// 推流结束（无论正常/异常），先确保 ffmpeg 退出，避免野进程占着摄像头
	_ = cmd.Process.Kill()
	_ = cmd.Wait()

	if err != nil {
		return fmt.Errorf("推送失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("网关返回 %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	log.Printf("网关已结束本次推流: %s", strings.TrimSpace(string(body)))
	return errors.New("网关关闭了推流连接")
}

// pumpStderr 把 ffmpeg 的日志打到标准错误，便于排查采集中断。
func pumpStderr(r io.Reader) {
	buf := make([]byte, 4096)
	var carry []byte
	for {
		n, err := r.Read(buf)
		if n > 0 {
			carry = append(carry, buf[:n]...)
			for {
				idx := -1
				for i, b := range carry {
					if b == '\n' {
						idx = i
						break
					}
				}
				if idx < 0 {
					break
				}
				line := strings.TrimSpace(string(carry[:idx]))
				carry = carry[idx+1:]
				if line != "" {
					log.Printf("ffmpeg| %s", line)
				}
			}
		}
		if err != nil {
			return
		}
	}
}

// reportProgress 每 5 秒打印一次推流码率。
func reportProgress(ctx context.Context, counter *countingReader, done <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	last := atomic.LoadUint64(&counter.n)
	lastAt := time.Now()
	for {
		select {
		case <-done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			total := atomic.LoadUint64(&counter.n)
			delta := total - last
			seconds := now.Sub(lastAt).Seconds()
			last, lastAt = total, now
			if seconds <= 0 {
				continue
			}
			log.Printf("推流中 %.2f Mbps（累计 %.2f MB）",
				float64(delta)*8/seconds/1e6, float64(total)/1024/1024)
		}
	}
}

// countingReader 统计已推送字节数。
type countingReader struct {
	r io.Reader
	n uint64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if n > 0 {
		atomic.AddUint64(&c.n, uint64(n))
	}
	return n, err
}

// buildFFmpegArgs 拼装采集 + 编码 + 封装参数。
func buildFFmpegArgs(cfg config, device string) []string {
	gop := cfg.fps
	if gop <= 0 {
		gop = 30
	}

	return []string{
		"-hide_banner", "-nostdin", "-loglevel", "warning",
		// ---- 采集：Windows DirectShow ----
		"-f", "dshow",
		"-rtbufsize", "256M",
		"-framerate", strconv.Itoa(cfg.fps),
		"-video_size", fmt.Sprintf("%dx%d", cfg.width, cfg.height),
		"-i", "video=" + device,
		// ---- 编码：低延迟 H.264（main profile 压缩效率明显优于 baseline）----
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-tune", "zerolatency",
		"-profile:v", "main",
		"-pix_fmt", "yuv420p",
		"-sc_threshold", "0",
		"-g", strconv.Itoa(gop),
		"-keyint_min", strconv.Itoa(gop),
		"-b:v", strconv.Itoa(cfg.bitrate) + "k",
		"-maxrate", strconv.Itoa(cfg.bitrate) + "k",
		"-bufsize", strconv.Itoa(cfg.bitrate*2) + "k",
		"-an", "-sn", "-dn",
		// ---- 封装：MPEG-TS over stdout，便于网关按块扇出 ----
		"-f", "mpegts",
		"-mpegts_flags", "+resend_headers",
		"-flush_packets", "1",
		"pipe:1",
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

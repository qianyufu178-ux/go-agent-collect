package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// enumerateCameras 调用 ffmpeg 列出 DirectShow 视频设备。
func enumerateCameras(ffmpeg string) ([]string, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("当前仅支持 Windows DirectShow 采集（当前系统: %s）", runtime.GOOS)
	}

	cmd := exec.Command(ffmpeg, "-hide_banner", "-list_devices", "true", "-f", "dshow", "-i", "dummy")
	cmd.SysProcAttr = hiddenProcAttr()

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// ffmpeg 打印完设备列表后一定以非零码退出，这里只在「完全没输出」时判失败。
	// 注意：FFmpeg 8 起不再打印 "DirectShow video devices" 分段标题，
	// 改为在设备行尾标注 (video)/(audio)，因此不能只用 "DirectShow" 判定成功。
	if err := cmd.Run(); err != nil && !looksLikeDeviceList(stderr.String()) {
		return nil, fmt.Errorf("调用 ffmpeg 失败: %w（输出: %s）", err, strings.TrimSpace(stderr.String()))
	}

	return parseDshowDevices(stderr.String()), nil
}

// looksLikeDeviceList 判断 stderr 是否已经包含设备列表（兼容新旧两种输出格式）。
func looksLikeDeviceList(output string) bool {
	return strings.Contains(output, "DirectShow") ||
		strings.Contains(output, "Alternative name") ||
		strings.Contains(output, "(video)") ||
		strings.Contains(output, "(audio)")
}

// listDevices 打印本机摄像头列表。
func listDevices(ffmpeg string) {
	devices, err := enumerateCameras(ffmpeg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "列出摄像头失败: %v\n", err)
		os.Exit(1)
	}
	if len(devices) == 0 {
		fmt.Println("未发现可用摄像头")
		return
	}

	fmt.Printf("发现 %d 个摄像头：\n", len(devices))
	for i, device := range devices {
		fmt.Printf("  [%d] %s\n", i+1, device)
	}
}

// parseDshowDevices 解析 ffmpeg -list_devices 的 stderr 输出，兼容两种格式：
//
//	FFmpeg ≤ 7：分段标题 + 纯名称行
//	  [dshow @ ...] DirectShow video devices (some may be both video and audio devices)
//	  [dshow @ ...]  "HD 720P webcam"
//	  [dshow @ ...] DirectShow audio devices
//	FFmpeg ≥ 8：没有分段标题，设备行尾直接标注类型
//	  [in#0 @ ...] "HD 720P webcam" (video)
//	  [in#0 @ ...] "麦克风 (USB microphone)" (audio)
func parseDshowDevices(output string) []string {
	devices := make([]string, 0, 4)

	inVideo := false
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		switch {
		case strings.Contains(trimmed, "DirectShow video devices"):
			inVideo = true
			continue
		case strings.Contains(trimmed, "DirectShow audio devices"):
			inVideo = false
			continue
		}

		name, ok := quotedName(trimmed)
		if !ok || strings.HasPrefix(name, "@device_pnp_") || strings.HasPrefix(name, "@device_cm_") {
			continue
		}

		// 新版本：行尾标注了类型，直接按标注判断，避免把音频设备当摄像头。
		if strings.HasSuffix(trimmed, "(video)") {
			devices = append(devices, name)
			continue
		}
		if strings.HasSuffix(trimmed, "(audio)") {
			continue
		}

		// 旧版本：靠分段标题判断当前是否处于视频段。
		if inVideo {
			devices = append(devices, name)
		}
	}

	return devices
}

// quotedName 取出 `[dshow @ ...] "名称"` 中的名称。
func quotedName(line string) (string, bool) {
	first := strings.Index(line, "\"")
	last := strings.LastIndex(line, "\"")
	if first < 0 || last <= first {
		return "", false
	}
	return line[first+1 : last], true
}

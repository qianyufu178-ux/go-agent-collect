package main

import (
	"reflect"
	"strings"
	"testing"
)

// dshowSample 是 ffmpeg -list_devices 在本机上的真实输出（含音频段与 Alternative name）。
const dshowSample = `[dshow @ 000002ca24759a40] DirectShow video devices (some may be both video and audio devices)
[dshow @ 000002ca24759a40]  "HD 720P webcam"
[dshow @ 000002ca24759a40]     Alternative name "@device_pnp_\\?\usb#vid_1bcf&pid_2701&mi_00#6&28afa389&0&0000#{65e8773d-8f56-11d0-a3b9-00a0c9223196}\global"
[dshow @ 000002ca24759a40]  "OBS Virtual Camera"
[dshow @ 000002ca24759a40] DirectShow audio devices
[dshow @ 000002ca24759a40]  "Microphone (USB Audio)"
[dshow @ 000002ca24759a40] Could not enumerate audio only devices (or none found).
dummy: Immediate exit requested`

// dshowSampleV8 是 FFmpeg 8.x 的真实输出：没有分段标题，改为在行尾标注 (video)/(audio)。
const dshowSampleV8 = `[in#0 @ 000001964e6df9c0] "HD 720P webcam" (video)
[in#0 @ 000001964e6df9c0]   Alternative name "@device_pnp_\\?\usb#vid_1bcf&pid_2701&mi_00#6&28afa389&0&0000#{65e8773d-8f56-11d0-a3b9-00a0c9223196}\global"
[in#0 @ 000001964e6df9c0] "麦克风 (USB microphone)" (audio)
[in#0 @ 000001964e6df9c0]   Alternative name "@device_cm_{33D9A762-90C8-11D0-BD43-00A0C911CE86}\wave_{66A4AB06-78A9-47C2-9400-F49733EAA4CA}"
dummy: Immediate exit requested`

// TestParseDshowDevices 验证只解析视频设备，且跳过 Alternative name。
func TestParseDshowDevices(t *testing.T) {
	got := parseDshowDevices(dshowSample)
	want := []string{"HD 720P webcam", "OBS Virtual Camera"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseDshowDevices() = %v, want %v", got, want)
	}
}

// TestParseDshowDevicesV8 验证 FFmpeg 8 的新格式：靠 (video)/(audio) 标注区分设备类型。
func TestParseDshowDevicesV8(t *testing.T) {
	got := parseDshowDevices(dshowSampleV8)
	want := []string{"HD 720P webcam"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseDshowDevices() = %v, want %v", got, want)
	}
}

// TestLooksLikeDeviceList 验证新旧两种输出都能被判定为「拿到了设备列表」。
func TestLooksLikeDeviceList(t *testing.T) {
	cases := []struct {
		name   string
		output string
		want   bool
	}{
		{"旧格式分段标题", dshowSample, true},
		{"新格式类型标注", dshowSampleV8, true},
		{"只有报错", "dummy: Immediate exit requested\nError opening input", false},
		{"空输出", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := looksLikeDeviceList(tc.output); got != tc.want {
				t.Fatalf("looksLikeDeviceList() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestParseDshowDevices_Empty 无视频设备时返回空切片。
func TestParseDshowDevices_Empty(t *testing.T) {
	got := parseDshowDevices("dummy: Immediate exit requested")
	if len(got) != 0 {
		t.Fatalf("parseDshowDevices() = %v, want empty", got)
	}
}

// TestBuildFFmpegArgs 验证关键采集/编码/封装参数都在。
func TestBuildFFmpegArgs(t *testing.T) {
	cfg := config{width: 1280, height: 720, fps: 30, bitrate: 2500}
	args := strings.Join(buildFFmpegArgs(cfg, "HD 720P webcam"), " ")

	for _, want := range []string{
		"-f dshow",
		"-video_size 1280x720",
		"-framerate 30",
		"-i video=HD 720P webcam",
		"-c:v libx264",
		"-tune zerolatency",
		"-g 30",
		"-b:v 2500k",
		"-f mpegts",
		"pipe:1",
	} {
		if !strings.Contains(args, want) {
			t.Errorf("args missing %q\nargs=%s", want, args)
		}
	}
}

// TestPublishURL 验证推流地址拼接。
func TestPublishURL(t *testing.T) {
	cases := []struct {
		gateway string
		stream  string
		want    string
	}{
		{"http://127.0.0.1:8082", "cam0", "http://127.0.0.1:8082/ingest/live/cam0"},
		{"http://127.0.0.1:8082/", "cam0", "http://127.0.0.1:8082/ingest/live/cam0"},
		{"http://gateway:9000", "cam1", "http://gateway:9000/ingest/live/cam1"},
	}

	for _, tc := range cases {
		cfg := config{gateway: tc.gateway, stream: tc.stream}
		if got := cfg.publishURL(); got != tc.want {
			t.Errorf("publishURL(%q, %q) = %q, want %q", tc.gateway, tc.stream, got, tc.want)
		}
	}
}

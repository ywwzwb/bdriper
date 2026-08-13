package gpu

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type EncoderInfo struct {
	Name      string `json:"name"`
	Codec     string `json:"codec"`
	Supported bool   `json:"supported"`
}

type GPUInfo struct {
	Vendor   string        `json:"vendor"`
	Model    string        `json:"model"`
	Encoders []EncoderInfo `json:"encoders"`
}

func DetectGPUs() []GPUInfo {
	var gpus []GPUInfo

	if gpu := detectNVIDIA(); gpu != nil {
		gpus = append(gpus, *gpu)
	}
	if gpu := detectIntelQSV(); gpu != nil {
		gpus = append(gpus, *gpu)
	}
	if gpu := detectAMD(); gpu != nil {
		gpus = append(gpus, *gpu)
	}

	return gpus
}

func detectNVIDIA() *GPUInfo {
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		return nil
	}

	cmd := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
	out, err := cmd.Output()
	if err != nil {
		slog.Debug("nvidia-smi query failed", "error", err)
		return nil
	}

	model := strings.TrimSpace(string(out))
	if model == "" {
		return nil
	}

	encoders := detectFFmpegEncoders("nvenc", []string{"h264_nvenc", "hevc_nvenc"})
	return &GPUInfo{
		Vendor:   "NVIDIA",
		Model:    model,
		Encoders: encoders,
	}
}

func detectIntelQSV() *GPUInfo {
	matches, err := filepath.Glob("/dev/dri/renderD*")
	if err != nil || len(matches) == 0 {
		return nil
	}

	model := "Intel Quick Sync"
	vendorPath := "/sys/class/drm/card0/device/vendor"
	if data, err := os.ReadFile(vendorPath); err == nil {
		vendorID := strings.TrimSpace(string(data))
		if vendorID == "0x8086" {
			if data, err := os.ReadFile("/sys/class/drm/card0/device/device"); err == nil {
				deviceID := strings.TrimSpace(string(data))
				model = "Intel " + vendorID + ":" + deviceID
			}
		}
	}

	encoders := detectFFmpegEncoders("qsv", []string{"h264_qsv", "hevc_qsv"})
	return &GPUInfo{
		Vendor:   "Intel",
		Model:    model,
		Encoders: encoders,
	}
}

func detectAMD() *GPUInfo {
	if _, err := exec.LookPath("vainfo"); err != nil {
		return nil
	}

	cmd := exec.Command("vainfo")
	if err := cmd.Run(); err != nil {
		slog.Debug("vainfo check failed", "error", err)
		return nil
	}

	encoders := detectFFmpegEncoders("amf", []string{"h264_amf", "hevc_amf"})
	return &GPUInfo{
		Vendor:   "AMD",
		Model:    "AMD GPU",
		Encoders: encoders,
	}
}

func detectFFmpegEncoders(pattern string, codecNames []string) []EncoderInfo {
	var encoders []EncoderInfo
	for _, name := range codecNames {
		encoders = append(encoders, EncoderInfo{
			Name:      name,
			Codec:     codecFromEncoder(name),
			Supported: ffmpegEncoderWorks(pattern, name),
		})
	}

	return encoders
}

// ffmpegEncoderWorks reports whether the given encoder can actually initialize
// and encode. Merely grepping `ffmpeg -encoders` is not enough: the encoder
// binary is compiled with support, but the runtime library (e.g.
// libnvidia-encode.so.1 for NVENC) may be missing inside the container. A tiny
// real encode is the only reliable check.
func ffmpegEncoderWorks(pattern, codec string) bool {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	args := []string{
		"-v", "error",
		"-f", "lavfi",
		"-i", "testsrc=duration=0.1:size=64x64:rate=5",
		"-frames:v", "1",
		"-c:v", codec,
		"-f", "null", "-",
	}

	// QSV needs an explicit device, otherwise the encoder cannot initialize.
	if pattern == "qsv" {
		args = append([]string{"-init_hw_device", "qsv=hw"}, args...)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		slog.Debug("ffmpeg encoder test failed", "codec", codec, "stderr", strings.TrimSpace(stderr.String()))
		return false
	}
	return true
}

func codecFromEncoder(name string) string {
	if strings.HasPrefix(name, "h264") {
		return "H.264"
	}
	if strings.HasPrefix(name, "hevc") {
		return "HEVC"
	}
	return name
}

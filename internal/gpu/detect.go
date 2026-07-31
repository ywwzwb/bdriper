package gpu

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil
	}

	cmd := exec.Command("ffmpeg", "-hide_banner", "-encoders")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return nil
		}
	}

	var encoders []EncoderInfo
	for _, name := range codecNames {
		supported := strings.Contains(string(out), name)
		encoders = append(encoders, EncoderInfo{
			Name:      name,
			Codec:     codecFromEncoder(name),
			Supported: supported,
		})
	}

	return encoders
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

package task

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type PipelineConfig struct {
	SourceFile      string
	OutputDir       string
	AudioTracks     []int
	SubtitleTracks  []int
	ChaptersEnabled bool
	VideoEncoder    string
	VideoParams     string
}

func extractStreams(cfg PipelineConfig) ([]*exec.Cmd, error) {
	var cmds []*exec.Cmd
	base := filepath.Join(cfg.OutputDir, filepath.Base(cfg.SourceFile))

	for _, idx := range cfg.AudioTracks {
		audioOut := fmt.Sprintf("%s_audio_%d.flac", base, idx)
		cmd := exec.Command("ffmpeg",
			"-i", cfg.SourceFile,
			"-map", fmt.Sprintf("0:%d", idx),
			"-c:a", "flac",
			"-y", audioOut,
		)
		cmds = append(cmds, cmd)
	}

	for _, idx := range cfg.SubtitleTracks {
		subOut := fmt.Sprintf("%s_subtitle_%d.sup", base, idx)
		cmd := exec.Command("ffmpeg",
			"-i", cfg.SourceFile,
			"-map", fmt.Sprintf("0:%d", idx),
			"-c:s", "copy",
			"-y", subOut,
		)
		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

func encodeVideo(cfg PipelineConfig) (*exec.Cmd, error) {
	videoOut := fmt.Sprintf("%s_video.265", filepath.Join(cfg.OutputDir, filepath.Base(cfg.SourceFile)))

	if cfg.VideoEncoder == "x264" || cfg.VideoEncoder == "x265" {
		var params map[string]any
		json.Unmarshal([]byte(cfg.VideoParams), &params)

		var flags []string
		for k, v := range params {
			switch val := v.(type) {
			case bool:
				if val {
					flags = append(flags, fmt.Sprintf("--%s", k))
				} else {
					flags = append(flags, fmt.Sprintf("--no-%s", k))
				}
			case float64:
				if val == float64(int(val)) {
					flags = append(flags, fmt.Sprintf("--%s", k), fmt.Sprintf("%d", int(val)))
				} else {
					flags = append(flags, fmt.Sprintf("--%s", k), fmt.Sprintf("%.1f", val))
				}
			case string:
				flags = append(flags, fmt.Sprintf("--%s", k), val)
			}
		}

		encoderArgs := strings.Join(flags, " ")
		return exec.Command("bash", "-c", fmt.Sprintf(
			"ffmpeg -i %s -f yuv4mpegpipe -pix_fmt yuv420p10le -strict -1 - 2>/dev/null | %s --y4m %s -o %s -",
			cfg.SourceFile, cfg.VideoEncoder, encoderArgs, videoOut,
		)), nil
	}

	return exec.Command("ffmpeg",
		"-hwaccel", "auto",
		"-i", cfg.SourceFile,
		"-c:v", cfg.VideoEncoder,
		"-y", videoOut,
	), nil
}

func muxMKV(videoFile string, outputDir string, sourceFile string) *exec.Cmd {
	base := filepath.Join(outputDir, filepath.Base(sourceFile))
	output := base + ".mkv"
	return exec.Command("mkvmerge",
		"-o", output,
		videoFile,
	)
}

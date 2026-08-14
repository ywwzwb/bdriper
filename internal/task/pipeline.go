package task

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// utf8Env returns an environment slice with a UTF-8 locale forced. Without it
// (e.g. POSIX/C locale, the default in minimal containers) mkvmerge v82 fails
// to parse paths containing multi-byte UTF-8 characters (CJK etc.), truncating
// them at the last ASCII path segment.
func utf8Env() []string {
	return append(os.Environ(), "LANG=C.UTF-8", "LC_ALL=C.UTF-8")
}

type PipelineConfig struct {
	SourceFile      string
	OutputDir       string
	TaskID          int64
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
		cmd := exec.Command("/usr/bin/ffmpeg",
			"-i", cfg.SourceFile,
			"-map", fmt.Sprintf("0:%d", idx),
			"-c:a", "flac",
			"-y", audioOut,
		)
		cmds = append(cmds, cmd)
	}

	for _, idx := range cfg.SubtitleTracks {
		subOut := fmt.Sprintf("%s_subtitle_%d.sup", base, idx)
		cmd := exec.Command("/usr/bin/ffmpeg",
			"-i", cfg.SourceFile,
			"-map", fmt.Sprintf("0:%d", idx),
			"-c:s", "copy",
			"-y", subOut,
		)
		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

func encodeVideo(cfg PipelineConfig) (*exec.Cmd, *exec.Cmd, error) {
	baseName := filepath.Base(cfg.SourceFile)
	ext := filepath.Ext(baseName)
	videoOut := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_t%d%s_video.265", baseName[:len(baseName)-len(ext)], cfg.TaskID, ext))

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

		ffmpegArgs := []string{
			"-i", cfg.SourceFile,
			"-f", "yuv4mpegpipe",
			"-pix_fmt", "yuv420p",
			"-strict", "-1",
			"-",
		}
		ffmpegCmd := exec.Command("/usr/bin/ffmpeg", ffmpegArgs...)

		encoderPath := "/usr/bin/" + cfg.VideoEncoder
		encoderArgs := []string{"-o", videoOut, "-"}
		if cfg.VideoEncoder == "x264" {
			encoderArgs = append([]string{"--demuxer", "y4m"}, encoderArgs...)
		} else {
			encoderArgs = append([]string{"--y4m"}, encoderArgs...)
		}
		encoderArgs = append(encoderArgs, flags...)
		encoderCmd := exec.Command(encoderPath, encoderArgs...)

		pr, pw, err := os.Pipe()
		if err != nil {
			return nil, nil, err
		}
		encoderCmd.Stdin = pr
		ffmpegCmd.Stdout = pw
		return ffmpegCmd, encoderCmd, nil
	}

	return nil, exec.Command("ffmpeg",
		"-hwaccel", "auto",
		"-i", cfg.SourceFile,
		"-c:v", cfg.VideoEncoder,
		"-y", videoOut,
	), nil
}

func muxMKV(videoFile string, outputDir string, sourceFile string, taskID int64) *exec.Cmd {
	baseName := filepath.Base(sourceFile)
	ext := filepath.Ext(baseName)
	output := filepath.Join(outputDir, fmt.Sprintf("%s_t%d%s.mkv", baseName[:len(baseName)-len(ext)], taskID, ext))
	// Take video from encoded file, everything else from source
	cmd := exec.Command("/usr/bin/mkvmerge",
		"-o", output,
		"--no-video", sourceFile,
		videoFile,
	)
	cmd.Env = utf8Env()
	return cmd
}

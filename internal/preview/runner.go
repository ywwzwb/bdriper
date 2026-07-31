package preview

import (
	"fmt"
	"os/exec"
)

func RunPreview(sourceFile, startTime string, duration int, outputFile string) (*exec.Cmd, error) {
	args := []string{
		"-ss", startTime,
		"-i", sourceFile,
		"-t", fmt.Sprintf("%d", duration),
		"-c:v", "libx264",
		"-crf", "23",
		"-preset", "fast",
		"-c:a", "copy",
		"-y",
		outputFile,
	}
	cmd := exec.Command("ffmpeg", args...)
	return cmd, cmd.Start()
}

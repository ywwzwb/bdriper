package preview

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

var timeRe = regexp.MustCompile(`time=(\d{2}):(\d{2}):(\d{2})\.(\d{2})`)

func RunPreview(sourceFile, startTime, encoder string, videoParams map[string]any, duration int, outputFile string) (*exec.Cmd, chan float64, error) {
	// Build encoder args from params
	var vcodec string
	var encArgs []string
	switch encoder {
	case "x264":
		vcodec = "libx264"
	case "x265":
		vcodec = "libx265"
	default:
		vcodec = encoder
	}

	if crf, ok := videoParams["crf"]; ok {
		encArgs = append(encArgs, "-crf", fmt.Sprint(crf))
	} else {
		encArgs = append(encArgs, "-crf", "23")
	}
	if preset, ok := videoParams["preset"]; ok {
		encArgs = append(encArgs, "-preset", fmt.Sprint(preset))
	} else {
		encArgs = append(encArgs, "-preset", "fast")
	}

	args := []string{
		"-ss", startTime,
		"-i", sourceFile,
		"-t", fmt.Sprintf("%d", duration),
		"-c:v", vcodec,
	}
	args = append(args, encArgs...)
	args = append(args,
		"-c:a", "aac",
		"-b:a", "128k",
		"-y",
		"-f", "matroska",
		outputFile,
	)
	cmd := exec.Command("ffmpeg", args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	ch := make(chan float64, 10)

	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(stderr)
		scanner.Split(scanLines)
		for scanner.Scan() {
			line := scanner.Text()
			matches := timeRe.FindStringSubmatch(line)
			if matches == nil {
				continue
			}
			h, _ := strconv.Atoi(matches[1])
			m, _ := strconv.Atoi(matches[2])
			s, _ := strconv.Atoi(matches[3])
			cs, _ := strconv.Atoi(matches[4])
			current := float64(h*3600+m*60+s) + float64(cs)/100.0
			if duration > 0 {
				pct := current / float64(duration)
				if pct > 1 {
					pct = 1
				}
				select {
				case ch <- pct:
				default:
				}
			}
		}
	}()

	return cmd, ch, cmd.Start()
}

func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i := 0; i < len(data); i++ {
		if data[i] == '\r' || data[i] == '\n' {
			start := i + 1
			for start < len(data) && (data[start] == '\r' || data[start] == '\n') {
				start++
			}
			return start, data[:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

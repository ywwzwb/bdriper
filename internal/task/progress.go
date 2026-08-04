package task

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
)

type ProgressInfo struct {
	Percent float64
	Frame   int64
	FPS     float64
	Bitrate string
	Speed   string
}

var (
	ffmpegFrameRe = regexp.MustCompile(`frame=\s*(\d+)`)
	x265PercentRe = regexp.MustCompile(`\[(\d+\.?\d*)%\]`)
	x264FrameRe   = regexp.MustCompile(`^(\d+)\s+frames:`)
)

func parseProgress(r io.Reader, totalFrames int64, ch chan<- ProgressInfo) {
	defer close(ch)
	scanner := bufio.NewScanner(r)
	scanner.Split(scanCRLF)
	for scanner.Scan() {
		line := scanner.Text()
		pi := ProgressInfo{}

		if matches := ffmpegFrameRe.FindStringSubmatch(line); matches != nil {
			pi.Frame, _ = strconv.ParseInt(matches[1], 10, 64)
			if totalFrames > 0 {
				pi.Percent = float64(pi.Frame) / float64(totalFrames) * 100
			}
		}

		if matches := x265PercentRe.FindStringSubmatch(line); matches != nil {
			pi.Percent, _ = strconv.ParseFloat(matches[1], 64)
		}

		if matches := x264FrameRe.FindStringSubmatch(line); matches != nil {
			frames, _ := strconv.ParseInt(matches[1], 10, 64)
			pi.Frame = frames
			if totalFrames > 0 {
				pi.Percent = float64(frames) / float64(totalFrames) * 100
			}
		}

		if pi.Percent > 0 || pi.Frame > 0 {
			ch <- pi
		}
	}
}

func scanCRLF(data []byte, atEOF bool) (int, []byte, error) {
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

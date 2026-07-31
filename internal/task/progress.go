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
	ffmpegTimeRe  = regexp.MustCompile(`time=(\d{2}):(\d{2}):(\d{2})\.(\d{2})`)
	x265PercentRe = regexp.MustCompile(`\[(\d+\.?\d*)%\]`)
	x264PercentRe = regexp.MustCompile(`(\d+\.?\d*).%`)
)

func parseProgress(r io.Reader, totalFrames int64, ch chan<- ProgressInfo) {
	scanner := bufio.NewScanner(r)
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

		if pi.Percent > 0 || pi.Frame > 0 {
			ch <- pi
		}
	}
}

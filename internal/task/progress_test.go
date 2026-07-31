package task

import (
	"strings"
	"testing"
)

func TestParseProgressFFmpeg(t *testing.T) {
	input := `frame=  150 fps=3.2 q=28.0 size=    1024kB time=00:00:05.00 bitrate=1600.0kbits/s speed=0.5x`
	r := strings.NewReader(input)
	ch := make(chan ProgressInfo, 10)

	go parseProgress(r, 300, ch)

	pi := <-ch
	if pi.Frame != 150 {
		t.Errorf("Frame = %d, want 150", pi.Frame)
	}
}

func TestParseProgressFFmpegWithPercent(t *testing.T) {
	input := `frame=  150 fps=3.2 q=28.0 size=    1024kB time=00:00:05.00 bitrate=1600.0kbits/s speed=0.5x`
	r := strings.NewReader(input)
	ch := make(chan ProgressInfo, 10)

	go parseProgress(r, 300, ch)

	pi := <-ch
	expectedPercent := (150.0 / 300.0) * 100
	if pi.Percent != expectedPercent {
		t.Errorf("Percent = %f, want %f", pi.Percent, expectedPercent)
	}
}

func TestParseProgressX265(t *testing.T) {
	input := `[45.2%] 1356/3000 frames, 4.51 fps, 2345.67 kb/s, eta 0:06:05`
	r := strings.NewReader(input)
	ch := make(chan ProgressInfo, 10)

	go parseProgress(r, 3000, ch)

	pi := <-ch
	if pi.Percent < 45 || pi.Percent > 46 {
		t.Errorf("Percent = %f, want ~45.2", pi.Percent)
	}
}

func TestX264PercentRegex(t *testing.T) {
	matches := x264PercentRe.FindStringSubmatch("50.5z% 1515/3000 frames,")
	if matches == nil {
		t.Fatal("x264PercentRe did not match")
	}
	if matches[1] != "50.5" {
		t.Errorf("captured percent = %q, want 50.5", matches[1])
	}
}

func TestParseProgressNoMatch(t *testing.T) {
	input := `some random log line with no progress info`
	r := strings.NewReader(input)
	ch := make(chan ProgressInfo, 2)

	go func() {
		parseProgress(r, 0, ch)
		close(ch)
	}()

	pi, ok := <-ch
	if ok {
		t.Errorf("unexpected progress event for non-matching line: %+v", pi)
	}
}

func TestParseProgressMultipleLines(t *testing.T) {
	input := `frame=    1 fps=0.0 q=0.0 size=       0kB time=N/A bitrate=N/A speed=N/A
frame=  100 fps=10.0 q=28.0 size=     512kB time=00:00:04.00 bitrate=1000.0kbits/s speed=0.5x
frame=  200 fps=10.0 q=28.0 size=    1024kB time=00:00:08.00 bitrate=1000.0kbits/s speed=0.5x`
	r := strings.NewReader(input)
	ch := make(chan ProgressInfo, 10)

	go func() {
		parseProgress(r, 300, ch)
		close(ch)
	}()

	var frames []int64
	for pi := range ch {
		frames = append(frames, pi.Frame)
	}
	if len(frames) != 3 {
		t.Fatalf("got %d progress events, want 3", len(frames))
	}
	if frames[0] != 1 {
		t.Errorf("first frame = %d, want 1", frames[0])
	}
	if frames[2] != 200 {
		t.Errorf("last frame = %d, want 200", frames[2])
	}
}

func TestParseProgressLargeFrameNumber(t *testing.T) {
	input := `frame=12345 fps=14.0 q=28.0 size=   12345kB time=00:08:13.80 bitrate=1500.0kbits/s speed=0.6x`
	r := strings.NewReader(input)
	ch := make(chan ProgressInfo, 10)

	go parseProgress(r, 0, ch)

	pi := <-ch
	if pi.Frame != 12345 {
		t.Errorf("Frame = %d, want 12345", pi.Frame)
	}
	if pi.Percent != 0 {
		t.Errorf("Percent with totalFrames=0 should be 0, got %f", pi.Percent)
	}
}

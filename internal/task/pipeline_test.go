package task

import (
	"os"
	"testing"
)

func TestEncodeVideoHardwareReturnsNilFFmpegCmd(t *testing.T) {
	cfg := PipelineConfig{
		SourceFile:   "00002.m2ts",
		OutputDir:    t.TempDir(),
		TaskID:       8,
		VideoEncoder: "hevc_nvenc",
		VideoParams:  "{}",
	}

	ffmpegCmd, videoCmd, err := encodeVideo(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if ffmpegCmd != nil {
		t.Fatalf("hardware encoder must return nil ffmpegCmd, got %v", ffmpegCmd)
	}
	if videoCmd == nil {
		t.Fatal("videoCmd must not be nil")
	}

	// Reproduces the panic from runner.go: ffmpegCmd.Stdout deref on a nil
	// *exec.Cmd. Must be guarded by a nil check before use.
	var stdout any
	if ffmpegCmd != nil {
		stdout = ffmpegCmd.Stdout
	}
	if pw, ok := stdout.(*os.File); ok && pw != nil {
		t.Fatal("unexpected pipe for hardware encoder")
	}
}

func TestEncodeVideoX265ReturnsFFmpegAndEncoder(t *testing.T) {
	cfg := PipelineConfig{
		SourceFile:   "00002.m2ts",
		OutputDir:    t.TempDir(),
		TaskID:       8,
		VideoEncoder: "x265",
		VideoParams:  `{"crf":16.5}`,
	}

	ffmpegCmd, videoCmd, err := encodeVideo(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if ffmpegCmd == nil {
		t.Fatal("x265 path must return a non-nil ffmpegCmd")
	}
	if videoCmd == nil {
		t.Fatal("x265 path must return a non-nil encoder cmd")
	}
}

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

func TestMuxMKVForcesUTF8Locale(t *testing.T) {
	cmd := muxMKV(
		"/mnt/动漫/[龙珠]/00002_t8.m2ts_video.265",
		"/mnt/动漫/[龙珠]",
		"/mnt/动漫/[龙珠]/DRAGON_BALL_01/BDMV/STREAM/00002.m2ts",
		8,
	)

	if cmd.Env == nil {
		t.Fatal("mux cmd must have explicit env")
	}
	env := map[string]string{}
	for _, kv := range cmd.Env {
		if i := indexByte(kv, '='); i >= 0 {
			env[kv[:i]] = kv[i+1:]
		}
	}
	if env["LANG"] != "C.UTF-8" || env["LC_ALL"] != "C.UTF-8" {
		t.Fatalf("expected UTF-8 locale in env, got LANG=%q LC_ALL=%q", env["LANG"], env["LC_ALL"])
	}
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

package wizard

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type BDMVFile struct {
	Path       string  `json:"path"`
	Duration   string  `json:"duration"`
	Resolution string  `json:"resolution"`
	FPS        float64 `json:"fps"`
	IsMain     bool    `json:"is_main"`
}

type BDMVInfo struct {
	DiscName string     `json:"disc_name"`
	Files    []BDMVFile `json:"files"`
}

type Stream struct {
	Index      int    `json:"index"`
	Codec      string `json:"codec"`
	Type       string `json:"type"`
	Language   string `json:"language"`
	Channels   int    `json:"channels,omitempty"`
	SampleRate string `json:"sample_rate,omitempty"`
}

type FileStreamInfo struct {
	Video    []Stream `json:"video"`
	Audio    []Stream `json:"audio"`
	Subtitle []Stream `json:"subtitle"`
}

func ParseBDMV(sourcePath string) (*BDMVInfo, error) {
	var streamDir string
	var metaRoot string

	// Pattern 1: path/BDMV/STREAM (sourcePath is the disc root)
	if _, err := os.Stat(filepath.Join(sourcePath, "BDMV", "STREAM")); err == nil {
		streamDir = filepath.Join(sourcePath, "BDMV", "STREAM")
		metaRoot = sourcePath
	} else if _, err := os.Stat(filepath.Join(sourcePath, "STREAM")); err == nil {
		// Pattern 2: path/STREAM (sourcePath IS the BDMV directory)
		streamDir = filepath.Join(sourcePath, "STREAM")
		metaRoot = filepath.Dir(sourcePath)
	} else {
		return nil, fmt.Errorf("not a valid BDMV directory: %s (expected BDMV/STREAM or STREAM subdirectory)", sourcePath)
	}

	entries, err := os.ReadDir(streamDir)
	if err != nil {
		return nil, fmt.Errorf("read STREAM dir: %w", err)
	}

	info := &BDMVInfo{
		DiscName: parseDiscName(metaRoot),
	}

	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".m2ts") {
			f := filepath.Join(streamDir, e.Name())
			bf, err := probeM2TS(f)
			if err != nil {
				continue
			}
			info.Files = append(info.Files, *bf)
		}
	}

	if len(info.Files) == 0 {
		return nil, fmt.Errorf("no .m2ts files found in %s", streamDir)
	}

	return info, nil
}

func parseDiscName(sourcePath string) string {
	metaFiles := []string{
		filepath.Join(sourcePath, "BDMV", "META", "DL", "bdmt_eng.xml"),
		filepath.Join(sourcePath, "BDMV", "META", "DL", "bdmt_jpn.xml"),
		filepath.Join(sourcePath, "META", "DL", "bdmt_eng.xml"),
	}
	for _, mf := range metaFiles {
		data, err := os.ReadFile(mf)
		if err != nil {
			continue
		}
		var meta struct {
			Extension struct {
				Name string `xml:"name"`
			} `xml:"extension"`
		}
		if err := xml.Unmarshal(data, &meta); err == nil && meta.Extension.Name != "" {
			return meta.Extension.Name
		}
	}
	return filepath.Base(sourcePath)
}

func probeM2TS(path string) (*BDMVFile, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var probe struct {
		Streams []struct {
			CodecType  string `json:"codec_type"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
			Duration   string `json:"duration"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	json.Unmarshal(out, &probe)

	dur := probe.Format.Duration
	durDisplay := "0:00"
	if d := parseDuration(dur); d > 0 {
		durDisplay = fmt.Sprintf("%.0f:%02.0f", d/60, modSeconds(d))
	}

	bf := &BDMVFile{
		Path:     path,
		Duration: durDisplay,
	}

	for _, s := range probe.Streams {
		if s.CodecType == "video" {
			bf.Resolution = fmt.Sprintf("%dx%d", s.Width, s.Height)
			bf.FPS = parseFPS(s.RFrameRate)
			bf.IsMain = true
		}
	}

	bf.IsMain = bf.IsMain && parseSeconds(dur) > 60
	return bf, nil
}

func GetFileStreams(m2tsPath string) (*FileStreamInfo, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_streams",
		m2tsPath,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var probe struct {
		Streams []struct {
			Index      int    `json:"index"`
			CodecType  string `json:"codec_type"`
			CodecName  string `json:"codec_name"`
			Channels   int    `json:"channels"`
			SampleRate string `json:"sample_rate"`
			Tags       struct {
				Language string `json:"language"`
			} `json:"tags"`
		} `json:"streams"`
	}
	json.Unmarshal(out, &probe)

	result := &FileStreamInfo{}
	for _, s := range probe.Streams {
		stream := Stream{
			Index:      s.Index,
			Codec:      s.CodecName,
			Type:       s.CodecType,
			Language:   s.Tags.Language,
			Channels:   s.Channels,
			SampleRate: s.SampleRate,
		}
		switch s.CodecType {
		case "video":
			result.Video = append(result.Video, stream)
		case "audio":
			result.Audio = append(result.Audio, stream)
		case "subtitle":
			result.Subtitle = append(result.Subtitle, stream)
		}
	}
	return result, nil
}

func parseFPS(rate string) float64 {
	parts := strings.Split(rate, "/")
	if len(parts) == 2 {
		num, _ := strconv.ParseFloat(parts[0], 64)
		den, _ := strconv.ParseFloat(parts[1], 64)
		if den > 0 {
			return num / den
		}
	}
	return 0
}

func parseSeconds(dur string) float64 {
	d, _ := strconv.ParseFloat(dur, 64)
	return d
}

func parseDuration(dur string) float64 {
	d, err := strconv.ParseFloat(dur, 64)
	if err != nil {
		return 0
	}
	return d
}

func modSeconds(d float64) float64 {
	return math.Mod(d, 60)
}

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
				// Fallback: ffprobe not available — just list file with basic info
				info2, _ := e.Info()
				bf = &BDMVFile{
					Path:       f,
					Duration:   fmt.Sprintf("%.0f MB", float64(info2.Size())/(1024*1024)),
					Resolution: "?",
					IsMain:     info2.Size() > 50*1024*1024, // >50MB = likely main content
				}
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

	// Try to enrich with language info from CLIPINF
	langMap := parseClipinfLanguages(m2tsPath)

	result := &FileStreamInfo{}
	audioIdx := 0
	subIdx := 0
	for _, s := range probe.Streams {
		lang := s.Tags.Language
		if lang == "" {
			switch s.CodecType {
			case "audio":
				if l, ok := langMap[fmt.Sprintf("audio_%d", audioIdx)]; ok {
					lang = l
				}
				audioIdx++
			case "subtitle":
				if l, ok := langMap[fmt.Sprintf("subtitle_%d", subIdx)]; ok {
					lang = l
				}
				subIdx++
			}
		}
		stream := Stream{
			Index:      s.Index,
			Codec:      s.CodecName,
			Type:       s.CodecType,
			Language:   lang,
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

func parseClipinfLanguages(m2tsPath string) map[string]string {
	m2tsName := filepath.Base(m2tsPath)
	baseName := strings.TrimSuffix(m2tsName, filepath.Ext(m2tsName))
	clpiPath := filepath.Join(filepath.Dir(filepath.Dir(m2tsPath)), "CLIPINF", baseName+".clpi")

	data, err := os.ReadFile(clpiPath)
	if err != nil {
		return nil
	}

	result := make(map[string]string)
	audioCount := 0
	subCount := 0

	// Scan for stream entry pattern:
	//   4 bytes: 0x00000000 (delimiter)
	//   1 byte:  length (0x11-0x12)
	//   1 byte:  stream_index
	//   1 byte:  0x15
	//   1 byte:  stream_coding_type (0x80-0x86 audio, 0x90/0x92 subtitle)
	//   For audio: 1 byte format + 3 bytes language
	//   For subtitle: 3 bytes language
	for i := 0; i < len(data)-10; i++ {
		if data[i] != 0x00 || data[i+1] != 0x00 || data[i+2] != 0x00 || data[i+3] != 0x00 {
			continue
		}
		entryLen := int(data[i+4])
		if entryLen < 5 || entryLen > 30 {
			continue
		}
		if data[i+6] != 0x15 {
			continue
		}
		codingType := data[i+7]
		isAudio := codingType >= 0x80 && codingType <= 0x86
		isSub := codingType == 0x90 || codingType == 0x92
		if !isAudio && !isSub {
			continue
		}

		var lang string
		if isAudio && i+11 < len(data) {
			lang = strings.ToLower(string(data[i+9 : i+12]))
		} else if isSub && i+11 < len(data) {
			lang = strings.ToLower(string(data[i+8 : i+11]))
		}
		if !isValidLang(lang) {
			continue
		}

		if isAudio {
			result[fmt.Sprintf("audio_%d", audioCount)] = lang
			audioCount++
		} else {
			result[fmt.Sprintf("subtitle_%d", subCount)] = lang
			subCount++
		}

		i += entryLen
	}
	return result
}

func isValidLang(s string) bool {
	if len(s) != 3 {
		return false
	}
	for _, c := range s {
		if c < 'a' || c > 'z' {
			return false
		}
	}
	return true
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

package db

import "database/sql"

type TranscodeConfig struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	EncoderType     string `json:"encoder_type"`
	VideoEncoder    string `json:"video_encoder"`
	VideoParams     string `json:"video_params"`
	AudioTracks     string `json:"audio_tracks"`
	SubtitleTracks  string `json:"subtitle_tracks"`
	ChaptersEnabled bool   `json:"chapters_enabled"`
	OutputMuxer     string `json:"output_muxer"`
	IsBuiltin       bool   `json:"is_builtin"`
	CreatedAt       string `json:"created_at"`
}

func CreateConfig(db *sql.DB, c *TranscodeConfig) (int64, error) {
	res, err := db.Exec(`INSERT INTO transcode_configs (name,encoder_type,video_encoder,video_params,audio_tracks,subtitle_tracks,chapters_enabled,output_muxer) VALUES (?,?,?,?,?,?,?,?)`,
		c.Name, c.EncoderType, c.VideoEncoder, c.VideoParams, c.AudioTracks, c.SubtitleTracks, c.ChaptersEnabled, c.OutputMuxer)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func ListConfigs(db *sql.DB) ([]TranscodeConfig, error) {
	rows, err := db.Query(`SELECT id,name,encoder_type,video_encoder,video_params,audio_tracks,subtitle_tracks,chapters_enabled,output_muxer,is_builtin,created_at FROM transcode_configs ORDER BY is_builtin DESC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cfgs := make([]TranscodeConfig, 0)
	for rows.Next() {
		var c TranscodeConfig
		rows.Scan(&c.ID, &c.Name, &c.EncoderType, &c.VideoEncoder, &c.VideoParams, &c.AudioTracks, &c.SubtitleTracks, &c.ChaptersEnabled, &c.OutputMuxer, &c.IsBuiltin, &c.CreatedAt)
		cfgs = append(cfgs, c)
	}
	return cfgs, nil
}

func GetConfig(db *sql.DB, id int64) (*TranscodeConfig, error) {
	c := &TranscodeConfig{}
	err := db.QueryRow(`SELECT id,name,encoder_type,video_encoder,video_params,audio_tracks,subtitle_tracks,chapters_enabled,output_muxer,is_builtin,created_at FROM transcode_configs WHERE id=?`, id).
		Scan(&c.ID, &c.Name, &c.EncoderType, &c.VideoEncoder, &c.VideoParams, &c.AudioTracks, &c.SubtitleTracks, &c.ChaptersEnabled, &c.OutputMuxer, &c.IsBuiltin, &c.CreatedAt)
	return c, err
}

func UpdateConfig(db *sql.DB, id int64, updates map[string]any) error {
	setClauses := ""
	args := []any{}
	for k, v := range updates {
		if setClauses != "" {
			setClauses += ", "
		}
		setClauses += k + " = ?"
		args = append(args, v)
	}
	args = append(args, id)
	_, err := db.Exec("UPDATE transcode_configs SET "+setClauses+" WHERE id=?", args...)
	return err
}

func DeleteConfig(db *sql.DB, id int64) error {
	_, err := db.Exec("DELETE FROM transcode_configs WHERE id=? AND is_builtin=0", id)
	return err
}

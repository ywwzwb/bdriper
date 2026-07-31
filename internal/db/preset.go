package db

type PresetTemplate struct {
	Name        string         `json:"name"`
	Encoder     string         `json:"encoder"`
	Mode        string         `json:"mode"`
	Params      map[string]any `json:"params"`
	Description string         `json:"description"`
	Builtin     bool           `json:"builtin"`
}

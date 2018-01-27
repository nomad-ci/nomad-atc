package nomadatc

type TaskConfig struct {
	Host    string   `json:"host"`
	Stream  int64    `json:"stream"`
	Path    string   `json:"path"`
	Args    []string `json:"args"`
	Outputs []string `json:"outputs"`
}

package config

type WindowSize struct {
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
}

type TaskConfig struct {
	Host           string      `json:"host"`
	Stream         int64       `json:"stream"`
	Path           string      `json:"path"`
	Args           []string    `json:"args"`
	Outputs        []string    `json:"outputs"`
	Inputs         []string    `json:"inputs"`
	Dir            string      `json:"dir"`
	Env            []string    `json:"env"`
	WaitForVolumes bool        `json:"wait_for_volumes"`
	UseTTY         bool        `json:"use_tty"`
	WindowSize     *WindowSize `json:"window_size"`
}

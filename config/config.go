package config

// WindowSize holds the initial window size of the terminal to use
type WindowSize struct {
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
}

// Volume indicates the path in the container represents a specific
// volume, identified by the handle.
type Volume struct {
	Path   string `json:"path"`
	Handle string `json:"handle"`
}

// TaskConfig is used by the driver to know what actions to take.
type TaskConfig struct {
	Host           string      `json:"host"`
	Stream         int64       `json:"stream"`
	Path           string      `json:"path"`
	Args           []string    `json:"args"`
	Outputs        []*Volume   `json:"outputs"`
	Inputs         []*Volume   `json:"inputs"`
	Dir            string      `json:"dir"`
	Env            []string    `json:"env"`
	WaitForVolumes bool        `json:"wait_for_volumes"`
	UseTTY         bool        `json:"use_tty"`
	WindowSize     *WindowSize `json:"window_size"`
}

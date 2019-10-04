package gourmet

type Config struct {
	InterfaceType   string `yaml:"type"`
	Interface       string
	Promiscuous     bool	 `yaml:",omitempty"`
	ConnTimeout     int    `yaml:"connection_timeout"`
	SnapLen         int    `yaml:"snapshot_length"`
	Bpf             string `yaml:",omitempty"`
	LogFile         string `yaml:"log_file"`
	UpdateAnalyzers bool `yaml:"update_analyzers"`
	Analyzers       map[string]interface{}
}

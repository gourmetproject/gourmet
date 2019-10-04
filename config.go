package gourmet

import (
	"github.com/ghodss/yaml"
	"errors"
	"fmt"
)

type Config struct {
	InterfaceType   string `json:"type"`
	Interface       string
	Promiscuous     bool
	ConnTimeout     int    `json:"connection_timeout"`
	SnapLen         int    `json:"snapshot_length"`
	Bpf             string
	LogFile         string `json:"log_file"`
	SkipUpdate      bool   `json:"skip_update"`
	Analyzers       map[string]interface{}
}

var (
	analyzerConfigs map[string]interface{}
)

func InitAnalyzerConfigs() {
	analyzerConfigs = make(map[string]interface{})
}

func GetAnalyzerConfig(key string) ([]byte, error) {
	val, ok := analyzerConfigs[key]; if !ok {
		return nil, errors.New(fmt.Sprintf("analyzer %s does not exist", key))
	}
	return yaml.Marshal(val)
}

func SetAnalyzerConfig(key string, config interface{}) {
	analyzerConfigs[key] = config
}
package gourmet

import (
	"errors"
	"fmt"
)

type Config struct {
	InterfaceType   string `yaml:"type"`
	Interface       string
	Promiscuous     bool
	ConnTimeout     int    `yaml:"connection_timeout"`
	SnapLen         int    `yaml:"snapshot_length"`
	Bpf             string
	LogFile         string `yaml:"log_file"`
	UpdateAnalyzers bool   `yaml:"update_analyzers"`
	Analyzers       map[string]interface{}
}

var (
	analyzerConfigs map[string]interface{}
)

func GetAnalyzerConfig(key string) (interface{}, error) {
	val, ok := analyzerConfigs[key]; if !ok {
		return nil, errors.New(fmt.Sprintf("analyzer %s does not exist", key))
	}
	return val, nil
}

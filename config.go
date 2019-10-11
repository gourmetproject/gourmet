package gourmet

import (
	"fmt"

	"github.com/ghodss/yaml"
)

// Config is the data structure used to expose Gourmet configuration settings to the user. Each of
// these fields have a default value, except for InterfaceType. For a list of default values and
// which values are allowed for each field, consult the web documentation at docs.gourmetproject.io
type Config struct {
	InterfaceType string `json:"type"`
	Interface     string
	Promiscuous   bool
	MaxCores      int `json:"max_cores"`
	ConnTimeout   int `json:"connection_timeout"`
	SnapLen       int `json:"snapshot_length"`
	Bpf           string
	LogFile       string `json:"log_file"`
	SkipUpdate    bool   `json:"skip_update"`
	Analyzers     map[string]interface{}
}

var (
	analyzerConfigs = make(map[string]interface{})
)

// GetAnalyzerConfig does a map lookup based on the analyzer's name. If the analyzer exists, then
// the configuration is returned as marshaled YAML bytes. It is the job of the analyzer to unmarshal
// these bytes back into the desired data structure for analyzer configuration.
func getAnalyzerConfig(key string) ([]byte, error) {
	val, ok := analyzerConfigs[key]
	if !ok {
		return nil, fmt.Errorf("analyzer %s does not exist", key)
	}
	return yaml.Marshal(val)
}

// SetAnalyzerConfig saves an analyzer configuration in the global Gourmet map. The config parameter
// is an arbitrary interface because Gourmet does not know each analyzer's config, and it is up to
// the analyzer to unmarshal the bytes returned by GetAnalyzerConfig and perform input validation.
// This design will most definitely change in the future.
func setAnalyzerConfig(key string, config interface{}) {
	analyzerConfigs[key] = config
}

package gourmet

import (
	"encoding/json"
	"os"
	"sync"
)

type Logger struct {
	encoder *json.Encoder
	mutex   sync.Mutex
}

func newLogger(logName string) (*Logger, error) {
	f, err := os.Create(logName)
	if err != nil {
		return nil, err
	}
	encoder := json.NewEncoder(f)
	return &Logger{
		encoder: encoder,
	}, nil
}


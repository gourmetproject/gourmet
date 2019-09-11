package gourmet

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

var (
	logger *Logger
)

type Logger struct {
	encoder *json.Encoder
	mutex   sync.Mutex
	connections chan *Connection
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

func (l *Logger) Log(c *Connection) {
	l.mutex.Lock()
	err := l.encoder.Encode(c)
	if err != nil {
		log.Println(err)
	}
	l.mutex.Unlock()
}


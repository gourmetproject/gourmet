package gourmet

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var (
	logger *Logger
)

type Logger struct {
	fileName string
	mutex   sync.Mutex
}

type LogFile struct {
	*sensorMetadata
	Connections []Connection
}

func newLogger(logName string, interfaceName string) (*Logger, error) {
	f, err := os.Create(logName)
	if err != nil {
		return nil, err
	}
	logFile := &LogFile{
		sensorMetadata: getSensorMetadata(interfaceName),
	}
	initJson, err := json.MarshalIndent(logFile, "", "  ")
	f.Write(initJson)
	return &Logger {
		fileName: logName,
	}, nil
}

func (l *Logger) Log(c Connection) {
	l.mutex.Lock()
	contents, err := ioutil.ReadFile(l.fileName)
	if err != nil {
		log.Println(err)
	}
	var logfile LogFile
	err = json.Unmarshal(contents, &logfile)
	if err != nil {
		log.Println(err)
	}
	logfile.Connections = append(logfile.Connections, c)
	newContents, err := json.MarshalIndent(logfile, "", "  ")
	if err != nil {
		log.Println(err)
	}
	err = ioutil.WriteFile(l.fileName, newContents, 0644)
	if err != nil {
		log.Println(err)
	}
	l.mutex.Unlock()
}


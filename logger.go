package gourmet

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	logger *Logger
)

type Logger struct {
	fileName string
	mutex    sync.Mutex
}

type LogFile struct {
	*SensorMetadata
	Connections []Connection
}

func newLogger(logName string, interfaceName string) (*Logger, error) {
	f, err := os.Create(logName)
	if err != nil {
		fmt.Println("meow")
		return nil, err
	}
	logFile := &LogFile{
		SensorMetadata: getSensorMetadata(interfaceName),
	}
	initJson, err := json.MarshalIndent(logFile, "", "  ")
	if err != nil {
		return nil, err
	}
	_, err = f.Write(initJson)
	if err != nil {
		return nil, err
	}
	return &Logger{
		fileName: logName,
	}, nil
}

func (l *Logger) Log(c Connection) {
	l.mutex.Lock()
	var logfile LogFile

	// Get the mode of the existing file so we don't modify it
	s, err := os.Stat(l.fileName)
	if err != nil {
		log.Println(err)
	}

	// Open the file in append mode only
	f, err := os.OpenFile(l.fileName, os.O_APPEND|os.O_WRONLY, s.Mode)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	logfile.Connections = append(logfile.Connections, c)
	newContents, err := json.MarshalIndent(logfile, "", "  ")
	if err != nil {
		log.Println(err)
	}
	_, err = f.Write(newContents)
	if err != nil {
		log.Println(err)
	}
	l.mutex.Unlock()
}

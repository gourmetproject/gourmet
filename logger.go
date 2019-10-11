package gourmet

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var (
	gLogger *logger
)

type logger struct {
	fileName string
	mutex    sync.Mutex
}

type logFile struct {
	SensorMetadata *sensorMetadata
	Connections    []Connection
}

func initLogger(logName string, interfaceName string) error {
	f, err := os.Create(logName)
	if err != nil {
		return err
	}
	logFile := &logFile{
		SensorMetadata: getSensorMetadata(interfaceName),
	}
	initJSON, err := json.MarshalIndent(logFile, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(initJSON)
	if err != nil {
		return err
	}
	gLogger = &logger{
		fileName: logName,
	}
	return nil
}

func (l *logger) log(c Connection) {
	l.mutex.Lock()
	contents, err := ioutil.ReadFile(l.fileName)
	if err != nil {
		log.Println(err)
	}
	var logfile logFile
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

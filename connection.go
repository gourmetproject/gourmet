package gourmet

import (
	"log"
	"time"
)

var (
	logger *Logger
)

type Connection struct {
	Timestamp       time.Time
	UID             uint64
	SourceIP        string
	SourcePort      string
	DestinationIP   string
	DestinationPort string
	Duration        time.Duration
	State          	string
}

func (cl *Logger) Log(c *Connection) {
	cl.mutex.Lock()
	err := cl.encoder.Encode(c)
	if err != nil {
		log.Println(err)
	}
	cl.mutex.Unlock()
}

func processTcpStream(ts *TcpStream) {
	c := &Connection{
		Timestamp: ts.startTime,
		UID: ts.net.FastHash() + ts.transport.FastHash(),
		SourceIP: ts.net.Src().String(),
		SourcePort: ts.transport.Src().String(),
		DestinationIP: ts.net.Dst().String(),
		DestinationPort: ts.transport.Dst().String(),
		Duration: ts.duration,
		State: ts.tcpState.String(),
	}
	logger.Log(c)
}

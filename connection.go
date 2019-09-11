package gourmet

import (
	"bytes"
	"time"
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
	payload         *bytes.Buffer
	Analyzers       map[string]interface{}
}

func newTcpConnection(ts *TcpStream) (c *Connection) {
	return &Connection{
		Timestamp: ts.startTime,
		UID: ts.net.FastHash() + ts.transport.FastHash(),
		SourceIP: ts.net.Src().String(),
		SourcePort: ts.transport.Src().String(),
		DestinationIP: ts.net.Dst().String(),
		DestinationPort: ts.transport.Dst().String(),
		Duration: ts.duration,
		State: ts.tcpState.String(),
		payload: ts.payload,
		Analyzers: make(map[string]interface{}),
	}
}

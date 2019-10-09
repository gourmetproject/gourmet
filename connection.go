package gourmet

import (
	"bytes"
	"time"
)

type Connection struct {
	Timestamp       time.Time
	UID             uint64
	SourceIP        string
	SourcePort      int
	DestinationIP   string
	DestinationPort int
	TransportType   string
	Duration        float64
	State          	string        `json:",omitempty"`
	Payload         *bytes.Buffer `json:"-"`
	Analyzers       map[string]interface{}
}

func (c *Connection) analyze() error{
	for _, analyzer := range analyzers {
		if analyzer.Filter(c) {
			result, err := analyzer.Analyze(c)
			if err != nil {
				return err
			}
			c.Analyzers[result.Key()] = result
		}
	}
	return nil
}

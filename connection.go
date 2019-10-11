package gourmet

import (
	"bytes"
	"time"
)

// Connection contains basic information about an IP connection, including application layer bytes.
// If the connection is TCP-based, then the Connection contains basic information about the reassembled
// stream of packets for that TCP session.
//
// A Connection is given to each Analyzer. The Result returned from an Analyzer is added to the
// Analyzers map for that Connection object. Once all Analyzers have been run against the Connection,
// it is marshaled as a JSON object into raw bytes and written to the log file.
type Connection struct {
	Timestamp       time.Time
	UID             uint64
	SourceIP        string
	SourcePort      int
	DestinationIP   string
	DestinationPort int
	TransportType   string
	Duration        float64
	State           string        `json:",omitempty"`
	Payload         *bytes.Buffer `json:"-"`
	Analyzers       map[string]interface{}
}

func (c *Connection) analyze() error {
	for _, analyzer := range registeredAnalyzers {
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

package gourmet

import (
	"bufio"
	"net/http"
)

type Http struct {
	Method string
	URL     string
	Headers map[string][]string
}

func (h *Http) Key() string {
	return "http"
}

type httpAnalyzer struct{}

func (ha *httpAnalyzer) Filter(c *Connection) bool {
	if c.SourcePort == "80" || c.DestinationPort == "80" {
		return true
	}
	return false
}

func (ha *httpAnalyzer) Analyze(c *Connection) (Result, error) {
	h := &Http{}
	r := bufio.NewReader(c.payload)
	req, err := http.ReadRequest(r)
	if err != nil {
		return nil, err
	}
	h.Method = req.Method
	h.URL = req.URL.String()
	h.Headers = req.Header
	return h, nil
}

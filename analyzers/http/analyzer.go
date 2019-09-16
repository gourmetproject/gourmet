package http

import (
	"bufio"
	"github.com/kvasirlabs/gourmet"
	"net/http"
)

type Request struct {
	Method                  string
	URL                     string
	Headers          http.Header
	ContentLength    int64
	TransferEncoding        []string
	Host                    string
}

type Response struct {
	Status                string
	Headers       http.Header
	ContentLength int64
}

type Http map[string]interface{}

func (h *Http) Key() string {
	return "http"
}

type httpAnalyzer struct{}

func NewHttpAnalyzer() *httpAnalyzer {
	return &httpAnalyzer{}
}

func (ha *httpAnalyzer) Filter(c *gourmet.Connection) bool {
	if c.SourcePort == 80 || c.DestinationPort == 80 {
		return true
	}
	return false
}

func (ha *httpAnalyzer) Analyze(c *gourmet.Connection) (gourmet.Result, error) {
	h := Http{}
	r := bufio.NewReader(c.Payload)
	req, err := http.ReadRequest(r)
	if err != nil {
		return nil, err
	}
	h["Request"] = Request{
		Method:               req.Method,
		URL:                  req.URL.String(),
		Headers:       req.Header,
		ContentLength: req.ContentLength,
		TransferEncoding:     req.TransferEncoding,
		Host:                 req.Host,
	}
	resp, err := http.ReadResponse(r, req)
	if err != nil {
		return nil, err
	}
	h["Response"] = Response{
		Status:                resp.Status,
		Headers:       resp.Header,
		ContentLength: resp.ContentLength,
	}
	return &h, nil
}

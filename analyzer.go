package gourmet

type Result interface{
	Key() string
}

type Analyzer interface {
	Filter(c *Connection) bool
	Analyze(c *Connection) (Result, error)
}

package gourmet

type Result interface{
	Key() string
}

type Analyzer interface {
	Filter(c *Connection) bool
	Analyze(c *Connection) (Result, error)
}

type analyzer struct {
	inUse bool
	Analyzer
}

type RegisteredAnalyzer int

const maxAnalyzers = 1000
var analyzers [maxAnalyzers]analyzer

func RegisterAnalyzer(num int, analyzer Analyzer) RegisteredAnalyzer {
	if num >= maxAnalyzers {
		panic("Analyzer value is too high")
	}
	if analyzers[num].inUse {
		panic("Analyzer type already exists.")
	}
	return OverrideAnalyzer(num, analyzer)
}

func OverrideAnalyzer(num int, a Analyzer) RegisteredAnalyzer {
	analyzers[num] = analyzer{
		inUse: true,
		Analyzer: a,
	}
	return RegisteredAnalyzer(num)
}

var (
	HttpAnalyzer = RegisterAnalyzer(1, &httpAnalyzer{})
)

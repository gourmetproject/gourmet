package gourmet

type RegisteredAnalyzer string

type Result interface{
	Key() string
}

type Analyzer interface {
	Filter(c *Connection) bool
	Analyze(c *Connection) (Result, error)
}

var registeredAnalyzers = make(map[string]Analyzer)

func RegisterAnalyzer(name string, a Analyzer) {
	if _, ok := registeredAnalyzers[name]; ok {
		panic("analyzer type already exists")
	}
	registeredAnalyzers[name] = a
}

func GetRegisteredAnalyzer(name string) Analyzer {
	return registeredAnalyzers[name]
}

func GetRegisteredAnalyzers() []Analyzer {
	ra := make([]Analyzer, len(registeredAnalyzers))
	i := 0
	for k := range registeredAnalyzers {
		ra[i] = registeredAnalyzers[k]
		i++
	}
	return ra
}

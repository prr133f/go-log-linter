package analyzer

import "golang.org/x/tools/go/analysis"

func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "loglinter",
		Doc:  "loglinter checks for common logging issues",
		Run:  run,
	}
}

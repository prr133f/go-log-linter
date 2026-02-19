package analyzer

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "loglinter",
		Doc:      "loglinter checks for common logging issues",
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer}, // Ускорение для golangci-lint
	}
}

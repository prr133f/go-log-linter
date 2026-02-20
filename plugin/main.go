package plugin

import (
	analyzer "github.com/prr133f/go-log-linter/analyzers/log-linter"
	"golang.org/x/tools/go/analysis"
)

func New(settings any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{analyzer.New()}, nil
}

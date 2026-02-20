package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	analyzer "github.com/prr133f/go-log-linter/analyzers/log-linter"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("loglinter", New)
}

func New(settings any) (register.LinterPlugin, error) {
	return LogLinterPlugin{}, nil
}

type LogLinterPlugin struct{}

func (p LogLinterPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{analyzer.New()}, nil
}

func (p LogLinterPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

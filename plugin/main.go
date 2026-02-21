package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	analyzer "github.com/prr133f/go-log-linter/analyzers/log-linter"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("loglinter", New)
}

type Settings struct {
	SensitivePatterns []string `json:"sensitivePatterns"`
}

func New(settings any) (register.LinterPlugin, error) {
	var s Settings
	if settings != nil {
		var err error
		s, err = register.DecodeSettings[Settings](settings)
		if err != nil {
			return nil, err
		}
	}

	return LogLinterPlugin{settings: s}, nil
}

type LogLinterPlugin struct {
	settings Settings
}

func (p LogLinterPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		analyzer.New(analyzer.Config{
			SensitivePatterns: p.settings.SensitivePatterns,
		}),
	}, nil
}

func (p LogLinterPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

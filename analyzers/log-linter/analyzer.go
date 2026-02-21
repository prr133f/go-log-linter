package analyzer

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// Config содержит настройки линтера.
type Config struct {
	// SensitivePatterns — подстроки имён переменных,
	// указывающие на потенциально чувствительные данные.
	SensitivePatterns []string
}

// defaultSensitivePatterns — паттерны по умолчанию.
var defaultSensitivePatterns = []string{
	"token",
	"password",
	"passwd",
	"secret",
	"apikey",
	"credential",
	"auth",
	"private",
}

func New(cfgs ...Config) *analysis.Analyzer {
	cfg := Config{
		SensitivePatterns: defaultSensitivePatterns,
	}
	if len(cfgs) > 0 && len(cfgs[0].SensitivePatterns) > 0 {
		cfg.SensitivePatterns = cfgs[0].SensitivePatterns
	}

	return &analysis.Analyzer{
		Name:     "loglinter",
		Doc:      "loglinter checks for common logging issues",
		Run:      makeRun(cfg),
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

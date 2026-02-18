package main

import (
	analyzer "github.com/prr133f/go-log-linter/analyzers/log-linter"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.New())
}

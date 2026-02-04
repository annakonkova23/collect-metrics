package main

import (
	"github.com/annakonkova23/collect-metrics/cmd/linter/analyscheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyscheck.CheckAnalyzer)
}

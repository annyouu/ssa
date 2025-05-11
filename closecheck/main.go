package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"
	"ssa/analyzer"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
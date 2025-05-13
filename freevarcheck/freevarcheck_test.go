package freevarcheck_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
	"ssa/freevarcheck"
)

func Test(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), freevarcheck.Analyzer, "a")
}
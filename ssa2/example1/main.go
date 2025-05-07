package main

import (
	"fmt"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var ssaAnalyzer = &analysis.Analyzer{
	Name: "ssainspect",
	Doc: "Print SSA from of go functions",
	Run: run,
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	s := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	for _, f := range s.SrcFuncs {
		fmt.Println("function:", f.Name())
		for _, b := range f.Blocks {
			fmt.Printf("Block %d\n", b.Index)
			for _, instr := range b.Instrs {
				fmt.Printf("\t\t%[1]T\t%[1]v(%[1]p)\n", instr)
				for _, v := range instr.Operands(nil) {
					if v != nil {
						fmt.Printf("\t\t\tOperand: %[1]T\t%[1]v(%[1]p)\n", *v)
					}
				}
			}
		}
	}
	return nil, nil
}

func main() {
	singlechecker.Main(ssaAnalyzer)
}
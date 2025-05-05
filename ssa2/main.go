package main

import (
	"go/token"
	"go/types"
	"log"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

var Analyzer = &analysis.Analyzer{
	Name: "nilerrcheck",
	Doc:  "if err != nil' blocksでnilを返しているエラーを検知する",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// packages.Loadを使って[]*packages.Packageを取得する
	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
		Fset: pass.Fset,
	}

	pkgs, err := packages.Load(cfg, pass.Pkg.Path())
	if err != nil {
		log.Fatal(err)
	}

	// SSAを構築する
	prog, ssaPkgs := ssautil.AllPackages(pkgs, 0)
	prog.Build()

	for _, ssaPkg := range ssaPkgs {
		if !ssaPkg.Pkg.Complete() {
			continue
		}

		for _, mem := range ssaPkg.Members {
			fn, ok := mem.(*ssa.Function)
			if !ok || fn.Blocks == nil {
				continue
			}

			for _, block := range fn.Blocks {
				if len(block.Instrs) == 0 {
					continue
				}

				// 最後の命令がifか
				ifInstr, ok := block.Instrs[len(block.Instrs)-1].(*ssa.If)
				if !ok || len(block.Succs) < 1 {
					continue
				}

				// 条件が "err != nil"か
				binOp, ok := ifInstr.Cond.(*ssa.BinOp)
				if !ok || binOp.Op != token.NEQ {
					continue
				}

				if !isErrorType(binOp.X.Type()) && !isErrorType(binOp.Y.Type()) {
					continue
				}

				// true分岐に return nil があるか
				trueBranch := block.Succs[0]
				for _, instr := range trueBranch.Instrs {
					ret, ok := instr.(*ssa.Return)
					if !ok || len(ret.Results) != 1 {
						continue
					}
					if c, ok := ret.Results[0].(*ssa.Const); ok && c.IsNil() {
						pass.Reportf(ret.Pos(), "err != nil のブロックで nil を返しています（バグの可能性）")
					}
				}
			}
		}
	}
	return nil, nil
}

// error型判定
func isErrorType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	return named.Obj().Name() == "error"
}

func main() {
	singlechecker.Main(Analyzer)
}
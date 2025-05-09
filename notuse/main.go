package main

import (
    "fmt"
    "go/token"
    "golang.org/x/tools/go/packages"
    "golang.org/x/tools/go/ssa"
    "golang.org/x/tools/go/ssa/ssautil"
)

func reportUnused(p *ssa.Parameter, fn *ssa.Function) {
	pos := fn.Prog.Fset.Position(p.Pos())
	fmt.Printf("使われていない変数を検出 %s in %s at %s\n", p.Name(), fn.String(), pos)
}


func main() {
	// ソースをローディングする
	cfg := &packages.Config{
		Mode: packages.LoadSyntax,
		Fset: token.NewFileSet(),
	}

	initial, err := packages.Load(cfg, "./target")
	if err != nil {
		panic(err)
	}

	// SSAプログラムとSSAパッケージの取得する
	prog, ssaPkgs := ssautil.AllPackages(initial, ssa.SanityCheckFunctions)
	prog.Build()

	// 各関数を走査する
	for _, ssaPkg := range ssaPkgs {
		// nilポインタを回避するため、nilならスキップする。
		if ssaPkg == nil {
			continue
		}
		
		for _, mem := range ssaPkg.Members {
			fn, ok := mem.(*ssa.Function)
			if !ok || fn.Blocks == nil {
				continue
			}

			// 各パラメータをチェックする
			for _, param := range fn.Params {
				refs := param.Referrers()
				if refs == nil {
					reportUnused(param, fn)
					continue
				}

				used := false
				for _, instr := range *refs {
					switch store := instr.(type) {
					case *ssa.Store:
						if store.Val == param {
							continue
						}
					default:
						used = true
					}
					if used {
						break
					}
				}
				
				if !used {
					reportUnused(param, fn)
				}
			}
		}
	}
}

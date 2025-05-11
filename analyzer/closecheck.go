package analyzer

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

var Analyzer = &analysis.Analyzer{
	Name: "closecheck",
	Doc: "同じチャネルを2回closeしている場所を探す",
	Run: run,
	Requires: []*analysis.Analyzer{buildssa.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	// SSA情報の取得する
	ssaResult := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	prog := ssaResult.Pkg.Prog

	// チャネルごとのclose命令を記録するマップ
	closeCalls := make(map[ssa.Value][]ssa.Instruction)

	// 全パッケージ内の全関数を調査する
	for _, pkg := range prog.AllPackages() {
		for _, member := range pkg.Members {
			fn, ok := member.(*ssa.Function)
			if !ok || fn.Blocks == nil {
				continue
			}

			for _, block := range fn.Blocks {
				for _, instr := range block.Instrs {
					callInst, ok := instr.(ssa.CallInstruction)
					if !ok {
						continue
					}

					common := callInst.Common()
					if common == nil {
						continue
					}

					// 組み込み関数closeのみ対象にする
					builtin, ok := common.Value.(*ssa.Builtin)
					if !ok || builtin.Name() != "close" {
						continue
					}

					if len(common.Args) != 1 {
						continue
					}

					ch := common.Args[0]
					closeCalls[ch] = append(closeCalls[ch], instr)
				}
			}
		}
	}

	// 同じチャネルを複数回closeしている箇所を報告する
	for ch, calls := range closeCalls {
		if len(calls) > 1 {
			first := calls[0]
			firstPos := pass.Fset.Position(first.Pos())
			
			for _, instr := range calls[1:] {
				pos := pass.Fset.Position(instr.Pos())
				pass.Reportf(instr.Pos(),
					"チャネル %s は %s で一度 close されていますが、%s で再度 close されています",
					ch.Name(), firstPos, pos)
			}
		}
	}	
	return nil, nil
}

package freevarcheck

import (
	"go/ast"
    "golang.org/x/tools/go/analysis"
    "golang.org/x/tools/go/analysis/passes/inspect"
    "golang.org/x/tools/go/ast/inspector"
)

// クロージャ内で自由変数変更を検出するAnalyzerである
var Analyzer = &analysis.Analyzer{
	Name: "freevarcheck",
	Doc: "クロージャ内で自由変数変更を検出する",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// ASTインスペクタの取得
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// クロージャのみを訪問
	nodeFilter := []ast.Node{(*ast.FuncLit)(nil)}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		lit := n.(*ast.FuncLit)

		// クロージャ内部のスコープ
		litScope := pass.TypesInfo.Scopes[lit.Body]

		// クロージャ本体をスキャンして代入文を探す
		ast.Inspect(lit.Body, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}

			for _, lhs := range assign.Lhs {
				id, ok := lhs.(*ast.Ident)
				if !ok || id.Name == "_" {
					continue
				}
				// クロージャ内で新規宣言された変数（:=）をスキップ
				if pass.TypesInfo.Defs[id] != nil {
					continue
				}

				obj := pass.TypesInfo.ObjectOf(id)
				if obj == nil {
					continue
				}
				// 変数がクロージャのスコープで定義されたものかチェック
				if litScope != nil && obj.Parent() == litScope {
					continue
				}

				// 上記以外は自由変数への書き込みとみなす
				pass.Reportf(id.Pos(), "自由変数への書き込みです %q", id.Name)
			}
			return true
		})
	})
	return nil, nil 
}
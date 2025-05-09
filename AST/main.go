package main

import (
	"go/parser"
	"go/token"
	"go/ast"
	"fmt"
)

func main() {
	src := `
		package main
		func f(n int) {
			n = 20
			println(n)
		}
	`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		panic(err)
	}

	// AST走査
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			if ident.Name == "n" {
				fmt.Printf("nを見つけました: %s\n", fset.Position(ident.Pos()))
			}
		}
		return true
	})
}
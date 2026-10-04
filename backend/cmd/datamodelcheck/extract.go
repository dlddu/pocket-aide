package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

var queryMethods = map[string]bool{"QueryContext": true, "QueryRowContext": true, "ExecContext": true, "PrepareContext": true}

type goFile struct {
	fset  *token.FileSet
	file  *ast.File
	calls map[int][]*ast.CallExpr
}

func parseGoFile(src []byte) *goFile {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return nil
	}
	g := &goFile{fset: fset, file: f, calls: map[int][]*ast.CallExpr{}}
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if sel, ok := c.Fun.(*ast.SelectorExpr); ok && queryMethods[sel.Sel.Name] {
				line := fset.Position(sel.Sel.Pos()).Line
				g.calls[line] = append(g.calls[line], c)
			}
		}
		return true
	})
	return g
}

func (g *goFile) sqlAt(line int) ([]string, string) {
	if g == nil {
		return nil, "Go 파일을 파싱하지 못했다"
	}
	calls := g.calls[line]
	if len(calls) == 0 {
		return nil, "이 줄에 database/sql 쿼리 호출식이 없다"
	}
	var out []string
	for _, c := range calls {
		sql, reason := g.callSQL(c)
		if reason != "" {
			return nil, reason
		}
		out = append(out, sql)
	}
	return out, ""
}

func (g *goFile) callSQL(c *ast.CallExpr) (string, string) {
	sel := c.Fun.(*ast.SelectorExpr)
	if id, ok := sel.X.(*ast.Ident); ok && sel.Sel.Name != "PrepareContext" {
		if prep := preparedBy(id); prep != nil {
			return g.callSQL(prep)
		}
	}
	if len(c.Args) < 2 {
		return "", "SQL 인자가 없다"
	}
	return resolveString(c.Args[1])
}

func preparedBy(id *ast.Ident) *ast.CallExpr {
	if id.Obj == nil || id.Obj.Kind != ast.Var {
		return nil
	}
	as, ok := id.Obj.Decl.(*ast.AssignStmt)
	if !ok || len(as.Rhs) != 1 {
		return nil
	}
	c, ok := as.Rhs[0].(*ast.CallExpr)
	if !ok {
		return nil
	}
	if sel, ok := c.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "PrepareContext" {
		return c
	}
	return nil
}

func resolveString(e ast.Expr) (string, string) {
	switch x := e.(type) {
	case *ast.BasicLit:
		if x.Kind != token.STRING {
			return "", "SQL 인자가 문자열 리터럴이 아니다"
		}
		s, err := strconv.Unquote(x.Value)
		if err != nil {
			return "", "문자열 리터럴을 해석하지 못했다"
		}
		return s, ""
	case *ast.ParenExpr:
		return resolveString(x.X)
	case *ast.BinaryExpr:
		if x.Op != token.ADD {
			return "", "SQL 인자가 문자열 연결이 아닌 식이다"
		}
		l, reason := resolveString(x.X)
		if reason != "" {
			return "", reason
		}
		r, reason := resolveString(x.Y)
		if reason != "" {
			return "", reason
		}
		return l + r, ""
	case *ast.Ident:
		if x.Obj != nil && x.Obj.Kind == ast.Con {
			if vs, ok := x.Obj.Decl.(*ast.ValueSpec); ok {
				for i, n := range vs.Names {
					if n.Name == x.Name && i < len(vs.Values) {
						return resolveString(vs.Values[i])
					}
				}
			}
		}
		return "", fmt.Sprintf("SQL 이 런타임 값 `%s` 에 달렸다", x.Name)
	case *ast.IndexExpr:
		return "", fmt.Sprintf("SQL 을 런타임에 `%s` 로 고른다", exprString(x))
	case *ast.CallExpr:
		return "", fmt.Sprintf("SQL 을 호출 `%s` 가 만든다", exprString(x.Fun))
	}
	return "", "SQL 인자를 정적으로 해석하지 못했다"
}

func exprString(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return exprString(x.X) + "." + x.Sel.Name
	case *ast.IndexExpr:
		return exprString(x.X) + "[" + exprString(x.Index) + "]"
	}
	return strings.TrimSpace(fmt.Sprintf("%T", e))
}

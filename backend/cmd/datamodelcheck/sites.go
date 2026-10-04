package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type site struct {
	File          string `json:"file"`
	Line          int    `json:"line"`
	Func          string `json:"func"`
	Shape         string `json:"shape,omitempty"`
	SQL           string `json:"sql,omitempty"`
	Unextractable string `json:"unextractable,omitempty"`

	shape shape
}

func (s site) String() string { return fmt.Sprintf("%s::%s (%d행)", s.File, s.Func, s.Line) }

func (s site) ID() string { return s.File + "::" + s.Func }

func listSites(root string, b *scopeBlock, sch *schema) ([]site, error) {
	args := append([]string{"-C", root, "ls-files", "--"}, b.Scope...)
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files 실패: %w", err)
	}
	var files []string
	for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if f != "" && !b.exclude.MatchString(f) {
			files = append(files, f)
		}
	}
	sort.Strings(files)

	var sites []site
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(root, f))
		if err != nil {
			return nil, err
		}
		if bytes.IndexByte(data, 0) >= 0 {
			continue
		}
		var funcs []funcSpan
		var gf *goFile
		lines := strings.Split(string(data), "\n")
		for i, l := range lines {
			if !b.site.MatchString(l) {
				continue
			}
			if funcs == nil && strings.HasSuffix(f, ".go") {
				funcs = goFuncs(data)
				gf = parseGoFile(data)
			}
			base := site{File: f, Line: i + 1, Func: enclosing(funcs, i+1)}
			if !strings.HasSuffix(f, ".go") {
				base.Unextractable = "Go 소스가 아니다"
				sites = append(sites, base)
				continue
			}
			sqls, reason := gf.sqlAt(i + 1)
			if reason != "" {
				base.Unextractable = reason
				sites = append(sites, base)
				continue
			}
			for _, q := range sqls {
				st := base
				st.SQL = normSpace(q)
				sh, err := extractShape(q, sch)
				if err != nil {
					st.Unextractable = "형태를 뽑지 못했다: " + err.Error()
				} else {
					st.shape, st.Shape = sh, sh.String()
				}
				sites = append(sites, st)
			}
		}
	}
	return sites, nil
}

type funcSpan struct {
	name       string
	start, end int
}

func goFuncs(src []byte) []funcSpan {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return []funcSpan{}
	}
	spans := []funcSpan{}
	for _, d := range file.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		name := fd.Name.Name
		if fd.Recv != nil && len(fd.Recv.List) > 0 {
			name = recvName(fd.Recv.List[0].Type) + "::" + name
		}
		spans = append(spans, funcSpan{name, fset.Position(fd.Pos()).Line, fset.Position(fd.End()).Line})
	}
	return spans
}

func recvName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.StarExpr:
		return recvName(t.X)
	case *ast.IndexExpr:
		return recvName(t.X)
	case *ast.IndexListExpr:
		return recvName(t.X)
	case *ast.Ident:
		return t.Name
	}
	return "?"
}

func enclosing(spans []funcSpan, line int) string {
	for _, s := range spans {
		if s.start <= line && line <= s.end {
			return s.name
		}
	}
	return "-"
}

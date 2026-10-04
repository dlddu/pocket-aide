package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
)

type invReport struct {
	Status     string      `json:"status"`
	Violations []violation `json:"violations"`
}

type schemaSummary struct {
	Tables  int `json:"tables"`
	Indexes int `json:"indexes_excluding_pk"`
	FKs     int `json:"foreign_keys"`
}

type report struct {
	Engine      string                `json:"engine"`
	ExitCode    int                   `json:"exit_code"`
	Block       *scopeBlock           `json:"block,omitempty"`
	Schema      *schemaSummary        `json:"schema,omitempty"`
	Invariants  map[string]*invReport `json:"invariants"`
	Sites       []site                `json:"sites"`
	Undecidable []string              `json:"undecidable"`
}

func driverVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, d := range bi.Deps {
			if d.Path == "modernc.org/sqlite" {
				return d.Path + " " + d.Version
			}
		}
	}
	return "modernc.org/sqlite"
}

func main() {
	asJSON := flag.Bool("json", false, "기계 판독형(JSON) 리포트")
	rootFlag := flag.String("root", "", "레포 루트(기본: git rev-parse --show-toplevel)")
	flag.Parse()

	root := *rootFlag
	if root == "" {
		out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "레포 루트를 찾지 못했다:", err)
			os.Exit(2)
		}
		root = strings.TrimSpace(string(out))
	}

	r := run(root)
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		_ = enc.Encode(r)
	} else {
		printHuman(os.Stdout, r)
	}
	os.Exit(r.ExitCode)
}

func run(root string) *report {
	r := &report{
		Engine:      "SQLite (" + driverVersion() + ")",
		Invariants:  map[string]*invReport{"1": {Status: "undecidable"}, "2": {Status: "undecidable"}, "3": {Status: "undecidable"}},
		Sites:       []site{},
		Undecidable: []string{},
	}
	finish := func() *report {
		for _, inv := range r.Invariants {
			if inv.Violations == nil {
				inv.Violations = []violation{}
			}
			sort.Slice(inv.Violations, func(i, j int) bool {
				a, b := inv.Violations[i], inv.Violations[j]
				if a.Kind != b.Kind {
					return a.Kind < b.Kind
				}
				return a.Subject < b.Subject
			})
			if inv.Status != "undecidable" {
				inv.Status = "ok"
				if len(inv.Violations) > 0 {
					inv.Status = "violated"
				}
			}
		}
		sort.Strings(r.Undecidable)
		switch {
		case len(r.Undecidable) > 0:
			r.ExitCode = 2
		case len(r.Invariants["1"].Violations)+len(r.Invariants["2"].Violations)+len(r.Invariants["3"].Violations) > 0:
			r.ExitCode = 1
		}
		return r
	}
	fail := func(format string, a ...any) *report {
		r.Undecidable = append(r.Undecidable, fmt.Sprintf(format, a...))
		return finish()
	}

	b, err := readBlock(filepath.Join(root, readmePath))
	if err != nil {
		return fail("C2 블록: %v", err)
	}
	r.Block = b
	if _, err := os.Stat(filepath.Join(root, b.Checker)); err != nil {
		return fail("C1 위치: 블록의 checker 경로 %s 가 없다", b.Checker)
	}

	s, err := observeSchema(root, b)
	if err != nil {
		return fail("C3 스키마 관측: %v", err)
	}
	r.Engine = s.Engine
	fks := 0
	for _, t := range s.Tables {
		fks += len(t.FKs)
	}
	r.Schema = &schemaSummary{Tables: len(s.Tables), Indexes: s.numIdx, FKs: fks}

	if v, err := checkERD(root, s); err != nil {
		r.Undecidable = append(r.Undecidable, "불변식 1: "+err.Error())
	} else {
		r.Invariants["1"].Status = ""
		r.Invariants["1"].Violations = v
	}

	sites, err := listSites(root, b)
	if err != nil {
		return fail("C4 지점 추출: %v", err)
	}
	r.Sites = sites

	if _, err := os.Stat(filepath.Join(root, qpPath)); err != nil {
		r.Invariants["2"].Status = ""
		r.Invariants["2"].Violations = append(r.Invariants["2"].Violations, violation{Invariant: "2", Kind: "catalog-missing", Subject: qpPath, Detail: fmt.Sprintf("카탈로그가 없다 — 지점 후보 %d개가 미등재", len(sites))})
		for _, st := range sites {
			r.Invariants["2"].Violations = append(r.Invariants["2"].Violations, violation{Invariant: "2", Kind: "query-unregistered", Subject: st.String()})
		}
		r.Invariants["3"].Status = ""
		r.Invariants["3"].Violations = append(r.Invariants["3"].Violations, violation{Invariant: "3", Kind: "catalog-missing", Subject: qpPath, Detail: fmt.Sprintf("지원 칸·미사용 인덱스 표가 없다 — 인덱스(PK 제외) %d개 미판정", s.numIdx)})
	} else {
		r.Undecidable = append(r.Undecidable, "불변식 2·3: 형태 추출(C4)과 플랜(C5) 판정이 아직 구현되지 않았다 — 판정 슬라이스 (2)·(3)이 체커를 확장한다")
	}
	return finish()
}

func printHuman(out io.Writer, r *report) {
	w := &strings.Builder{}
	fmt.Fprintf(w, "datamodelcheck — 엔진 %s\n", r.Engine)
	if r.Schema != nil {
		fmt.Fprintf(w, "스키마(빈 DB · 마이그레이션 전부 적용): 테이블 %d · 인덱스(PK 제외) %d · FK %d\n", r.Schema.Tables, r.Schema.Indexes, r.Schema.FKs)
	}
	fmt.Fprintf(w, "지점 후보(site): %d\n", len(r.Sites))
	names := map[string]string{"ok": "정합", "violated": "위반", "undecidable": "판정 불가"}
	for _, k := range []string{"1", "2", "3"} {
		inv := r.Invariants[k]
		fmt.Fprintf(w, "\n불변식 %s: %s", k, names[inv.Status])
		if len(inv.Violations) > 0 {
			fmt.Fprintf(w, " %d건", len(inv.Violations))
		}
		fmt.Fprintln(w)
		for _, v := range inv.Violations {
			fmt.Fprintf(w, "  - [%s] %s", v.Kind, v.Subject)
			if v.Detail != "" {
				fmt.Fprintf(w, " — %s", v.Detail)
			}
			fmt.Fprintln(w)
			if v.Expected != "" {
				for _, l := range strings.Split(strings.TrimRight(v.Expected, "\n"), "\n") {
					fmt.Fprintf(w, "      기대: %s\n", l)
				}
			}
		}
	}
	for _, u := range r.Undecidable {
		fmt.Fprintf(w, "\n판정 불가: %s\n", u)
	}
	fmt.Fprintf(w, "\n종료 코드 %d\n", r.ExitCode)
	_, _ = io.WriteString(out, w.String())
}

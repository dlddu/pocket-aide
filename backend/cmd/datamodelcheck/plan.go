package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	noSupport    = "지원 없음"
	fullScanCell = "풀스캔 허용"
)

type access struct {
	Table  string `json:"table"`
	Index  string `json:"index,omitempty"`
	Full   bool   `json:"full_scan"`
	Detail string `json:"detail"`
}

type sqlPlan struct {
	Pattern  string   `json:"pattern"`
	Site     string   `json:"site"`
	SQL      string   `json:"sql"`
	Support  string   `json:"support"`
	Accesses []access `json:"accesses"`
	Notes    []string `json:"notes"`
}

var (
	planRowRe   = regexp.MustCompile(`^(SEARCH|SCAN) (\S+)(?: AS (\S+))?(.*)$`)
	planIndexRe = regexp.MustCompile(`USING (?:COVERING )?INDEX (\S+)`)
	aliasRe     = regexp.MustCompile(`(?i)\b(?:FROM|JOIN|UPDATE|INTO)\s+"?([a-z_][a-z0-9_]*)"?(?:\s+(?:AS\s+)?"?([a-z_][a-z0-9_]*)"?)?`)
	fullScanRe  = regexp.MustCompile(`^` + fullScanCell + `\((F\d+)\): \S`)
	aliasStop   = map[string]bool{"where": true, "join": true, "left": true, "right": true, "inner": true, "outer": true, "cross": true, "natural": true, "on": true, "using": true, "set": true, "values": true, "select": true, "order": true, "group": true, "limit": true, "default": true, "as": true, "union": true, "having": true, "returning": true}
)

func placeholders(q string) int {
	n, quote := 0, byte(0)
	for i := 0; i < len(q); i++ {
		c := q[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '\'' || c == '"' || c == '`':
			quote = c
		case c == '?':
			n++
		}
	}
	return n
}

func aliases(q string, sch *schema) map[string]string {
	out := map[string]string{}
	for _, m := range aliasRe.FindAllStringSubmatch(q, -1) {
		t := strings.ToLower(m[1])
		if sch.byName[t] == nil {
			continue
		}
		out[t] = t
		if a := strings.ToLower(m[2]); a != "" && !aliasStop[a] {
			out[a] = t
		}
	}
	return out
}

func (s *schema) indexLabel(t *table, name string) (string, bool) {
	if name == t.pkIndex {
		return "PK(" + t.Name + ")", true
	}
	for _, ix := range t.Indexes {
		if ix.Name == name {
			return ix.Label, true
		}
	}
	return "", false
}

func explain(sch *schema, q string) ([]access, []string, error) {
	rows, err := sch.db.Query("EXPLAIN QUERY PLAN "+q, make([]any, placeholders(q))...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()
	names := aliases(q, sch)
	var acc []access
	var notes []string
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			return nil, nil, err
		}
		m := planRowRe.FindStringSubmatch(detail)
		if m == nil {
			notes = append(notes, detail)
			continue
		}
		name := strings.ToLower(m[2])
		if m[3] != "" {
			name = strings.ToLower(m[3])
		}
		if strings.HasPrefix(name, "(") {
			notes = append(notes, detail)
			continue
		}
		tn, ok := names[name]
		if !ok && sch.byName[name] != nil {
			tn, ok = name, true
		}
		if !ok {
			return nil, nil, fmt.Errorf("플랜 행 %q 의 %s 를 테이블로 되돌리지 못했다", detail, name)
		}
		t := sch.byName[tn]
		a := access{Table: tn, Detail: detail, Full: true}
		switch {
		case strings.Contains(m[4], "AUTOMATIC"):
		case m[1] == "SEARCH" && (strings.Contains(m[4], "USING INTEGER PRIMARY KEY") || strings.Contains(m[4], "USING PRIMARY KEY")):
			a.Index, a.Full = "PK("+tn+")", false
		default:
			if im := planIndexRe.FindStringSubmatch(m[4]); im != nil {
				l, ok := sch.indexLabel(t, im[1])
				if !ok {
					return nil, nil, fmt.Errorf("플랜 행 %q 의 인덱스 %s 가 스키마에 없다", detail, im[1])
				}
				a.Index = l
				a.Full = m[1] != "SEARCH"
			}
		}
		acc = append(acc, a)
	}
	return acc, notes, rows.Err()
}

func supportOf(sh string, acc []access) string {
	tables := map[string]bool{}
	if f := strings.Fields(sh); len(f) > 1 {
		for _, t := range strings.Split(f[1], ",") {
			tables[t] = true
		}
	}
	labels := map[string][]string{}
	var order []string
	for _, a := range acc {
		if !tables[a.Table] {
			continue
		}
		if a.Full {
			return noSupport
		}
		if _, ok := labels[a.Table]; !ok {
			order = append(order, a.Table)
		}
		if !contains(labels[a.Table], a.Index) {
			labels[a.Table] = append(labels[a.Table], a.Index)
		}
	}
	if len(order) == 0 {
		return noAccess
	}
	sort.Strings(order)
	if len(tables) == 1 {
		return strings.Join(labels[order[0]], ", ")
	}
	parts := make([]string, len(order))
	for i, t := range order {
		parts[i] = t + ": " + strings.Join(labels[t], ", ")
	}
	return strings.Join(parts, "; ")
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

type planInput struct{ site, sql string }

func judgePlans(c *catalog, sites []site, sch *schema, criteria map[string]bool, expectedRow func(shape, support string) string) ([]violation, []sqlPlan) {
	var inv3 []violation
	add := func(kind, subject, detail, expected string) {
		inv3 = append(inv3, violation{Invariant: "3", Kind: kind, Subject: subject, Detail: detail, Expected: expected})
	}
	inputs := map[string][]planInput{}
	for _, s := range sites {
		if s.Unextractable == "" {
			inputs[s.Shape] = append(inputs[s.Shape], planInput{s.ID(), s.SQL})
		}
	}
	byID := map[string]patternRow{}
	for _, r := range c.patterns {
		byID[r.ID] = r
	}
	for _, m := range c.manual {
		if r, ok := byID[m.Pattern]; ok {
			inputs[r.Shape] = append(inputs[r.Shape], planInput{m.Site, normSpace(m.SQL)})
		}
	}

	plans := []sqlPlan{}
	used := map[string]bool{}
	for _, r := range c.patterns {
		in := inputs[r.Shape]
		got := map[string][]string{}
		for _, p := range in {
			acc, notes, err := explain(sch, p.sql)
			if err != nil {
				add("plan-error", r.ID, fmt.Sprintf("%s 의 SQL 을 플랜하지 못했다: %v", p.site, err), "")
				continue
			}
			for _, a := range acc {
				if a.Index != "" {
					used[a.Table+"\x00"+a.Index] = true
				}
			}
			sup := supportOf(r.Shape, acc)
			if acc == nil {
				acc = []access{}
			}
			if notes == nil {
				notes = []string{}
			}
			plans = append(plans, sqlPlan{Pattern: r.ID, Site: p.site, SQL: p.sql, Support: sup, Accesses: acc, Notes: notes})
			if !contains(got[sup], p.site) {
				got[sup] = append(got[sup], p.site)
			}
		}
		if len(got) == 0 {
			continue
		}
		if len(got) > 1 {
			var parts []string
			for sup, ss := range got {
				sort.Strings(ss)
				parts = append(parts, sup+" ← "+strings.Join(ss, ", "))
			}
			sort.Strings(parts)
			add("split-plan", r.ID, "같은 패턴의 지점들이 서로 다른 판정을 받았다: "+strings.Join(parts, " / "), "")
			continue
		}
		var want string
		for sup := range got {
			want = sup
		}
		fm := fullScanRe.FindStringSubmatch(r.Support)
		switch {
		case want == noSupport && fm != nil && criteria[fm[1]]:
		case want == noSupport && fm != nil:
			add("criterion-unknown", r.ID, fmt.Sprintf("%s 는 %s 표에 없다", fm[1], criteriaPath), "")
		case want == noSupport && r.Support == noSupport:
			add("no-support", r.ID, "엔진이 풀스캔한다 — 지원 인덱스를 더하는 마이그레이션이나 fullscan-criteria.md 풀스캔 허용 기준(사람의 결정)이 닫는다", "")
		case want == noSupport:
			add("support-mismatch", r.ID, fmt.Sprintf("지원 칸 %s — 엔진은 풀스캔", r.Support), expectedRow(r.Shape, noSupport))
		case r.Support != want:
			add("support-mismatch", r.ID, fmt.Sprintf("지원 칸 %s — 엔진 판정 %s", r.Support, want), expectedRow(r.Shape, want))
		}
	}
	sort.SliceStable(plans, func(i, j int) bool {
		if plans[i].Pattern != plans[j].Pattern {
			return plans[i].Pattern < plans[j].Pattern
		}
		if plans[i].Site != plans[j].Site {
			return plans[i].Site < plans[j].Site
		}
		return plans[i].SQL < plans[j].SQL
	})

	type ix struct{ label, table string }
	var unusedWant []ix
	for _, t := range sch.Tables {
		for _, i := range t.Indexes {
			if !used[t.Name+"\x00"+i.Label] {
				unusedWant = append(unusedWant, ix{i.Label, t.Name})
			}
		}
	}
	if !c.hasUnused {
		add("unused-table-missing", qpPath, fmt.Sprintf("「미사용 인덱스」 표(%s)가 없다", strings.Join(unusedHeader, " | ")), "")
	}
	doc := map[ix]bool{}
	for _, u := range c.unused {
		doc[ix{u.Index, u.Table}] = true
	}
	for _, u := range unusedWant {
		if !doc[u] {
			add("unused-unregistered", u.label, "어떤 패턴의 플랜도 쓰지 않는다 — 등재만 하고 지우지 않는다", unusedLine(u.label, u.table, "-"))
		}
		delete(doc, u)
	}
	for u := range doc {
		add("unused-stale", u.label, "플랜이 쓰는 인덱스이거나 스키마에 없는 인덱스다 — 행을 지운다", "")
	}
	return inv3, plans
}

func unusedLine(label, table, note string) string {
	return fmt.Sprintf("| `%s` | `%s` | %s |", label, table, escapeCell(note))
}

var (
	criteriaHeader = []string{"ID", "기준", "관측 가능한 근거"}
	criterionIDRe  = regexp.MustCompile(`^F\d+$`)
)

func readCriteria(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	in, seen := false, false
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			in = false
			continue
		}
		cs := pipeCells(line)
		switch {
		case sameHeader(cs, criteriaHeader):
			if seen {
				return nil, fmt.Errorf("기준 표(%s)가 둘 이상이다", strings.Join(criteriaHeader, " | "))
			}
			in, seen = true, true
		case !in || strings.HasPrefix(strings.ReplaceAll(line, " ", ""), "|---"):
		case len(cs) != 3 || !criterionIDRe.MatchString(cs[0]):
			return nil, fmt.Errorf("기준 표 행은 F<n> | 기준 | 근거 세 칸이다: %s", line)
		default:
			out[cs[0]] = true
		}
	}
	if !seen {
		return nil, fmt.Errorf("기준 표(%s)가 없다", strings.Join(criteriaHeader, " | "))
	}
	return out, nil
}

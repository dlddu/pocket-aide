package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

const noAccess = "—"

type patternRow struct {
	ID, Shape, Support string
	Sites              []string
}

type manualRow struct {
	Site, Pattern, SQL, Reason string
}

type unusedRow struct {
	Index, Table, Note string
}

type catalog struct {
	patterns  []patternRow
	manual    []manualRow
	unused    []unusedRow
	hasUnused bool
}

var (
	patternHeader = []string{"ID", "형태", "지원 인덱스 또는 허용 사유", "호출 지점"}
	manualHeader  = []string{"호출 지점", "패턴 ID", "대표 SQL", "추출 불가 사유"}
	unusedHeader  = []string{"인덱스", "테이블", "비고"}
	idRe          = regexp.MustCompile(`^Q-\d{2,}$`)
	shapeRe       = regexp.MustCompile(`^(select|insert|update|delete|upsert) [a-z_]+(,[a-z_]+)* \| (-|eq\(.+\)) \| (-|range\(.+\)) \| (-|order\(.+\))$`)
)

func pipeCells(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	if strings.HasSuffix(line, "|") && !strings.HasSuffix(line, `\|`) {
		line = line[:len(line)-1]
	}
	var out []string
	var cur strings.Builder
	for i := 0; i < len(line); i++ {
		switch {
		case line[i] == '\\' && i+1 < len(line) && line[i+1] == '|':
			cur.WriteByte('|')
			i++
		case line[i] == '|':
			out = append(out, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(line[i])
		}
	}
	return append(out, strings.TrimSpace(cur.String()))
}

func sameHeader(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func parseCatalog(path string) (*catalog, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	c := &catalog{}
	var mode string
	seen := map[string]bool{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "|") {
			mode = ""
			continue
		}
		cs := pipeCells(line)
		switch {
		case sameHeader(cs, patternHeader):
			mode = "pattern"
		case sameHeader(cs, manualHeader):
			mode = "manual"
		case sameHeader(cs, unusedHeader):
			mode = "unused"
		case strings.HasPrefix(strings.ReplaceAll(line, " ", ""), "|---"):
			continue
		case mode == "pattern":
			if len(cs) != 4 {
				return nil, fmt.Errorf("패턴 표 행의 칸 수가 4가 아니다: %s", line)
			}
			r := patternRow{ID: cs[0], Shape: unquote(cs[1]), Support: cs[2]}
			for _, s := range strings.Split(unquote(cs[3]), ",") {
				if s = strings.TrimSpace(s); s != "" {
					r.Sites = append(r.Sites, s)
				}
			}
			c.patterns = append(c.patterns, r)
		case mode == "manual":
			if len(cs) != 4 {
				return nil, fmt.Errorf("수동 형태 표 행의 칸 수가 4가 아니다: %s", line)
			}
			c.manual = append(c.manual, manualRow{Site: unquote(cs[0]), Pattern: cs[1], SQL: unquote(cs[2]), Reason: cs[3]})
		case mode == "unused":
			if len(cs) != 3 {
				return nil, fmt.Errorf("미사용 인덱스 표 행의 칸 수가 3이 아니다: %s", line)
			}
			c.unused = append(c.unused, unusedRow{Index: unquote(cs[0]), Table: unquote(cs[1]), Note: cs[2]})
		default:
			return nil, fmt.Errorf("어느 표에도 속하지 않는 표 행: %s", line)
		}
		if mode != "" && (sameHeader(cs, patternHeader) || sameHeader(cs, manualHeader) || sameHeader(cs, unusedHeader)) {
			if seen[mode] {
				return nil, fmt.Errorf("%s 표가 둘 이상이다", mode)
			}
			seen[mode] = true
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	c.hasUnused = seen["unused"]
	if !seen["pattern"] {
		return nil, fmt.Errorf("패턴 표(%s)가 없다", strings.Join(patternHeader, " | "))
	}
	if !seen["manual"] {
		return nil, fmt.Errorf("수동 형태 표(%s)가 없다", strings.Join(manualHeader, " | "))
	}
	return c, nil
}

func escapeCell(s string) string { return strings.ReplaceAll(s, "|", `\|`) }

func supportFor(accessless bool) string {
	if accessless {
		return noAccess
	}
	return noSupport
}

func patternLine(id, shape, support string, sites []string) string {
	q := make([]string, len(sites))
	for i, s := range sites {
		q[i] = "`" + s + "`"
	}
	return fmt.Sprintf("| %s | `%s` | %s | %s |", id, escapeCell(shape), support, strings.Join(q, ", "))
}

func manualLine(site, id, sql, reason string) string {
	return fmt.Sprintf("| `%s` | %s | `%s` | %s |", site, id, escapeCell(sql), escapeCell(reason))
}

type pair struct{ site, shape string }

func checkCatalog(c *catalog, sites []site, sch *schema, criteria map[string]bool) (inv2, inv3 []violation, plans []sqlPlan) {
	add := func(kind, subject, detail, expected string) {
		inv2 = append(inv2, violation{Invariant: "2", Kind: kind, Subject: subject, Detail: detail, Expected: expected})
	}

	code := map[pair]bool{}
	codeSites := map[string]map[string]bool{}
	accessless := map[string]bool{}
	unextractable := map[string]string{}
	sqlFor := map[string]string{}
	for _, s := range sites {
		if s.Unextractable != "" {
			if _, ok := unextractable[s.ID()]; !ok {
				unextractable[s.ID()] = s.Unextractable
			}
			continue
		}
		if _, ok := sqlFor[s.Shape]; !ok {
			sqlFor[s.Shape] = s.SQL
		}
		code[pair{s.ID(), s.Shape}] = true
		if codeSites[s.Shape] == nil {
			codeSites[s.Shape] = map[string]bool{}
		}
		codeSites[s.Shape][s.ID()] = true
		accessless[s.Shape] = s.shape.accessless
	}

	byID := map[string]*patternRow{}
	byShape := map[string][]string{}
	for i := range c.patterns {
		r := &c.patterns[i]
		if !idRe.MatchString(r.ID) {
			add("pattern-id", r.ID, "ID 는 Q-<두 자리 이상 숫자>", "")
		}
		if byID[r.ID] != nil {
			add("pattern-id-duplicate", r.ID, "", "")
		}
		byID[r.ID] = r
		if !shapeRe.MatchString(r.Shape) {
			add("shape-malformed", r.ID, "형태 "+r.Shape, "")
		}
		byShape[r.Shape] = append(byShape[r.Shape], r.ID)
	}
	for sh, ids := range byShape {
		if len(ids) > 1 {
			sort.Strings(ids)
			add("ambiguous-shape", sh, "같은 형태의 행: "+strings.Join(ids, ", "), "")
		}
	}

	manualPairs := map[pair]bool{}
	manualSites := map[string]bool{}
	for _, m := range c.manual {
		manualSites[m.Site] = true
		r := byID[m.Pattern]
		if r == nil {
			add("manual-pattern-unknown", m.Site, "패턴 ID "+m.Pattern+" 가 패턴 표에 없다", "")
			continue
		}
		if _, ok := unextractable[m.Site]; !ok {
			add("manual-extra", m.Site, "체커가 추출 불가로 보고하지 않은 지점이다", "")
		}
		sh, err := extractShape(m.SQL, sch)
		switch {
		case err != nil:
			add("manual-sql", m.Site, "대표 SQL 에서 형태를 뽑지 못했다: "+err.Error(), "")
		case sh.String() != r.Shape:
			add("manual-shape-mismatch", m.Site, fmt.Sprintf("대표 SQL 의 형태 %s ≠ %s 의 형태 %s", sh.String(), r.ID, r.Shape), "")
		}
		manualPairs[pair{m.Site, r.Shape}] = true
		accessless[r.Shape] = accessless[r.Shape] || err == nil && sh.accessless
		if _, ok := sqlFor[r.Shape]; !ok && err == nil {
			sqlFor[r.Shape] = normSpace(m.SQL)
		}
	}
	var missingManual []string
	for id := range unextractable {
		if !manualSites[id] {
			missingManual = append(missingManual, id)
		}
	}
	sort.Strings(missingManual)
	for _, id := range missingManual {
		add("manual-missing", id, unextractable[id], manualLine(id, "Q-??", "<대표 SQL>", unextractable[id]))
	}

	doc := map[pair]string{}
	for _, r := range c.patterns {
		for _, s := range r.Sites {
			doc[pair{s, r.Shape}] = r.ID
		}
	}
	idFor := func(shape string) string {
		if ids := byShape[shape]; len(ids) > 0 {
			return ids[0]
		}
		return "Q-??"
	}
	guess := func(shape string) string {
		if q, ok := sqlFor[shape]; ok && sch != nil && sch.db != nil {
			if acc, _, err := explain(sch, q); err == nil {
				return supportOf(shape, acc)
			}
		}
		return supportFor(accessless[shape])
	}
	rowFor := func(shape, support string) string {
		set := map[string]bool{}
		for s := range codeSites[shape] {
			set[s] = true
		}
		for p := range manualPairs {
			if p.shape == shape {
				set[p.site] = true
			}
		}
		var ss []string
		for s := range set {
			ss = append(ss, s)
		}
		sort.Strings(ss)
		return patternLine(idFor(shape), shape, support, ss)
	}
	expectedRow := func(shape string) string { return rowFor(shape, guess(shape)) }
	var unreg []pair
	for p := range code {
		if _, ok := doc[p]; !ok {
			unreg = append(unreg, p)
		}
	}
	for p := range manualPairs {
		if _, ok := doc[p]; !ok {
			unreg = append(unreg, p)
		}
	}
	sort.Slice(unreg, func(i, j int) bool {
		if unreg[i].site != unreg[j].site {
			return unreg[i].site < unreg[j].site
		}
		return unreg[i].shape < unreg[j].shape
	})
	for _, p := range unreg {
		add("query-unregistered", p.site, "형태 "+p.shape, expectedRow(p.shape))
	}

	for _, r := range c.patterns {
		live := 0
		for _, s := range r.Sites {
			p := pair{s, r.Shape}
			if code[p] || manualPairs[p] {
				live++
			} else {
				add("dead-site", s, r.ID+" 의 형태 "+r.Shape+" 를 가진 쿼리가 이 지점에 없다", "")
			}
		}
		if live == 0 {
			add("dead-pattern", r.ID, "살아 있는 지점이 없다 — 행을 지운다", "")
		}
	}
	inv3, plans = judgePlans(c, sites, sch, criteria, rowFor)
	return inv2, inv3, plans
}

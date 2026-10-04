package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type violation struct {
	Invariant string `json:"invariant"`
	Kind      string `json:"kind"`
	Subject   string `json:"subject"`
	Detail    string `json:"detail,omitempty"`
	Expected  string `json:"expected,omitempty"`
}

func boolCell(b bool) string {
	if b {
		return "YES"
	}
	return "NO"
}

func keyCell(t *table, c column) string {
	var parts []string
	if c.PKOrd > 0 {
		pks := pkCols(t)
		var inOrder []string
		for _, col := range t.Columns {
			if col.PKOrd > 0 {
				inOrder = append(inOrder, col.Name)
			}
		}
		if strings.Join(pks, ",") == strings.Join(inOrder, ",") {
			parts = append(parts, "PK")
		} else {
			parts = append(parts, fmt.Sprintf("PK(%d)", c.PKOrd))
		}
	}
	for _, fk := range c.FKs {
		parts = append(parts, "FK → "+fk)
	}
	return strings.Join(parts, ", ")
}

func columnRow(t *table, c column) string {
	return fmt.Sprintf("| `%s` | %s | %s | %s |", c.Name, c.Type, boolCell(c.Nullable), keyCell(t, c))
}

func indexRow(ix index) string {
	cols := make([]string, len(ix.Cols))
	for i, c := range ix.Cols {
		cols[i] = "`" + c + "`"
	}
	cond := "-"
	if ix.Cond != "-" {
		cond = "`" + ix.Cond + "`"
	}
	return fmt.Sprintf("| `%s` | %s | %s | %s |", ix.Label, strings.Join(cols, ", "), boolCell(ix.Unique), cond)
}

func renderSection(t *table) string {
	var b strings.Builder
	fmt.Fprintf(&b, "### `%s`\n\n의미: <doc 주석 링크>\n\n| 컬럼 | 타입 | NULL | 키 |\n| --- | --- | --- | --- |\n", t.Name)
	for _, c := range t.Columns {
		b.WriteString(columnRow(t, c) + "\n")
	}
	if len(t.Indexes) > 0 {
		b.WriteString("\n| 인덱스 | 컬럼 | UNIQUE | 조건 |\n| --- | --- | --- | --- |\n")
		for _, ix := range t.Indexes {
			b.WriteString(indexRow(ix) + "\n")
		}
	}
	return b.String()
}

func expectedRelations(s *schema) []string {
	var out []string
	for _, t := range s.Tables {
		for _, fk := range t.FKs {
			left, right := "|o", "o{"
			if fk.NotNull {
				left = "||"
			}
			if fk.UniqueFK {
				right = "o|"
			}
			out = append(out, fmt.Sprintf("%s %s--%s %s : %s", fk.Ref, left, right, t.Name, strings.Join(fk.Cols, ",")))
		}
	}
	sort.Strings(out)
	return out
}

type erdSection struct {
	meaning string
	cols    map[string]string
	idx     map[string]string
}

type erdDoc struct {
	hasMermaid bool
	entities   map[string]bool
	relations  map[string]bool
	sections   map[string]*erdSection
}

var (
	relRe     = regexp.MustCompile(`^([A-Za-z_][\w]*)\s+(\|\||\|o|o\||\}o|o\{|\}\|)(--|\.\.)(\|\||\|o|o\||o\{|\}o|\|\{)\s+([A-Za-z_][\w]*)\s*:\s*"?([^"]*?)"?$`)
	entityRe  = regexp.MustCompile(`^([A-Za-z_][\w]*)$`)
	headingRe = regexp.MustCompile("^### `([^`]+)`\\s*$")
)

func cells(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

func unquote(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "`", "")) }

func parseERD(path string) (*erdDoc, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	d := &erdDoc{entities: map[string]bool{}, relations: map[string]bool{}, sections: map[string]*erdSection{}}
	var (
		inMermaid bool
		cur       *erdSection
		tableKind string
		header    bool
	)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		trim := strings.TrimSpace(line)
		if inMermaid {
			if strings.HasPrefix(trim, "```") {
				inMermaid = false
				continue
			}
			if trim == "" || trim == "erDiagram" || strings.HasPrefix(trim, "%%") {
				continue
			}
			if m := relRe.FindStringSubmatch(trim); m != nil {
				d.entities[m[1]] = true
				d.entities[m[5]] = true
				d.relations[fmt.Sprintf("%s %s%s%s %s : %s", m[1], m[2], m[3], m[4], m[5], strings.TrimSpace(m[6]))] = true
			} else if m := entityRe.FindStringSubmatch(trim); m != nil {
				d.entities[m[1]] = true
			} else {
				return nil, fmt.Errorf("erd.md mermaid 줄을 해석하지 못했다: %q", trim)
			}
			continue
		}
		if strings.HasPrefix(trim, "```mermaid") {
			inMermaid, d.hasMermaid = true, true
			continue
		}
		if strings.HasPrefix(trim, "#") {
			cur, tableKind = nil, ""
			if m := headingRe.FindStringSubmatch(trim); m != nil {
				cur = &erdSection{cols: map[string]string{}, idx: map[string]string{}}
				d.sections[m[1]] = cur
			}
			continue
		}
		if cur == nil {
			continue
		}
		if strings.HasPrefix(trim, "의미:") {
			cur.meaning = strings.TrimSpace(strings.TrimPrefix(trim, "의미:"))
			continue
		}
		if !strings.HasPrefix(trim, "|") {
			tableKind = ""
			continue
		}
		c := cells(trim)
		if tableKind == "" {
			switch c[0] {
			case "컬럼":
				tableKind = "col"
			case "인덱스":
				tableKind = "idx"
			default:
				tableKind = "other"
			}
			header = true
			continue
		}
		if header {
			header = false
			continue
		}
		switch tableKind {
		case "col":
			if len(c) < 4 {
				return nil, fmt.Errorf("erd.md 컬럼 행의 칸이 모자란다: %q", trim)
			}
			cur.cols[unquote(c[0])] = fmt.Sprintf("%s|%s|%s", strings.ToUpper(normSpace(c[1])), c[2], normSpace(strings.ReplaceAll(c[3], "->", "→")))
		case "idx":
			if len(c) < 4 {
				return nil, fmt.Errorf("erd.md 인덱스 행의 칸이 모자란다: %q", trim)
			}
			cols := strings.Split(unquote(c[1]), ",")
			for i := range cols {
				cols[i] = normSpace(cols[i])
			}
			cond := unquote(c[3])
			if cond == "" {
				cond = "-"
			}
			ix := index{Cols: cols, Unique: c[2] == "YES", Cond: cond}
			cur.idx[ix.key()] = unquote(c[0])
		}
	}
	return d, sc.Err()
}

var (
	linkRe    = regexp.MustCompile("\\[`?([A-Za-z_][\\w.]*)`?\\]\\(([^)\\s]+)\\)")
	docLineRe = regexp.MustCompile(`^\s*//`)
)

func checkMeaning(root, table, meaning string) *violation {
	v := &violation{Invariant: "1", Subject: table}
	if meaning == "" {
		v.Kind, v.Detail = "meaning-missing", "`의미:` 줄이 없다"
		return v
	}
	m := linkRe.FindStringSubmatch(meaning)
	if m == nil {
		v.Kind, v.Detail = "meaning-link", "`의미:` 가 doc 주석 링크가 아니다: "+meaning
		return v
	}
	name := m[1][strings.LastIndex(m[1], ".")+1:]
	target := filepath.Join(root, docDir, m[2])
	data, err := os.ReadFile(target)
	if err != nil {
		v.Kind, v.Detail = "meaning-link", "링크 대상 파일이 없다: "+m[2]
		return v
	}
	declRe := regexp.MustCompile(`^(type|package) ` + regexp.QuoteMeta(name) + `\b`)
	lines := strings.Split(string(data), "\n")
	for i, l := range lines {
		if declRe.MatchString(l) {
			if i > 0 && docLineRe.MatchString(lines[i-1]) {
				return nil
			}
			v.Kind, v.Detail = "meaning-undocumented", fmt.Sprintf("%s 의 %s 에 doc 주석이 없다", m[2], name)
			return v
		}
	}
	v.Kind, v.Detail = "meaning-link", fmt.Sprintf("%s 에 type/package %s 선언이 없다", m[2], name)
	return v
}

func checkERD(root string, s *schema) ([]violation, error) {
	path := filepath.Join(root, erdPath)
	var out []violation
	if _, err := os.Stat(path); err != nil {
		out = append(out, violation{Invariant: "1", Kind: "erd-missing", Subject: erdPath, Detail: "erd.md 가 없다"})
		for _, t := range s.Tables {
			out = append(out, violation{Invariant: "1", Kind: "table-missing", Subject: t.Name, Expected: renderSection(t)})
		}
		return out, nil
	}
	d, err := parseERD(path)
	if err != nil {
		return nil, err
	}

	if !d.hasMermaid {
		out = append(out, violation{Invariant: "1", Kind: "diagram-missing", Subject: erdPath, Expected: "```mermaid\nerDiagram\n    " + strings.Join(expectedRelations(s), "\n    ") + "\n```"})
	} else {
		for _, t := range s.Tables {
			if !d.entities[t.Name] {
				out = append(out, violation{Invariant: "1", Kind: "diagram-entity-missing", Subject: t.Name})
			}
		}
		for e := range d.entities {
			if s.byName[e] == nil {
				out = append(out, violation{Invariant: "1", Kind: "diagram-entity-extra", Subject: e})
			}
		}
		want := map[string]bool{}
		for _, r := range expectedRelations(s) {
			want[r] = true
			if !d.relations[r] {
				out = append(out, violation{Invariant: "1", Kind: "diagram-relation-missing", Subject: r, Expected: r})
			}
		}
		for r := range d.relations {
			if !want[r] {
				out = append(out, violation{Invariant: "1", Kind: "diagram-relation-extra", Subject: r})
			}
		}
	}

	for _, t := range s.Tables {
		sec := d.sections[t.Name]
		if sec == nil {
			out = append(out, violation{Invariant: "1", Kind: "table-missing", Subject: t.Name, Expected: renderSection(t)})
			continue
		}
		if v := checkMeaning(root, t.Name, sec.meaning); v != nil {
			out = append(out, *v)
		}
		seen := map[string]bool{}
		for _, c := range t.Columns {
			seen[c.Name] = true
			want := fmt.Sprintf("%s|%s|%s", c.Type, boolCell(c.Nullable), keyCell(t, c))
			got, ok := sec.cols[c.Name]
			switch {
			case !ok:
				out = append(out, violation{Invariant: "1", Kind: "column-missing", Subject: t.Name + "." + c.Name, Expected: columnRow(t, c)})
			case got != want:
				out = append(out, violation{Invariant: "1", Kind: "column-mismatch", Subject: t.Name + "." + c.Name, Detail: "ERD " + got, Expected: columnRow(t, c)})
			}
		}
		for name := range sec.cols {
			if !seen[name] {
				out = append(out, violation{Invariant: "1", Kind: "column-extra", Subject: t.Name + "." + name})
			}
		}
		seenIdx := map[string]bool{}
		for _, ix := range t.Indexes {
			seenIdx[ix.key()] = true
			label, ok := sec.idx[ix.key()]
			switch {
			case !ok:
				out = append(out, violation{Invariant: "1", Kind: "index-missing", Subject: t.Name + " " + ix.Label, Expected: indexRow(ix)})
			case label != ix.Label:
				out = append(out, violation{Invariant: "1", Kind: "index-label", Subject: t.Name + " " + ix.Label, Detail: "ERD " + label, Expected: indexRow(ix)})
			}
		}
		for k, label := range sec.idx {
			if !seenIdx[k] {
				out = append(out, violation{Invariant: "1", Kind: "index-extra", Subject: t.Name + " " + label})
			}
		}
	}
	for name := range d.sections {
		if s.byName[name] == nil {
			out = append(out, violation{Invariant: "1", Kind: "table-extra", Subject: name})
		}
	}
	return out, nil
}

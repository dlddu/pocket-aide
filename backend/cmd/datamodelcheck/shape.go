package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type tokKind int

const (
	tWord tokKind = iota
	tValue
	tPunct
	tSub
)

type tok struct {
	kind tokKind
	text string
	sub  int
}

func (t tok) is(words ...string) bool {
	if t.kind != tWord {
		return false
	}
	for _, w := range words {
		if strings.EqualFold(t.text, w) {
			return true
		}
	}
	return false
}

func (t tok) isP(p string) bool { return t.kind == tPunct && t.text == p }

var sqlKeywords = map[string]bool{}

func init() {
	for _, k := range strings.Fields("SELECT FROM WHERE AND OR NOT IN IS NULL LIKE BETWEEN ORDER BY GROUP HAVING LIMIT OFFSET ASC DESC JOIN INNER LEFT RIGHT OUTER CROSS NATURAL ON AS INSERT INTO VALUES UPDATE SET DELETE REPLACE CONFLICT DO NOTHING IGNORE ABORT FAIL ROLLBACK RETURNING UNION ALL INTERSECT EXCEPT WITH DISTINCT CASE WHEN THEN ELSE END EXISTS TRUE FALSE USING") {
		sqlKeywords[k] = true
	}
}

func isKeyword(t tok) bool { return t.kind == tWord && sqlKeywords[strings.ToUpper(t.text)] }

func tokenize(sql string) ([]tok, error) {
	var out []tok
	rs := []rune(sql)
	for i := 0; i < len(rs); {
		c := rs[i]
		switch {
		case unicode.IsSpace(c):
			i++
		case c == '-' && i+1 < len(rs) && rs[i+1] == '-':
			for i < len(rs) && rs[i] != '\n' {
				i++
			}
		case c == '\'':
			j := i + 1
			for ; j < len(rs); j++ {
				if rs[j] == '\'' {
					if j+1 < len(rs) && rs[j+1] == '\'' {
						j++
						continue
					}
					break
				}
			}
			if j >= len(rs) {
				return nil, fmt.Errorf("닫히지 않은 문자열")
			}
			out = append(out, tok{kind: tValue, text: string(rs[i : j+1])})
			i = j + 1
		case c == '"' || c == '`':
			j := i + 1
			for j < len(rs) && rs[j] != c {
				j++
			}
			if j >= len(rs) {
				return nil, fmt.Errorf("닫히지 않은 식별자")
			}
			out = append(out, tok{kind: tWord, text: strings.ToLower(string(rs[i+1 : j]))})
			i = j + 1
		case c == '?':
			j := i + 1
			for j < len(rs) && unicode.IsDigit(rs[j]) {
				j++
			}
			out = append(out, tok{kind: tValue, text: string(rs[i:j])})
			i = j
		case unicode.IsDigit(c):
			j := i
			for j < len(rs) && (unicode.IsDigit(rs[j]) || rs[j] == '.') {
				j++
			}
			out = append(out, tok{kind: tValue, text: string(rs[i:j])})
			i = j
		case c == '_' || unicode.IsLetter(c):
			j := i
			for j < len(rs) && (rs[j] == '_' || unicode.IsLetter(rs[j]) || unicode.IsDigit(rs[j])) {
				j++
			}
			out = append(out, tok{kind: tWord, text: string(rs[i:j])})
			i = j
		default:
			if i+1 < len(rs) {
				two := string(rs[i : i+2])
				if two == "<=" || two == ">=" || two == "!=" || two == "<>" || two == "==" || two == "||" {
					out = append(out, tok{kind: tPunct, text: two})
					i += 2
					continue
				}
			}
			out = append(out, tok{kind: tPunct, text: string(c)})
			i++
		}
	}
	for i := range out {
		if out[i].kind == tWord && !isKeyword(out[i]) {
			out[i].text = strings.ToLower(out[i].text)
		}
		if out[i].is("NULL", "TRUE", "FALSE") {
			out[i].kind = tValue
		}
	}
	return out, nil
}

type scope struct {
	toks []tok
}

type tableRef struct {
	name, alias string
	scope       int
}

type colRef struct {
	qual, name string
	scope      int
}

type pred struct {
	op    string
	left  colRef
	right *colRef
	expr  string
}

type shapeParser struct {
	scopes  []scope
	tables  []tableRef
	preds   []pred
	order   []string
	op      string
	conflct []colRef
	schema  *schema
}

type shape struct {
	Op     string
	Tables []string
	Eq     []string
	Range  []string
	Order  []string

	accessless bool
}

func list(name string, xs []string) string {
	if len(xs) == 0 {
		return "-"
	}
	return name + "(" + strings.Join(xs, ", ") + ")"
}

func (s shape) String() string {
	return fmt.Sprintf("%s %s | %s | %s | %s", s.Op, strings.Join(s.Tables, ","), list("eq", s.Eq), list("range", s.Range), list("order", s.Order))
}

func splitScopes(ts []tok) ([]scope, error) {
	var scopes []scope
	var walk func(ts []tok) ([]tok, error)
	walk = func(ts []tok) ([]tok, error) {
		var out []tok
		for i := 0; i < len(ts); i++ {
			if ts[i].isP("(") && i+1 < len(ts) && ts[i+1].is("SELECT") {
				depth, j := 0, i
				for ; j < len(ts); j++ {
					if ts[j].isP("(") {
						depth++
					} else if ts[j].isP(")") {
						depth--
						if depth == 0 {
							break
						}
					}
				}
				if j >= len(ts) {
					return nil, fmt.Errorf("괄호가 닫히지 않았다")
				}
				inner, err := walk(ts[i+1 : j])
				if err != nil {
					return nil, err
				}
				scopes = append(scopes, scope{toks: inner})
				out = append(out, tok{kind: tSub, text: "<sub>", sub: len(scopes) - 1})
				i = j
				continue
			}
			out = append(out, ts[i])
		}
		return out, nil
	}
	top, err := walk(ts)
	if err != nil {
		return nil, err
	}
	return append([]scope{{toks: top}}, scopes...), nil
}

func extractShape(sql string, sch *schema) (shape, error) {
	ts, err := tokenize(sql)
	if err != nil {
		return shape{}, err
	}
	for len(ts) > 0 && ts[len(ts)-1].isP(";") {
		ts = ts[:len(ts)-1]
	}
	for _, t := range ts {
		if t.is("WITH", "UNION", "INTERSECT", "EXCEPT") {
			return shape{}, fmt.Errorf("%s 는 형태로 합치지 않는다", strings.ToUpper(t.text))
		}
		if t.isP(";") {
			return shape{}, fmt.Errorf("문장이 둘 이상이다")
		}
	}
	scopes, err := splitScopes(ts)
	if err != nil {
		return shape{}, err
	}
	p := &shapeParser{scopes: scopes, schema: sch}
	if err := p.statement(0); err != nil {
		return shape{}, err
	}
	for i := 1; i < len(scopes); i++ {
		if err := p.selectScope(i, scopes[i].toks, false); err != nil {
			return shape{}, err
		}
	}
	return p.build()
}

func (p *shapeParser) statement(si int) error {
	ts := p.scopes[si].toks
	if len(ts) == 0 {
		return fmt.Errorf("빈 SQL")
	}
	switch {
	case ts[0].is("SELECT"):
		p.op = "select"
		return p.selectScope(si, ts, true)
	case ts[0].is("INSERT", "REPLACE"):
		return p.insert(si, ts)
	case ts[0].is("UPDATE"):
		p.op = "update"
		return p.update(si, ts)
	case ts[0].is("DELETE"):
		p.op = "delete"
		if len(ts) < 3 || !ts[1].is("FROM") {
			return fmt.Errorf("DELETE FROM 이 아니다")
		}
		i, err := p.tableAt(si, ts, 2)
		if err != nil {
			return err
		}
		return p.tail(si, ts[i:], false)
	}
	return fmt.Errorf("연산을 알 수 없다: %s", ts[0].text)
}

func (p *shapeParser) tableAt(si int, ts []tok, i int) (int, error) {
	if i >= len(ts) || ts[i].kind != tWord || isKeyword(ts[i]) {
		return i, fmt.Errorf("테이블 이름이 없다")
	}
	ref := tableRef{name: ts[i].text, scope: si}
	i++
	if i+1 < len(ts) && ts[i].isP(".") {
		ref.name = ts[i+1].text
		i += 2
	}
	if i < len(ts) && ts[i].is("AS") {
		i++
	}
	if i < len(ts) && ts[i].kind == tWord && !isKeyword(ts[i]) {
		ref.alias = ts[i].text
		i++
	}
	p.tables = append(p.tables, ref)
	return i, nil
}

func (p *shapeParser) insert(si int, ts []tok) error {
	p.op = "insert"
	i := 1
	if ts[0].is("REPLACE") {
		p.op = "upsert"
	} else if i+1 < len(ts) && ts[i].is("OR") {
		if ts[i+1].is("REPLACE") {
			p.op = "upsert"
		}
		i += 2
	}
	if i >= len(ts) || !ts[i].is("INTO") {
		return fmt.Errorf("INSERT INTO 가 아니다")
	}
	i, err := p.tableAt(si, ts, i+1)
	if err != nil {
		return err
	}
	if i < len(ts) && ts[i].isP("(") {
		depth := 0
		for ; i < len(ts); i++ {
			if ts[i].isP("(") {
				depth++
			} else if ts[i].isP(")") {
				depth--
				if depth == 0 {
					i++
					break
				}
			}
		}
	}
	rest := ts[i:]
	end := len(rest)
	for j := 0; j+1 < len(rest); j++ {
		if rest[j].is("ON") && rest[j+1].is("CONFLICT") {
			end = j
			p.op = "upsert"
			k := j + 2
			if k < len(rest) && rest[k].isP("(") {
				for k++; k < len(rest) && !rest[k].isP(")"); k++ {
					if rest[k].kind == tWord && !isKeyword(rest[k]) {
						p.conflct = append(p.conflct, colRef{name: rest[k].text, scope: si})
					}
				}
			}
			break
		}
	}
	for j := 0; j < end; j++ {
		if rest[j].is("RETURNING") {
			end = j
			break
		}
	}
	rest = rest[:end]
	if len(rest) > 0 && rest[0].is("SELECT") {
		return p.selectScope(si, rest, false)
	}
	if len(rest) > 0 && !rest[0].is("VALUES", "DEFAULT") {
		return fmt.Errorf("INSERT 본문을 알 수 없다: %s", rest[0].text)
	}
	return nil
}

func (p *shapeParser) update(si int, ts []tok) error {
	i := 1
	if i+1 < len(ts) && ts[i].is("OR") {
		i += 2
	}
	i, err := p.tableAt(si, ts, i)
	if err != nil {
		return err
	}
	if i >= len(ts) || !ts[i].is("SET") {
		return fmt.Errorf("UPDATE … SET 이 아니다")
	}
	for ; i < len(ts) && !ts[i].is("WHERE", "RETURNING"); i++ {
	}
	return p.tail(si, ts[i:], false)
}

func clauseEnd(ts []tok, i int, stops ...string) int {
	depth := 0
	for ; i < len(ts); i++ {
		switch {
		case ts[i].isP("("):
			depth++
		case ts[i].isP(")"):
			depth--
		case depth == 0 && ts[i].is(stops...):
			return i
		}
	}
	return i
}

var clauseStops = []string{"FROM", "WHERE", "GROUP", "HAVING", "ORDER", "LIMIT", "OFFSET", "RETURNING", "JOIN", "INNER", "LEFT", "RIGHT", "CROSS", "NATURAL", "OUTER", "ON"}

func (p *shapeParser) selectScope(si int, ts []tok, outer bool) error {
	if len(ts) == 0 || !ts[0].is("SELECT") {
		return fmt.Errorf("SELECT 가 아니다")
	}
	i := clauseEnd(ts, 1, "FROM")
	if i >= len(ts) {
		return nil
	}
	i, err := p.tableAt(si, ts, i+1)
	if err != nil {
		return err
	}
	for i < len(ts) {
		if ts[i].isP(",") {
			if i, err = p.tableAt(si, ts, i+1); err != nil {
				return err
			}
			continue
		}
		j := i
		for j < len(ts) && ts[j].is("INNER", "LEFT", "RIGHT", "CROSS", "NATURAL", "OUTER") {
			j++
		}
		if j >= len(ts) || !ts[j].is("JOIN") {
			break
		}
		if i, err = p.tableAt(si, ts, j+1); err != nil {
			return err
		}
		if i < len(ts) && ts[i].is("ON") {
			e := clauseEnd(ts, i+1, clauseStops...)
			if err := p.conds(si, ts[i+1:e]); err != nil {
				return err
			}
			i = e
		}
	}
	return p.tail(si, ts[i:], outer)
}

func (p *shapeParser) tail(si int, ts []tok, outer bool) error {
	for i := 0; i < len(ts); {
		switch {
		case ts[i].is("WHERE"), ts[i].is("HAVING"):
			e := clauseEnd(ts, i+1, "GROUP", "HAVING", "ORDER", "LIMIT", "OFFSET", "RETURNING")
			if ts[i].is("WHERE") {
				if err := p.conds(si, ts[i+1:e]); err != nil {
					return err
				}
			}
			i = e
		case ts[i].is("ORDER") && i+1 < len(ts) && ts[i+1].is("BY"):
			e := clauseEnd(ts, i+2, "LIMIT", "OFFSET", "RETURNING")
			if outer {
				p.orderBy(si, ts[i+2:e])
			}
			i = e
		case ts[i].is("GROUP", "LIMIT", "OFFSET", "RETURNING"):
			i = clauseEnd(ts, i+1, "HAVING", "ORDER", "LIMIT", "OFFSET", "RETURNING")
			if i < len(ts) && ts[i].is("GROUP", "LIMIT", "OFFSET", "RETURNING") {
				i++
			}
		default:
			return fmt.Errorf("해석하지 못한 절: %s", ts[i].text)
		}
	}
	return nil
}

func splitTop(ts []tok, sep string) [][]tok {
	var out [][]tok
	depth, start, between := 0, 0, false
	for i, t := range ts {
		switch {
		case t.isP("("):
			depth++
		case t.isP(")"):
			depth--
		case depth == 0 && t.is("BETWEEN"):
			between = true
		case depth == 0 && sep == "AND" && t.is("AND") && between:
			between = false
		case depth == 0 && (sep == "," && t.isP(",") || sep != "," && t.is(sep)):
			out = append(out, ts[start:i])
			start = i + 1
		}
	}
	return append(out, ts[start:])
}

func wrapped(ts []tok) bool {
	if len(ts) < 2 || !ts[0].isP("(") || !ts[len(ts)-1].isP(")") {
		return false
	}
	depth := 0
	for i, t := range ts {
		if t.isP("(") {
			depth++
		} else if t.isP(")") {
			depth--
			if depth == 0 && i != len(ts)-1 {
				return false
			}
		}
	}
	return true
}

func (p *shapeParser) conds(si int, ts []tok) error {
	if len(splitTop(ts, "OR")) > 1 {
		return nil
	}
	for _, c := range splitTop(ts, "AND") {
		if len(c) == 0 {
			return fmt.Errorf("빈 조건")
		}
		if wrapped(c) {
			if err := p.conds(si, c[1:len(c)-1]); err != nil {
				return err
			}
			continue
		}
		p.atom(si, c)
	}
	return nil
}

func operand(si int, ts []tok) (*colRef, string) {
	switch {
	case len(ts) == 1 && ts[0].kind == tWord && !isKeyword(ts[0]):
		return &colRef{name: ts[0].text, scope: si}, ""
	case len(ts) == 3 && ts[0].kind == tWord && ts[1].isP(".") && ts[2].kind == tWord:
		return &colRef{qual: ts[0].text, name: ts[2].text, scope: si}, ""
	}
	for _, t := range ts {
		if t.kind == tWord && !isKeyword(t) {
			return nil, exprText(ts)
		}
	}
	return nil, ""
}

func exprText(ts []tok) string {
	var b strings.Builder
	for i, t := range ts {
		if i > 0 && t.kind != tPunct && ts[i-1].kind != tPunct {
			b.WriteByte(' ')
		}
		if t.kind == tWord || t.kind == tValue && !strings.HasPrefix(t.text, "'") {
			b.WriteString(strings.ToLower(t.text))
		} else {
			b.WriteString(t.text)
		}
	}
	return b.String()
}

func (p *shapeParser) atom(si int, c []tok) {
	if c[0].is("NOT", "EXISTS") {
		return
	}
	for i, t := range c {
		var op string
		var n int
		switch {
		case t.isP("=") || t.isP("=="):
			op, n = "eq", 1
		case t.is("IS"):
			if i+1 < len(c) && c[i+1].is("NOT") {
				return
			}
			op, n = "eq", 1
		case t.is("IN"):
			op, n = "eq", 1
		case t.is("NOT"):
			return
		case t.isP("<") || t.isP("<=") || t.isP(">") || t.isP(">=") || t.is("BETWEEN", "LIKE"):
			op, n = "range", 1
		case t.isP("!=") || t.isP("<>"):
			return
		default:
			continue
		}
		left, lexpr := operand(si, c[:i])
		right, _ := operand(si, c[i+n:])
		if left == nil && lexpr == "" && right != nil && op == "eq" {
			p.preds = append(p.preds, pred{op: op, left: *right})
			return
		}
		if left == nil && lexpr == "" {
			return
		}
		pr := pred{op: op, expr: lexpr}
		if left != nil {
			pr.left = *left
		}
		if op == "eq" && left != nil && right != nil && !t.is("IN", "IS") {
			pr.right = right
		}
		p.preds = append(p.preds, pr)
		return
	}
}

func (p *shapeParser) orderBy(si int, ts []tok) {
	for _, item := range splitTop(ts, ",") {
		dir := "asc"
		if n := len(item); n > 0 && item[n-1].is("ASC", "DESC") {
			dir = strings.ToLower(item[n-1].text)
			item = item[:n-1]
		}
		if c, e := operand(si, item); c != nil {
			p.order = append(p.order, "\x00"+c.qual+"\x00"+c.name+"\x00"+fmt.Sprint(si)+"\x00"+dir)
		} else {
			p.order = append(p.order, e+" "+dir)
		}
	}
}

func (p *shapeParser) resolve(c colRef) (table string, err error) {
	if c.qual != "" {
		for _, t := range p.tables {
			if t.alias == c.qual {
				return t.name, nil
			}
		}
		for _, t := range p.tables {
			if t.name == c.qual {
				return t.name, nil
			}
		}
		return "", fmt.Errorf("별칭 %s 를 찾지 못했다", c.qual)
	}
	var inScope []string
	for _, t := range p.tables {
		if t.scope == c.scope {
			inScope = append(inScope, t.name)
		}
	}
	if len(inScope) == 1 {
		return inScope[0], nil
	}
	var owners []string
	for _, name := range inScope {
		if p.schema != nil {
			if t := p.schema.byName[name]; t != nil {
				for _, col := range t.Columns {
					if col.Name == c.name {
						owners = append(owners, name)
					}
				}
			}
		}
	}
	if len(owners) == 1 {
		return owners[0], nil
	}
	return "", fmt.Errorf("컬럼 %s 의 테이블을 정하지 못했다", c.name)
}

func uniqSorted(xs []string) []string {
	sort.Strings(xs)
	out := xs[:0]
	for i, x := range xs {
		if i == 0 || x != xs[i-1] {
			out = append(out, x)
		}
	}
	return out
}

func (p *shapeParser) build() (shape, error) {
	s := shape{Op: p.op, accessless: (p.op == "insert" || p.op == "upsert") && len(p.preds) == 0 && len(p.scopes) == 1}
	set := map[string]bool{}
	for _, t := range p.tables {
		if !set[t.name] {
			set[t.name] = true
			s.Tables = append(s.Tables, t.name)
		}
	}
	sort.Strings(s.Tables)
	multi := len(s.Tables) > 1
	name := func(c colRef) (string, error) {
		t, err := p.resolve(c)
		if err != nil {
			return "", err
		}
		if multi {
			return t + "." + c.name, nil
		}
		return c.name, nil
	}
	for _, pr := range p.preds {
		var l string
		if pr.expr != "" {
			l = pr.expr
		} else {
			var err error
			if l, err = name(pr.left); err != nil {
				return shape{}, err
			}
		}
		if pr.right != nil {
			r, err := name(*pr.right)
			if err != nil {
				return shape{}, err
			}
			if r < l {
				l, r = r, l
			}
			l += "=" + r
		}
		if pr.op == "eq" {
			s.Eq = append(s.Eq, l)
		} else {
			s.Range = append(s.Range, l)
		}
	}
	for _, c := range p.conflct {
		n, err := name(c)
		if err != nil {
			return shape{}, err
		}
		s.Eq = append(s.Eq, n)
	}
	s.Eq, s.Range = uniqSorted(s.Eq), uniqSorted(s.Range)
	for _, o := range p.order {
		if strings.HasPrefix(o, "\x00") {
			f := strings.Split(o, "\x00")
			var si int
			_, _ = fmt.Sscan(f[3], &si)
			n, err := name(colRef{qual: f[1], name: f[2], scope: si})
			if err != nil {
				return shape{}, err
			}
			o = n + " " + f[4]
		}
		s.Order = append(s.Order, o)
	}
	return s, nil
}

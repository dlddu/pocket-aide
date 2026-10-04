package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	msqlite "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

type column struct {
	Name     string
	Type     string
	Nullable bool
	PKOrd    int
	FKs      []string
}

type index struct {
	Label  string
	Cols   []string
	Unique bool
	Cond   string
}

func (ix index) key() string {
	return fmt.Sprintf("%s|%t|%s", strings.Join(ix.Cols, ","), ix.Unique, ix.Cond)
}

type foreignKey struct {
	Cols     []string
	Ref      string
	RefCols  []string
	NotNull  bool
	UniqueFK bool
}

type table struct {
	Name    string
	Columns []column
	Indexes []index
	FKs     []foreignKey
}

type schema struct {
	Tables []*table
	Engine string
	byName map[string]*table
	numIdx int
}

var wsRe = regexp.MustCompile(`\s+`)

func normSpace(s string) string { return strings.TrimSpace(wsRe.ReplaceAllString(s, " ")) }

func quoteIdent(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

func observeSchema(root string, b *scopeBlock) (*schema, error) {
	dir := filepath.Join(root, b.Migrations)
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return nil, fmt.Errorf("마이그레이션 디렉터리 %s 가 없다", b.Migrations)
	}
	tmp, err := os.MkdirTemp("", "datamodelcheck-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	conn, err := sql.Open("sqlite", "file:"+filepath.Join(tmp, "empty.db")+"?_pragma=foreign_keys(ON)")
	if err != nil {
		return nil, fmt.Errorf("엔진을 열지 못했다: %w", err)
	}
	defer func() { _ = conn.Close() }()

	src, err := iofs.New(os.DirFS(dir), ".")
	if err != nil {
		return nil, fmt.Errorf("마이그레이션 소스: %w", err)
	}
	drv, err := msqlite.WithInstance(conn, &msqlite.Config{})
	if err != nil {
		return nil, fmt.Errorf("마이그레이션 드라이버: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "sqlite", drv)
	if err != nil {
		return nil, fmt.Errorf("마이그레이션 준비: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("마이그레이션 적용 실패: %w", err)
	}

	s := &schema{byName: map[string]*table{}}
	var ver string
	if err := conn.QueryRow(`SELECT sqlite_version()`).Scan(&ver); err != nil {
		return nil, err
	}
	s.Engine = "SQLite " + ver + " (" + driverVersion() + ")"

	type master struct{ name, sql string }
	var tables []master
	rows, err := conn.Query(`SELECT name, sql FROM sqlite_master WHERE type = 'table' ORDER BY name`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var mt master
		if err := rows.Scan(&mt.name, &mt.sql); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if !b.schemaExclude.MatchString(mt.name) {
			tables = append(tables, mt)
		}
	}
	_ = rows.Close()

	for _, mt := range tables {
		t, err := readTable(conn, mt.name, mt.sql)
		if err != nil {
			return nil, fmt.Errorf("테이블 %s 카탈로그: %w", mt.name, err)
		}
		s.Tables = append(s.Tables, t)
		s.byName[t.Name] = t
	}
	for _, t := range s.Tables {
		s.numIdx += len(t.Indexes)
		for _, fk := range t.FKs {
			for _, c := range fk.Cols {
				for i := range t.Columns {
					if t.Columns[i].Name == c && len(fk.Cols) == 1 {
						t.Columns[i].FKs = append(t.Columns[i].FKs, fk.Ref+"."+fk.RefCols[0])
					}
				}
			}
		}
	}
	_, _ = m.Close()
	return s, nil
}

func readTable(conn *sql.DB, name, ddl string) (*table, error) {
	t := &table{Name: name}
	withoutRowid := strings.Contains(strings.ToUpper(normSpace(ddl)), "WITHOUT ROWID")

	rows, err := conn.Query(`PRAGMA table_info(` + quoteIdent(name) + `)`)
	if err != nil {
		return nil, err
	}
	var pkCount int
	for rows.Next() {
		var (
			cid, notnull, pk int
			cname, ctype     string
			dflt             sql.NullString
		)
		if err := rows.Scan(&cid, &cname, &ctype, &notnull, &dflt, &pk); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if pk > 0 {
			pkCount++
		}
		t.Columns = append(t.Columns, column{Name: cname, Type: strings.ToUpper(normSpace(ctype)), Nullable: notnull == 0, PKOrd: pk})
	}
	_ = rows.Close()
	for i := range t.Columns {
		c := &t.Columns[i]
		if pkCount == 1 && c.PKOrd == 1 && c.Type == "INTEGER" && !withoutRowid {
			c.Nullable = false
		}
	}

	if err := readIndexes(conn, t); err != nil {
		return nil, err
	}
	if err := readFKs(conn, t); err != nil {
		return nil, err
	}
	return t, nil
}

func readIndexes(conn *sql.DB, t *table) error {
	type il struct {
		name    string
		unique  bool
		origin  string
		partial bool
	}
	var list []il
	rows, err := conn.Query(`PRAGMA index_list(` + quoteIdent(t.Name) + `)`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			seq, unique, partial int
			name, origin         string
		)
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			_ = rows.Close()
			return err
		}
		list = append(list, il{name, unique == 1, origin, partial == 1})
	}
	_ = rows.Close()

	for _, x := range list {
		if x.origin == "pk" {
			continue
		}
		var ddl sql.NullString
		if err := conn.QueryRow(`SELECT sql FROM sqlite_master WHERE type = 'index' AND name = ?`, x.name).Scan(&ddl); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		exprs, cond := splitIndexDDL(ddl.String)

		xr, err := conn.Query(`PRAGMA index_xinfo(` + quoteIdent(x.name) + `)`)
		if err != nil {
			return err
		}
		var cols []string
		for xr.Next() {
			var (
				seqno, cid, desc, key int
				cname                 sql.NullString
				coll                  sql.NullString
			)
			if err := xr.Scan(&seqno, &cid, &cname, &desc, &coll, &key); err != nil {
				_ = xr.Close()
				return err
			}
			if key == 0 {
				continue
			}
			c := cname.String
			if cid == -2 {
				if seqno < len(exprs) {
					c = exprs[seqno]
				} else {
					c = "<expr>"
				}
			}
			if desc == 1 {
				c += " DESC"
			}
			cols = append(cols, c)
		}
		_ = xr.Close()

		ix := index{Cols: cols, Unique: x.unique, Cond: "-"}
		if x.partial {
			ix.Cond = cond
		}
		if x.origin == "c" {
			ix.Label = x.name
		} else {
			plain := make([]string, len(cols))
			for i, c := range cols {
				plain[i] = t.Name + "." + strings.TrimSuffix(c, " DESC")
			}
			ix.Label = "UNIQUE(" + strings.Join(plain, ",") + ")"
		}
		t.Indexes = append(t.Indexes, ix)
	}
	sort.Slice(t.Indexes, func(i, j int) bool { return t.Indexes[i].Label < t.Indexes[j].Label })
	return nil
}

func splitIndexDDL(ddl string) ([]string, string) {
	if ddl == "" {
		return nil, "-"
	}
	open := strings.Index(ddl, "(")
	if open < 0 {
		return nil, "-"
	}
	depth, end := 0, -1
	for i := open; i < len(ddl); i++ {
		switch ddl[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				end = i
			}
		}
		if end >= 0 {
			break
		}
	}
	if end < 0 {
		return nil, "-"
	}
	var parts []string
	depth, start := 0, open+1
	for i := open + 1; i < end; i++ {
		switch ddl[i] {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, normSpace(ddl[start:i]))
				start = i + 1
			}
		}
	}
	parts = append(parts, normSpace(ddl[start:end]))
	cond := "-"
	rest := ddl[end+1:]
	if i := strings.Index(strings.ToUpper(rest), "WHERE"); i >= 0 {
		cond = normSpace(rest[i+len("WHERE"):])
	}
	return parts, cond
}

func readFKs(conn *sql.DB, t *table) error {
	rows, err := conn.Query(`PRAGMA foreign_key_list(` + quoteIdent(t.Name) + `)`)
	if err != nil {
		return err
	}
	byID := map[int]*foreignKey{}
	var ids []int
	for rows.Next() {
		var (
			id, seq                 int
			ref, from               string
			to                      sql.NullString
			onUpd, onDel, matchRule string
		)
		if err := rows.Scan(&id, &seq, &ref, &from, &to, &onUpd, &onDel, &matchRule); err != nil {
			_ = rows.Close()
			return err
		}
		fk, ok := byID[id]
		if !ok {
			fk = &foreignKey{Ref: ref}
			byID[id] = fk
			ids = append(ids, id)
		}
		fk.Cols = append(fk.Cols, from)
		fk.RefCols = append(fk.RefCols, to.String)
	}
	_ = rows.Close()
	sort.Ints(ids)

	for _, id := range ids {
		fk := byID[id]
		fk.NotNull = true
		for _, c := range fk.Cols {
			for _, col := range t.Columns {
				if col.Name == c && col.Nullable {
					fk.NotNull = false
				}
			}
		}
		fk.UniqueFK = sameSet(fk.Cols, pkCols(t))
		for _, ix := range t.Indexes {
			if ix.Unique && ix.Cond == "-" && sameSet(fk.Cols, ix.Cols) {
				fk.UniqueFK = true
			}
		}
		t.FKs = append(t.FKs, *fk)
	}
	return nil
}

func pkCols(t *table) []string {
	var cols []column
	for _, c := range t.Columns {
		if c.PKOrd > 0 {
			cols = append(cols, c)
		}
	}
	sort.Slice(cols, func(i, j int) bool { return cols[i].PKOrd < cols[j].PKOrd })
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = c.Name
	}
	return out
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]int{}
	for _, x := range a {
		seen[x]++
	}
	for _, x := range b {
		seen[x]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
}

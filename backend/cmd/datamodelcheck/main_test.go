package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureReadme = "# 데이터 모델\n\n```data-model-scope\nmigrations: m\nchecker: chk\nscope: src\nexclude: _test\\.go$\nsite: \\.(QueryContext|ExecContext)\\(\nschema-exclude: ^(schema_migrations|sqlite_sequence)$\nsql: \\bFROM [a-z_]+\\b\n```\n"

const fixtureCriteria = "# 풀스캔 허용 기준\n\n| ID | 기준 | 관측 가능한 근거 |\n| --- | --- | --- |\n| F1 | 픽스처 기준 | 픽스처 |\n"

const fixtureMigration = `CREATE TABLE parent (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);
CREATE TABLE child (
    id INTEGER PRIMARY KEY,
    parent_id INTEGER NOT NULL REFERENCES parent(id),
    note text,
    at INTEGER NOT NULL
);
CREATE INDEX idx_child_parent_at ON child(parent_id, at DESC);
CREATE INDEX idx_child_note ON child(note) WHERE note IS NOT NULL;
CREATE TABLE solo (
    parent_id INTEGER PRIMARY KEY REFERENCES parent(id)
);
`

var fixtureSource = strings.Join([]string{
	"// Package src is a fixture.",
	"package src",
	"",
	"// Parent is one row of the parent table.",
	"type Parent struct{}",
	"",
	"type Child struct{}",
	"",
	"func (s *Store) List() {",
	"\ts.db.QueryContext(ctx, `SELECT id FROM child`)",
	"}",
	"",
	"func helper() {",
	"\tdb.ExecContext(ctx, `DELETE FROM parent`)",
	"}",
	"",
}, "\n")

const fixtureERD = "# ERD\n\n```mermaid\nerDiagram\n    parent ||--o{ child : parent_id\n    parent ||--o| solo : parent_id\n```\n\n" +
	"### `child`\n\n의미: [`src.Parent`](../../src/a.go)\n\n| 컬럼 | 타입 | NULL | 키 |\n| --- | --- | --- | --- |\n| `id` | INTEGER | NO | PK |\n| `parent_id` | INTEGER | NO | FK → parent.id |\n| `note` | TEXT | YES |  |\n| `at` | INTEGER | NO |  |\n\n" +
	"| 인덱스 | 컬럼 | UNIQUE | 조건 |\n| --- | --- | --- | --- |\n| `idx_child_note` | `note` | NO | `note IS NOT NULL` |\n| `idx_child_parent_at` | `parent_id`, `at DESC` | NO | - |\n\n" +
	"### `parent`\n\n의미: [`src.Parent`](../../src/a.go)\n\n| 컬럼 | 타입 | NULL | 키 |\n| --- | --- | --- | --- |\n| `id` | INTEGER | NO | PK |\n| `name` | TEXT | NO |  |\n\n" +
	"| 인덱스 | 컬럼 | UNIQUE | 조건 |\n| --- | --- | --- | --- |\n| `UNIQUE(parent.name)` | `name` | YES | - |\n\n" +
	"### `solo`\n\n의미: [`src`](../../src/a.go)\n\n| 컬럼 | 타입 | NULL | 키 |\n| --- | --- | --- | --- |\n| `parent_id` | INTEGER | NO | PK, FK → parent.id |\n"

func writeFixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	base := map[string]string{
		"docs/data-model/README.md":            fixtureReadme,
		"docs/data-model/fullscan-criteria.md": fixtureCriteria,
		"docs/data-model/erd.md":               fixtureERD,
		"m/0001_init.up.sql":                   fixtureMigration,
		"m/0001_init.down.sql":                 "DROP TABLE solo; DROP TABLE child; DROP TABLE parent;\n",
		"chk/check.txt":                        "x\n",
		"src/a.go":                             fixtureSource,
		"src/a_test.go":                        "package src\n\nfunc t() { db.QueryContext(ctx, `SELECT 1`) }\n",
	}
	for k, v := range files {
		if v == "" {
			delete(base, k)
		} else {
			base[k] = v
		}
	}
	for name, body := range base {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return root
}

func kinds(r *report, inv string) map[string]bool {
	out := map[string]bool{}
	for _, v := range r.Invariants[inv].Violations {
		out[v.Kind] = true
	}
	return out
}

func TestConsistentERD(t *testing.T) {
	r := run(writeFixture(t, nil))
	if len(r.Undecidable) != 0 {
		t.Fatalf("undecidable: %v", r.Undecidable)
	}
	if got := r.Invariants["1"].Violations; len(got) != 0 {
		t.Fatalf("invariant 1 violations: %+v", got)
	}
	if r.ExitCode != 1 || !kinds(r, "2")["catalog-missing"] || !kinds(r, "3")["catalog-missing"] {
		t.Fatalf("want exit 1 with catalog-missing, got %d %+v", r.ExitCode, r.Invariants)
	}
	if r.Schema.Tables != 3 || r.Schema.Indexes != 3 || r.Schema.FKs != 2 {
		t.Fatalf("schema summary %+v", r.Schema)
	}
	want := []site{
		{File: "src/a.go", Line: 10, Func: "Store::List", Shape: "select child | - | - | -", SQL: "SELECT id FROM child"},
		{File: "src/a.go", Line: 14, Func: "helper", Shape: "delete parent | - | - | -", SQL: "DELETE FROM parent"},
	}
	if len(r.Sites) != len(want) {
		t.Fatalf("sites %+v", r.Sites)
	}
	for i := range want {
		got, w := r.Sites[i], want[i]
		if got.File != w.File || got.Line != w.Line || got.Func != w.Func || got.Shape != w.Shape || got.SQL != w.SQL || got.Unextractable != "" {
			t.Fatalf("site %d = %+v, want %+v", i, got, w)
		}
	}
}

func TestERDDrift(t *testing.T) {
	cases := []struct {
		name, old, new, kind string
	}{
		{"column type", "| `name` | TEXT | NO |  |", "| `name` | INTEGER | NO |  |", "column-mismatch"},
		{"column null", "| `note` | TEXT | YES |  |", "| `note` | TEXT | NO |  |", "column-mismatch"},
		{"column extra", "| `at` | INTEGER | NO |  |\n", "| `at` | INTEGER | NO |  |\n| `ghost` | TEXT | YES |  |\n", "column-extra"},
		{"column missing", "| `at` | INTEGER | NO |  |\n", "", "column-missing"},
		{"fk dropped", "| `parent_id` | INTEGER | NO | FK → parent.id |", "| `parent_id` | INTEGER | NO |  |", "column-mismatch"},
		{"index label", "| `UNIQUE(parent.name)` |", "| `idx_parent_name` |", "index-label"},
		{"index order", "`parent_id`, `at DESC`", "`at DESC`, `parent_id`", "index-missing"},
		{"partial cond", "`note IS NOT NULL`", "-", "index-missing"},
		{"relation cardinality", "parent ||--o| solo", "parent ||--o{ solo", "diagram-relation-missing"},
		{"relation extra", "    parent ||--o{ child : parent_id\n", "    parent ||--o{ child : parent_id\n    parent ||--o{ child : note\n", "diagram-relation-extra"},
		{"entity extra", "erDiagram\n", "erDiagram\n    ghost\n", "diagram-entity-extra"},
		{"table missing", "### `solo`", "### `solo_old`", "table-missing"},
		{"meaning undocumented", "[`src.Parent`](../../src/a.go)\n\n| 컬럼 | 타입 | NULL | 키 |\n| --- | --- | --- | --- |\n| `id` | INTEGER | NO | PK |\n| `name`", "[`src.Child`](../../src/a.go)\n\n| 컬럼 | 타입 | NULL | 키 |\n| --- | --- | --- | --- |\n| `id` | INTEGER | NO | PK |\n| `name`", "meaning-undocumented"},
		{"meaning broken link", "[`src`](../../src/a.go)", "[`src`](../../src/b.go)", "meaning-link"},
		{"meaning missing", "의미: [`src`](../../src/a.go)\n", "", "meaning-missing"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if strings.Count(fixtureERD, c.old) != 1 {
				t.Fatalf("fixture anchor %q is not unique", c.old)
			}
			r := run(writeFixture(t, map[string]string{"docs/data-model/erd.md": strings.Replace(fixtureERD, c.old, c.new, 1)}))
			if r.ExitCode != 1 || !kinds(r, "1")[c.kind] {
				t.Fatalf("want %s with exit 1, got %d %+v %v", c.kind, r.ExitCode, r.Invariants["1"].Violations, r.Undecidable)
			}
		})
	}
}

func TestUndecidable(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
	}{
		{"no block", map[string]string{"docs/data-model/README.md": "# 데이터 모델\n"}},
		{"broken regexp", map[string]string{"docs/data-model/README.md": strings.Replace(fixtureReadme, `\.(QueryContext|ExecContext)\(`, `\.(QueryContext`, 1)}},
		{"checker path missing", map[string]string{"chk/check.txt": ""}},
		{"migration fails", map[string]string{"m/0002_bad.up.sql": "CREATE TABLE parent (id INTEGER);\n", "m/0002_bad.down.sql": "SELECT 1;\n"}},
		{"unparsable diagram", map[string]string{"docs/data-model/erd.md": strings.Replace(fixtureERD, "erDiagram\n", "erDiagram\n    parent }}--{{ child\n", 1)}},
		{"catalog without pattern table", map[string]string{"docs/data-model/query-patterns.md": "# 쿼리 패턴\n"}},
		{"catalog stray row", map[string]string{"docs/data-model/query-patterns.md": fixtureCatalog + "\n| x | y |\n"}},
		{"criteria file missing", map[string]string{"docs/data-model/fullscan-criteria.md": ""}},
		{"criteria table missing", map[string]string{"docs/data-model/fullscan-criteria.md": strings.Replace(fixtureCriteria, "| ID | 기준 | 관측 가능한 근거 |", "| ID | 기준 |", 1)}},
		{"criteria row malformed", map[string]string{"docs/data-model/fullscan-criteria.md": strings.Replace(fixtureCriteria, "| F1 |", "| X1 |", 1)}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := run(writeFixture(t, c.files))
			if r.ExitCode != 2 || len(r.Undecidable) == 0 {
				t.Fatalf("want exit 2, got %d %v", r.ExitCode, r.Undecidable)
			}
		})
	}
}

func TestSchemaChangeWins(t *testing.T) {
	r := run(writeFixture(t, map[string]string{
		"m/0002_more.up.sql":   "ALTER TABLE parent ADD COLUMN email TEXT;\nCREATE UNIQUE INDEX idx_parent_email ON parent(lower(email));\n",
		"m/0002_more.down.sql": "SELECT 1;\n",
	}))
	k := kinds(r, "1")
	if r.ExitCode != 1 || !k["column-missing"] || !k["index-missing"] {
		t.Fatalf("got %d %+v", r.ExitCode, r.Invariants["1"].Violations)
	}
	for _, v := range r.Invariants["1"].Violations {
		if v.Kind == "index-missing" && v.Expected != "| `idx_parent_email` | `lower(email)` | YES | - |" {
			t.Fatalf("expected row %q", v.Expected)
		}
	}
}

const fixtureCatalog = "# 쿼리 패턴\n\n| ID | 형태 | 지원 인덱스 또는 허용 사유 | 호출 지점 |\n| --- | --- | --- | --- |\n" +
	"| Q-01 | `select child \\| - \\| - \\| -` | 풀스캔 허용(F1): 픽스처 | `src/a.go::Store::List` |\n" +
	"| Q-02 | `delete parent \\| - \\| - \\| -` | 풀스캔 허용(F1): 픽스처 | `src/a.go::helper` |\n" +
	"| Q-03 | `insert child \\| - \\| - \\| -` | — | `src/b.go::Store::Add` |\n" +
	"| Q-04 | `select parent \\| eq(name) \\| - \\| -` | UNIQUE(parent.name) | `src/b.go::Store::Find` |\n" +
	"\n| 호출 지점 | 패턴 ID | 대표 SQL | 추출 불가 사유 |\n| --- | --- | --- | --- |\n" +
	"| `src/b.go::Store::Add` | Q-03 | `INSERT INTO child (parent_id, at) VALUES (?, ?)` | 테이블을 런타임에 고른다 |\n" +
	"\n| 인덱스 | 테이블 | 비고 |\n| --- | --- | --- |\n" +
	"| `idx_child_note` | `child` | - |\n"

var fixtureDynamic = strings.Join([]string{
	"package src",
	"",
	"const cols = `id, name`",
	"",
	"func (s *Store) Add(tbl string) {",
	"\ts.db.ExecContext(ctx, `INSERT INTO `+tbl+` (parent_id, at) VALUES (?, ?)`, 1, 2)",
	"}",
	"",
	"func (s *Store) Find() {",
	"\tstmt, _ := s.db.PrepareContext(ctx, `SELECT `+cols+` FROM parent WHERE name = ?`)",
	"\tstmt.QueryContext(ctx, \"x\")",
	"}",
	"",
}, "\n")

func catalogFixture(t *testing.T, files map[string]string) *report {
	t.Helper()
	base := map[string]string{"docs/data-model/query-patterns.md": fixtureCatalog, "src/b.go": fixtureDynamic}
	for k, v := range files {
		base[k] = v
	}
	return run(writeFixture(t, base))
}

func TestCatalogConsistent(t *testing.T) {
	r := catalogFixture(t, nil)
	if len(r.Undecidable) != 0 {
		t.Fatalf("undecidable: %v", r.Undecidable)
	}
	if r.Invariants["2"].Status != "ok" {
		t.Fatalf("invariant 2: %+v", r.Invariants["2"].Violations)
	}
	if r.ExitCode != 0 || r.Invariants["3"].Status != "ok" {
		t.Fatalf("want exit 0, got %d %+v", r.ExitCode, r.Invariants["3"].Violations)
	}
	sup := map[string]string{}
	for _, p := range r.Plans {
		sup[p.Pattern] = p.Support
	}
	if sup["Q-01"] != noSupport || sup["Q-02"] != noSupport || sup["Q-03"] != noAccess || sup["Q-04"] != "UNIQUE(parent.name)" {
		t.Fatalf("plans %+v", r.Plans)
	}
	var stmt, dyn int
	for _, s := range r.Sites {
		switch {
		case s.File == "src/b.go" && s.Shape == "select parent | eq(name) | - | -":
			stmt++
		case s.File == "src/b.go" && s.Unextractable != "" && strings.Contains(s.Unextractable, "`tbl`"):
			dyn++
		}
	}
	if len(r.Sites) != 4 || stmt != 1 || dyn != 1 {
		t.Fatalf("sites %+v", r.Sites)
	}
}

func TestCatalogDrift(t *testing.T) {
	cases := []struct {
		name, file, old, new, inv, kind string
	}{
		{"pattern row removed", "docs/data-model/query-patterns.md", "| Q-02 | `delete parent \\| - \\| - \\| -` | 풀스캔 허용(F1): 픽스처 | `src/a.go::helper` |\n", "", "2", "query-unregistered"},
		{"site on wrong row", "docs/data-model/query-patterns.md", "`src/a.go::Store::List` |", "`src/a.go::Store::List`, `src/a.go::helper` |", "2", "dead-site"},
		{"dead pattern", "docs/data-model/query-patterns.md", "| Q-02 | `delete parent \\| - \\| - \\| -` | 풀스캔 허용(F1): 픽스처 | `src/a.go::helper` |\n", "| Q-02 | `delete parent \\| - \\| - \\| -` | 풀스캔 허용(F1): 픽스처 | `src/a.go::helper` |\n| Q-05 | `delete child \\| - \\| - \\| -` | 지원 없음 | `src/a.go::gone` |\n", "2", "dead-pattern"},
		{"ambiguous shape", "docs/data-model/query-patterns.md", "| Q-02 | `delete parent \\| - \\| - \\| -` | 풀스캔 허용(F1): 픽스처 | `src/a.go::helper` |\n", "| Q-02 | `delete parent \\| - \\| - \\| -` | 풀스캔 허용(F1): 픽스처 | `src/a.go::helper` |\n| Q-05 | `delete parent \\| - \\| - \\| -` | 풀스캔 허용(F1): 픽스처 | `src/a.go::helper` |\n", "2", "ambiguous-shape"},
		{"shape drift", "docs/data-model/query-patterns.md", "`select parent \\| eq(name) \\| - \\| -`", "`select parent \\| eq(id) \\| - \\| -`", "2", "query-unregistered"},
		{"malformed shape", "docs/data-model/query-patterns.md", "`select child \\| - \\| - \\| -`", "`select child`", "2", "shape-malformed"},
		{"bad id", "docs/data-model/query-patterns.md", "| Q-02 |", "| P2 |", "2", "pattern-id"},
		{"manual row removed", "docs/data-model/query-patterns.md", "| `src/b.go::Store::Add` | Q-03 | `INSERT INTO child (parent_id, at) VALUES (?, ?)` | 테이블을 런타임에 고른다 |\n", "", "2", "manual-missing"},
		{"manual row extra", "docs/data-model/query-patterns.md", "| `src/b.go::Store::Add` | Q-03 |", "| `src/a.go::helper` | Q-02 | `DELETE FROM parent` | x |\n| `src/b.go::Store::Add` | Q-03 |", "2", "manual-extra"},
		{"manual shape mismatch", "docs/data-model/query-patterns.md", "`INSERT INTO child (parent_id, at) VALUES (?, ?)`", "`INSERT INTO parent (name) VALUES (?)`", "2", "manual-shape-mismatch"},
		{"manual unknown pattern", "docs/data-model/query-patterns.md", "| `src/b.go::Store::Add` | Q-03 |", "| `src/b.go::Store::Add` | Q-09 |", "2", "manual-pattern-unknown"},
		{"manual sql broken", "docs/data-model/query-patterns.md", "`INSERT INTO child (parent_id, at) VALUES (?, ?)`", "`SELECT 1 UNION SELECT 2`", "2", "manual-sql"},
		{"new query in code", "src/b.go", "func (s *Store) Find() {", "func (s *Store) Drop() {\n\ts.db.ExecContext(ctx, `DELETE FROM child WHERE parent_id = ?`, 1)\n}\n\nfunc (s *Store) Find() {", "2", "query-unregistered"},
		{"code query became dynamic", "src/a.go", "\tdb.ExecContext(ctx, `DELETE FROM parent`)", "\tdb.ExecContext(ctx, q)", "2", "manual-missing"},
		{"support dash on read", "docs/data-model/query-patterns.md", "| UNIQUE(parent.name) |", "| — |", "3", "support-mismatch"},
		{"support missing dash on write", "docs/data-model/query-patterns.md", "| Q-03 | `insert child \\| - \\| - \\| -` | — |", "| Q-03 | `insert child \\| - \\| - \\| -` | PK(child) |", "3", "support-mismatch"},
		{"support names other index", "docs/data-model/query-patterns.md", "| UNIQUE(parent.name) |", "| PK(parent) |", "3", "support-mismatch"},
		{"index cell but engine scans", "docs/data-model/query-patterns.md", "`select child \\| - \\| - \\| -` | 풀스캔 허용(F1): 픽스처 |", "`select child \\| - \\| - \\| -` | idx_child_parent_at |", "3", "support-mismatch"},
		{"full scan allowed but engine uses index", "docs/data-model/query-patterns.md", "| UNIQUE(parent.name) |", "| 풀스캔 허용(F1): 픽스처 |", "3", "support-mismatch"},
		{"no support documented", "docs/data-model/query-patterns.md", "`select child \\| - \\| - \\| -` | 풀스캔 허용(F1): 픽스처 |", "`select child \\| - \\| - \\| -` | 지원 없음 |", "3", "no-support"},
		{"unknown criterion", "docs/data-model/query-patterns.md", "`select child \\| - \\| - \\| -` | 풀스캔 허용(F1): 픽스처 |", "`select child \\| - \\| - \\| -` | 풀스캔 허용(F9): 픽스처 |", "3", "criterion-unknown"},
		{"unjudged cell", "docs/data-model/query-patterns.md", "| UNIQUE(parent.name) |", "| 미판정 |", "3", "support-mismatch"},
		{"unused row removed", "docs/data-model/query-patterns.md", "| `idx_child_note` | `child` | - |\n", "", "3", "unused-unregistered"},
		{"used index listed as unused", "docs/data-model/query-patterns.md", "| `idx_child_note` | `child` | - |\n", "| `idx_child_note` | `child` | - |\n| `UNIQUE(parent.name)` | `parent` | - |\n", "3", "unused-stale"},
		{"unused table removed", "docs/data-model/query-patterns.md", "\n| 인덱스 | 테이블 | 비고 |\n| --- | --- | --- |\n| `idx_child_note` | `child` | - |\n", "", "3", "unused-table-missing"},
		{"same shape different plans", "src/b.go", "func (s *Store) Find() {", "func (s *Store) FindCI() {\n\ts.db.QueryContext(ctx, `SELECT id FROM parent WHERE name = ? COLLATE NOCASE`, 1)\n}\n\nfunc (s *Store) Find() {", "3", "split-plan"},
		{"manual sql does not plan", "docs/data-model/query-patterns.md", "`INSERT INTO child (parent_id, at) VALUES (?, ?)`", "`INSERT INTO child (parent_id, nope) VALUES (?, ?)`", "3", "plan-error"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := map[string]string{"docs/data-model/query-patterns.md": fixtureCatalog, "src/a.go": fixtureSource, "src/b.go": fixtureDynamic}[c.file]
			if strings.Count(src, c.old) != 1 {
				t.Fatalf("fixture anchor %q is not unique", c.old)
			}
			r := catalogFixture(t, map[string]string{c.file: strings.Replace(src, c.old, c.new, 1)})
			if r.ExitCode != 1 || !kinds(r, c.inv)[c.kind] {
				t.Fatalf("want %s in invariant %s with exit 1, got %d %+v %+v %v", c.kind, c.inv, r.ExitCode, r.Invariants["2"].Violations, r.Invariants["3"].Violations, r.Undecidable)
			}
		})
	}
}

func TestShapes(t *testing.T) {
	cases := []struct{ sql, want string }{
		{"SELECT id FROM a WHERE user_id = ? ORDER BY created_at DESC, id DESC", "select a | eq(user_id) | - | order(created_at desc, id desc)"},
		{"SELECT * FROM a WHERE user_id = ? AND id < ? ORDER BY id DESC LIMIT ?", "select a | eq(user_id) | range(id) | order(id desc)"},
		{"SELECT * FROM a WHERE user_id = ? AND (? = 0 OR id = ?) ORDER BY id", "select a | eq(user_id) | - | order(id asc)"},
		{"SELECT * FROM a WHERE x = ? OR y = ?", "select a | - | - | -"},
		{"SELECT s.id FROM b s JOIN a r ON r.id = s.a_id WHERE r.user_id = ? ORDER BY s.a_id, s.pos", "select a,b | eq(a.id=b.a_id, a.user_id) | - | order(b.a_id asc, b.pos asc)"},
		{"SELECT c.day, COUNT(*) FROM c JOIN b s ON s.id = c.b_id WHERE s.a_id = ? AND c.day BETWEEN ? AND ? GROUP BY c.day", "select b,c | eq(b.a_id, b.id=c.b_id) | range(c.day) | -"},
		{"SELECT id FROM u WHERE id NOT IN (SELECT user_id FROM x WHERE repo = ?) ORDER BY id", "select u,x | eq(x.repo) | - | order(u.id asc)"},
		{"DELETE FROM b WHERE id = ? AND a_id IN (SELECT id FROM a WHERE id = ? AND user_id = ?)", "delete a,b | eq(a.id, a.user_id, b.a_id, b.id) | - | -"},
		{"INSERT INTO b (a_id, pos, t) SELECT r.id, COALESCE((SELECT MAX(pos) + 1 FROM b WHERE a_id = r.id), 0), ? FROM a r WHERE r.id = ? AND r.user_id = ?", "insert a,b | eq(a.id, a.id=b.a_id, a.user_id) | - | -"},
		{"INSERT INTO t (user_id, token) VALUES (?, ?) ON CONFLICT(token) DO UPDATE SET user_id = excluded.user_id", "upsert t | eq(token) | - | -"},
		{"INSERT OR IGNORE INTO c (b_id, day) VALUES (?, ?)", "insert c | - | - | -"},
		{"INSERT OR REPLACE INTO c (b_id, day) VALUES (?, ?)", "upsert c | - | - | -"},
		{"REPLACE INTO c (b_id, day) VALUES (?, ?)", "upsert c | - | - | -"},
		{"UPDATE h SET at = CASE WHEN at IS NULL THEN unixepoch() ELSE at END WHERE id = ? AND user_id = ?", "update h | eq(id, user_id) | - | -"},
		{"SELECT * FROM u WHERE lower(email) = ? AND deleted_at IS NULL AND name != ? AND note LIKE ?", "select u | eq(deleted_at, lower(email)) | range(note) | -"},
		{"SELECT * FROM t ORDER BY done_at IS NOT NULL, due IS NULL, due", "select t | - | - | order(done_at is not null asc, due is null asc, due asc)"},
		{"SELECT \"Id\" FROM \"T\" WHERE \"User_Id\" = ?", "select t | eq(user_id) | - | -"},
	}
	for _, c := range cases {
		got, err := extractShape(c.sql, nil)
		if err != nil {
			t.Fatalf("%s: %v", c.sql, err)
		}
		if got.String() != c.want {
			t.Fatalf("%s:\n got %s\nwant %s", c.sql, got.String(), c.want)
		}
	}
	for _, bad := range []string{"SELECT 1 UNION SELECT 2", "WITH x AS (SELECT 1) SELECT * FROM x", "SELECT * FROM a; SELECT * FROM b", "PRAGMA foo", "SELECT a.x FROM a JOIN b ON a.id = b.id WHERE y = ?"} {
		if _, err := extractShape(bad, nil); err == nil {
			t.Fatalf("%s: want error", bad)
		}
	}
}

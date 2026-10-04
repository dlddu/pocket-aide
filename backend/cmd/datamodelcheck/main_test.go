package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const fixtureReadme = "# 데이터 모델\n\n```data-model-scope\nmigrations: m\nchecker: chk\nscope: src\nexclude: _test\\.go$\nsite: \\.(QueryContext|ExecContext)\\(\nschema-exclude: ^(schema_migrations|sqlite_sequence)$\nsql: \\bFROM [a-z_]+\\b\n```\n"

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
		"docs/data-model/README.md": fixtureReadme,
		"docs/data-model/erd.md":    fixtureERD,
		"m/0001_init.up.sql":        fixtureMigration,
		"m/0001_init.down.sql":      "DROP TABLE solo; DROP TABLE child; DROP TABLE parent;\n",
		"chk/check.txt":             "x\n",
		"src/a.go":                  fixtureSource,
		"src/a_test.go":             "package src\n\nfunc t() { db.QueryContext(ctx, `SELECT 1`) }\n",
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
	want := []site{{"src/a.go", 10, "Store::List"}, {"src/a.go", 14, "helper"}}
	if len(r.Sites) != len(want) {
		t.Fatalf("sites %+v", r.Sites)
	}
	for i := range want {
		if r.Sites[i] != want[i] {
			t.Fatalf("site %d = %+v, want %+v", i, r.Sites[i], want[i])
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
		{"catalog present", map[string]string{"docs/data-model/query-patterns.md": "# 쿼리 패턴\n"}},
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

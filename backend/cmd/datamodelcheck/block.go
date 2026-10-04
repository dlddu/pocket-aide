package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	docDir     = "docs/data-model"
	readmePath = docDir + "/README.md"
	erdPath    = docDir + "/erd.md"
	qpPath     = docDir + "/query-patterns.md"
	fence      = "```data-model-scope"
)

type scopeBlock struct {
	Migrations    string   `json:"migrations"`
	Checker       string   `json:"checker"`
	Scope         []string `json:"scope"`
	Exclude       string   `json:"exclude"`
	Site          string   `json:"site"`
	SQL           string   `json:"sql"`
	SchemaExclude string   `json:"schema_exclude"`

	exclude       *regexp.Regexp
	site          *regexp.Regexp
	schemaExclude *regexp.Regexp
}

func readBlock(path string) (*scopeBlock, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("README 를 읽지 못했다: %w", err)
	}
	defer func() { _ = f.Close() }()

	vals := map[string][]string{}
	in, found := false, false
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case !in && !found && (line == fence || strings.HasPrefix(line, fence+" ")):
			in, found = true, true
		case in && strings.HasPrefix(line, "```"):
			in = false
		case in && strings.TrimSpace(line) != "":
			key, val, ok := strings.Cut(line, ":")
			if !ok {
				continue
			}
			vals[key] = append(vals[key], strings.TrimLeft(val, " \t"))
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("README 를 읽지 못했다: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("%s 에 %s 블록이 없다", readmePath, fence)
	}

	first := func(k string) string {
		if len(vals[k]) == 0 {
			return ""
		}
		return vals[k][0]
	}
	joined := func(k string) string { return strings.Join(vals[k], "|") }

	b := &scopeBlock{
		Migrations:    first("migrations"),
		Checker:       first("checker"),
		Scope:         strings.Fields(strings.Join(vals["scope"], " ")),
		Exclude:       joined("exclude"),
		Site:          joined("site"),
		SQL:           joined("sql"),
		SchemaExclude: joined("schema-exclude"),
	}
	if b.Exclude == "" {
		b.Exclude = "^$"
	}
	if b.SchemaExclude == "" {
		b.SchemaExclude = "^$"
	}
	for _, k := range []struct{ key, val string }{
		{"migrations", b.Migrations}, {"checker", b.Checker}, {"site", b.Site}, {"sql", b.SQL},
	} {
		if k.val == "" {
			return nil, fmt.Errorf("블록 키 %s 가 비어 있다", k.key)
		}
	}
	if len(b.Scope) == 0 {
		return nil, fmt.Errorf("블록 키 scope 가 비어 있다")
	}
	for _, r := range []struct {
		key string
		src string
		dst **regexp.Regexp
	}{
		{"exclude", b.Exclude, &b.exclude},
		{"site", b.Site, &b.site},
		{"sql", b.SQL, nil},
		{"schema-exclude", b.SchemaExclude, &b.schemaExclude},
	} {
		re, err := regexp.Compile(r.src)
		if err != nil {
			return nil, fmt.Errorf("블록 키 %s 의 정규식이 깨졌다: %w", r.key, err)
		}
		if r.dst != nil {
			*r.dst = re
		}
	}
	return b, nil
}

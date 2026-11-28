// Reading the few things a reviewer needs out of what the author actually
// wrote (PRD-AI채팅 AC10, 검토 자료).
//
// The agreed turn (platform AC-N9) has no column for "this was decided" and
// none for an alternative's advantages and drawbacks, and that schema is not
// ours to extend. But an author who has weighed two options does write them
// down — in the turn's own `response`, or in the parent's `explanation`. So this
// reads those two fields, and only where the author marked what they were
// writing.
//
// "Marked" is the whole point. Two shapes count, both of which a person has to
// choose deliberately:
//
//  1. a Markdown table whose header names 대안 · 장점 · 단점, and
//  2. a 「대안별 장단점」 section whose items carry 장점/단점 labels.
//
// Decisions are read the same way, from a 「결정 사항」 section and nowhere else.
//
// Everything outside those markers stays unread. A paragraph of explanation that
// happens to sound conclusive is not a decision; a sentence on a choice button
// is not a trade-off. Guessing either would put a position in the author's mouth
// that they never took, which is worse for a reviewer than an honest blank.
package chat

import "strings"

// tradeoffHeadings start a 「대안별 장단점」 section. The bare 「장단점」 is
// accepted because authors write it, and both are explicit enough that nothing
// is being inferred from ordinary prose.
var tradeoffHeadings = []string{"대안별 장단점", "장단점"}

// decisionHeadings start a 「결정 사항」 section. Only a heading that says the
// author is recording decisions counts — never a paragraph that reads like one.
var decisionHeadings = []string{"결정 사항", "결정사항", "확정 사항", "확정사항"}

// proLabels and conLabels mark one side of an alternative inside a section.
var (
	proLabels = []string{"장점", "좋은 점", "pros", "pro"}
	conLabels = []string{"단점", "아쉬운 점", "cons", "con"}
)

// tradeoffColumns are the header cells a table has to name to be read as a
// trade-off table. The alternative column accepts the words authors use for it.
var (
	altColumns = []string{"대안", "선택지", "안", "옵션", "option", "alternative"}
	proColumns = proLabels
	conColumns = conLabels
)

// materialsFromText is everything the two marked shapes yield from one field's
// text, tagged with where it was written.
type materialsFromText struct {
	tradeoffs []TaskTradeoff
	decisions []TaskDecision
}

// readMaterialsText reads one turn field. source is the field's own name
// ("response" or "explanation") so a reviewer can see which one said it.
func readMaterialsText(text string, turnID int64, source string) materialsFromText {
	out := materialsFromText{}
	if strings.TrimSpace(text) == "" {
		return out
	}
	lines := strings.Split(text, "\n")
	for _, block := range markedSections(lines) {
		switch block.kind {
		case sectionTradeoffs:
			out.tradeoffs = append(out.tradeoffs, tradeoffsOfSection(block.lines, turnID, source)...)
		case sectionDecisions:
			out.decisions = append(out.decisions, decisionsOfSection(block.lines, turnID, source)...)
		}
	}
	// A trade-off table is explicit on its own — its header says what every
	// column is — so it is read whether or not somebody put a heading above it.
	// A table found inside a section above is already accounted for and is not
	// read twice.
	for _, table := range looseTradeoffTables(lines, markedSectionRanges(lines)) {
		out.tradeoffs = append(out.tradeoffs, tradeoffsOfTable(table, turnID, source)...)
	}
	return out
}

type sectionKind int

const (
	sectionNone sectionKind = iota
	sectionTradeoffs
	sectionDecisions
)

type markedSection struct {
	kind  sectionKind
	lines []string
	from  int
	to    int
}

// headingKind is what this line declares, if it declares anything. A heading is
// a line whose own text is the marker — `## 대안별 장단점`, `**결정 사항**`,
// `결정 사항:` — rather than a sentence that merely contains the words, so a
// paragraph mentioning trade-offs does not open a section.
func headingKind(line string) sectionKind {
	bare := strings.TrimSpace(line)
	bare = strings.TrimLeft(bare, "#>-*0123456789.) \t")
	bare = strings.Trim(bare, "*_`:：[]()【】「」 \t")
	bare = strings.TrimSpace(bare)
	if bare == "" {
		return sectionNone
	}
	for _, h := range decisionHeadings {
		if strings.EqualFold(bare, h) {
			return sectionDecisions
		}
	}
	for _, h := range tradeoffHeadings {
		if strings.EqualFold(bare, h) {
			return sectionTradeoffs
		}
	}
	return sectionNone
}

// markedSections splits the text into the marked sections it contains. A
// section runs to the next marker or the next Markdown heading, so an author's
// following section is not swallowed into this one.
func markedSections(lines []string) []markedSection {
	out := []markedSection{}
	for i := 0; i < len(lines); i++ {
		kind := headingKind(lines[i])
		if kind == sectionNone {
			continue
		}
		end := len(lines)
		for j := i + 1; j < len(lines); j++ {
			if headingKind(lines[j]) != sectionNone || isMarkdownHeading(lines[j]) {
				end = j
				break
			}
		}
		out = append(out, markedSection{kind: kind, lines: lines[i+1 : end], from: i, to: end})
		i = end - 1
	}
	return out
}

func markedSectionRanges(lines []string) [][2]int {
	out := [][2]int{}
	for _, s := range markedSections(lines) {
		out = append(out, [2]int{s.from, s.to})
	}
	return out
}

func isMarkdownHeading(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

// decisionsOfSection reads the items under a 「결정 사항」 heading. Each item is
// one decision, kept as the author's own words — this never summarises.
func decisionsOfSection(lines []string, turnID int64, source string) []TaskDecision {
	out := []TaskDecision{}
	for _, raw := range lines {
		text := strings.TrimSpace(raw)
		if text == "" || isTableRule(text) {
			continue
		}
		text = strings.TrimSpace(strings.TrimLeft(text, "-*0123456789.) \t"))
		text = strings.TrimSpace(strings.Trim(text, "|"))
		if text == "" {
			continue
		}
		out = append(out, TaskDecision{Text: text, TurnID: turnID, Source: source})
	}
	return out
}

// tradeoffsOfSection reads a 「대안별 장단점」 section: a table if it holds one,
// otherwise items whose alternative is named and whose sides are labelled.
func tradeoffsOfSection(lines []string, turnID int64, source string) []TaskTradeoff {
	if table := tableIn(lines); table != nil {
		return tradeoffsOfTable(table, turnID, source)
	}
	out := []TaskTradeoff{}
	var current *TaskTradeoff
	flush := func() {
		// An alternative with neither side written is not a trade-off — it is a
		// name. Keeping it would show an empty row that reads as "no advantages
		// and no drawbacks", which the records do not say.
		if current != nil && (len(current.Pros) > 0 || len(current.Cons) > 0) {
			out = append(out, *current)
		}
		current = nil
	}
	for _, raw := range lines {
		text := strings.TrimSpace(raw)
		if text == "" {
			continue
		}
		item := strings.TrimSpace(strings.TrimLeft(text, "-*0123456789.) \t"))
		if item == "" {
			continue
		}
		if label, value, ok := labelled(item, proLabels); ok {
			if current != nil && value != "" {
				current.Pros = append(current.Pros, value)
			}
			_ = label
			continue
		}
		if label, value, ok := labelled(item, conLabels); ok {
			if current != nil && value != "" {
				current.Cons = append(current.Cons, value)
			}
			_ = label
			continue
		}
		// Not a side, so it names the next alternative.
		flush()
		name := strings.TrimSpace(strings.Trim(item, ":："))
		if name == "" {
			continue
		}
		current = &TaskTradeoff{Alternative: name, TurnID: turnID, Source: source}
	}
	flush()
	return out
}

// labelled splits `장점: 빠르다` or `장점 — 빠르다` into its label and value.
// The label has to start the line: a sentence that mentions 장점 in the middle
// is prose, not a marked side.
func labelled(item string, labels []string) (string, string, bool) {
	for _, label := range labels {
		if len(item) < len(label) || !strings.EqualFold(item[:len(label)], label) {
			continue
		}
		rest := strings.TrimSpace(item[len(label):])
		for _, sep := range []string{":", "：", "—", "-", "–", "="} {
			if strings.HasPrefix(rest, sep) {
				return label, strings.TrimSpace(strings.TrimPrefix(rest, sep)), true
			}
		}
		// `장점` on its own line labels what follows on the same line after a
		// space, and nothing else.
		if rest != "" {
			return label, rest, true
		}
		return label, "", true
	}
	return "", "", false
}

// --- tables ---

type mdTable struct {
	header []string
	rows   [][]string
}

func isTableRow(line string) bool {
	t := strings.TrimSpace(line)
	return strings.HasPrefix(t, "|") && strings.Count(t, "|") >= 2
}

// isTableRule is the `| --- | --- |` line under a header.
func isTableRule(line string) bool {
	t := strings.TrimSpace(line)
	if !isTableRow(t) {
		return false
	}
	for _, cell := range splitRow(t) {
		trimmed := strings.Trim(cell, ": \t")
		if trimmed == "" || strings.Trim(trimmed, "-") != "" {
			return false
		}
	}
	return true
}

func splitRow(line string) []string {
	t := strings.TrimSpace(line)
	t = strings.TrimPrefix(t, "|")
	t = strings.TrimSuffix(t, "|")
	parts := strings.Split(t, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

// tableIn is the first trade-off table in these lines, or nil. A table is only
// read when its header names the alternative column and at least one side —
// that header is what makes the columns' meaning the author's statement rather
// than this code's assumption.
func tableIn(lines []string) *mdTable {
	for i := 0; i < len(lines); i++ {
		if !isTableRow(lines[i]) || isTableRule(lines[i]) {
			continue
		}
		header := splitRow(lines[i])
		if !namesTradeoffColumns(header) {
			continue
		}
		table := &mdTable{header: header}
		for j := i + 1; j < len(lines); j++ {
			if !isTableRow(lines[j]) {
				break
			}
			if isTableRule(lines[j]) {
				continue
			}
			table.rows = append(table.rows, splitRow(lines[j]))
		}
		return table
	}
	return nil
}

// looseTradeoffTables are trade-off tables outside every marked section, so a
// table that is its own marker is read without a heading above it.
func looseTradeoffTables(lines []string, skip [][2]int) []*mdTable {
	inSkip := func(i int) bool {
		for _, r := range skip {
			if i >= r[0] && i < r[1] {
				return true
			}
		}
		return false
	}
	out := []*mdTable{}
	for i := 0; i < len(lines); i++ {
		if inSkip(i) || !isTableRow(lines[i]) || isTableRule(lines[i]) {
			continue
		}
		header := splitRow(lines[i])
		if !namesTradeoffColumns(header) {
			continue
		}
		table := &mdTable{header: header}
		end := i + 1
		for j := i + 1; j < len(lines); j++ {
			if !isTableRow(lines[j]) {
				break
			}
			end = j + 1
			if isTableRule(lines[j]) {
				continue
			}
			table.rows = append(table.rows, splitRow(lines[j]))
		}
		out = append(out, table)
		i = end - 1
	}
	return out
}

func columnIndex(header []string, names []string) int {
	for i, cell := range header {
		bare := strings.TrimSpace(strings.Trim(cell, "*_`"))
		for _, n := range names {
			if strings.EqualFold(bare, n) {
				return i
			}
		}
	}
	return -1
}

func namesTradeoffColumns(header []string) bool {
	if columnIndex(header, altColumns) < 0 {
		return false
	}
	return columnIndex(header, proColumns) >= 0 || columnIndex(header, conColumns) >= 0
}

// tradeoffsOfTable reads one row per alternative. A column the table does not
// have is left unrecorded rather than filled from the other one.
func tradeoffsOfTable(table *mdTable, turnID int64, source string) []TaskTradeoff {
	alt := columnIndex(table.header, altColumns)
	pro := columnIndex(table.header, proColumns)
	con := columnIndex(table.header, conColumns)
	out := []TaskTradeoff{}
	for _, row := range table.rows {
		if alt >= len(row) {
			continue
		}
		name := strings.TrimSpace(strings.Trim(row[alt], "*_`"))
		if name == "" {
			continue
		}
		entry := TaskTradeoff{Alternative: name, TurnID: turnID, Source: source}
		entry.Pros = cellValues(row, pro)
		entry.Cons = cellValues(row, con)
		if len(entry.Pros) == 0 && len(entry.Cons) == 0 {
			continue
		}
		out = append(out, entry)
	}
	return out
}

// cellValues is one cell split on the line breaks authors write inside a table
// cell. An em dash or a dash alone is how a table says "none here", so it reads
// as unrecorded rather than as a value.
func cellValues(row []string, at int) []string {
	if at < 0 || at >= len(row) {
		return nil
	}
	cell := strings.TrimSpace(row[at])
	if cell == "" {
		return nil
	}
	cell = strings.ReplaceAll(cell, "<br/>", "\n")
	cell = strings.ReplaceAll(cell, "<br />", "\n")
	cell = strings.ReplaceAll(cell, "<br>", "\n")
	out := []string{}
	for _, part := range strings.Split(cell, "\n") {
		value := strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(part), "-*· \t"))
		if value == "" || value == "—" || value == "–" {
			continue
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

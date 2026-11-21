package chat

import (
	"strings"
	"testing"
)

// The transcript arrives in arbitrary pieces; marker lines split across
// reads still frame exactly one turn, and text outside it is not the reply.
func TestTurnFollowerAcrossSplitReads(t *testing.T) {
	transcript := "[session-platform: plugin skills refreshed]\n" +
		"\x1esession-platform input seq=4 at=x\nhello\n\x1esession-platform output seq=4\n" +
		"device code: ABCD\nanswer\n" +
		"\x1esession-platform end seq=4 status=ok at=x\n" +
		"\x1esession-platform input seq=5 at=x\nlater\n"
	for size := 1; size <= len(transcript); size++ {
		f := &turnFollower{base: 100}
		var emitted strings.Builder
		done := false
		for i := 0; i < len(transcript) && !done; i += size {
			end := min(i+size, len(transcript))
			done = f.feed([]byte(transcript[i:end]), func(s string) { emitted.WriteString(s) })
		}
		if !done {
			t.Fatalf("chunk %d: the turn never ended", size)
		}
		if got := f.reply.String(); got != "device code: ABCD\nanswer\n" || emitted.String() != got {
			t.Fatalf("chunk %d: reply %q emitted %q", size, got, emitted.String())
		}
		wantEnd := int64(100 + strings.Index(transcript, "\x1esession-platform input seq=5"))
		if f.end != wantEnd || f.startedAt != int64(100+strings.Index(transcript, "\x1esession-platform input seq=4")) {
			t.Fatalf("chunk %d: end %d startedAt %d", size, f.end, f.startedAt)
		}
	}
}

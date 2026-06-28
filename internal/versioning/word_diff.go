package versioning

import (
	"fmt"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

// ErrWordDiffUnsupported is returned when the active versioner cannot produce
// word-level diffs.
var ErrWordDiffUnsupported = fmt.Errorf("word-level diff not supported")

// WordDiffText returns a unified diff with one token per line for word-level
// comparison. Used by CoW and other non-git backends.
func WordDiffText(from, to, fromLabel, toLabel string) (string, error) {
	a := tokenizeWords(from)
	b := tokenizeWords(to)
	ud := difflib.UnifiedDiff{
		A:        a,
		B:        b,
		FromFile: fromLabel,
		ToFile:   toLabel,
		Context:  1,
	}
	return difflib.GetUnifiedDiffString(ud)
}

func tokenizeWords(s string) []string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil
	}
	out := make([]string, len(fields))
	copy(out, fields)
	return out
}

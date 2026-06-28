package memory

import (
	"fmt"
	"io"
	"strings"
	"time"
)

func parseFrontmatterDate(fm map[string]any, key string) (time.Time, bool) {
	val, ok := fm[key]
	if !ok {
		return time.Time{}, false
	}
	switch v := val.(type) {
	case string:
		for _, layout := range []string{"2006-01-02", time.RFC3339, "2006-01-02T15:04:05Z"} {
			if t, err := time.Parse(layout, v); err == nil {
				return t, true
			}
		}
	case time.Time:
		return v, true
	}
	return time.Time{}, false
}

func coveragePercent(totalEpisodic, totalUnmerged int) float64 {
	if totalEpisodic == 0 {
		return 0
	}
	merged := totalEpisodic - totalUnmerged
	return float64(merged) / float64(totalEpisodic) * 100
}

// WriteHealthMetrics prints coverage, freshness, and scope summary lines.
func (r *Report) WriteHealthMetrics(w io.Writer) {
	fmt.Fprintf(w, "coverage:                %.1f%%\n", r.CoveragePct)
	fmt.Fprintf(w, "avg age (active pages):  %.1f days\n", r.AvgAgeDays)
	fmt.Fprintf(w, "expired pages:           %d\n", r.ExpiredCount)
	fmt.Fprintf(w, "contested pages:         %d\n", r.ContestedCount)
	fmt.Fprintf(w, "contradictions:          %d\n", r.Contradictions)
	if len(r.ScopeCounts) == 0 {
		fmt.Fprintln(w, "scope breakdown:         (none)")
		return
	}
	fmt.Fprintln(w, "scope breakdown:")
	keys := scopeCountKeys(r.ScopeCounts)
	for _, k := range keys {
		label := k
		if label == "" {
			label = "(empty)"
		}
		fmt.Fprintf(w, "  %s: %d\n", label, r.ScopeCounts[k])
	}
}

func scopeCountKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Stable sort for CLI/MCP output.
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if strings.Compare(keys[i], keys[j]) > 0 {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

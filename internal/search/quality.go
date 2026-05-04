package search

func QualityScore(fm map[string]any, daysSinceUpdate float64) float64 {
	score := 0.0
	if wordCount(fm) >= 100 {
		score += 0.20
	}
	if hasFrontmatter(fm) {
		score += 0.20
	}
	if linkCount(fm) > 0 {
		score += 0.15
	}
	if backlinkCount(fm) > 0 {
		score += 0.15
	}
	if daysSinceUpdate >= 0 && daysSinceUpdate < 90 {
		score += 0.15
	}
	if headingCount(fm) > 0 {
		score += 0.10
	}
	if hasField(fm, "tags") {
		score += 0.05
	}
	return score
}

func wordCount(fm map[string]any) int {
	return toInt(fm["_word_count"])
}

func linkCount(fm map[string]any) int {
	return toInt(fm["_link_count"])
}

func backlinkCount(fm map[string]any) int {
	return toInt(fm["_backlink_count"])
}

func headingCount(fm map[string]any) int {
	return toInt(fm["_heading_count"])
}

func hasFrontmatter(fm map[string]any) bool {
	if v, ok := fm["_has_frontmatter"]; ok {
		switch b := v.(type) {
		case bool:
			return b
		case float64:
			return b != 0
		case int:
			return b != 0
		}
	}
	return false
}

func hasField(fm map[string]any, key string) bool {
	_, ok := fm[key]
	return ok
}

func toInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case int32:
		return int(n)
	}
	return 0
}

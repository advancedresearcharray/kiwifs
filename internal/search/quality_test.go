package search

import "testing"

func TestQualityScoreFullPage(t *testing.T) {
	fm := map[string]any{
		"_word_count":      200,
		"_has_frontmatter": true,
		"_link_count":      5,
		"_backlink_count":  3,
		"_heading_count":   4,
		"tags":             []string{"go", "test"},
	}
	score := QualityScore(fm, 10)
	if score < 0.95 {
		t.Fatalf("expected score near 1.0 for fully-featured page, got %.2f", score)
	}
}

func TestQualityScoreEmptyPage(t *testing.T) {
	fm := map[string]any{}
	score := QualityScore(fm, 200)
	if score > 0.1 {
		t.Fatalf("expected score near 0.0 for empty page, got %.2f", score)
	}
}

func TestQualityScorePartial(t *testing.T) {
	fm := map[string]any{
		"_word_count":      150,
		"_has_frontmatter": true,
		"_link_count":      0,
		"_backlink_count":  0,
		"_heading_count":   2,
	}
	score := QualityScore(fm, 30)
	if score < 0.4 || score > 0.8 {
		t.Fatalf("expected partial score 0.4-0.8, got %.2f", score)
	}
}

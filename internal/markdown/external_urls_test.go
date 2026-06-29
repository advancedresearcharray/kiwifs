package markdown

import "testing"

func TestExtractExternalURLs_SkipsCodeBlocksAndInlineCode(t *testing.T) {
	body := `See https://good.example.com/page for docs.

` + "```" + `
const u = "https://ignored-in-fence.example.com";
` + "```" + `

Also check ` + "`https://ignored-inline.example.com`" + ` and [link](https://linked.example.com/path).
`
	got := ExtractExternalURLs(body)
	want := map[string]bool{
		"https://good.example.com/page":    true,
		"https://linked.example.com/path":  true,
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want keys %v", got, want)
	}
	for _, u := range got {
		if !want[u] {
			t.Fatalf("unexpected url %q in %v", u, got)
		}
	}
}

func TestExtractExternalURLs_AngleLinks(t *testing.T) {
	body := "Docs at <https://docs.example.com/guide>."
	got := ExtractExternalURLs(body)
	if len(got) != 1 || got[0] != "https://docs.example.com/guide" {
		t.Fatalf("got %v", got)
	}
}

func TestExtractExternalURLs_TrimsTrailingPunctuation(t *testing.T) {
	body := "Visit https://example.com/page)."
	got := ExtractExternalURLs(body)
	if len(got) != 1 || got[0] != "https://example.com/page" {
		t.Fatalf("got %v", got)
	}
}

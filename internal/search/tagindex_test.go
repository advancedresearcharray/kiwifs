package search

import (
	"sort"
	"testing"
)

func TestTagIndex_UpdateAndQuery(t *testing.T) {
	idx := NewTagIndex()
	idx.Update("pages/auth.md", []string{"security", "core"})
	idx.Update("pages/payments.md", []string{"billing", "core"})

	got := idx.ByTag("core")
	sort.Strings(got)
	if len(got) != 2 {
		t.Fatalf("ByTag(core) = %v, want 2 results", got)
	}

	tags := idx.TagsFor("pages/auth.md")
	sort.Strings(tags)
	if len(tags) != 2 || tags[0] != "core" || tags[1] != "security" {
		t.Errorf("TagsFor(auth.md) = %v, want [core security]", tags)
	}
}

func TestTagIndex_Remove(t *testing.T) {
	idx := NewTagIndex()
	idx.Update("pages/auth.md", []string{"security", "core"})
	idx.Update("pages/payments.md", []string{"billing", "core"})

	idx.Remove("pages/auth.md")

	got := idx.ByTag("core")
	if len(got) != 1 || got[0] != "pages/payments.md" {
		t.Errorf("after remove, ByTag(core) = %v, want [pages/payments.md]", got)
	}

	got = idx.ByTag("security")
	if len(got) != 0 {
		t.Errorf("after remove, ByTag(security) = %v, want empty", got)
	}

	tags := idx.TagsFor("pages/auth.md")
	if len(tags) != 0 {
		t.Errorf("after remove, TagsFor(auth.md) = %v, want empty", tags)
	}
}

func TestTagIndex_MultipleFiles(t *testing.T) {
	idx := NewTagIndex()
	idx.Update("a.md", []string{"shared"})
	idx.Update("b.md", []string{"shared"})
	idx.Update("c.md", []string{"shared"})

	got := idx.ByTag("shared")
	if len(got) != 3 {
		t.Errorf("ByTag(shared) = %v, want 3 results", got)
	}
}

func TestTagIndex_ReplaceOverwrite(t *testing.T) {
	idx := NewTagIndex()
	idx.Update("pages/auth.md", []string{"security", "core"})
	idx.Update("pages/auth.md", []string{"auth", "core"})

	got := idx.ByTag("security")
	if len(got) != 0 {
		t.Errorf("after overwrite, ByTag(security) = %v, want empty", got)
	}

	got = idx.ByTag("auth")
	if len(got) != 1 || got[0] != "pages/auth.md" {
		t.Errorf("after overwrite, ByTag(auth) = %v, want [pages/auth.md]", got)
	}

	got = idx.ByTag("core")
	if len(got) != 1 {
		t.Errorf("after overwrite, ByTag(core) = %v, want 1 result", got)
	}

	tags := idx.TagsFor("pages/auth.md")
	sort.Strings(tags)
	if len(tags) != 2 || tags[0] != "auth" || tags[1] != "core" {
		t.Errorf("TagsFor(auth.md) = %v, want [auth core]", tags)
	}
}

func TestTagIndex_All(t *testing.T) {
	idx := NewTagIndex()
	idx.Update("a.md", []string{"x", "y"})
	idx.Update("b.md", []string{"y", "z"})

	all := idx.All()
	if len(all["y"]) != 2 {
		t.Errorf("All()[y] = %v, want 2 entries", all["y"])
	}
	if len(all["x"]) != 1 {
		t.Errorf("All()[x] = %v, want 1 entry", all["x"])
	}
}

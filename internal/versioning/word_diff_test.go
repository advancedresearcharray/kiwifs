package versioning

import (
	"context"
	"testing"
)

func TestWordDiffText(t *testing.T) {
	diff, err := WordDiffText("hello world today", "hello brave world", "a", "b")
	if err != nil {
		t.Fatal(err)
	}
	if diff == "" {
		t.Fatal("expected non-empty diff")
	}
	if !containsAll(diff, "hello", "world") {
		t.Fatalf("diff missing tokens: %q", diff)
	}
}

func TestCowWordDiff(t *testing.T) {
	dir := t.TempDir()
	c, err := NewCow(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	path := "note.md"
	writeRoot(t, dir, path, "hello world today")
	if err := c.Commit(ctx, path, "tester", "v1"); err != nil {
		t.Fatal(err)
	}
	writeRoot(t, dir, path, "hello brave world today")
	if err := c.Commit(ctx, path, "tester", "v2"); err != nil {
		t.Fatal(err)
	}
	vs, err := c.Log(ctx, path)
	if err != nil || len(vs) < 2 {
		t.Fatalf("log: %v len=%d", err, len(vs))
	}
	diff, err := c.WordDiff(ctx, path, vs[1].Hash, vs[0].Hash)
	if err != nil {
		t.Fatal(err)
	}
	if diff == "" {
		t.Fatal("expected word diff")
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !containsSubstring(s, p) {
			return false
		}
	}
	return true
}

func containsSubstring(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexString(s, sub) >= 0)
}

func indexString(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

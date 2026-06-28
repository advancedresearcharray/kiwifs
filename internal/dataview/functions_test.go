package dataview

import (
	"strings"
	"testing"
)

func TestDaysAgoCompiler(t *testing.T) {
	fn, ok := funcRegistry["days_ago"]
	if !ok {
		t.Fatal("days_ago not registered")
	}
	sql, _, err := fn([]compiledArg{{SQL: "7", Params: nil}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, "datetime('now'") || !strings.Contains(sql, "days") {
		t.Fatalf("unexpected SQL: %s", sql)
	}
}

func TestParseDaysAgoExpr(t *testing.T) {
	expr, err := ParseExpr("days_ago(7)")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := expr.(*FuncCall); !ok {
		t.Fatalf("expected *FuncCall, got %T", expr)
	}
}

func TestNowCompiler(t *testing.T) {
	fn, ok := funcRegistry["now"]
	if !ok {
		t.Fatal("now not registered")
	}
	sql, _, err := fn(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, "strftime('%Y-%m-%dT%H:%M:%SZ', 'now')") {
		t.Fatalf("unexpected SQL: %s", sql)
	}
}

func TestDateCompiler(t *testing.T) {
	fn, ok := funcRegistry["date"]
	if !ok {
		t.Fatal("date not registered")
	}
	sql, _, err := fn([]compiledArg{{SQL: "?", Params: []any{"2026-01-01"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, "date(?)") {
		t.Fatalf("unexpected SQL: %s", sql)
	}
}

func TestParseTemporalFuncCaseInsensitive(t *testing.T) {
	for _, input := range []string{"NOW()", "Date(\"2026-01-01\")", "DAYS_AGO(3)"} {
		if _, err := ParseExpr(input); err != nil {
			t.Fatalf("ParseExpr(%q): %v", input, err)
		}
	}
}

package dataview

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestParseExpr_TemporalFunctions(t *testing.T) {
	cases := []string{
		`NOW()`,
		`DATE("2026-01-01")`,
		`created < NOW()`,
		`published_at > DATE("2026-06-15")`,
		`created BETWEEN DATE("2026-01-01") AND NOW()`,
		`NOT (due BETWEEN DATE("2026-04-01") AND DATE("2026-06-01"))`,
	}
	for _, input := range cases {
		if _, err := ParseExpr(input); err != nil {
			t.Fatalf("ParseExpr(%q): %v", input, err)
		}
	}
}

func TestCompileSQL_TemporalFunctions(t *testing.T) {
	cases := []struct {
		where string
		want  []string
	}{
		{
			where: `created < NOW()`,
			want:  []string{"strftime('%Y-%m-%dT%H:%M:%SZ', 'now')", "$.created"},
		},
		{
			where: `published_at > DATE("2026-01-01")`,
			want:  []string{"date(?)", "$.published_at", "2026-01-01"},
		},
		{
			where: `created BETWEEN DATE("2026-01-01") AND NOW()`,
			want:  []string{"BETWEEN", "date(?)", "strftime('%Y-%m-%dT%H:%M:%SZ', 'now')", "2026-01-01"},
		},
	}
	for _, tc := range cases {
		expr, err := ParseExpr(tc.where)
		if err != nil {
			t.Fatalf("parse %q: %v", tc.where, err)
		}
		plan := &QueryPlan{Type: "table", Where: expr, Limit: 50}
		sql, args, err := CompileSQL(plan)
		if err != nil {
			t.Fatalf("compile %q: %v", tc.where, err)
		}
		for _, fragment := range tc.want {
			if !strings.Contains(sql, fragment) && !containsArg(args, fragment) {
				t.Fatalf("compile %q: missing %q in sql=%q args=%v", tc.where, fragment, sql, args)
			}
		}
	}
}

func containsArg(args []any, want string) bool {
	for _, a := range args {
		if fmtString(a) == want {
			return true
		}
	}
	return false
}

func fmtString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func TestIntegration_TemporalDateFilter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`TABLE name FROM "students/" WHERE last_active > DATE("2026-04-01") SORT name ASC`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("got %d rows, want 2 (Priya and Amit)", len(result.Rows))
	}
}

func TestIntegration_TemporalBetween(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`TABLE name FROM "students/" WHERE last_active BETWEEN DATE("2026-04-01") AND DATE("2026-04-30")`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(result.Rows))
	}
}

func TestIntegration_TemporalNow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`TABLE name FROM "students/" WHERE last_active < NOW()`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 3 {
		t.Fatalf("got %d rows, want 3 historical records", len(result.Rows))
	}
}

func TestIntegration_TemporalDaysAgo(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`TABLE name FROM "students/" WHERE last_active > days_ago(365)`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 3 {
		t.Fatalf("got %d rows, want 3 within last year", len(result.Rows))
	}
}

func TestIntegration_TaskTemporalDue(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`TASK WHERE due > DATE("2026-04-01")`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("got %d rows, want 1 task with due after 2026-04-01", len(result.Rows))
	}
	if result.Rows[0]["text"] != "Send email" {
		t.Fatalf("unexpected task: %v", result.Rows[0]["text"])
	}
}

func TestIntegration_TaskBetweenDates(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`TASK WHERE due BETWEEN DATE("2026-04-01") AND DATE("2026-06-01")`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(result.Rows))
	}
}

func TestEvalDateLiteral_Malformed(t *testing.T) {
	task := taskRow{}
	if got := evalDateLiteral(&Literal{Value: "not-a-date"}, task); got != nil {
		t.Fatalf("expected nil for malformed date, got %v", got)
	}
}

func TestNormalizeComparableTime_Timezone(t *testing.T) {
	tm, ok := normalizeComparableTime("2026-06-15T10:00:00+05:30")
	if !ok {
		t.Fatal("expected parse success")
	}
	if tm.Location() != time.UTC {
		t.Fatalf("expected UTC, got %v", tm.Location())
	}
	if tm.Format("2006-01-02") != "2026-06-15" {
		t.Fatalf("unexpected normalized date: %s", tm.Format("2006-01-02"))
	}
}

func TestIntegration_MalformedDate_NoMatch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`TABLE name FROM "students/" WHERE last_active > DATE("not-a-date")`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 0 {
		t.Fatalf("malformed DATE should match nothing, got %d rows", len(result.Rows))
	}
}

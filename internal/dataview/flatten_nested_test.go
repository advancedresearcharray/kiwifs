package dataview

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"testing"
)

func setupEventLogDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS file_meta (
		path TEXT PRIMARY KEY,
		frontmatter TEXT NOT NULL DEFAULT '{}',
		tasks TEXT NOT NULL DEFAULT '[]',
		updated_at TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatal(err)
	}

	fm, _ := json.Marshal(map[string]any{
		"entries": []any{
			map[string]any{"event_type": "user.signup", "actor": "alice", "timestamp": "2026-01-01T10:00:00Z"},
			map[string]any{"event_type": "user.login", "actor": "bob", "timestamp": "2026-01-02T11:00:00Z"},
		},
	})
	_, err = db.Exec(`INSERT INTO file_meta(path, frontmatter, tasks, updated_at) VALUES (?, ?, ?, ?)`,
		"events/log1.md", string(fm), "[]", "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	fm2, _ := json.Marshal(map[string]any{"title": "no entries"})
	_, err = db.Exec(`INSERT INTO file_meta(path, frontmatter, tasks, updated_at) VALUES (?, ?, ?, ?)`,
		"events/empty.md", string(fm2), "[]", "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	fm3, _ := json.Marshal(map[string]any{"entries": "not-an-array"})
	_, err = db.Exec(`INSERT INTO file_meta(path, frontmatter, tasks, updated_at) VALUES (?, ?, ?, ?)`,
		"events/bad.md", string(fm3), "[]", "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func TestIntegration_FlattenNestedObjects(t *testing.T) {
	db := setupEventLogDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`TABLE entries.event_type, entries.actor FROM "events/" FLATTEN entries`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(result.Rows))
	}
	if result.Rows[0]["entries.event_type"] != "user.signup" && result.Rows[0]["entries.event_type"] != "user.login" {
		t.Errorf("unexpected event_type: %v", result.Rows[0]["entries.event_type"])
	}
}

func TestIntegration_FlattenNestedWhere(t *testing.T) {
	db := setupEventLogDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`TABLE entries.event_type FROM "events/" FLATTEN entries WHERE entries.event_type = "user.signup"`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(result.Rows))
	}
	if result.Rows[0]["entries.event_type"] != "user.signup" {
		t.Errorf("event_type = %v, want user.signup", result.Rows[0]["entries.event_type"])
	}
}

func TestIntegration_FlattenNestedCount(t *testing.T) {
	db := setupEventLogDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`COUNT FROM "events/" FLATTEN entries WHERE entries.event_type = "user.signup"`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 {
		t.Fatalf("count = %d, want 1", result.Total)
	}
}

func TestIntegration_FlattenNestedSort(t *testing.T) {
	db := setupEventLogDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`TABLE entries.event_type FROM "events/" FLATTEN entries SORT entries.timestamp DESC`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(result.Rows))
	}
	if result.Rows[0]["entries.event_type"] != "user.login" {
		t.Errorf("first row = %v, want user.login", result.Rows[0]["entries.event_type"])
	}
}

func TestIntegration_FlattenMissingField(t *testing.T) {
	db := setupEventLogDB(t)
	defer db.Close()
	exec := NewExecutor(db)

	result, err := exec.Query(context.Background(),
		`TABLE entries.event_type FROM "events/" FLATTEN entries`, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("got %d rows, want 2 (missing/non-array fields produce zero rows)", len(result.Rows))
	}
}

func TestCompileSQL_FlattenDotNotation(t *testing.T) {
	plan := &QueryPlan{
		Type:    "table",
		Fields:  []FieldSpec{{Expr: "entries.event_type"}, {Expr: "entries.actor"}},
		Flatten: "entries",
		Limit:   50,
	}
	sql, _, err := CompileSQL(plan)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"json_each(file_meta.frontmatter, '$.entries')",
		"json_extract(_flat.value, '$.event_type')",
		"json_extract(_flat.value, '$.actor')",
	} {
		if !strings.Contains(sql, want) {
			t.Errorf("sql = %q, missing %q", sql, want)
		}
	}
}

func TestCompileSQL_FlattenDotNotationWhere(t *testing.T) {
	plan := &QueryPlan{
		Type:    "count",
		Flatten: "entries",
		Where: &BinaryExpr{
			Left:  &FieldRef{Path: "entries.event_type"},
			Op:    OpEq,
			Right: &Literal{Value: "user.signup"},
		},
		Limit: 50,
	}
	sql, _, err := CompileSQL(plan)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(sql, "json_extract(_flat.value, '$.event_type') = ?") {
		t.Errorf("sql = %q, missing flatten WHERE on nested field", sql)
	}
}

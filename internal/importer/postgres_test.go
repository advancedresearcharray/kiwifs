package importer

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPostgresImporterIntegration(t *testing.T) {
	requireDocker(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("kiwi_test"),
		postgres.WithUsername("kiwi"),
		postgres.WithPassword("secret"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		_ = pgContainer.Terminate(context.Background())
	})

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open seed db: %v", err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, `
		CREATE TABLE sample_rows (
			id SERIAL PRIMARY KEY,
			label TEXT NOT NULL,
			qty INTEGER,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			amount NUMERIC(10,2)
		);
		INSERT INTO sample_rows (label, qty, active, amount) VALUES
			('alpha', 10, true, 19.99),
			('beta', 0, false, 0.00),
			('gamma', 5, true, 3.50);
		ANALYZE sample_rows;
	`)
	if err != nil {
		t.Fatalf("seed table: %v", err)
	}

	src, err := NewPostgres(dsn, "sample_rows", "", nil)
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}
	defer src.Close()

	tables, err := BrowsePostgresTables(ctx, src)
	if err != nil {
		t.Fatalf("BrowsePostgresTables: %v", err)
	}
	foundTable := false
	for _, tbl := range tables {
		if tbl.Name == "sample_rows" {
			foundTable = true
			// pg_class.reltuples is -1 until ANALYZE; after seeding we expect a non-negative estimate.
			if tbl.EstimatedCount < 0 {
				t.Fatalf("unexpected estimated count after ANALYZE: %d", tbl.EstimatedCount)
			}
		}
	}
	if !foundTable {
		t.Fatalf("sample_rows not listed: %+v", tables)
	}

	records, errs := src.Stream(ctx)
	var got []Record
	for rec := range records {
		got = append(got, rec)
	}
	for err := range errs {
		if err != nil {
			t.Fatalf("stream error: %v", err)
		}
	}
	if len(got) != 3 {
		t.Fatalf("records=%d, want 3", len(got))
	}

	byPK := map[string]Record{}
	for _, rec := range got {
		byPK[rec.PrimaryKey] = rec
		if rec.Table != "sample_rows" {
			t.Fatalf("table=%q, want sample_rows", rec.Table)
		}
		if rec.SourceDSN != "postgres" {
			t.Fatalf("source dsn=%q, want postgres", rec.SourceDSN)
		}
	}

	alpha, ok := byPK["1"]
	if !ok {
		t.Fatalf("missing pk=1 record: %+v", got)
	}
	if alpha.Fields["label"] != "alpha" {
		t.Fatalf("label=%v, want alpha", alpha.Fields["label"])
	}
	if alpha.Fields["qty"] != int64(10) {
		t.Fatalf("qty=%T %v, want int64(10)", alpha.Fields["qty"], alpha.Fields["qty"])
	}
	if alpha.Fields["active"] != true {
		t.Fatalf("active=%v, want true", alpha.Fields["active"])
	}
	if alpha.Fields["amount"] == nil {
		t.Fatal("expected amount field")
	}
	createdAt, ok := alpha.Fields["created_at"].(string)
	if !ok || createdAt == "" {
		t.Fatalf("created_at should be RFC3339 string, got %T %v", alpha.Fields["created_at"], alpha.Fields["created_at"])
	}
	if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
		t.Fatalf("created_at not RFC3339: %v (%q)", err, createdAt)
	}

	filtered, err := NewPostgres(dsn, "sample_rows", "", []string{"label"})
	if err != nil {
		t.Fatalf("NewPostgres filtered: %v", err)
	}
	defer filtered.Close()

	filteredRecords, filteredErrs := filtered.Stream(ctx)
	var filteredGot Record
	for rec := range filteredRecords {
		filteredGot = rec
		break
	}
	for err := range filteredErrs {
		if err != nil {
			t.Fatalf("filtered stream error: %v", err)
		}
	}
	if _, ok := filteredGot.Fields["label"]; !ok {
		t.Fatalf("expected label in filtered fields: %+v", filteredGot.Fields)
	}
	if _, ok := filteredGot.Fields["qty"]; ok {
		t.Fatalf("qty should be filtered out: %+v", filteredGot.Fields)
	}
	if filteredGot.PrimaryKey == "" {
		t.Fatal("expected primary key on filtered record")
	}

	customQuery, err := NewPostgres(dsn, "", "SELECT label FROM sample_rows WHERE id = 2", nil)
	if err != nil {
		t.Fatalf("NewPostgres custom query: %v", err)
	}
	defer customQuery.Close()

	customRecords, customErrs := customQuery.Stream(ctx)
	var customGot Record
	for rec := range customRecords {
		customGot = rec
		break
	}
	for err := range customErrs {
		if err != nil {
			t.Fatalf("custom query stream error: %v", err)
		}
	}
	if customGot.Fields["label"] != "beta" {
		t.Fatalf("custom query label=%v, want beta", customGot.Fields["label"])
	}

	pipe, store := testPipeline(t)
	stats, err := Run(ctx, src, pipe, Options{Actor: "test"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if stats.Imported != 3 {
		t.Fatalf("imported=%d, want 3", stats.Imported)
	}

	content, err := store.Read(ctx, "sample_rows/1.md")
	if err != nil {
		t.Fatalf("read imported file: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, "label: alpha") {
		t.Fatalf("missing alpha in frontmatter: %s", s)
	}
	if !strings.Contains(s, "_source: sample_rows") {
		t.Fatalf("missing _source: %s", s)
	}
}

package importer

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
)

func TestMySQLImporterIntegration(t *testing.T) {
	requireDocker(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8",
		mysql.WithDatabase("kiwi_test"),
		mysql.WithUsername("kiwi"),
		mysql.WithPassword("secret"),
	)
	if err != nil {
		t.Fatalf("start mysql container: %v", err)
	}
	t.Cleanup(func() {
		_ = mysqlContainer.Terminate(context.Background())
	})

	dsn, err := mysqlContainer.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}
	if !strings.Contains(dsn, "parseTime=true") {
		t.Fatalf("dsn missing parseTime=true: %s", dsn)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open seed db: %v", err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, `
		CREATE TABLE sample_rows (
			id INT AUTO_INCREMENT PRIMARY KEY,
			label VARCHAR(64) NOT NULL,
			qty INT,
			active BOOLEAN DEFAULT true,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			amount DECIMAL(10,2)
		)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO sample_rows (label, qty, active, amount) VALUES
			('alpha', 10, true, 19.99),
			('beta', 0, false, 0.00),
			('gamma', 5, true, 3.50)`)
	if err != nil {
		t.Fatalf("seed rows: %v", err)
	}

	src, err := NewMySQL(dsn, "sample_rows", "", nil)
	if err != nil {
		t.Fatalf("NewMySQL: %v", err)
	}
	defer src.Close()

	tables, err := BrowseMySQLTables(ctx, src)
	if err != nil {
		t.Fatalf("BrowseMySQLTables: %v", err)
	}
	foundTable := false
	for _, tbl := range tables {
		if tbl.Name == "sample_rows" {
			foundTable = true
			if tbl.EstimatedCount < 0 {
				t.Fatalf("unexpected estimated count: %d", tbl.EstimatedCount)
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
		if rec.SourceDSN != "mysql" {
			t.Fatalf("source dsn=%q, want mysql", rec.SourceDSN)
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
		t.Fatalf("created_at should be RFC3339 string with parseTime=true, got %T %v", alpha.Fields["created_at"], alpha.Fields["created_at"])
	}
	if _, err := time.Parse(time.RFC3339, createdAt); err != nil {
		t.Fatalf("created_at not RFC3339: %v (%q)", err, createdAt)
	}

	filtered, err := NewMySQL(dsn, "sample_rows", "", []string{"label"})
	if err != nil {
		t.Fatalf("NewMySQL filtered: %v", err)
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

	customQuery, err := NewMySQL(dsn, "", "SELECT label FROM sample_rows WHERE id = 2", nil)
	if err != nil {
		t.Fatalf("NewMySQL custom query: %v", err)
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
}

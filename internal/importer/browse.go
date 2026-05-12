package importer

import (
	"context"
	"fmt"
)

// BrowseTable represents a table/collection discovered during browse.
type BrowseTable struct {
	Name           string `json:"name"`
	EstimatedCount int64  `json:"estimated_count,omitempty"`
}

// BrowsePostgresTables lists tables in the public schema of a Postgres database.
func BrowsePostgresTables(ctx context.Context, src *PostgresSource) ([]BrowseTable, error) {
	rows, err := src.DB().QueryContext(ctx, `
		SELECT table_name,
		       COALESCE(
		           (SELECT reltuples::bigint FROM pg_class
		            WHERE relname = t.table_name AND relnamespace =
		              (SELECT oid FROM pg_namespace WHERE nspname = t.table_schema)),
		           0
		       ) AS estimated_count
		FROM information_schema.tables t
		WHERE table_schema = current_schema()
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name`)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []BrowseTable
	for rows.Next() {
		var t BrowseTable
		if err := rows.Scan(&t.Name, &t.EstimatedCount); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

// BrowseMySQLTables lists tables in the current MySQL database.
func BrowseMySQLTables(ctx context.Context, src *MySQLSource) ([]BrowseTable, error) {
	rows, err := src.DB().QueryContext(ctx, `
		SELECT TABLE_NAME, TABLE_ROWS
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME`)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []BrowseTable
	for rows.Next() {
		var t BrowseTable
		if err := rows.Scan(&t.Name, &t.EstimatedCount); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

// BrowseMongoCollections lists collections in the configured MongoDB database.
func BrowseMongoCollections(ctx context.Context, src *MongoSource) ([]BrowseTable, error) {
	names, err := src.Client().Database(src.DatabaseName()).ListCollectionNames(ctx, map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}

	tables := make([]BrowseTable, len(names))
	for i, name := range names {
		tables[i] = BrowseTable{Name: name}
	}
	return tables, nil
}

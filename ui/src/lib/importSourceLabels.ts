/**
 * Human-readable labels for import source type keys (API / CLI slugs).
 * Single source of truth for the import wizard and connection lists.
 *
 * Phase 1: 14 sources (7 builtin file, 3 native network, 4 Airbyte-powered)
 */

export type ImportSourceBackend = "builtin" | "native" | "airbyte";

export type ImportSourceType =
  // File-based (builtin)
  | "markdown" | "obsidian" | "csv" | "json" | "jsonl" | "yaml" | "excel" | "sqlite"
  // Native network (Go driver, no Airbyte needed)
  | "postgres" | "mysql" | "mongodb"
  // Airbyte-powered (migrating from legacy / new)
  | "firestore" | "notion" | "airtable" | "firebase-rtdb";

export type ImportSourceOption = {
  type: ImportSourceType;
  label: string;
  description: string;
  backend: ImportSourceBackend;
};

export const IMPORT_SOURCE_OPTIONS: ImportSourceOption[] = [
  // File-based (builtin — no external dependency)
  { type: "markdown", label: "Markdown", description: "Folder of .md files", backend: "builtin" },
  { type: "obsidian", label: "Obsidian", description: "Obsidian vault", backend: "builtin" },
  { type: "csv", label: "CSV", description: "CSV file", backend: "builtin" },
  { type: "json", label: "JSON", description: "JSON file", backend: "builtin" },
  { type: "jsonl", label: "JSON Lines", description: "JSONL file", backend: "builtin" },
  { type: "yaml", label: "YAML", description: "YAML file", backend: "builtin" },
  { type: "excel", label: "Excel", description: "Excel spreadsheet (.xlsx)", backend: "builtin" },
  { type: "sqlite", label: "SQLite", description: "SQLite database", backend: "builtin" },
  // Native network (Go driver, simple DSN/URI)
  { type: "postgres", label: "PostgreSQL", description: "PostgreSQL database", backend: "native" },
  { type: "mysql", label: "MySQL", description: "MySQL / MariaDB", backend: "native" },
  { type: "mongodb", label: "MongoDB", description: "MongoDB collection", backend: "native" },
  // Airbyte-powered (complex auth / API churn — migrating from legacy)
  { type: "firestore", label: "Firestore", description: "Google Cloud Firestore", backend: "airbyte" },
  { type: "firebase-rtdb", label: "Firebase RTDB", description: "Firebase Realtime Database", backend: "airbyte" },
  { type: "notion", label: "Notion", description: "Notion workspace", backend: "airbyte" },
  { type: "airtable", label: "Airtable", description: "Airtable base", backend: "airbyte" },
];

// --- Future sources (uncomment when ready) ---
// { type: "dynamodb", label: "DynamoDB", description: "AWS DynamoDB table", backend: "airbyte" },
// { type: "elasticsearch", label: "Elasticsearch", description: "Elasticsearch index", backend: "airbyte" },
// { type: "cockroachdb", label: "CockroachDB", description: "CockroachDB database", backend: "airbyte" },
// { type: "gsheets", label: "Google Sheets", description: "Google Spreadsheets", backend: "airbyte" },
// { type: "confluence", label: "Confluence", description: "Atlassian Confluence", backend: "airbyte" },
// { type: "github", label: "GitHub", description: "GitHub repos & issues", backend: "airbyte" },
// { type: "jira", label: "Jira", description: "Atlassian Jira", backend: "airbyte" },
// { type: "slack", label: "Slack", description: "Slack messages & channels", backend: "airbyte" },
// { type: "linear", label: "Linear", description: "Linear issues & projects", backend: "airbyte" },
// { type: "s3", label: "Amazon S3", description: "AWS S3 bucket", backend: "airbyte" },
// { type: "gcs", label: "Google Cloud Storage", description: "GCS bucket", backend: "airbyte" },

const AIRBYTE_SOURCE_SET = new Set(
  IMPORT_SOURCE_OPTIONS.filter((o) => o.backend === "airbyte").map((o) => o.type),
);

export function isAirbyteSourceType(t: string | null | undefined): t is ImportSourceType {
  return t != null && AIRBYTE_SOURCE_SET.has(t as ImportSourceType);
}

export function isBuiltinSourceType(t: string | null | undefined): t is ImportSourceType {
  return t != null && IMPORT_SOURCE_OPTIONS.some((o) => o.type === t && (o.backend === "builtin" || o.backend === "native"));
}

export function importSourceOption(type: string): ImportSourceOption | undefined {
  return IMPORT_SOURCE_OPTIONS.find((o) => o.type === type);
}

/**
 * Pretty label for a backend/connection `from` slug (e.g. "firestore" → "Firestore").
 */
export function sourceTypeLabel(from: string | null | undefined): string {
  if (from == null || from === "") return "Unknown source";
  const key = from.trim().toLowerCase();
  const row = IMPORT_SOURCE_OPTIONS.find((o) => o.type === key);
  if (row) return row.label;
  return key.replace(/[-_]/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

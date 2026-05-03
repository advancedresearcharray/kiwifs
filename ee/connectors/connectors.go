// Package connectors provides enterprise data source integrations.
//
// Gated behind [license.FeatureConnectors] and [license.FeatureVectorExt].
//
// The core open-source importer (internal/importer) supports file-based sources:
// CSV, JSON, JSONL, YAML, Excel, Obsidian vaults. This package adds:
//   - SaaS connectors: Notion, Airtable, Google Sheets, Confluence
//   - Database connectors: PostgreSQL, MySQL, MongoDB, DynamoDB, Redis, Elasticsearch
//   - Confluence migration tool (hierarchy, attachments, formatting)
//   - Scheduled / automated import sync (cron-style)
//   - External vector store connectors: Qdrant, pgvector, Pinecone, Weaviate, Milvus
package connectors

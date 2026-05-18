package importer

import "strings"

// AirbyteRegistry maps source type names to their Airbyte Docker image.
// These track the official Airbyte source connector images.
var AirbyteRegistry = map[string]string{
	// Databases
	"postgres":      "airbyte/source-postgres:latest",
	"mysql":         "airbyte/source-mysql:latest",
	"mongodb":       "airbyte/source-mongodb-v2:latest",
	"firestore":     "airbyte/source-firestore:latest",
	"dynamodb":      "airbyte/source-dynamodb:latest",
	"elasticsearch": "airbyte/source-elasticsearch:latest",
	"redis":         "airbyte/source-redis:latest",
	"mssql":         "airbyte/source-mssql:latest",
	"cockroachdb":   "airbyte/source-cockroachdb:latest",

	// SaaS / APIs
	"notion":     "airbyte/source-notion:latest",
	"airtable":   "airbyte/source-airtable:latest",
	"confluence": "airbyte/source-confluence:latest",
	"gsheets":    "airbyte/source-google-sheets:latest",
	"hubspot":    "airbyte/source-hubspot:latest",
	"salesforce": "airbyte/source-salesforce:latest",
	"stripe":     "airbyte/source-stripe:latest",
	"github":     "airbyte/source-github:latest",
	"jira":       "airbyte/source-jira:latest",
	"slack":      "airbyte/source-slack:latest",
	"zendesk":    "airbyte/source-zendesk-support:latest",
	"intercom":   "airbyte/source-intercom:latest",
	"linear":     "airbyte/source-linear:latest",

	// Files / Cloud Storage
	"s3":  "airbyte/source-s3:latest",
	"gcs": "airbyte/source-gcs:latest",
}

// BuiltinSources are handled natively by KiwiFS without Docker/Airbyte.
// These are local file-based sources that need no network auth.
var BuiltinSources = map[string]bool{
	"markdown": true,
	"obsidian": true,
	"csv":      true,
	"json":     true,
	"jsonl":    true,
	"excel":    true,
	"yaml":     true,
	"sqlite":   true,
}

// LookupAirbyteImage returns the Docker image for a given source name.
// Returns empty string if not found in registry.
func LookupAirbyteImage(sourceType string) string {
	sourceType = strings.ToLower(strings.TrimSpace(sourceType))
	if img, ok := AirbyteRegistry[sourceType]; ok {
		return img
	}
	// Allow full image references (e.g. "airbyte/source-custom:v1.2.3")
	if strings.Contains(sourceType, "/") {
		return sourceType
	}
	// Try with "source-" prefix for custom connectors
	if strings.HasPrefix(sourceType, "airbyte/") {
		return sourceType
	}
	return ""
}

// IsBuiltinSource returns true if the source is handled natively.
func IsBuiltinSource(sourceType string) bool {
	return BuiltinSources[strings.ToLower(strings.TrimSpace(sourceType))]
}

// ListAvailableSources returns all available source names grouped by type.
func ListAvailableSources(dockerAvailable bool) map[string][]string {
	result := map[string][]string{
		"builtin": make([]string, 0),
	}

	for name := range BuiltinSources {
		result["builtin"] = append(result["builtin"], name)
	}

	if dockerAvailable {
		result["airbyte"] = make([]string, 0)
		for name := range AirbyteRegistry {
			result["airbyte"] = append(result["airbyte"], name)
		}
	}

	return result
}

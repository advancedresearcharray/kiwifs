import { useCallback, useState } from "react";
import { ArrowLeft, ArrowRight, CheckCircle, Loader2, Upload } from "lucide-react";
import { Button } from "./ui/button";
import { api } from "../lib/api";

type SourceType = "firestore" | "postgres" | "mysql" | "mongodb" | "notion" | "airtable" | "csv";

const SOURCE_OPTIONS: { type: SourceType; label: string; icon: string; description: string }[] = [
  { type: "firestore", label: "Firestore", icon: "\uD83D\uDD25", description: "Google Cloud Firestore" },
  { type: "postgres", label: "PostgreSQL", icon: "\uD83D\uDC18", description: "PostgreSQL database" },
  { type: "mysql", label: "MySQL", icon: "\uD83D\uDC2C", description: "MySQL / MariaDB" },
  { type: "mongodb", label: "MongoDB", icon: "\uD83C\uDF43", description: "MongoDB collection" },
  { type: "notion", label: "Notion", icon: "\u270D\uFE0F", description: "Notion database" },
  { type: "airtable", label: "Airtable", icon: "\uD83D\uDCCA", description: "Airtable base" },
  { type: "csv", label: "CSV File", icon: "\uD83D\uDCC4", description: "CSV / JSON file" },
];

type WizardState = {
  step: number;
  sourceType: SourceType | null;
  // Connection params
  project: string;
  dsn: string;
  uri: string;
  database: string;
  databaseId: string;
  baseId: string;
  tableId: string;
  apiKey: string;
  credentials: unknown | null;
  // Browse results
  tables: { name: string; estimated_count?: number }[];
  selectedTable: string;
  // Config
  prefix: string;
  idColumn: string;
  // Results
  previews: { path: string; frontmatter: Record<string, unknown>; body_preview: string }[];
  importResult: { imported: number; skipped: number; errors: string[] } | null;
};

const initialState: WizardState = {
  step: 1,
  sourceType: null,
  project: "",
  dsn: "",
  uri: "",
  database: "",
  databaseId: "",
  baseId: "",
  tableId: "",
  apiKey: "",
  credentials: null,
  tables: [],
  selectedTable: "",
  prefix: "",
  idColumn: "",
  previews: [],
  importResult: null,
};

export function KiwiImportWizard({
  onClose,
  onComplete,
}: {
  onClose: () => void;
  onComplete: () => void;
}) {
  const [state, setState] = useState<WizardState>(initialState);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const update = useCallback((partial: Partial<WizardState>) => {
    setState((prev) => ({ ...prev, ...partial }));
    setError(null);
  }, []);

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => {
      try {
        const json = JSON.parse(reader.result as string);
        update({ credentials: json, project: json.project_id || state.project });
      } catch {
        setError("Invalid JSON file");
      }
    };
    reader.readAsText(file);
  };

  const handleBrowse = async () => {
    setLoading(true);
    setError(null);
    try {
      const params: Record<string, unknown> = { from: state.sourceType };
      if (state.sourceType === "firestore") {
        params.project = state.project;
        if (state.credentials) params.credentials = state.credentials;
      } else if (state.sourceType === "postgres" || state.sourceType === "mysql") {
        params.dsn = state.dsn;
      } else if (state.sourceType === "mongodb") {
        params.uri = state.uri || state.dsn;
        params.database = state.database;
      }

      const resp = await api.importBrowse(params as any);
      update({ tables: resp.tables, step: 3 });
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  };

  const handlePreview = async () => {
    setLoading(true);
    setError(null);
    try {
      const params: Record<string, unknown> = {
        from: state.sourceType,
        limit: 5,
      };
      if (state.sourceType === "firestore") {
        params.project = state.project;
        params.collection = state.selectedTable;
        if (state.credentials) params.credentials = state.credentials;
      } else if (state.sourceType === "postgres" || state.sourceType === "mysql") {
        params.dsn = state.dsn;
        params.table = state.selectedTable;
      } else if (state.sourceType === "mongodb") {
        params.uri = state.uri || state.dsn;
        params.database = state.database;
        params.collection = state.selectedTable;
      } else if (state.sourceType === "notion") {
        params.database_id = state.databaseId;
        params.api_key = state.apiKey;
      } else if (state.sourceType === "airtable") {
        params.base_id = state.baseId;
        params.table_id = state.tableId || state.selectedTable;
        params.api_key = state.apiKey;
      }

      const resp = await api.importPreview(params as any);
      update({ previews: resp.records, step: 4 });
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  };

  const handleImport = async () => {
    setLoading(true);
    setError(null);
    try {
      const params: Record<string, unknown> = {
        from: state.sourceType,
        prefix: state.prefix || state.selectedTable || state.sourceType,
      };
      if (state.idColumn) params.id_column = state.idColumn;

      if (state.sourceType === "firestore") {
        params.project = state.project;
        params.collection = state.selectedTable;
        if (state.credentials) params.credentials = state.credentials;
      } else if (state.sourceType === "postgres" || state.sourceType === "mysql") {
        params.dsn = state.dsn;
        params.table = state.selectedTable;
      } else if (state.sourceType === "mongodb") {
        params.uri = state.uri || state.dsn;
        params.database = state.database;
        params.collection = state.selectedTable;
      } else if (state.sourceType === "notion") {
        params.database_id = state.databaseId;
        params.api_key = state.apiKey;
      } else if (state.sourceType === "airtable") {
        params.base_id = state.baseId;
        params.table_id = state.tableId || state.selectedTable;
        params.api_key = state.apiKey;
      }

      const result = await api.importRun(params as any);
      update({ importResult: result, step: 6 });
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  };

  // Can we skip browse? (Notion/Airtable don't have browse)
  const skipsBrowse = state.sourceType === "notion" || state.sourceType === "airtable" || state.sourceType === "csv";

  return (
    <div className="max-w-3xl mx-auto p-6">
      {/* Header */}
      <div className="flex items-center gap-2 mb-6">
        <button
          type="button"
          onClick={state.step === 1 ? onClose : () => update({ step: state.step - 1 })}
          className="text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
        </button>
        <h1 className="text-lg font-semibold">Connect a data source</h1>
        <span className="text-xs text-muted-foreground ml-auto">Step {state.step} of 6</span>
      </div>

      {error && (
        <div className="bg-destructive/10 text-destructive text-sm px-3 py-2 rounded-md mb-4">
          {error}
        </div>
      )}

      {/* Step 1: Pick source type */}
      {state.step === 1 && (
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
          {SOURCE_OPTIONS.map((opt) => (
            <button
              key={opt.type}
              type="button"
              onClick={() => update({ sourceType: opt.type, step: 2 })}
              className="border border-border rounded-lg p-4 hover:bg-accent/50 text-left transition-colors"
            >
              <div className="text-2xl mb-2">{opt.icon}</div>
              <div className="font-medium text-sm">{opt.label}</div>
              <div className="text-xs text-muted-foreground">{opt.description}</div>
            </button>
          ))}
        </div>
      )}

      {/* Step 2: Connection details */}
      {state.step === 2 && (
        <div className="space-y-4">
          <h2 className="font-medium">Connection details for {state.sourceType}</h2>

          {state.sourceType === "firestore" && (
            <>
              <label className="block">
                <span className="text-sm font-medium">Project ID</span>
                <input
                  type="text"
                  value={state.project}
                  onChange={(e) => update({ project: e.target.value })}
                  placeholder="my-firebase-project"
                  className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
                />
              </label>
              <label className="block">
                <span className="text-sm font-medium">Service Account JSON</span>
                <div className="mt-1 border border-dashed border-border rounded-md p-4 text-center">
                  <Upload className="h-6 w-6 mx-auto mb-2 text-muted-foreground" />
                  <input type="file" accept=".json" onChange={handleFileUpload} className="text-sm" />
                  <p className="text-xs text-muted-foreground mt-1">
                    Credentials are held in memory only, never stored to disk
                  </p>
                </div>
                {state.credentials && (
                  <p className="text-xs text-green-600 mt-1">Service account loaded</p>
                )}
              </label>
            </>
          )}

          {(state.sourceType === "postgres" || state.sourceType === "mysql") && (
            <label className="block">
              <span className="text-sm font-medium">Connection DSN</span>
              <input
                type="text"
                value={state.dsn}
                onChange={(e) => update({ dsn: e.target.value })}
                placeholder={state.sourceType === "postgres"
                  ? "postgres://user:pass@host:5432/dbname"
                  : "user:pass@tcp(host:3306)/dbname"}
                className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
              />
            </label>
          )}

          {state.sourceType === "mongodb" && (
            <>
              <label className="block">
                <span className="text-sm font-medium">Connection URI</span>
                <input
                  type="text"
                  value={state.uri || state.dsn}
                  onChange={(e) => update({ uri: e.target.value })}
                  placeholder="mongodb://user:pass@host:27017"
                  className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
                />
              </label>
              <label className="block">
                <span className="text-sm font-medium">Database</span>
                <input
                  type="text"
                  value={state.database}
                  onChange={(e) => update({ database: e.target.value })}
                  placeholder="mydb"
                  className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
                />
              </label>
            </>
          )}

          {state.sourceType === "notion" && (
            <>
              <label className="block">
                <span className="text-sm font-medium">Notion API Key</span>
                <input
                  type="password"
                  value={state.apiKey}
                  onChange={(e) => update({ apiKey: e.target.value })}
                  placeholder="ntn_..."
                  className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
                />
              </label>
              <label className="block">
                <span className="text-sm font-medium">Database ID</span>
                <input
                  type="text"
                  value={state.databaseId}
                  onChange={(e) => update({ databaseId: e.target.value })}
                  placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
                  className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
                />
              </label>
            </>
          )}

          {state.sourceType === "airtable" && (
            <>
              <label className="block">
                <span className="text-sm font-medium">Airtable API Key</span>
                <input
                  type="password"
                  value={state.apiKey}
                  onChange={(e) => update({ apiKey: e.target.value })}
                  placeholder="pat..."
                  className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
                />
              </label>
              <label className="block">
                <span className="text-sm font-medium">Base ID</span>
                <input
                  type="text"
                  value={state.baseId}
                  onChange={(e) => update({ baseId: e.target.value })}
                  placeholder="app..."
                  className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
                />
              </label>
              <label className="block">
                <span className="text-sm font-medium">Table ID</span>
                <input
                  type="text"
                  value={state.tableId}
                  onChange={(e) => update({ tableId: e.target.value })}
                  placeholder="tbl..."
                  className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
                />
              </label>
            </>
          )}

          <div className="flex justify-end gap-2 pt-4">
            <Button variant="outline" onClick={onClose}>Cancel</Button>
            {skipsBrowse ? (
              <Button onClick={() => { update({ selectedTable: state.databaseId || state.tableId || "data" }); handlePreview(); }} disabled={loading}>
                {loading && <Loader2 className="h-4 w-4 animate-spin mr-1.5" />}
                Preview
                <ArrowRight className="h-4 w-4 ml-1.5" />
              </Button>
            ) : (
              <Button onClick={handleBrowse} disabled={loading}>
                {loading && <Loader2 className="h-4 w-4 animate-spin mr-1.5" />}
                Browse
                <ArrowRight className="h-4 w-4 ml-1.5" />
              </Button>
            )}
          </div>
        </div>
      )}

      {/* Step 3: Select table/collection */}
      {state.step === 3 && (
        <div className="space-y-4">
          <h2 className="font-medium">Select a table or collection</h2>
          {state.tables.length === 0 ? (
            <p className="text-muted-foreground text-sm">No tables found. Check your connection details.</p>
          ) : (
            <div className="border border-border rounded-lg divide-y divide-border max-h-80 overflow-auto">
              {state.tables.map((t) => (
                <button
                  key={t.name}
                  type="button"
                  onClick={() => { update({ selectedTable: t.name }); handlePreview(); }}
                  className="w-full px-4 py-3 text-left hover:bg-accent/50 flex items-center justify-between"
                >
                  <span className="font-medium text-sm">{t.name}</span>
                  {t.estimated_count != null && t.estimated_count > 0 && (
                    <span className="text-xs text-muted-foreground">~{t.estimated_count.toLocaleString()} rows</span>
                  )}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Step 4: Preview */}
      {state.step === 4 && (
        <div className="space-y-4">
          <h2 className="font-medium">Preview: {state.selectedTable}</h2>
          <p className="text-sm text-muted-foreground">
            Showing {state.previews.length} sample records as they will appear in your knowledge base.
          </p>

          <div className="space-y-2 max-h-80 overflow-auto">
            {state.previews.map((rec, i) => (
              <div key={i} className="border border-border rounded-md p-3">
                <div className="text-xs font-mono text-muted-foreground mb-1">{rec.path}</div>
                <div className="text-sm whitespace-pre-wrap">{rec.body_preview}</div>
              </div>
            ))}
          </div>

          <div className="flex justify-end gap-2 pt-4">
            <Button variant="outline" onClick={() => update({ step: 3 })}>Back</Button>
            <Button onClick={() => update({ step: 5 })}>
              Configure
              <ArrowRight className="h-4 w-4 ml-1.5" />
            </Button>
          </div>
        </div>
      )}

      {/* Step 5: Configure prefix + ID column */}
      {state.step === 5 && (
        <div className="space-y-4">
          <h2 className="font-medium">Import configuration</h2>

          <label className="block">
            <span className="text-sm font-medium">Target folder prefix</span>
            <input
              type="text"
              value={state.prefix || state.selectedTable}
              onChange={(e) => update({ prefix: e.target.value })}
              placeholder={state.selectedTable}
              className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Files will be created as <code>{state.prefix || state.selectedTable}/&lt;id&gt;.md</code>
            </p>
          </label>

          <label className="block">
            <span className="text-sm font-medium">ID column (optional)</span>
            <input
              type="text"
              value={state.idColumn}
              onChange={(e) => update({ idColumn: e.target.value })}
              placeholder="Auto-detect (primary key or document ID)"
              className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
            />
          </label>

          <div className="flex justify-end gap-2 pt-4">
            <Button variant="outline" onClick={() => update({ step: 4 })}>Back</Button>
            <Button onClick={handleImport} disabled={loading}>
              {loading && <Loader2 className="h-4 w-4 animate-spin mr-1.5" />}
              Import
            </Button>
          </div>
        </div>
      )}

      {/* Step 6: Results */}
      {state.step === 6 && state.importResult && (
        <div className="text-center py-8">
          <CheckCircle className="h-12 w-12 mx-auto mb-4 text-green-500" />
          <h2 className="text-lg font-semibold mb-2">Import complete</h2>
          <div className="text-sm text-muted-foreground space-y-1 mb-6">
            <div><strong>{state.importResult.imported}</strong> documents imported</div>
            {state.importResult.skipped > 0 && (
              <div><strong>{state.importResult.skipped}</strong> unchanged (skipped)</div>
            )}
            {state.importResult.errors.length > 0 && (
              <div className="text-destructive"><strong>{state.importResult.errors.length}</strong> errors</div>
            )}
          </div>
          <Button onClick={onComplete}>Done</Button>
        </div>
      )}
    </div>
  );
}

import { useCallback, useEffect, useState } from "react";
import { ArrowLeft, ArrowRight, CheckCircle, Loader2, AlertCircle, RefreshCw } from "lucide-react";
import { Button } from "./ui/button";
import { api, type AirbyteSpecProperty, type AirbyteSpecResponse, type AirbyteStream } from "../lib/api";
import { SourceIcon } from "./SourceIcon";

type SourceType = "markdown" | "firestore" | "postgres" | "mysql" | "mongodb" | "notion" | "airtable" | "csv" | "json" | "sqlite";

const SOURCE_OPTIONS: { type: SourceType; label: string; description: string }[] = [
  { type: "markdown", label: "Markdown", description: "Folder of .md files" },
  { type: "firestore", label: "Firestore", description: "Google Cloud Firestore" },
  { type: "postgres", label: "PostgreSQL", description: "PostgreSQL database" },
  { type: "mysql", label: "MySQL", description: "MySQL / MariaDB" },
  { type: "mongodb", label: "MongoDB", description: "MongoDB collection" },
  { type: "notion", label: "Notion", description: "Notion database" },
  { type: "airtable", label: "Airtable", description: "Airtable base" },
  { type: "csv", label: "CSV", description: "CSV file" },
  { type: "json", label: "JSON", description: "JSON / JSONL file" },
  { type: "sqlite", label: "SQLite", description: "SQLite database" },
];

const AIRBYTE_SOURCES: SourceType[] = ["firestore", "postgres", "mysql", "mongodb"];
const SIMPLE_SOURCES: SourceType[] = ["notion", "airtable"];

function isAirbyteSource(t: SourceType | null): boolean {
  return t != null && AIRBYTE_SOURCES.includes(t);
}

type WizardState = {
  step: number;
  sourceType: SourceType | null;
  // Airbyte dynamic config
  airbyteSpec: AirbyteSpecResponse | null;
  airbyteConfig: Record<string, unknown>;
  airbyteStreams: AirbyteStream[];
  airbyteDockerAvailable: boolean | null;
  // Simple source params (Notion, Airtable)
  databaseId: string;
  baseId: string;
  tableId: string;
  apiKey: string;
  // File/path params
  path: string;
  file: string;
  db: string;
  // Browse/stream results
  selectedStreams: string[];
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
  airbyteSpec: null,
  airbyteConfig: {},
  airbyteStreams: [],
  airbyteDockerAvailable: null,
  databaseId: "",
  baseId: "",
  tableId: "",
  apiKey: "",
  path: "",
  file: "",
  db: "",
  selectedStreams: [],
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
  const [specLoading, setSpecLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const update = useCallback((partial: Partial<WizardState>) => {
    setState((prev) => ({ ...prev, ...partial }));
    setError(null);
  }, []);

  // When an Airbyte source is selected, fetch its spec
  useEffect(() => {
    if (!state.sourceType || !isAirbyteSource(state.sourceType) || state.step !== 2) return;
    if (state.airbyteSpec) return;

    let cancelled = false;
    setSpecLoading(true);
    api.importAirbyteSpec(state.sourceType).then((spec) => {
      if (!cancelled) {
        update({ airbyteSpec: spec, airbyteDockerAvailable: true });
        setSpecLoading(false);
      }
    }).catch((err) => {
      if (!cancelled) {
        setSpecLoading(false);
        const errStr = String(err);
        if (errStr.includes("docker") || errStr.includes("Docker")) {
          update({ airbyteDockerAvailable: false });
        } else {
          setError(`Failed to load connector spec: ${errStr}`);
        }
      }
    });
    return () => { cancelled = true; };
  }, [state.sourceType, state.step, state.airbyteSpec, update]);

  const handleAirbyteCheck = async () => {
    if (!state.sourceType) return;
    setLoading(true);
    setError(null);
    try {
      const result = await api.importAirbyteCheck(state.sourceType, state.airbyteConfig as Record<string, unknown>);
      if (result.status === "SUCCEEDED") {
        // Connection valid — discover streams
        const catalog = await api.importAirbyteDiscover(state.sourceType, state.airbyteConfig as Record<string, unknown>);
        update({ airbyteStreams: catalog.streams, step: 3 });
      } else {
        setError(`Connection failed: ${result.message || "Unknown error"}`);
      }
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

      if (isAirbyteSource(state.sourceType)) {
        params.via = "airbyte";
        params.airbyte_config = state.airbyteConfig;
        params.streams = state.selectedStreams.length > 0 ? state.selectedStreams : undefined;
      } else if (state.sourceType === "markdown") {
        params.path = state.path;
      } else if (state.sourceType === "notion") {
        params.database_id = state.databaseId;
        params.api_key = state.apiKey;
      } else if (state.sourceType === "airtable") {
        params.base_id = state.baseId;
        params.table_id = state.tableId || state.selectedTable;
        params.api_key = state.apiKey;
      } else if (state.sourceType === "csv" || state.sourceType === "json") {
        params.file = state.file;
      } else if (state.sourceType === "sqlite") {
        params.db = state.db;
        params.table = state.selectedTable;
      }

      const result = await api.importRun(params as any);
      update({ importResult: result, step: 6 });
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

      if (isAirbyteSource(state.sourceType)) {
        params.via = "airbyte";
        params.airbyte_config = state.airbyteConfig;
        params.streams = state.selectedStreams.length > 0 ? state.selectedStreams : undefined;
      } else if (state.sourceType === "markdown") {
        params.path = state.path;
      } else if (state.sourceType === "notion") {
        params.database_id = state.databaseId;
        params.api_key = state.apiKey;
      } else if (state.sourceType === "airtable") {
        params.base_id = state.baseId;
        params.table_id = state.tableId || state.selectedTable;
        params.api_key = state.apiKey;
      } else if (state.sourceType === "csv" || state.sourceType === "json") {
        params.file = state.file;
      } else if (state.sourceType === "sqlite") {
        params.db = state.db;
        params.table = state.selectedTable;
      }

      const resp = await api.importPreview(params as any);
      update({ previews: resp.records, step: 4 });
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  };

  const totalSteps = isAirbyteSource(state.sourceType) ? 6 : 5;

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
        <span className="text-xs text-muted-foreground ml-auto">Step {state.step} of {totalSteps}</span>
      </div>

      {error && (
        <div className="bg-destructive/10 text-destructive text-sm px-3 py-2 rounded-md mb-4 flex items-start gap-2">
          <AlertCircle className="h-4 w-4 mt-0.5 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Step 1: Pick source type */}
      {state.step === 1 && (
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
          {SOURCE_OPTIONS.map((opt) => (
            <button
              key={opt.type}
              type="button"
              onClick={() => update({ sourceType: opt.type, step: 2, airbyteSpec: null, airbyteConfig: {} })}
              className="border border-border rounded-lg p-4 hover:bg-accent/50 text-left transition-colors"
            >
              <div className="mb-2"><SourceIcon source={opt.type} size={28} /></div>
              <div className="font-medium text-sm">{opt.label}</div>
              <div className="text-xs text-muted-foreground">{opt.description}</div>
            </button>
          ))}
        </div>
      )}

      {/* Step 2: Connection config */}
      {state.step === 2 && state.sourceType && (
        <div className="space-y-4">
          {isAirbyteSource(state.sourceType) ? (
            <AirbyteConfigForm
              sourceType={state.sourceType}
              spec={state.airbyteSpec}
              specLoading={specLoading}
              config={state.airbyteConfig}
              dockerAvailable={state.airbyteDockerAvailable}
              onConfigChange={(cfg) => update({ airbyteConfig: cfg })}
              onConnect={handleAirbyteCheck}
              connecting={loading}
              onCancel={onClose}
            />
          ) : SIMPLE_SOURCES.includes(state.sourceType) ? (
            <SimpleSourceForm
              sourceType={state.sourceType}
              state={state}
              update={update}
              onCancel={onClose}
              onNext={() => {
                const tableName = state.databaseId || state.tableId || "data";
                update({ selectedTable: tableName });
                handlePreview();
              }}
              loading={loading}
            />
          ) : (
            <FileSourceForm
              sourceType={state.sourceType}
              state={state}
              update={update}
              onCancel={onClose}
              onNext={() => {
                let tableName = "data";
                if (state.sourceType === "markdown") {
                  tableName = state.path.split(/[/\\]/).filter(Boolean).pop() || "docs";
                } else if (state.sourceType === "csv" || state.sourceType === "json") {
                  tableName = state.file.split(/[/\\]/).filter(Boolean).pop()?.replace(/\.\w+$/, "") || "data";
                } else if (state.sourceType === "sqlite") {
                  tableName = state.selectedTable || "data";
                }
                update({ selectedTable: tableName });
                handlePreview();
              }}
              loading={loading}
            />
          )}
        </div>
      )}

      {/* Step 3: Select streams (Airbyte) or tables */}
      {state.step === 3 && (
        <div className="space-y-4">
          <h2 className="font-medium">Select streams to import</h2>
          <p className="text-sm text-muted-foreground">
            Choose which collections or tables to sync into your knowledge base.
          </p>
          {state.airbyteStreams.length === 0 ? (
            <p className="text-muted-foreground text-sm">No streams found. Check your connection.</p>
          ) : (
            <div className="border border-border rounded-lg divide-y divide-border max-h-80 overflow-auto">
              {state.airbyteStreams.map((stream) => {
                const selected = state.selectedStreams.includes(stream.name);
                return (
                  <label
                    key={stream.name}
                    className="flex items-center gap-3 px-4 py-3 hover:bg-accent/50 cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={selected}
                      onChange={() => {
                        const next = selected
                          ? state.selectedStreams.filter((s) => s !== stream.name)
                          : [...state.selectedStreams, stream.name];
                        update({ selectedStreams: next });
                      }}
                      className="rounded border-border"
                    />
                    <div className="flex-1 min-w-0">
                      <span className="font-medium text-sm">{stream.name}</span>
                      {stream.namespace && (
                        <span className="text-xs text-muted-foreground ml-2">{stream.namespace}</span>
                      )}
                    </div>
                    {stream.supported_sync_modes && (
                      <span className="text-xs text-muted-foreground">
                        {stream.supported_sync_modes.join(", ")}
                      </span>
                    )}
                  </label>
                );
              })}
            </div>
          )}
          <div className="flex justify-between items-center pt-4">
            <span className="text-xs text-muted-foreground">
              {state.selectedStreams.length === 0
                ? "All streams will be imported"
                : `${state.selectedStreams.length} selected`}
            </span>
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => update({ step: 2 })}>Back</Button>
              <Button onClick={() => {
                const table = state.selectedStreams[0] || state.airbyteStreams[0]?.name || state.sourceType || "data";
                update({ selectedTable: table, step: 4 });
              }}>
                Configure
                <ArrowRight className="h-4 w-4 ml-1.5" />
              </Button>
            </div>
          </div>
        </div>
      )}

      {/* Step 4: Configure prefix + ID column */}
      {state.step === 4 && (
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
            <Button variant="outline" onClick={() => update({ step: 3 })}>Back</Button>
            <Button onClick={handlePreview} disabled={loading}>
              {loading && <Loader2 className="h-4 w-4 animate-spin mr-1.5" />}
              Preview
              <ArrowRight className="h-4 w-4 ml-1.5" />
            </Button>
          </div>
        </div>
      )}

      {/* Step 5: Preview */}
      {state.step === 5 && (
        <div className="space-y-4">
          <h2 className="font-medium">Preview</h2>
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

/**
 * Dynamic form rendered from Airbyte connector spec (JSON Schema).
 */
function AirbyteConfigForm({
  sourceType,
  spec,
  specLoading,
  config,
  dockerAvailable,
  onConfigChange,
  onConnect,
  connecting,
  onCancel,
}: {
  sourceType: SourceType;
  spec: AirbyteSpecResponse | null;
  specLoading: boolean;
  config: Record<string, unknown>;
  dockerAvailable: boolean | null;
  onConfigChange: (cfg: Record<string, unknown>) => void;
  onConnect: () => void;
  connecting: boolean;
  onCancel: () => void;
}) {
  if (dockerAvailable === false) {
    return (
      <div className="text-center py-8 space-y-3">
        <AlertCircle className="h-10 w-10 mx-auto text-amber-500" />
        <h2 className="font-medium">Docker required</h2>
        <p className="text-sm text-muted-foreground max-w-md mx-auto">
          Importing from {sourceType} uses Airbyte connectors which require Docker.
          Please install and start Docker Desktop, then try again.
        </p>
        <Button variant="outline" onClick={onCancel}>Cancel</Button>
      </div>
    );
  }

  if (specLoading || !spec) {
    return (
      <div className="flex flex-col items-center justify-center py-12 gap-3">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        <p className="text-sm text-muted-foreground">Loading {sourceType} connector...</p>
      </div>
    );
  }

  const schema = spec.connectionSpecification;
  const properties = schema.properties || {};
  const required = new Set(schema.required || []);

  const sortedKeys = Object.entries(properties)
    .sort(([, a], [, b]) => (a.order ?? 999) - (b.order ?? 999))
    .map(([k]) => k);

  const setField = (key: string, value: unknown) => {
    onConfigChange({ ...config, [key]: value });
  };

  return (
    <>
      <div className="flex items-center gap-2 mb-1">
        <SourceIcon source={sourceType} size={20} />
        <h2 className="font-medium">Connect to {sourceType}</h2>
      </div>
      <p className="text-xs text-muted-foreground mb-4">
        Powered by Airbyte &mdash; credentials are sent directly to the connector container, never stored.
      </p>

      <div className="space-y-4">
        {sortedKeys.map((key) => {
          const prop = properties[key];
          if (prop.const !== undefined) return null;
          return (
            <SpecField
              key={key}
              fieldKey={key}
              prop={prop}
              value={config[key]}
              isRequired={required.has(key)}
              onChange={(v) => setField(key, v)}
            />
          );
        })}
      </div>

      <div className="flex justify-end gap-2 pt-6">
        <Button variant="outline" onClick={onCancel}>Cancel</Button>
        <Button onClick={onConnect} disabled={connecting}>
          {connecting ? (
            <><Loader2 className="h-4 w-4 animate-spin mr-1.5" /> Checking...</>
          ) : (
            <><RefreshCw className="h-4 w-4 mr-1.5" /> Test &amp; Discover</>
          )}
        </Button>
      </div>
    </>
  );
}

/**
 * Renders a single field from an Airbyte JSON Schema property.
 */
function SpecField({
  fieldKey,
  prop,
  value,
  isRequired,
  onChange,
}: {
  fieldKey: string;
  prop: AirbyteSpecProperty;
  value: unknown;
  isRequired: boolean;
  onChange: (v: unknown) => void;
}) {
  const label = prop.title || fieldKey.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
  const isSecret = prop.airbyte_secret === true;
  const inputType = isSecret ? "password" : "text";

  // oneOf → select between option groups
  if (prop.oneOf && prop.oneOf.length > 0) {
    return (
      <OneOfField
        fieldKey={fieldKey}
        prop={prop}
        value={value as Record<string, unknown> | undefined}
        isRequired={isRequired}
        onChange={onChange}
      />
    );
  }

  // enum → select dropdown
  if (prop.enum && prop.enum.length > 0) {
    return (
      <label className="block">
        <span className="text-sm font-medium">
          {label}{isRequired && <span className="text-destructive ml-0.5">*</span>}
        </span>
        <select
          value={(value as string) ?? prop.default ?? ""}
          onChange={(e) => onChange(e.target.value)}
          className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
        >
          <option value="">Select...</option>
          {prop.enum.map((opt) => (
            <option key={opt} value={opt}>{opt}</option>
          ))}
        </select>
        {prop.description && <p className="text-xs text-muted-foreground mt-1">{prop.description}</p>}
      </label>
    );
  }

  // boolean → checkbox
  if (prop.type === "boolean") {
    return (
      <label className="flex items-center gap-2">
        <input
          type="checkbox"
          checked={value as boolean ?? prop.default ?? false}
          onChange={(e) => onChange(e.target.checked)}
          className="rounded border-border"
        />
        <span className="text-sm font-medium">{label}</span>
        {prop.description && <span className="text-xs text-muted-foreground">— {prop.description}</span>}
      </label>
    );
  }

  // integer / number
  if (prop.type === "integer" || prop.type === "number") {
    return (
      <label className="block">
        <span className="text-sm font-medium">
          {label}{isRequired && <span className="text-destructive ml-0.5">*</span>}
        </span>
        <input
          type="number"
          value={(value as number) ?? prop.default ?? ""}
          onChange={(e) => onChange(e.target.value ? Number(e.target.value) : undefined)}
          className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
        />
        {prop.description && <p className="text-xs text-muted-foreground mt-1">{prop.description}</p>}
      </label>
    );
  }

  // object with properties → nested group (simplified: JSON textarea)
  if (prop.type === "object" && prop.properties) {
    return (
      <label className="block">
        <span className="text-sm font-medium">
          {label}{isRequired && <span className="text-destructive ml-0.5">*</span>}
        </span>
        <textarea
          value={typeof value === "string" ? value : JSON.stringify(value ?? {}, null, 2)}
          onChange={(e) => {
            try { onChange(JSON.parse(e.target.value)); } catch { /* user still typing */ }
          }}
          placeholder="{}"
          rows={4}
          className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
        />
        {prop.description && <p className="text-xs text-muted-foreground mt-1">{prop.description}</p>}
      </label>
    );
  }

  // Default: text/password input
  return (
    <label className="block">
      <span className="text-sm font-medium">
        {label}{isRequired && <span className="text-destructive ml-0.5">*</span>}
      </span>
      <input
        type={inputType}
        value={(value as string) ?? ""}
        onChange={(e) => onChange(e.target.value || undefined)}
        placeholder={prop.default != null ? String(prop.default) : undefined}
        className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
      />
      {prop.description && <p className="text-xs text-muted-foreground mt-1">{prop.description}</p>}
    </label>
  );
}

/**
 * Handles `oneOf` spec properties — renders a select to pick the variant, then renders its fields.
 */
function OneOfField({
  fieldKey,
  prop,
  value,
  isRequired,
  onChange,
}: {
  fieldKey: string;
  prop: AirbyteSpecProperty;
  value: Record<string, unknown> | undefined;
  isRequired: boolean;
  onChange: (v: unknown) => void;
}) {
  const label = prop.title || fieldKey.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
  const options = prop.oneOf || [];

  // Determine which option is selected by matching const fields
  const selectedIdx = options.findIndex((opt) => {
    if (!opt.properties) return false;
    return Object.entries(opt.properties).some(
      ([k, v]) => v.const !== undefined && value?.[k] === v.const
    );
  });

  const activeIdx = selectedIdx >= 0 ? selectedIdx : 0;
  const activeOption = options[activeIdx];

  const selectOption = (idx: number) => {
    const opt = options[idx];
    if (!opt?.properties) return;
    const base: Record<string, unknown> = {};
    for (const [k, v] of Object.entries(opt.properties)) {
      if (v.const !== undefined) base[k] = v.const;
      else if (v.default !== undefined) base[k] = v.default;
    }
    onChange(base);
  };

  return (
    <div className="space-y-3 border border-border rounded-lg p-4">
      <label className="block">
        <span className="text-sm font-medium">
          {label}{isRequired && <span className="text-destructive ml-0.5">*</span>}
        </span>
        <select
          value={activeIdx}
          onChange={(e) => selectOption(Number(e.target.value))}
          className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
        >
          {options.map((opt, i) => (
            <option key={i} value={i}>{opt.title || `Option ${i + 1}`}</option>
          ))}
        </select>
      </label>

      {activeOption?.properties && (
        <div className="space-y-3 pl-2 border-l-2 border-border">
          {Object.entries(activeOption.properties)
            .filter(([, v]) => v.const === undefined)
            .map(([k, v]) => (
              <SpecField
                key={k}
                fieldKey={k}
                prop={v}
                value={(value as Record<string, unknown>)?.[k]}
                isRequired={activeOption.required?.includes(k) ?? false}
                onChange={(newVal) => {
                  onChange({ ...(value || {}), [k]: newVal });
                }}
              />
            ))}
        </div>
      )}
    </div>
  );
}

/**
 * Form for simple API-key sources (Notion, Airtable).
 */
function SimpleSourceForm({
  sourceType,
  state,
  update,
  onCancel,
  onNext,
  loading,
}: {
  sourceType: SourceType;
  state: WizardState;
  update: (p: Partial<WizardState>) => void;
  onCancel: () => void;
  onNext: () => void;
  loading: boolean;
}) {
  return (
    <>
      <h2 className="font-medium">Connect to {sourceType}</h2>

      {sourceType === "notion" && (
        <>
          <label className="block">
            <span className="text-sm font-medium">Notion API Key<span className="text-destructive ml-0.5">*</span></span>
            <input
              type="password"
              value={state.apiKey}
              onChange={(e) => update({ apiKey: e.target.value })}
              placeholder="ntn_..."
              className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
            />
          </label>
          <label className="block">
            <span className="text-sm font-medium">Database ID<span className="text-destructive ml-0.5">*</span></span>
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

      {sourceType === "airtable" && (
        <>
          <label className="block">
            <span className="text-sm font-medium">Airtable API Key<span className="text-destructive ml-0.5">*</span></span>
            <input
              type="password"
              value={state.apiKey}
              onChange={(e) => update({ apiKey: e.target.value })}
              placeholder="pat..."
              className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
            />
          </label>
          <label className="block">
            <span className="text-sm font-medium">Base ID<span className="text-destructive ml-0.5">*</span></span>
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
        <Button variant="outline" onClick={onCancel}>Cancel</Button>
        <Button onClick={onNext} disabled={loading}>
          {loading && <Loader2 className="h-4 w-4 animate-spin mr-1.5" />}
          Preview
          <ArrowRight className="h-4 w-4 ml-1.5" />
        </Button>
      </div>
    </>
  );
}

/**
 * Form for file-based sources (Markdown, CSV, JSON, SQLite).
 */
function FileSourceForm({
  sourceType,
  state,
  update,
  onCancel,
  onNext,
  loading,
}: {
  sourceType: SourceType;
  state: WizardState;
  update: (p: Partial<WizardState>) => void;
  onCancel: () => void;
  onNext: () => void;
  loading: boolean;
}) {
  return (
    <>
      <h2 className="font-medium">Configure {sourceType} import</h2>

      {sourceType === "markdown" && (
        <label className="block">
          <span className="text-sm font-medium">Directory Path<span className="text-destructive ml-0.5">*</span></span>
          <input
            type="text"
            value={state.path}
            onChange={(e) => update({ path: e.target.value })}
            placeholder="/path/to/docs or ./docs"
            className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
          />
          <p className="text-xs text-muted-foreground mt-1">
            Path to the folder containing your markdown files
          </p>
        </label>
      )}

      {(sourceType === "csv" || sourceType === "json") && (
        <label className="block">
          <span className="text-sm font-medium">File Path<span className="text-destructive ml-0.5">*</span></span>
          <input
            type="text"
            value={state.file}
            onChange={(e) => update({ file: e.target.value })}
            placeholder={sourceType === "csv" ? "/path/to/data.csv" : "/path/to/data.json"}
            className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
          />
          <p className="text-xs text-muted-foreground mt-1">
            {sourceType === "csv" ? "Path to the CSV file" : "Path to the JSON or JSONL file"}
          </p>
        </label>
      )}

      {sourceType === "sqlite" && (
        <>
          <label className="block">
            <span className="text-sm font-medium">Database File<span className="text-destructive ml-0.5">*</span></span>
            <input
              type="text"
              value={state.db}
              onChange={(e) => update({ db: e.target.value })}
              placeholder="/path/to/database.db"
              className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm font-mono"
            />
          </label>
          <label className="block">
            <span className="text-sm font-medium">Table Name</span>
            <input
              type="text"
              value={state.selectedTable}
              onChange={(e) => update({ selectedTable: e.target.value })}
              placeholder="my_table"
              className="mt-1 block w-full rounded-md border border-border bg-background px-3 py-2 text-sm"
            />
          </label>
        </>
      )}

      <div className="flex justify-end gap-2 pt-4">
        <Button variant="outline" onClick={onCancel}>Cancel</Button>
        <Button onClick={onNext} disabled={loading}>
          {loading && <Loader2 className="h-4 w-4 animate-spin mr-1.5" />}
          Preview
          <ArrowRight className="h-4 w-4 ml-1.5" />
        </Button>
      </div>
    </>
  );
}

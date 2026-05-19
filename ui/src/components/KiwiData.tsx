import { useCallback, useEffect, useState } from "react";
import { ArrowLeft, Database, Pause, Play, Plus, RefreshCw, Trash2 } from "lucide-react";
import { Button } from "./ui/button";
import { api, type ImportConnection } from "../lib/api";
import { sourceTypeLabel } from "../lib/importSourceLabels";
import { KiwiImportWizard } from "./KiwiImportWizard";
import { SourceIcon } from "./SourceIcon";

/** Prefer title case / brand labels when the saved name repeats the raw `from` slug. */
function connectionDisplayName(conn: ImportConnection): string {
  const fromKey = conn.from.trim().toLowerCase();
  const raw = (conn.name ?? "").trim();
  if (!raw) return sourceTypeLabel(conn.from);
  const lower = raw.toLowerCase();
  if (lower === fromKey) return sourceTypeLabel(conn.from);
  if (lower.startsWith(`${fromKey} `)) {
    return `${sourceTypeLabel(conn.from)}${raw.slice(fromKey.length)}`;
  }
  return raw;
}

/**
 * KiwiData — the data sources management view.
 *
 * Shows saved import connections with their status and provides
 * entry points to the import wizard and re-import functionality.
 */
export function KiwiData({ onClose }: { onClose: () => void }) {
  const [connections, setConnections] = useState<ImportConnection[]>([]);
  const [loading, setLoading] = useState(true);
  const [wizardOpen, setWizardOpen] = useState(false);
  const [selectedConn, setSelectedConn] = useState<ImportConnection | null>(null);
  const [reImporting, setReImporting] = useState<string | null>(null);

  const fetchConnections = useCallback(async () => {
    try {
      const conns = await api.importConnections();
      setConnections(conns ?? []);
    } catch {
      // Connection store may not be available (e.g. fresh install)
      setConnections([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchConnections();
  }, [fetchConnections]);

  const handleDelete = async (id: string) => {
    if (!confirm("Remove this data source connection?")) return;
    try {
      await api.importDeleteConnection(id);
      setConnections((prev) => prev.filter((c) => c.id !== id));
      if (selectedConn?.id === id) setSelectedConn(null);
    } catch (err) {
      alert(`Failed to delete: ${err}`);
    }
  };

  const handleReImport = async (conn: ImportConnection) => {
    setReImporting(conn.id);
    try {
      const result = await api.importRunConnection(conn.id);
      alert(`Sync complete: ${result.imported} imported, ${result.skipped} skipped`);
      fetchConnections();
    } catch (err) {
      alert(`Sync failed: ${err}`);
    } finally {
      setReImporting(null);
    }
  };

  const handleToggleSync = async (conn: ImportConnection, enabled: boolean) => {
    try {
      await api.importToggleSync(conn.id, enabled);
      fetchConnections();
    } catch (err) {
      alert(`Failed to toggle sync: ${err}`);
    }
  };

  if (wizardOpen) {
    return (
      <KiwiImportWizard
        onClose={() => setWizardOpen(false)}
        onComplete={() => {
          setWizardOpen(false);
          fetchConnections();
        }}
      />
    );
  }

  // Detail view
  if (selectedConn) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <button
          type="button"
          onClick={() => setSelectedConn(null)}
          className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground mb-4"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Data Sources
        </button>

        <div className="flex items-center gap-3 mb-2">
          <SourceIcon source={selectedConn.from} size={28} />
          <h1 className="text-xl font-semibold">{connectionDisplayName(selectedConn)}</h1>
          {selectedConn.sync_enabled && (
            <span className={`inline-flex items-center gap-1 text-xs font-medium px-2 py-0.5 rounded-full ${
              selectedConn.sync_status === "running" ? "bg-blue-500/10 text-blue-600" :
              selectedConn.sync_status === "error" ? "bg-red-500/10 text-red-600" :
              "bg-emerald-500/10 text-emerald-600"
            }`}>
              <RefreshCw className={`h-3 w-3 ${selectedConn.sync_status === "running" ? "animate-spin" : ""}`} />
              {selectedConn.sync_status === "running" ? "Syncing..." :
               selectedConn.sync_status === "error" ? "Sync error" :
               `Auto-sync every ${selectedConn.sync_interval || "1h"}`}
            </span>
          )}
        </div>
        <p className="text-sm text-muted-foreground mb-4">
          {selectedConn.prefix && <>Prefix: <code className="bg-muted px-1 rounded">{selectedConn.prefix}/</code> &middot; </>}
          {selectedConn.last_stats && <>{selectedConn.last_stats.imported} docs</>}
          {selectedConn.last_run && <> &middot; last synced {new Date(selectedConn.last_run).toLocaleString()}</>}
          {selectedConn.next_sync && selectedConn.sync_enabled && <> &middot; next sync {timeAgo(selectedConn.next_sync).replace(" ago", "")}</>}
        </p>

        {selectedConn.sync_error && (
          <div className="text-xs text-destructive bg-destructive/5 border border-destructive/20 rounded-lg px-3 py-2 mb-4">
            Last sync error: {selectedConn.sync_error}
          </div>
        )}

        <div className="flex gap-2 mb-6">
          <Button
            size="sm"
            onClick={() => handleReImport(selectedConn)}
            disabled={reImporting === selectedConn.id}
          >
            {reImporting === selectedConn.id ? (
              <RefreshCw className="h-3.5 w-3.5 animate-spin mr-1.5" />
            ) : (
              <Play className="h-3.5 w-3.5 mr-1.5" />
            )}
            Sync now
          </Button>
          {selectedConn.sync_enabled && (
            <Button size="sm" variant="outline" onClick={() => handleToggleSync(selectedConn, false)}>
              <Pause className="h-3.5 w-3.5 mr-1.5" />
              Pause sync
            </Button>
          )}
          {!selectedConn.sync_enabled && (
            <Button size="sm" variant="outline" onClick={() => handleToggleSync(selectedConn, true)}>
              <RefreshCw className="h-3.5 w-3.5 mr-1.5" />
              Resume sync
            </Button>
          )}
          <Button size="sm" variant="destructive" onClick={() => handleDelete(selectedConn.id)}>
            <Trash2 className="h-3.5 w-3.5 mr-1.5" />
            Remove
          </Button>
        </div>

        {/* Connection details */}
        <div className="border border-border rounded-lg p-4 text-sm space-y-2">
          <div><strong>Type:</strong> {sourceTypeLabel(selectedConn.from)}</div>
          {selectedConn.project && <div><strong>Project:</strong> {selectedConn.project}</div>}
          {selectedConn.collection && <div><strong>Collection:</strong> {selectedConn.collection}</div>}
          {selectedConn.table && <div><strong>Table:</strong> {selectedConn.table}</div>}
          {selectedConn.database && <div><strong>Database:</strong> {selectedConn.database}</div>}
          {selectedConn.dsn && <div><strong>DSN:</strong> <code className="bg-muted px-1 rounded text-xs">{selectedConn.dsn.replace(/:[^@]+@/, ":***@")}</code></div>}
        </div>
      </div>
    );
  }

  // List view
  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-2">
          <Database className="h-5 w-5 text-muted-foreground" />
          <h1 className="text-xl font-semibold">Data Sources</h1>
        </div>
        <Button variant="ghost" size="sm" onClick={onClose} className="text-muted-foreground">
          Close
        </Button>
      </div>

      {loading ? (
        <div className="flex justify-center py-12">
          <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      ) : connections.length === 0 ? (
        <div className="text-center py-12">
          <Database className="h-12 w-12 mx-auto mb-3 text-muted-foreground/50" />
          <p className="text-muted-foreground mb-4">No data sources connected yet</p>
          <Button onClick={() => setWizardOpen(true)} className="gap-2">
            <Plus className="h-4 w-4" />
            Connect a source
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          {connections.map((conn) => (
            <div
              key={conn.id}
              className="border border-border rounded-lg p-4 hover:bg-accent/50 cursor-pointer transition-colors group"
              onClick={() => setSelectedConn(conn)}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <SourceIcon source={conn.from} size={22} />
                  <div>
                    <div className="font-medium flex items-center gap-2">
                      {connectionDisplayName(conn)}
                      {conn.sync_enabled && (
                        <span className={`inline-flex items-center gap-1 text-[10px] font-medium px-1.5 py-0.5 rounded-full ${
                          conn.sync_status === "running" ? "bg-blue-500/10 text-blue-600 dark:text-blue-400" :
                          conn.sync_status === "error" ? "bg-red-500/10 text-red-600 dark:text-red-400" :
                          "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                        }`}>
                          {conn.sync_status === "running" ? <RefreshCw className="h-2.5 w-2.5 animate-spin" /> : <RefreshCw className="h-2.5 w-2.5" />}
                          {conn.sync_status === "running" ? "Syncing" : conn.sync_status === "error" ? "Error" : "Auto-sync"}
                        </span>
                      )}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      <span>{sourceTypeLabel(conn.from)}</span>
                      {conn.prefix && (
                        <>
                          {" "}&middot; <code>{conn.prefix}/</code>
                        </>
                      )}
                      {conn.last_stats && (
                        <>
                          {" "}&middot; {conn.last_stats.imported} docs
                        </>
                      )}
                      {" "}&middot; {conn.last_run ? `synced ${timeAgo(conn.last_run)}` : "never synced"}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                  <Button
                    size="sm"
                    variant="ghost"
                    title="Sync now"
                    onClick={(e) => { e.stopPropagation(); handleReImport(conn); }}
                    disabled={reImporting === conn.id}
                  >
                    <RefreshCw className={`h-3.5 w-3.5 ${reImporting === conn.id ? "animate-spin" : ""}`} />
                  </Button>
                  <Button
                    size="sm"
                    variant="ghost"
                    title="Remove"
                    onClick={(e) => { e.stopPropagation(); handleDelete(conn.id); }}
                  >
                    <Trash2 className="h-3.5 w-3.5 text-destructive" />
                  </Button>
                </div>
              </div>
            </div>
          ))}

          <Button
            variant="outline"
            onClick={() => setWizardOpen(true)}
            className="w-full gap-2 mt-4"
          >
            <Plus className="h-4 w-4" />
            Connect a source
          </Button>
        </div>
      )}
    </div>
  );
}

function timeAgo(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return "just now";
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffH = Math.floor(diffMin / 60);
  if (diffH < 24) return `${diffH}h ago`;
  const diffD = Math.floor(diffH / 24);
  return `${diffD}d ago`;
}

// KiwiBases — A database view over frontmatter queries with Table, Card, List,
// and Map layouts. Views are saved server-side and executed on demand.

import { useCallback, useEffect, useState } from "react";
import {
  ArrowLeft,
  Filter,
  Grid2X2,
  LayoutList,
  Loader2,
  MapPin,
  Plus,
  Search as SearchIcon,
  SortAsc,
  Table2,
} from "lucide-react";
import { api } from "@kw/lib/api";
import { cn } from "@kw/lib/cn";
import { Button } from "@kw/components/ui/button";
import { Input } from "@kw/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@kw/components/ui/select";
import { BasesTable, type ViewColumn, type ViewRow } from "./bases/BasesTable";
import { BasesCards } from "./bases/BasesCards";
import { BasesList } from "./bases/BasesList";
import { BasesMap } from "./bases/BasesMap";
import { BasesFilterPanel, type ViewFilter } from "./bases/BasesFilterPanel";
import { BasesSortPanel, type SortKey } from "./bases/BasesSortPanel";
import { BasesViewDialog, type ViewDefinition } from "./bases/BasesViewDialog";

type LayoutKind = "table" | "cards" | "list" | "map";

type Props = {
  onClose: () => void;
  onNavigate: (path: string) => void;
};

type SavedView = {
  name: string;
  query: string;
  layout: LayoutKind;
  columns: ViewColumn[];
  filters: ViewFilter[];
  sort: SortKey[];
  group_by?: string;
};

export function KiwiBases({ onClose, onNavigate }: Props) {
  // Views list
  const [views, setViews] = useState<SavedView[]>([]);
  const [viewsLoading, setViewsLoading] = useState(true);
  const [activeViewName, setActiveViewName] = useState<string | null>(null);

  // View data
  const [data, setData] = useState<ViewRow[]>([]);
  const [dataLoading, setDataLoading] = useState(false);

  // Layout
  const [layout, setLayout] = useState<LayoutKind>("table");
  const [cardSize] = useState<"small" | "medium" | "large">("medium");

  // Columns (extracted from view or data)
  const [columns, setColumns] = useState<ViewColumn[]>([]);

  // Filters & sorts
  const [filters, setFilters] = useState<ViewFilter[]>([]);
  const [conjunction, setConjunction] = useState<"and" | "or">("and");
  const [sorts, setSorts] = useState<SortKey[]>([]);
  const [filterOpen, setFilterOpen] = useState(false);
  const [sortOpen, setSortOpen] = useState(false);

  // Client-side search
  const [search, setSearch] = useState("");

  // View dialog
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [editingView, setEditingView] = useState<ViewDefinition | null>(null);

  // Properties for filter/sort selectors
  const properties = columns.map((c) => c.key);

  // Load views list
  useEffect(() => {
    setViewsLoading(true);
    api
      .listViews()
      .then((r) => {
        const v = r.views || [];
        setViews(v as unknown as SavedView[]);
        if (v.length > 0 && !activeViewName) {
          setActiveViewName(v[0].name);
        }
      })
      .catch(() => setViews([]))
      .finally(() => setViewsLoading(false));
  }, []);

  // Execute active view
  const executeView = useCallback(
    async (viewName: string) => {
      setDataLoading(true);
      try {
        const result = await api.executeView(viewName);
        const rows: ViewRow[] = (result.rows || []).map(
          (r: Record<string, unknown>) => ({
            path: String(r.path || ""),
            title: String(r.title || r.path || ""),
            ...r,
          }),
        );
        setData(rows);

        // Auto-detect columns from data keys
        if (rows.length > 0) {
          const keys = new Set<string>();
          for (const row of rows.slice(0, 20)) {
            for (const key of Object.keys(row)) {
              keys.add(key);
            }
          }
          // Always put title first, path second
          const ordered: ViewColumn[] = [];
          if (keys.has("title")) {
            ordered.push({ key: "title", label: "Title" });
            keys.delete("title");
          }
          if (keys.has("path")) {
            ordered.push({ key: "path", label: "Path" });
            keys.delete("path");
          }
          for (const k of Array.from(keys).sort()) {
            ordered.push({ key: k, label: k.charAt(0).toUpperCase() + k.slice(1) });
          }
          setColumns(ordered);
        }
      } catch {
        setData([]);
      } finally {
        setDataLoading(false);
      }
    },
    [],
  );

  // When active view changes, load its config and execute
  useEffect(() => {
    if (!activeViewName) return;
    const view = views.find((v) => v.name === activeViewName);
    if (view) {
      setLayout(view.layout || "table");
      if (view.columns?.length) setColumns(view.columns);
      if (view.filters?.length) setFilters(view.filters);
    }
    executeView(activeViewName);
  }, [activeViewName, executeView]);

  // Client-side search filtering
  const filteredData = search
    ? data.filter((row) => {
        const q = search.toLowerCase();
        return (
          row.title.toLowerCase().includes(q) ||
          row.path.toLowerCase().includes(q) ||
          Object.values(row).some(
            (v) => typeof v === "string" && v.toLowerCase().includes(q),
          )
        );
      })
    : data;

  // Handle new view save
  function handleSaveView(view: ViewDefinition) {
    api
      .saveView(view.name, view as unknown as Record<string, unknown>)
      .then(() => {
        setViews((prev) => {
          const exists = prev.findIndex((v) => v.name === view.name);
          if (exists >= 0) {
            const next = [...prev];
            next[exists] = view as unknown as SavedView;
            return next;
          }
          return [...prev, view as unknown as SavedView];
        });
        setActiveViewName(view.name);
      })
      .catch(() => {});
  }

  const LAYOUT_ICONS: Record<LayoutKind, typeof Table2> = {
    table: Table2,
    cards: Grid2X2,
    list: LayoutList,
    map: MapPin,
  };

  return (
    <div className="h-full flex flex-col">
      {/* ── Top bar ── */}
      <div className="flex flex-wrap items-center gap-2 px-3 sm:px-6 py-3 border-b border-border bg-card">
        <Button variant="outline" size="sm" onClick={onClose}>
          <ArrowLeft className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Back</span>
        </Button>
        <div className="font-semibold text-sm">Bases</div>

        {/* View selector */}
        {views.length > 0 && (
          <Select
            value={activeViewName || ""}
            onValueChange={(v) => setActiveViewName(v)}
          >
            <SelectTrigger className="h-8 w-40 text-sm">
              <SelectValue placeholder="Select view" />
            </SelectTrigger>
            <SelectContent>
              {views.map((v) => (
                <SelectItem key={v.name} value={v.name}>
                  {v.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}

        <Button
          variant="outline"
          size="sm"
          className="gap-1"
          onClick={() => {
            setEditingView(null);
            setViewDialogOpen(true);
          }}
        >
          <Plus className="h-3.5 w-3.5" /> New view
        </Button>

        <div className="ml-auto flex items-center gap-2">
          {/* Layout switcher */}
          <div className="flex items-center border border-border rounded-md">
            {(Object.keys(LAYOUT_ICONS) as LayoutKind[]).map((l) => {
              const Icon = LAYOUT_ICONS[l];
              return (
                <button
                  key={l}
                  type="button"
                  className={cn(
                    "h-8 w-8 grid place-items-center text-muted-foreground transition-colors",
                    layout === l && "bg-accent text-accent-foreground",
                    "hover:text-foreground",
                  )}
                  onClick={() => setLayout(l)}
                  title={l}
                >
                  <Icon className="h-3.5 w-3.5" />
                </button>
              );
            })}
          </div>

          {/* Filter */}
          <Button
            variant={filterOpen ? "secondary" : "outline"}
            size="sm"
            className="gap-1"
            onClick={() => {
              setFilterOpen((v) => !v);
              setSortOpen(false);
            }}
          >
            <Filter className="h-3.5 w-3.5" />
            {filters.length > 0 && (
              <span className="text-xs">({filters.length})</span>
            )}
          </Button>

          {/* Sort */}
          <Button
            variant={sortOpen ? "secondary" : "outline"}
            size="sm"
            className="gap-1"
            onClick={() => {
              setSortOpen((v) => !v);
              setFilterOpen(false);
            }}
          >
            <SortAsc className="h-3.5 w-3.5" />
            {sorts.length > 0 && (
              <span className="text-xs">({sorts.length})</span>
            )}
          </Button>

          {/* Search */}
          <div className="relative">
            <SearchIcon className="h-3.5 w-3.5 absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none" />
            <Input
              type="text"
              placeholder="Search..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="h-8 pl-7 w-32 sm:w-48 text-sm"
            />
          </div>
        </div>
      </div>

      {/* ── Filter / Sort panels ── */}
      <div className="relative">
        {filterOpen && (
          <div className="absolute top-2 left-4 z-20">
            <BasesFilterPanel
              filters={filters}
              onChange={setFilters}
              conjunction={conjunction}
              onConjunctionChange={setConjunction}
              properties={properties}
              onClose={() => setFilterOpen(false)}
            />
          </div>
        )}
        {sortOpen && (
          <div className="absolute top-2 right-4 z-20">
            <BasesSortPanel
              sorts={sorts}
              onChange={setSorts}
              properties={properties}
              onClose={() => setSortOpen(false)}
            />
          </div>
        )}
      </div>

      {/* ── Content ── */}
      <div className="flex-1 overflow-auto">
        {viewsLoading || dataLoading ? (
          <div className="flex items-center justify-center h-64 text-muted-foreground">
            <Loader2 className="h-5 w-5 animate-spin mr-2" />
            Loading...
          </div>
        ) : views.length === 0 && !viewsLoading ? (
          <div className="flex flex-col items-center justify-center h-64 text-muted-foreground gap-3">
            <p className="text-sm">No views yet. Create one to get started.</p>
            <Button
              variant="outline"
              onClick={() => {
                setEditingView(null);
                setViewDialogOpen(true);
              }}
            >
              <Plus className="h-4 w-4 mr-1" /> Create view
            </Button>
          </div>
        ) : layout === "table" ? (
          <BasesTable
            columns={columns}
            data={filteredData}
            onNavigate={onNavigate}
          />
        ) : layout === "cards" ? (
          <BasesCards
            columns={columns}
            data={filteredData}
            onNavigate={onNavigate}
            cardSize={cardSize}
          />
        ) : layout === "list" ? (
          <BasesList
            columns={columns}
            data={filteredData}
            onNavigate={onNavigate}
          />
        ) : layout === "map" ? (
          <BasesMap data={filteredData} onNavigate={onNavigate} />
        ) : null}
      </div>

      {/* ── View dialog ── */}
      <BasesViewDialog
        open={viewDialogOpen}
        onOpenChange={setViewDialogOpen}
        onSave={handleSaveView}
        initial={editingView}
      />
    </div>
  );
}

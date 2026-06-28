import { useEffect, useState, useCallback } from "react";
import {
  BarChart3,
  TrendingUp,
  TrendingDown,
  Minus,
  Search,
  SearchX,
  Eye,
  FileQuestion,
  X,
} from "lucide-react";
import {
  api,
  type OverviewStats,
  type AnalyticsViewsResponse,
  type AnalyticsTrendsResponse,
  type AnalyticsContentGapsResponse,
  type AnalyticsSourcesResponse,
  type TimePoint,
} from "@kw/lib/api";
import { Button } from "./ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./ui/select";

type Props = {
  onClose: () => void;
  onNavigate: (path: string) => void;
};

// --- Sparkline (pure SVG, ~30 lines, no dependency) ---

function Sparkline({
  data,
  width = 120,
  height = 32,
  className = "",
}: {
  data: TimePoint[];
  width?: number;
  height?: number;
  className?: string;
}) {
  if (data.length < 2) return null;
  const counts = data.map((d) => d.count);
  const max = Math.max(...counts, 1);
  const min = Math.min(...counts, 0);
  const range = max - min || 1;
  const points = counts
    .map((v, i) => {
      const x = (i / (counts.length - 1)) * width;
      const y = height - ((v - min) / range) * (height - 4) - 2;
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(" ");

  return (
    <svg
      width={width}
      height={height}
      viewBox={`0 0 ${width} ${height}`}
      className={className}
      aria-hidden
    >
      <polyline
        points={points}
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="text-primary"
      />
    </svg>
  );
}

// --- Delta indicator ---

function Delta({ value, suffix = "%" }: { value: number; suffix?: string }) {
  if (value === 0) {
    return (
      <span className="inline-flex items-center gap-0.5 text-xs text-muted-foreground">
        <Minus className="h-3 w-3" />
        0{suffix}
      </span>
    );
  }
  const positive = value > 0;
  return (
    <span
      className={`inline-flex items-center gap-0.5 text-xs ${
        positive ? "text-green-600 dark:text-green-400" : "text-red-600 dark:text-red-400"
      }`}
    >
      {positive ? (
        <TrendingUp className="h-3 w-3" />
      ) : (
        <TrendingDown className="h-3 w-3" />
      )}
      {positive ? "+" : ""}
      {value.toFixed(1)}
      {suffix}
    </span>
  );
}

// --- Stacked source bar ---

const SOURCE_COLORS: Record<string, string> = {
  ui: "bg-blue-500",
  api: "bg-emerald-500",
  mcp: "bg-amber-500",
  s3: "bg-purple-500",
  webdav: "bg-rose-500",
};

function SourceBar({ sources }: { sources: Record<string, number> }) {
  const total = Object.values(sources).reduce((a, b) => a + b, 0);
  if (total === 0) {
    return <p className="text-sm text-muted-foreground">No data.</p>;
  }
  const entries = Object.entries(sources).sort((a, b) => b[1] - a[1]);
  return (
    <div>
      <div className="flex h-3 w-full rounded overflow-hidden gap-px">
        {entries.map(([src, count]) => (
          <div
            key={src}
            className={`${SOURCE_COLORS[src] ?? "bg-gray-400"} transition-all`}
            style={{ width: `${(count / total) * 100}%` }}
            title={`${src}: ${count.toLocaleString()}`}
          />
        ))}
      </div>
      <div className="flex flex-wrap gap-3 mt-2">
        {entries.map(([src, count]) => (
          <span key={src} className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <span
              className={`inline-block w-2.5 h-2.5 rounded-sm ${SOURCE_COLORS[src] ?? "bg-gray-400"}`}
            />
            {src} ({count.toLocaleString()})
          </span>
        ))}
      </div>
    </div>
  );
}

// --- Main dashboard ---

export function KiwiAnalytics({ onClose, onNavigate }: Props) {
  const [period, setPeriod] = useState("7d");
  const [overview, setOverview] = useState<OverviewStats | null>(null);
  const [views, setViews] = useState<AnalyticsViewsResponse | null>(null);
  const [trends, setTrends] = useState<AnalyticsTrendsResponse | null>(null);
  const [gaps, setGaps] = useState<AnalyticsContentGapsResponse | null>(null);
  const [sources, setSources] = useState<AnalyticsSourcesResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(
    (p: string) => {
      setLoading(true);
      setError(null);
      Promise.all([
        api.analyticsOverview(p).catch(() => null),
        api.analyticsViews(p).catch(() => null),
        api.analyticsTrends(p).catch(() => null),
        api.analyticsContentGaps().catch(() => null),
        api.analyticsSources(p).catch(() => null),
      ])
        .then(([ov, vi, tr, cg, sr]) => {
          setOverview(ov);
          setViews(vi);
          setTrends(tr);
          setGaps(cg);
          setSources(sr);
        })
        .catch((e) => setError(String(e)))
        .finally(() => setLoading(false));
    },
    [],
  );

  useEffect(() => {
    load(period);
  }, [period, load]);

  const handleDismiss = async (query: string) => {
    try {
      await api.dismissContentGap(query);
      // Refresh content gaps
      const cg = await api.analyticsContentGaps();
      setGaps(cg);
    } catch {
      // silently ignore
    }
  };

  return (
    <div className="flex flex-col h-full bg-background">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-border shrink-0">
        <div className="flex items-center gap-2">
          <BarChart3 className="h-5 w-5 text-primary" />
          <h1 className="text-lg font-semibold">Analytics</h1>
        </div>
        <div className="flex items-center gap-2">
          <Select value={period} onValueChange={setPeriod}>
            <SelectTrigger className="w-[100px] h-8 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="7d">7 days</SelectItem>
              <SelectItem value="30d">30 days</SelectItem>
              <SelectItem value="90d">90 days</SelectItem>
            </SelectContent>
          </Select>
          <Button variant="ghost" size="icon" onClick={onClose} aria-label="Close">
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto kiwi-scroll p-4 md:p-6 max-w-4xl mx-auto w-full">
        {loading && (
          <p className="text-sm text-muted-foreground">Loading analytics...</p>
        )}
        {error && <p className="text-sm text-destructive">{error}</p>}

        {!loading && !error && !overview && !views && !trends && !gaps && !sources && (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <BarChart3 className="h-10 w-10 text-muted-foreground/40 mb-3" />
            <p className="text-sm text-muted-foreground">
              Analytics data is not available yet.
            </p>
            <p className="text-xs text-muted-foreground mt-1">
              The KiwiFS instance may need to be updated, or no page views have been recorded.
            </p>
          </div>
        )}

        {!loading && !error && (overview || views || trends || gaps || sources) && (
          <div className="space-y-6">
            {/* Summary cards */}
            {overview && (
              <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                <Card>
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-xs font-medium text-muted-foreground flex items-center gap-1">
                      <Eye className="h-3.5 w-3.5" /> Views
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="p-4 pt-0">
                    <p className="text-2xl font-bold tabular-nums">
                      {overview.total_views.toLocaleString()}
                    </p>
                    <Delta value={overview.views_delta_percent} />
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-xs font-medium text-muted-foreground flex items-center gap-1">
                      <Search className="h-3.5 w-3.5" /> Searches
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="p-4 pt-0">
                    <p className="text-2xl font-bold tabular-nums">
                      {overview.total_searches.toLocaleString()}
                    </p>
                    <Delta value={overview.searches_delta_percent} />
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-xs font-medium text-muted-foreground">
                      Search success
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="p-4 pt-0">
                    <p className="text-2xl font-bold tabular-nums">
                      {(overview.search_success_rate * 100).toFixed(0)}%
                    </p>
                    <Delta value={overview.success_rate_delta_pp} suffix="pp" />
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="p-4 pb-2">
                    <CardTitle className="text-xs font-medium text-muted-foreground">
                      Unique pages
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="p-4 pt-0">
                    <p className="text-2xl font-bold tabular-nums">
                      {overview.unique_pages_viewed.toLocaleString()}
                    </p>
                    <Delta value={overview.unique_pages_delta_percent} />
                  </CardContent>
                </Card>
              </div>
            )}

            {/* View sparkline + top pages */}
            {views && (
              <section>
                <h2 className="text-sm font-medium mb-3 flex items-center gap-2">
                  <BarChart3 className="h-4 w-4" />
                  Page views
                </h2>
                {views.time_series.length > 1 && (
                  <Sparkline data={views.time_series} width={600} height={48} className="w-full mb-4" />
                )}
                {views.top_pages.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No views recorded yet.</p>
                ) : (
                  <ul className="space-y-1">
                    {views.top_pages.slice(0, 10).map((row) => (
                      <li key={row.path}>
                        <button
                          type="button"
                          onClick={() => {
                            onClose();
                            onNavigate(row.path);
                          }}
                          className="flex w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-sm hover:bg-accent transition-colors text-left"
                        >
                          <span className="truncate font-medium">{row.path}</span>
                          <span className="shrink-0 text-muted-foreground tabular-nums">
                            {row.count.toLocaleString()} view{row.count === 1 ? "" : "s"}
                          </span>
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
              </section>
            )}

            {/* Trending pages */}
            {trends && trends.trending.length > 0 && (
              <section>
                <h2 className="text-sm font-medium mb-3 flex items-center gap-2">
                  <TrendingUp className="h-4 w-4" />
                  Trending pages
                </h2>
                <ul className="space-y-1">
                  {trends.trending.slice(0, 8).map((row) => (
                    <li key={row.path}>
                      <button
                        type="button"
                        onClick={() => {
                          onClose();
                          onNavigate(row.path);
                        }}
                        className="flex w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-sm hover:bg-accent transition-colors text-left"
                      >
                        <span className="truncate font-medium">{row.path}</span>
                        <span className="shrink-0">
                          <Delta value={row.delta_percent} />
                        </span>
                      </button>
                    </li>
                  ))}
                </ul>
              </section>
            )}

            {/* Declining pages */}
            {trends && trends.declining.length > 0 && (
              <section>
                <h2 className="text-sm font-medium mb-3 flex items-center gap-2">
                  <TrendingDown className="h-4 w-4" />
                  Declining pages
                </h2>
                <ul className="space-y-1">
                  {trends.declining.slice(0, 5).map((row) => (
                    <li key={row.path}>
                      <button
                        type="button"
                        onClick={() => {
                          onClose();
                          onNavigate(row.path);
                        }}
                        className="flex w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-sm hover:bg-accent transition-colors text-left"
                      >
                        <span className="truncate font-medium">{row.path}</span>
                        <span className="shrink-0">
                          <Delta value={row.delta_percent} />
                        </span>
                      </button>
                    </li>
                  ))}
                </ul>
              </section>
            )}

            {/* Content gaps */}
            {gaps && gaps.results.length > 0 && (
              <section>
                <h2 className="text-sm font-medium mb-3 flex items-center gap-2">
                  <FileQuestion className="h-4 w-4" />
                  Content gaps
                </h2>
                <p className="text-xs text-muted-foreground mb-2">
                  Searches that returned zero results. Create a page or dismiss if irrelevant.
                </p>
                <ul className="space-y-1">
                  {gaps.results.slice(0, 10).map((row) => (
                    <li
                      key={`${row.search_type}:${row.query}`}
                      className="flex items-center justify-between gap-3 rounded-md px-3 py-2 text-sm bg-muted/40"
                    >
                      <span className="truncate font-mono text-xs">{row.query}</span>
                      <div className="flex items-center gap-2 shrink-0">
                        <span className="text-muted-foreground tabular-nums text-xs">
                          {row.count}x
                        </span>
                        <Button
                          variant="ghost"
                          size="sm"
                          className="h-6 px-2 text-xs"
                          onClick={() => handleDismiss(row.query)}
                        >
                          Dismiss
                        </Button>
                      </div>
                    </li>
                  ))}
                </ul>
              </section>
            )}

            {/* Source breakdown */}
            {sources && Object.keys(sources.sources).length > 0 && (
              <section>
                <h2 className="text-sm font-medium mb-3 flex items-center gap-2">
                  <SearchX className="h-4 w-4" />
                  Source breakdown
                </h2>
                <SourceBar sources={sources.sources} />
              </section>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

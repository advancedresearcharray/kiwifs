import { useEffect, useState } from "react";
import { BarChart3, SearchX, X } from "lucide-react";
import { api, type AnalyticsResponse } from "@kw/lib/api";
import { Button } from "./ui/button";

type Props = {
  onClose: () => void;
  onNavigate: (path: string) => void;
};

export function KiwiEngagement({ onClose, onNavigate }: Props) {
  const [data, setData] = useState<AnalyticsResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    api
      .analytics()
      .then((r) => {
        if (!cancelled) setData(r);
      })
      .catch((e) => {
        if (!cancelled) setError(String(e));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="flex flex-col h-full bg-background">
      <div className="flex items-center justify-between px-4 py-3 border-b border-border shrink-0">
        <div className="flex items-center gap-2">
          <BarChart3 className="h-5 w-5 text-primary" />
          <h1 className="text-lg font-semibold">Engagement analytics</h1>
        </div>
        <Button variant="ghost" size="icon" onClick={onClose} aria-label="Close">
          <X className="h-4 w-4" />
        </Button>
      </div>

      <div className="flex-1 overflow-auto kiwi-scroll p-4 md:p-6 max-w-3xl mx-auto w-full">
        {loading && (
          <p className="text-sm text-muted-foreground">Loading analytics…</p>
        )}
        {error && (
          <p className="text-sm text-destructive">{error}</p>
        )}
        {data && !loading && !error && (
          <div className="space-y-8">
            <section>
              <h2 className="text-sm font-medium text-muted-foreground uppercase tracking-wide mb-2">
                Overview
              </h2>
              <p className="text-3xl font-bold tabular-nums">
                {data.engagement.total_views.toLocaleString()}
              </p>
              <p className="text-sm text-muted-foreground mt-1">
                total page views across {data.total_pages.toLocaleString()} indexed pages
              </p>
            </section>

            <section>
              <h2 className="text-sm font-medium mb-3 flex items-center gap-2">
                <BarChart3 className="h-4 w-4" />
                Most viewed pages
              </h2>
              {data.engagement.top_viewed.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No views recorded yet. Open pages in the wiki UI to start tracking.
                </p>
              ) : (
                <ul className="space-y-1">
                  {data.engagement.top_viewed.map((row) => (
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

            <section>
              <h2 className="text-sm font-medium mb-3 flex items-center gap-2">
                <SearchX className="h-4 w-4" />
                Failed searches
              </h2>
              <p className="text-xs text-muted-foreground mb-2">
                Queries that returned zero results (helps find content gaps).
              </p>
              {data.engagement.failed_searches.length === 0 ? (
                <p className="text-sm text-muted-foreground">No failed searches recorded.</p>
              ) : (
                <ul className="space-y-1">
                  {data.engagement.failed_searches.map((row) => (
                    <li
                      key={`${row.search_type}:${row.query}`}
                      className="flex items-center justify-between gap-3 rounded-md px-3 py-2 text-sm bg-muted/40"
                    >
                      <span className="truncate font-mono text-xs">{row.query}</span>
                      <span className="shrink-0 text-muted-foreground tabular-nums">
                        {row.count}× · {row.search_type}
                      </span>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          </div>
        )}
      </div>
    </div>
  );
}

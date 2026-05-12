// KiwiTimeline — Chronological feed of changes across the knowledge base.
// Shows events grouped by day with filters and infinite scroll.

import { useCallback, useEffect, useRef, useState } from "react";
import {
  ArrowLeft,
  Edit3,
  Loader2,
  Rss,
  Trash2,
} from "lucide-react";
import {
  formatDistanceToNow,
  format,
  isToday,
  isYesterday,
  parseISO,
} from "date-fns";
import { api, type TimelineEvent } from "@kw/lib/api";
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
import { titleize } from "@kw/lib/paths";

type Props = {
  onClose: () => void;
  onNavigate: (path: string) => void;
};

type TimeRange = "24h" | "7d" | "30d" | "all";

const PAGE_SIZE = 50;

function dateLabel(iso: string): string {
  const d = parseISO(iso);
  if (isToday(d)) return "Today";
  if (isYesterday(d)) return "Yesterday";
  return format(d, "MMMM d, yyyy");
}

function groupByDate(events: TimelineEvent[]): Map<string, TimelineEvent[]> {
  const groups = new Map<string, TimelineEvent[]>();
  for (const ev of events) {
    const label = dateLabel(ev.timestamp);
    const list = groups.get(label);
    if (list) {
      list.push(ev);
    } else {
      groups.set(label, [ev]);
    }
  }
  return groups;
}

const EVENT_ICONS: Record<string, typeof Edit3> = {
  write: Edit3,
  delete: Trash2,
  rename: Edit3,
};

export function KiwiTimeline({ onClose, onNavigate }: Props) {
  const [events, setEvents] = useState<TimelineEvent[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);

  // Filters
  const [typeFilter, setTypeFilter] = useState<string>("all");
  const [actorFilter, setActorFilter] = useState<string>("all");
  const [prefixFilter, setPrefixFilter] = useState("");
  const [range, setRange] = useState<TimeRange>("7d");
  const [actors, setActors] = useState<string[]>([]);

  const scrollRef = useRef<HTMLDivElement>(null);

  // Load actors list
  useEffect(() => {
    api
      .getTimelineActors()
      .then((r) => setActors(r.actors || []))
      .catch(() => setActors([]));
  }, []);

  // Load events
  const loadEvents = useCallback(
    async (offset = 0, append = false) => {
      if (offset === 0) setLoading(true);
      else setLoadingMore(true);

      try {
        const params: Parameters<typeof api.getTimeline>[0] = {
          limit: PAGE_SIZE,
          offset,
          range,
        };
        if (typeFilter !== "all") params.type = typeFilter;
        if (actorFilter !== "all") params.actor = actorFilter;
        if (prefixFilter.trim()) params.prefix = prefixFilter.trim();

        const result = await api.getTimeline(params);
        if (append) {
          setEvents((prev) => [...prev, ...(result.events || [])]);
        } else {
          setEvents(result.events || []);
        }
        setTotal(result.total);
      } catch {
        if (!append) setEvents([]);
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [typeFilter, actorFilter, prefixFilter, range],
  );

  useEffect(() => {
    loadEvents(0, false);
  }, [loadEvents]);

  // Infinite scroll
  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    const onScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = el;
      if (
        scrollHeight - scrollTop - clientHeight < 200 &&
        !loadingMore &&
        events.length < total
      ) {
        loadEvents(events.length, true);
      }
    };
    el.addEventListener("scroll", onScroll, { passive: true });
    return () => el.removeEventListener("scroll", onScroll);
  }, [events.length, total, loadingMore, loadEvents]);

  const grouped = groupByDate(events);

  return (
    <div className="h-full flex flex-col">
      {/* Toolbar */}
      <div className="flex flex-wrap items-center gap-2 px-3 sm:px-6 py-3 border-b border-border bg-card">
        <Button variant="outline" size="sm" onClick={onClose}>
          <ArrowLeft className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Back</span>
        </Button>
        <div className="font-semibold text-sm">Timeline</div>
        <div className="text-xs text-muted-foreground hidden sm:block">
          {total} event{total !== 1 ? "s" : ""}
        </div>

        <div className="ml-auto flex items-center gap-2 flex-wrap">
          {/* Type filter */}
          <div className="flex items-center border border-border rounded-md text-xs">
            {["all", "write", "delete"].map((t) => (
              <button
                key={t}
                type="button"
                className={cn(
                  "px-2.5 py-1 transition-colors capitalize",
                  typeFilter === t
                    ? "bg-accent text-accent-foreground"
                    : "text-muted-foreground hover:text-foreground",
                )}
                onClick={() => setTypeFilter(t)}
              >
                {t === "all" ? "All" : t === "write" ? "Writes" : "Deletes"}
              </button>
            ))}
          </div>

          {/* Actor filter */}
          {actors.length > 0 && (
            <Select
              value={actorFilter}
              onValueChange={setActorFilter}
            >
              <SelectTrigger className="h-8 w-32 text-sm">
                <SelectValue placeholder="All actors" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All actors</SelectItem>
                {actors.map((a) => (
                  <SelectItem key={a} value={a}>
                    {a}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}

          {/* Path prefix */}
          <Input
            type="text"
            placeholder="Path prefix..."
            value={prefixFilter}
            onChange={(e) => setPrefixFilter(e.target.value)}
            className="h-8 w-28 sm:w-36 text-sm"
          />

          {/* Time range */}
          <Select value={range} onValueChange={(v) => setRange(v as TimeRange)}>
            <SelectTrigger className="h-8 w-28 text-sm">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="24h">Last 24h</SelectItem>
              <SelectItem value="7d">Last 7d</SelectItem>
              <SelectItem value="30d">Last 30d</SelectItem>
              <SelectItem value="all">All time</SelectItem>
            </SelectContent>
          </Select>

          {/* RSS */}
          <a
            href="/api/kiwi/feed.xml"
            target="_blank"
            rel="noopener noreferrer"
            className="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1"
          >
            <Rss className="h-3 w-3" /> Subscribe
          </a>
        </div>
      </div>

      {/* Events */}
      <div ref={scrollRef} className="flex-1 overflow-auto kiwi-scroll">
        {loading ? (
          <div className="flex items-center justify-center h-64 text-muted-foreground">
            <Loader2 className="h-5 w-5 animate-spin mr-2" /> Loading...
          </div>
        ) : events.length === 0 ? (
          <div className="flex items-center justify-center h-64 text-muted-foreground text-sm">
            No events match the current filters.
          </div>
        ) : (
          <div className="max-w-3xl mx-auto py-4 px-4">
            {Array.from(grouped.entries()).map(([dateStr, dayEvents]) => (
              <div key={dateStr} className="mb-6">
                <div className="sticky top-0 bg-background/95 backdrop-blur z-10 text-xs font-medium text-muted-foreground uppercase tracking-wider py-2 border-b border-border/50 mb-2">
                  {dateStr}
                </div>
                <div className="space-y-1">
                  {dayEvents.map((ev, i) => {
                    const Icon = EVENT_ICONS[ev.type] || Edit3;
                    const d = parseISO(ev.timestamp);
                    return (
                      <div
                        key={`${ev.path}-${ev.timestamp}-${i}`}
                        className="flex items-start gap-3 py-2 px-2 rounded-md hover:bg-accent/30 transition-colors group"
                      >
                        <div
                          className={cn(
                            "mt-0.5 h-6 w-6 rounded-full grid place-items-center shrink-0",
                            ev.type === "delete"
                              ? "bg-destructive/10 text-destructive"
                              : "bg-primary/10 text-primary",
                          )}
                        >
                          <Icon className="h-3 w-3" />
                        </div>
                        <div className="min-w-0 flex-1">
                          <div className="flex items-center gap-2 flex-wrap">
                            <button
                              type="button"
                              className="font-medium text-sm hover:underline truncate"
                              onClick={() => onNavigate(ev.path)}
                            >
                              {ev.title || titleize(ev.path)}
                            </button>
                            <span className="text-xs text-muted-foreground capitalize">
                              {ev.type}
                            </span>
                          </div>
                          <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                            <span>{ev.actor}</span>
                            <span>·</span>
                            <time
                              dateTime={ev.timestamp}
                              title={format(d, "PPpp")}
                            >
                              {formatDistanceToNow(d, { addSuffix: true })}
                            </time>
                          </div>
                          {ev.message && (
                            <div className="text-xs text-muted-foreground mt-0.5 truncate">
                              {ev.message}
                            </div>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}

            {/* Load more indicator */}
            {loadingMore && (
              <div className="flex items-center justify-center py-4 text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin mr-2" /> Loading more...
              </div>
            )}
            {events.length >= total && events.length > 0 && (
              <div className="text-center text-xs text-muted-foreground py-4">
                End of timeline
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

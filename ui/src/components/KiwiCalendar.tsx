// KiwiCalendar — Monthly calendar view for pages with date-based frontmatter fields.

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ArrowLeft,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Loader2,
} from "lucide-react";
import { format, parseISO, startOfMonth } from "date-fns";
import { api } from "@kw/lib/api";
import { cn } from "@kw/lib/cn";
import {
  buildMonthRangeQuery,
  defaultSelectedDateKey,
  discoverDateFields,
  groupEntriesByDate,
  monthLabel,
  parseCalendarQueryRows,
  parseMonthInput,
  shiftMonth,
  todayKey,
  weekGridAnchor,
  type CalendarPageEntry,
} from "@kw/lib/calendarView";
import { Button } from "@kw/components/ui/button";
import { Input } from "@kw/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@kw/components/ui/select";
import { CalendarGrid } from "./calendar/CalendarGrid";
import { CalendarDayPanel } from "./calendar/CalendarDayPanel";

type Props = {
  onClose: () => void;
  onNavigate: (path: string) => void;
};

const DATE_FIELD_STORAGE_KEY = "kiwifs-calendar-date-field";

function readStoredDateField(): string {
  try {
    return localStorage.getItem(DATE_FIELD_STORAGE_KEY) || "date";
  } catch {
    return "date";
  }
}

export function KiwiCalendar({ onClose, onNavigate }: Props) {
  const [month, setMonth] = useState(() => startOfMonth(new Date()));
  const [dateField, setDateField] = useState(readStoredDateField);
  const [dateFields, setDateFields] = useState<string[]>(["date"]);
  const [entries, setEntries] = useState<CalendarPageEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedDateKey, setSelectedDateKey] = useState<string | null>(todayKey());
  const [isMobile, setIsMobile] = useState(
    () => typeof window !== "undefined" && window.innerWidth < 768,
  );

  useEffect(() => {
    const mq = window.matchMedia("(max-width: 767px)");
    const onChange = (e: MediaQueryListEvent) => setIsMobile(e.matches);
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  useEffect(() => {
    let cancelled = false;
    api
      .meta({ limit: 500 })
      .then((result) => {
        if (cancelled) return;
        const discovered = discoverDateFields(result.results ?? []);
        setDateFields(discovered);
        setDateField((current) =>
          discovered.includes(current) ? current : (discovered[0] ?? "date"),
        );
      })
      .catch(() => {
        if (!cancelled) setDateFields(["date", "due", "created"]);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const loadMonth = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.query(buildMonthRangeQuery(dateField, month));
      setEntries(parseCalendarQueryRows(data));
    } catch (err) {
      setEntries([]);
      setError(err instanceof Error ? err.message : "Failed to load calendar data");
    } finally {
      setLoading(false);
    }
  }, [dateField, month]);

  useEffect(() => {
    void loadMonth();
  }, [loadMonth]);

  useEffect(() => {
    try {
      localStorage.setItem(DATE_FIELD_STORAGE_KEY, dateField);
    } catch {
      /* ignore */
    }
  }, [dateField]);

  useEffect(() => {
    setSelectedDateKey((current) => {
      if (!current) return defaultSelectedDateKey(month);
      try {
        const parsed = parseISO(current);
        if (
          !Number.isNaN(parsed.getTime()) &&
          parsed.getFullYear() === month.getFullYear() &&
          parsed.getMonth() === month.getMonth()
        ) {
          return current;
        }
      } catch {
        /* fall through */
      }
      return defaultSelectedDateKey(month);
    });
  }, [month]);

  const byDate = useMemo(() => groupEntriesByDate(entries), [entries]);
  const weekAnchor = useMemo(
    () => weekGridAnchor(selectedDateKey, month),
    [selectedDateKey, month],
  );
  const selectedEntries = selectedDateKey ? byDate.get(selectedDateKey) ?? [] : [];

  const monthPickerValue = format(month, "yyyy-MM");

  return (
    <div className="h-full flex flex-col">
      <div className="flex flex-wrap items-center gap-2 px-3 sm:px-6 py-3 border-b border-border bg-card">
        <Button variant="outline" size="sm" onClick={onClose}>
          <ArrowLeft className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Back</span>
        </Button>
        <div className="flex items-center gap-2">
          <CalendarDays className="h-4 w-4 text-muted-foreground" />
          <div className="font-semibold text-sm">Calendar</div>
        </div>
        <div className="text-xs text-muted-foreground hidden sm:block">
          {entries.length} page{entries.length === 1 ? "" : "s"} in {monthLabel(month)}
        </div>

        <div className="ml-auto flex flex-wrap items-center gap-2">
          <Select value={dateField} onValueChange={setDateField}>
            <SelectTrigger className="h-8 w-[9.5rem] text-xs">
              <SelectValue placeholder="Date field" />
            </SelectTrigger>
            <SelectContent>
              {dateFields.map((field) => (
                <SelectItem key={field} value={field}>
                  {field}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <div className="flex items-center gap-1">
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              aria-label="Previous month"
              onClick={() => setMonth((m) => shiftMonth(m, -1))}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Input
              type="month"
              value={monthPickerValue}
              onChange={(e) => {
                const parsed = parseMonthInput(e.target.value);
                if (parsed) setMonth(parsed);
              }}
              className="h-8 w-[9.5rem] text-xs"
              aria-label="Select month"
            />
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              aria-label="Next month"
              onClick={() => setMonth((m) => shiftMonth(m, 1))}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-8"
              onClick={() => {
                const now = startOfMonth(new Date());
                setMonth(now);
                setSelectedDateKey(todayKey());
              }}
            >
              Today
            </Button>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-hidden">
        {loading ? (
          <div className="flex h-full items-center justify-center text-muted-foreground">
            <Loader2 className="h-5 w-5 animate-spin mr-2" />
            Loading calendar…
          </div>
        ) : error ? (
          <div className="flex h-full items-center justify-center text-sm text-destructive px-6 text-center">
            {error}
          </div>
        ) : (
          <div
            className={cn(
              "h-full gap-4 p-3 sm:p-6",
              isMobile
                ? "flex flex-col overflow-auto"
                : "grid grid-cols-[minmax(0,1fr)_20rem] overflow-hidden",
            )}
          >
            <div className={cn(isMobile ? "shrink-0" : "overflow-auto kiwi-scroll")}>
              <CalendarGrid
                month={month}
                weekAnchor={weekAnchor}
                byDate={byDate}
                selectedDateKey={selectedDateKey}
                weekOnly={isMobile}
                onSelectDate={setSelectedDateKey}
              />
            </div>
            <div className={cn(isMobile ? "min-h-[16rem]" : "overflow-hidden")}>
              {selectedDateKey ? (
                <CalendarDayPanel
                  dateKey={selectedDateKey}
                  entries={selectedEntries}
                  onNavigate={onNavigate}
                />
              ) : (
                <div className="flex h-full items-center justify-center rounded-lg border border-dashed border-border text-sm text-muted-foreground">
                  Select a day to view pages
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

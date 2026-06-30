import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ArrowLeft,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Loader2,
} from "lucide-react";
import { api } from "@kw/lib/api";
import { cn } from "@kw/lib/cn";
import {
  addMonths,
  buildCalendarQuery,
  buildCalendarQueryRange,
  buildMonthGrid,
  dayAfter,
  DEFAULT_DATE_FIELDS,
  detectDateFields,
  entryDotColor,
  formatYearMonth,
  groupByDate,
  MONTH_NAMES,
  pageTitleFromPath,
  parseCalendarResponse,
  parseYearMonth,
  todayISO,
  WEEKDAYS,
  weekDateKeys,
  type CalendarEntry,
} from "@kw/lib/calendarView";
import { titleize } from "@kw/lib/paths";
import { Button } from "@kw/components/ui/button";
import { Badge } from "@kw/components/ui/badge";
import { Card, CardContent } from "@kw/components/ui/card";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@kw/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@kw/components/ui/select";

type Props = {
  onClose: () => void;
  onNavigate: (path: string) => void;
  isMobile?: boolean;
};

const DATE_FIELD_STORAGE_KEY = "kiwifs-calendar-date-field";

function loadSavedDateField(): string {
  try {
    const saved = localStorage.getItem(DATE_FIELD_STORAGE_KEY);
    if (saved) return saved;
  } catch {
    // ignore
  }
  return DEFAULT_DATE_FIELDS[0]!;
}

export function KiwiCalendar({ onClose, onNavigate, isMobile = false }: Props) {
  const [yearMonth, setYearMonth] = useState(() => {
    const t = new Date();
    return formatYearMonth(t.getFullYear(), t.getMonth() + 1);
  });
  const [dateField, setDateField] = useState(loadSavedDateField);
  const [dateFieldOptions, setDateFieldOptions] = useState<string[]>([...DEFAULT_DATE_FIELDS]);
  const [entries, setEntries] = useState<CalendarEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [truncated, setTruncated] = useState(false);
  const [openDay, setOpenDay] = useState<string | null>(null);

  const { year, month } = parseYearMonth(yearMonth);
  const byDate = useMemo(() => groupByDate(entries), [entries]);

  // Discover date fields from frontmatter samples
  useEffect(() => {
    api
      .meta({ limit: 100 })
      .then((resp) => {
        const samples = (resp.results ?? []).map((r) => r.frontmatter);
        const fields = detectDateFields(samples);
        setDateFieldOptions(fields);
        if (!fields.includes(dateField)) {
          setDateField(fields[0] ?? DEFAULT_DATE_FIELDS[0]!);
        }
      })
      .catch(() => {
        setDateFieldOptions([...DEFAULT_DATE_FIELDS]);
      });
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const loadMonth = useCallback(async () => {
    setLoading(true);
    setError(null);
    setTruncated(false);
    try {
      let dql: string;
      if (isMobile) {
        const today = new Date();
        const anchorDay = Math.min(today.getDate(), new Date(year, month, 0).getDate());
        const week = weekDateKeys(new Date(year, month - 1, anchorDay));
        dql = buildCalendarQueryRange(dateField, week[0]!, dayAfter(week[6]!));
      } else {
        dql = buildCalendarQuery(dateField, yearMonth);
      }
      const data = await api.query(dql, { limit: 200 });
      setEntries(parseCalendarResponse(data, dateField));
      setTruncated(data.has_more);
    } catch (e) {
      setEntries([]);
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }, [dateField, yearMonth, isMobile, year, month]);

  useEffect(() => {
    loadMonth();
  }, [loadMonth]);

  const handleDateFieldChange = (field: string) => {
    setDateField(field);
    try {
      localStorage.setItem(DATE_FIELD_STORAGE_KEY, field);
    } catch {
      // ignore
    }
  };

  const goToday = () => {
    const t = new Date();
    setYearMonth(formatYearMonth(t.getFullYear(), t.getMonth() + 1));
  };

  const monthPickerOptions = useMemo(() => {
    const options: { value: string; label: string }[] = [];
    const baseYear = year - 1;
    for (let y = baseYear; y <= baseYear + 3; y++) {
      for (let m = 1; m <= 12; m++) {
        const ym = formatYearMonth(y, m);
        options.push({ value: ym, label: `${MONTH_NAMES[m - 1]} ${y}` });
      }
    }
    return options;
  }, [year]);

  const weekKeys = useMemo(() => {
    const today = new Date();
    const anchorDay = Math.min(today.getDate(), new Date(year, month, 0).getDate());
    return weekDateKeys(new Date(year, month - 1, anchorDay));
  }, [year, month]);

  const renderDayPopover = (dateStr: string, hits: CalendarEntry[], dayNum: number) => {
    const visible = hits.slice(0, 3);
    const overflow = hits.length - visible.length;

    return (
      <Popover
        open={openDay === dateStr}
        onOpenChange={(open) => setOpenDay(open ? dateStr : null)}
      >
        <PopoverTrigger asChild>
          <button
            type="button"
            className={cn(
              "relative flex min-h-[4.5rem] w-full flex-col items-start rounded-md border border-transparent p-1.5 text-left text-xs transition-colors",
              "hover:border-border hover:bg-accent/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
              dateStr === todayISO() && "ring-1 ring-primary",
              hits.length > 0 && "bg-accent/20",
            )}
          >
            <span className="mb-1 font-medium">{dayNum}</span>
            <div className="flex flex-wrap gap-0.5">
              {visible.map((entry) => (
                <span
                  key={entry.path}
                  className="inline-block h-1.5 w-1.5 rounded-full"
                  style={{ backgroundColor: entryDotColor(entry) }}
                  title={pageTitleFromPath(entry.path, entry.title)}
                />
              ))}
              {overflow > 0 && (
                <Badge variant="secondary" className="h-4 px-1 text-[9px] leading-none">
                  +{overflow}
                </Badge>
              )}
            </div>
          </button>
        </PopoverTrigger>
        <PopoverContent className="w-80 p-0" align="start">
          <div className="border-b border-border px-3 py-2 text-sm font-semibold">
            {dateStr}
          </div>
          <div className="max-h-64 overflow-y-auto p-2 space-y-1.5">
            {hits.map((entry) => (
              <Card
                key={entry.path}
                className="cursor-pointer border-border/60 shadow-none hover:bg-accent/30"
                onClick={() => {
                  setOpenDay(null);
                  onNavigate(entry.path);
                }}
              >
                <CardContent className="p-2.5">
                  <div className="flex items-start gap-2">
                    <span
                      className="mt-1.5 inline-block h-2 w-2 shrink-0 rounded-full"
                      style={{ backgroundColor: entryDotColor(entry) }}
                    />
                    <div className="min-w-0 flex-1">
                      <div className="truncate text-sm font-medium">
                        {pageTitleFromPath(entry.path, entry.title)}
                      </div>
                      <div className="truncate text-xs text-muted-foreground">{entry.path}</div>
                      {(entry.state || entry.tags?.length) && (
                        <div className="mt-1 flex flex-wrap gap-1">
                          {entry.state && (
                            <Badge variant="outline" className="text-[10px]">
                              {entry.state}
                            </Badge>
                          )}
                          {entry.tags?.slice(0, 3).map((tag) => (
                            <Badge key={tag} variant="secondary" className="text-[10px]">
                              {tag}
                            </Badge>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </PopoverContent>
      </Popover>
    );
  };

  const renderMonthGrid = () => {
    const cells = buildMonthGrid(yearMonth);
    return (
      <div className="grid grid-cols-7 gap-1">
        {WEEKDAYS.map((d) => (
          <div key={d} className="py-1 text-center text-xs font-medium text-muted-foreground">
            {isMobile ? d.slice(0, 1) : d}
          </div>
        ))}
        {cells.map((day, i) => {
          if (day == null) {
            return <div key={`empty-${i}`} className="min-h-[4.5rem]" />;
          }
          const dateStr = `${yearMonth}-${String(day).padStart(2, "0")}`;
          const hits = byDate.get(dateStr) ?? [];
          return (
            <div key={dateStr}>{renderDayPopover(dateStr, hits, day)}</div>
          );
        })}
      </div>
    );
  };

  const renderWeekView = () => (
    <div className="space-y-2">
      {weekKeys.map((dateStr) => {
        const hits = byDate.get(dateStr) ?? [];
        const weekday = WEEKDAYS[new Date(`${dateStr}T12:00:00`).getDay() === 0 ? 6 : new Date(`${dateStr}T12:00:00`).getDay() - 1];
        return (
          <div
            key={dateStr}
            className={cn(
              "rounded-lg border border-border p-2",
              dateStr === todayISO() && "ring-1 ring-primary",
            )}
          >
            <div className="mb-1 flex items-center justify-between text-xs">
              <span className="font-semibold">{weekday} {dateStr.slice(5).replace("-", "/")}</span>
              {hits.length > 0 && (
                <span className="text-muted-foreground">{hits.length} page{hits.length !== 1 ? "s" : ""}</span>
              )}
            </div>
            {hits.length === 0 ? (
              <div className="text-xs text-muted-foreground py-1">No pages</div>
            ) : (
              <div className="space-y-1">
                {hits.map((entry) => (
                  <button
                    key={entry.path}
                    type="button"
                    className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm hover:bg-accent/40"
                    onClick={() => onNavigate(entry.path)}
                  >
                    <span
                      className="inline-block h-2 w-2 shrink-0 rounded-full"
                      style={{ backgroundColor: entryDotColor(entry) }}
                    />
                    <span className="truncate">{pageTitleFromPath(entry.path, entry.title)}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );

  return (
    <div className="flex h-full flex-col">
      <div className="flex flex-wrap items-center gap-2 border-b border-border bg-card px-3 py-3 sm:px-6">
        <Button variant="outline" size="sm" onClick={onClose}>
          <ArrowLeft className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Back</span>
        </Button>
        <div className="flex items-center gap-1.5 text-sm font-semibold">
          <CalendarDays className="h-4 w-4 text-muted-foreground" />
          Calendar
        </div>
        <div className="text-xs text-muted-foreground hidden sm:block">
          {entries.length} page{entries.length !== 1 ? "s" : ""} in {MONTH_NAMES[month - 1]} {year}
        </div>

        <div className="ml-auto flex flex-wrap items-center gap-2">
          <Select value={dateField} onValueChange={handleDateFieldChange}>
            <SelectTrigger className="h-8 w-36 text-xs">
              <SelectValue placeholder="Date field" />
            </SelectTrigger>
            <SelectContent>
              {dateFieldOptions.map((field) => (
                <SelectItem key={field} value={field}>
                  {titleize(field.replace(/_/g, " "))}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <div className="flex items-center gap-0.5">
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              onClick={() => setYearMonth((ym) => addMonths(ym, -1))}
              aria-label="Previous month"
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <Button variant="outline" size="sm" className="h-8 text-xs" onClick={goToday}>
              Today
            </Button>
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              onClick={() => setYearMonth((ym) => addMonths(ym, 1))}
              aria-label="Next month"
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>

          <Select value={yearMonth} onValueChange={setYearMonth}>
            <SelectTrigger className="h-8 w-40 text-xs hidden md:flex">
              <SelectValue />
            </SelectTrigger>
            <SelectContent className="max-h-64">
              {monthPickerOptions.map((opt) => (
                <SelectItem key={opt.value} value={opt.value}>
                  {opt.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-3 sm:p-6">
        {loading ? (
          <div className="flex items-center justify-center gap-2 py-16 text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            Loading calendar…
          </div>
        ) : error ? (
          <div className="rounded-md border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">
            {error}
          </div>
        ) : (
          <>
            {truncated && (
              <div className="mb-3 rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-xs text-amber-800 dark:text-amber-200">
                Showing first 200 pages for this range. Narrow the date field or month if entries are missing.
              </div>
            )}
            {isMobile ? renderWeekView() : renderMonthGrid()}
          </>
        )}
      </div>
    </div>
  );
}

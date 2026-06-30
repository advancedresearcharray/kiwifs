import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ArrowLeft,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Loader2,
} from "lucide-react";
import { format, startOfMonth } from "date-fns";
import { api } from "@kw/lib/api";
import { cn } from "@kw/lib/cn";
import {
  addMonths,
  buildFieldDiscoveryQuery,
  buildMobileWeekCells,
  buildMonthGrid,
  buildMonthQuery,
  buildWeekQuery,
  CALENDAR_DATE_FIELD_STORAGE_KEY,
  discoverDateFields,
  formatMonthLabel,
  groupPagesByDate,
  isTodayDate,
  pageDotClass,
  parseCalendarRows,
  type CalendarDayCell,
  type CalendarPage,
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
};

const WEEKDAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
const MAX_VISIBLE_DOTS = 3;

function loadStoredDateField(): string {
  try {
    return localStorage.getItem(CALENDAR_DATE_FIELD_STORAGE_KEY) ?? "date";
  } catch {
    return "date";
  }
}

function pageTitle(page: CalendarPage): string {
  return page.title?.trim() || titleize(page.path.replace(/\.md$/i, "").split("/").pop() ?? page.path);
}

type DayCellProps = {
  cell: CalendarDayCell;
  pages: CalendarPage[];
  onNavigate: (path: string) => void;
};

function CalendarDayCell({ cell, pages, onNavigate }: DayCellProps) {
  const [open, setOpen] = useState(false);
  const visibleDots = pages.slice(0, MAX_VISIBLE_DOTS);
  const overflow = pages.length - MAX_VISIBLE_DOTS;

  const handleActivate = () => {
    if (pages.length === 0) return;
    if (pages.length === 1) {
      onNavigate(pages[0]!.path);
      return;
    }
    setOpen(true);
  };

  const cellButton = (
    <button
      type="button"
      onClick={handleActivate}
      disabled={pages.length === 0}
      className={cn(
        "relative flex min-h-[4.5rem] sm:min-h-[5.5rem] flex-col rounded-md border border-border p-1.5 text-left transition-colors",
        cell.inMonth ? "bg-card" : "bg-muted/30 text-muted-foreground",
        pages.length > 0 && "hover:bg-accent/40 cursor-pointer",
        isTodayDate(cell.date) && "ring-2 ring-primary/60",
      )}
    >
      <span className="text-xs font-medium">{format(cell.date, "d")}</span>
      {pages.length > 0 && (
        <div className="mt-auto flex flex-wrap items-center gap-1 pt-1">
          {visibleDots.map((page) => (
            <span
              key={page.path}
              className={cn("h-2 w-2 rounded-full", pageDotClass(page))}
              title={pageTitle(page)}
            />
          ))}
          {overflow > 0 && (
            <Badge variant="secondary" className="h-4 px-1 text-[10px] leading-none">
              +{overflow}
            </Badge>
          )}
        </div>
      )}
    </button>
  );

  if (pages.length <= 1) {
    return cellButton;
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>{cellButton}</PopoverTrigger>
      <PopoverContent align="start" className="w-80 p-2">
        <div className="mb-2 text-xs font-medium text-muted-foreground">
          {format(cell.date, "EEEE, MMMM d, yyyy")}
        </div>
        <div className="flex max-h-64 flex-col gap-2 overflow-y-auto">
          {pages.map((page) => (
            <Card
              key={page.path}
              className="cursor-pointer transition-colors hover:bg-accent/30"
              onClick={() => {
                setOpen(false);
                onNavigate(page.path);
              }}
            >
              <CardContent className="flex items-start gap-2 p-3">
                <span className={cn("mt-1.5 h-2 w-2 shrink-0 rounded-full", pageDotClass(page))} />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm font-medium">{pageTitle(page)}</div>
                  <div className="truncate text-xs text-muted-foreground">{page.path}</div>
                  {(page.state || page.status || page.tags?.length) && (
                    <div className="mt-1 flex flex-wrap gap-1">
                      {(page.state ?? page.status) && (
                        <Badge variant="outline" className="text-[10px]">
                          {page.state ?? page.status}
                        </Badge>
                      )}
                      {page.tags?.slice(0, 3).map((tag) => (
                        <Badge key={tag} variant="secondary" className="text-[10px]">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  );
}

export function KiwiCalendar({ onClose, onNavigate }: Props) {
  const [viewDate, setViewDate] = useState(() => startOfMonth(new Date()));
  const [dateField, setDateField] = useState(loadStoredDateField);
  const [fieldOptions, setFieldOptions] = useState<string[]>(["date"]);
  const [pages, setPages] = useState<CalendarPage[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isMobile, setIsMobile] = useState(
    () => typeof window !== "undefined" && window.innerWidth < 768,
  );

  useEffect(() => {
    const mq = window.matchMedia("(max-width: 767px)");
    const onChange = (e: MediaQueryListEvent) => setIsMobile(e.matches);
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  const loadFields = useCallback(async () => {
    try {
      const result = await api.query(buildFieldDiscoveryQuery());
      const fields = discoverDateFields(result.rows ?? []);
      setFieldOptions(fields);
      if (!fields.includes(dateField)) {
        setDateField(fields[0] ?? "date");
      }
    } catch {
      setFieldOptions(["date"]);
    }
  }, [dateField]);

  useEffect(() => {
    loadFields();
  }, [loadFields]);

  const loadPages = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const dql = isMobile
        ? buildWeekQuery(dateField, viewDate)
        : buildMonthQuery(dateField, viewDate.getFullYear(), viewDate.getMonth());
      const result = await api.query(dql);
      setPages(parseCalendarRows(result.rows ?? [], dateField));
    } catch (err) {
      setPages([]);
      setError(err instanceof Error ? err.message : "Failed to load calendar");
    } finally {
      setLoading(false);
    }
  }, [dateField, isMobile, viewDate]);

  useEffect(() => {
    loadPages();
  }, [loadPages]);

  useEffect(() => {
    try {
      localStorage.setItem(CALENDAR_DATE_FIELD_STORAGE_KEY, dateField);
    } catch {
      // ignore storage errors
    }
  }, [dateField]);

  const grouped = useMemo(() => groupPagesByDate(pages), [pages]);
  const gridCells = useMemo(
    () =>
      isMobile
        ? buildMobileWeekCells(viewDate, viewDate)
        : buildMonthGrid(viewDate),
    [isMobile, viewDate],
  );

  const monthInputValue = format(viewDate, "yyyy-MM");

  return (
    <div className="flex h-full flex-col">
      <div className="flex flex-wrap items-center gap-2 border-b border-border bg-card px-3 py-3 sm:px-6">
        <Button variant="outline" size="sm" onClick={onClose}>
          <ArrowLeft className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Back</span>
        </Button>
        <div className="flex items-center gap-1.5 text-sm font-semibold">
          <CalendarDays className="h-4 w-4" />
          Calendar
        </div>

        <div className="ml-auto flex flex-wrap items-center gap-2">
          <Select value={dateField} onValueChange={setDateField}>
            <SelectTrigger className="h-8 w-36 text-sm">
              <SelectValue placeholder="Date field" />
            </SelectTrigger>
            <SelectContent>
              {fieldOptions.map((field) => (
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
              onClick={() => setViewDate((d) => addMonths(d, -1))}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <input
              type="month"
              value={monthInputValue}
              onChange={(e) => {
                const [year, month] = e.target.value.split("-").map(Number);
                if (year && month) {
                  setViewDate(startOfMonth(new Date(year, month - 1, 1)));
                }
              }}
              className="h-8 rounded-md border border-border bg-background px-2 text-sm"
              aria-label="Month picker"
            />
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              aria-label="Next month"
              onClick={() => setViewDate((d) => addMonths(d, 1))}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>

          <Button
            variant="outline"
            size="sm"
            onClick={() => setViewDate(startOfMonth(new Date()))}
          >
            Today
          </Button>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-3 sm:p-6">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold">{formatMonthLabel(viewDate)}</h2>
          {loading && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
        </div>

        {error && (
          <div className="mb-4 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {error}
          </div>
        )}

        <div className="grid grid-cols-7 gap-1 sm:gap-2">
          {WEEKDAY_LABELS.map((label) => (
            <div
              key={label}
              className="px-1 py-1 text-center text-xs font-medium text-muted-foreground"
            >
              {label}
            </div>
          ))}
          {gridCells.map((cell) => (
            <CalendarDayCell
              key={cell.iso}
              cell={cell}
              pages={grouped.get(cell.iso) ?? []}
              onNavigate={onNavigate}
            />
          ))}
        </div>
      </div>
    </div>
  );
}

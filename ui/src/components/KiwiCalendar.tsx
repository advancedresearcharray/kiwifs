// KiwiCalendar — monthly grid of pages keyed by frontmatter date fields.

import { useCallback, useEffect, useMemo, useState, type ReactNode } from "react";
import {
  ArrowLeft,
  ArrowLeftRight,
  ChevronLeft,
  ChevronRight,
  Loader2,
} from "lucide-react";
import { format, parseISO } from "date-fns";
import { api } from "@kw/lib/api";
import { cn } from "@kw/lib/cn";
import { titleize } from "@kw/lib/paths";
import {
  buildMonthGrid,
  buildMonthQuery,
  dateKeyFromCell,
  discoverDateFields,
  groupPagesByDate,
  MONTH_NAMES,
  pageDotColor,
  parseCalendarRows,
  weekDateKeys,
  WEEKDAY_LABELS,
  type CalendarPage,
  type CalendarPagesByDate,
} from "@kw/lib/calendarView";
import { Button } from "@kw/components/ui/button";
import { Badge } from "@kw/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@kw/components/ui/card";
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

const MAX_DOTS = 3;

function DayPopover({
  dateKey,
  pages,
  onNavigate,
  children,
}: {
  dateKey: string;
  pages: CalendarPage[];
  onNavigate: (path: string) => void;
  children: ReactNode;
}) {
  if (pages.length === 0) return <>{children}</>;

  return (
    <Popover>
      <PopoverTrigger asChild>{children}</PopoverTrigger>
      <PopoverContent className="w-80 p-0" align="start">
        <Card className="border-0 shadow-none">
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">
              {format(parseISO(dateKey), "EEEE, MMMM d, yyyy")}
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 max-h-64 overflow-auto kiwi-scroll">
            {pages.map((page) => (
              <button
                key={page.path}
                type="button"
                className="w-full text-left rounded-md border border-border p-2 hover:bg-accent/40 transition-colors"
                onClick={() => onNavigate(page.path)}
              >
                <div className="flex items-start gap-2">
                  <span
                    className="mt-1.5 h-2 w-2 shrink-0 rounded-full"
                    style={{ backgroundColor: pageDotColor(page) }}
                  />
                  <div className="min-w-0 flex-1">
                    <div className="font-medium text-sm truncate">
                      {page.title || titleize(page.path)}
                    </div>
                    <div className="text-xs text-muted-foreground truncate">{page.path}</div>
                    {(page.status || page.tags?.length) && (
                      <div className="flex flex-wrap gap-1 mt-1">
                        {page.status && (
                          <Badge variant="secondary" className="text-[10px] px-1.5 py-0">
                            {page.status}
                          </Badge>
                        )}
                        {page.tags?.slice(0, 3).map((tag) => (
                          <Badge key={tag} variant="outline" className="text-[10px] px-1.5 py-0">
                            {tag}
                          </Badge>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </button>
            ))}
          </CardContent>
        </Card>
      </PopoverContent>
    </Popover>
  );
}

function DayCell({
  dateKey,
  day,
  pages,
  isToday,
  onNavigate,
}: {
  dateKey: string;
  day: number;
  pages: CalendarPage[];
  isToday: boolean;
  onNavigate: (path: string) => void;
}) {
  const count = pages.length;
  const visible = pages.slice(0, MAX_DOTS);
  const overflow = count - visible.length;

  const cell = (
    <div
      className={cn(
        "relative min-h-[4.5rem] rounded-md border border-border/60 p-1.5 text-left transition-colors",
        count > 0 && "hover:bg-accent/30 cursor-pointer",
        isToday && "ring-2 ring-primary ring-offset-1 ring-offset-background",
      )}
    >
      <div className={cn("text-xs font-medium mb-1", isToday && "text-primary")}>{day}</div>
      {count > 0 && (
        <div className="flex flex-wrap items-center gap-0.5">
          {visible.map((page) => (
            <span
              key={page.path}
              className="h-1.5 w-1.5 rounded-full shrink-0"
              style={{ backgroundColor: pageDotColor(page) }}
              title={page.title || page.path}
            />
          ))}
          {overflow > 0 && (
            <span className="text-[9px] text-muted-foreground font-medium">+{overflow}</span>
          )}
        </div>
      )}
      {count > MAX_DOTS && (
        <span className="absolute bottom-1 right-1 text-[9px] bg-muted text-muted-foreground rounded px-1">
          {count}
        </span>
      )}
    </div>
  );

  if (count === 1) {
    return (
      <button type="button" className="w-full" onClick={() => onNavigate(pages[0]!.path)}>
        {cell}
      </button>
    );
  }

  return (
    <DayPopover dateKey={dateKey} pages={pages} onNavigate={onNavigate}>
      <button type="button" className="w-full">
        {cell}
      </button>
    </DayPopover>
  );
}

function WeekRow({
  dateKeys,
  byDate,
  onNavigate,
}: {
  dateKeys: string[];
  byDate: CalendarPagesByDate;
  onNavigate: (path: string) => void;
}) {
  const todayKey = new Date().toISOString().slice(0, 10);

  return (
    <div className="grid grid-cols-7 gap-1">
      {dateKeys.map((dateKey) => {
        const day = Number(dateKey.slice(8, 10));
        const pages = byDate.get(dateKey) ?? [];
        return (
          <DayCell
            key={dateKey}
            dateKey={dateKey}
            day={day}
            pages={pages}
            isToday={dateKey === todayKey}
            onNavigate={onNavigate}
          />
        );
      })}
    </div>
  );
}

export function KiwiCalendar({ onClose, onNavigate, isMobile = false }: Props) {
  const now = new Date();
  const [year, setYear] = useState(now.getFullYear());
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [dateField, setDateField] = useState("date");
  const [dateFields, setDateFields] = useState<string[]>(["date"]);
  const [byDate, setByDate] = useState<CalendarPagesByDate>(new Map());
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    discoverDateFields((dql) => api.query(dql))
      .then(setDateFields)
      .catch(() => setDateFields(["date"]));
  }, []);

  const loadMonth = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const dql = buildMonthQuery(dateField, year, month);
      const data = await api.query(dql, { limit: 500 });
      const pages = parseCalendarRows(data, dateField);
      setByDate(groupPagesByDate(pages));
    } catch (err) {
      setByDate(new Map());
      setError(err instanceof Error ? err.message : "Failed to load calendar data");
    } finally {
      setLoading(false);
    }
  }, [dateField, year, month]);

  useEffect(() => {
    loadMonth();
  }, [loadMonth]);

  const goToday = () => {
    const t = new Date();
    setYear(t.getFullYear());
    setMonth(t.getMonth() + 1);
  };

  const shiftMonth = (delta: number) => {
    let m = month + delta;
    let y = year;
    while (m < 1) {
      m += 12;
      y -= 1;
    }
    while (m > 12) {
      m -= 12;
      y += 1;
    }
    setYear(y);
    setMonth(m);
  };

  const monthCells = useMemo(() => buildMonthGrid(year, month), [year, month]);
  const todayKey = new Date().toISOString().slice(0, 10);
  const weekKeys = useMemo(() => {
    const today = new Date();
    const anchor =
      year === today.getFullYear() && month === today.getMonth() + 1
        ? today
        : new Date(year, month - 1, 1);
    return weekDateKeys(anchor);
  }, [year, month]);

  const monthPickerValue = `${year}-${String(month).padStart(2, "0")}`;

  return (
    <div className="h-full flex flex-col">
      <div className="flex flex-wrap items-center gap-2 px-3 sm:px-6 py-3 border-b border-border bg-card">
        <Button variant="outline" size="sm" onClick={onClose}>
          <ArrowLeft className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Back</span>
        </Button>
        <div className="font-semibold text-sm flex items-center gap-1.5">
          <ArrowLeftRight className="h-3.5 w-3.5 text-muted-foreground" />
          Calendar
        </div>

        <div className="ml-auto flex items-center gap-2 flex-wrap">
          <Select value={dateField} onValueChange={setDateField}>
            <SelectTrigger className="h-8 w-36 text-sm">
              <SelectValue placeholder="Date field" />
            </SelectTrigger>
            <SelectContent>
              {dateFields.map((f) => (
                <SelectItem key={f} value={f}>
                  {f}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <div className="flex items-center gap-1">
            <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => shiftMonth(-1)}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <input
              type="month"
              value={monthPickerValue}
              onChange={(e) => {
                const [y, m] = e.target.value.split("-").map(Number);
                if (y && m) {
                  setYear(y);
                  setMonth(m);
                }
              }}
              className="h-8 rounded-md border border-border bg-background px-2 text-sm"
            />
            <Button variant="outline" size="icon" className="h-8 w-8" onClick={() => shiftMonth(1)}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>

          <Button variant="outline" size="sm" onClick={goToday}>
            Today
          </Button>
        </div>
      </div>

      <div className="flex-1 overflow-auto kiwi-scroll p-4">
        <div className="max-w-5xl mx-auto">
          <h2 className="text-lg font-semibold mb-4">
            {MONTH_NAMES[month - 1]} {year}
          </h2>

          {loading ? (
            <div className="flex items-center justify-center h-48 text-muted-foreground">
              <Loader2 className="h-5 w-5 animate-spin mr-2" /> Loading...
            </div>
          ) : error ? (
            <div className="text-sm text-destructive text-center py-12">{error}</div>
          ) : isMobile ? (
            <div>
              <div className="grid grid-cols-7 gap-1 mb-1">
                {WEEKDAY_LABELS.map((d) => (
                  <div key={d} className="text-center text-xs text-muted-foreground font-medium py-1">
                    {d.slice(0, 1)}
                  </div>
                ))}
              </div>
              <WeekRow dateKeys={weekKeys} byDate={byDate} onNavigate={onNavigate} />
            </div>
          ) : (
            <div>
              <div className="grid grid-cols-7 gap-1 mb-1">
                {WEEKDAY_LABELS.map((d) => (
                  <div key={d} className="text-center text-xs text-muted-foreground font-medium py-1">
                    {d}
                  </div>
                ))}
              </div>
              <div className="grid grid-cols-7 gap-1">
                {monthCells.map((day, i) => {
                  if (day == null) {
                    return <div key={`empty-${i}`} className="min-h-[4.5rem]" />;
                  }
                  const dateKey = dateKeyFromCell(year, month, day);
                  const pages = byDate.get(dateKey) ?? [];
                  return (
                    <DayCell
                      key={dateKey}
                      dateKey={dateKey}
                      day={day}
                      pages={pages}
                      isToday={dateKey === todayKey}
                      onNavigate={onNavigate}
                    />
                  );
                })}
              </div>
            </div>
          )}

          {!loading && !error && byDate.size === 0 && (
            <div className="text-center text-sm text-muted-foreground py-8">
              No pages with <code className="text-xs">{dateField}</code> in{" "}
              {MONTH_NAMES[month - 1]} {year}.
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

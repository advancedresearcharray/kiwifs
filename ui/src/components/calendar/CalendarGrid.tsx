import { useMemo } from "react";
import { format } from "date-fns";
import { cn } from "@kw/lib/cn";
import {
  MAX_VISIBLE_DOTS,
  WEEKDAY_LABELS,
  buildMonthGrid,
  buildWeekGrid,
  pageDotColor,
  todayKey,
  type CalendarPageEntry,
} from "@kw/lib/calendarView";
import { Badge } from "@kw/components/ui/badge";

type Props = {
  month: Date;
  byDate: Map<string, CalendarPageEntry[]>;
  selectedDateKey: string | null;
  weekOnly?: boolean;
  onSelectDate: (dateKey: string) => void;
};

export function CalendarGrid({
  month,
  byDate,
  selectedDateKey,
  weekOnly = false,
  onSelectDate,
}: Props) {
  const cells = useMemo(
    () => (weekOnly ? buildWeekGrid(month) : buildMonthGrid(month)),
    [month, weekOnly],
  );
  const today = todayKey();

  return (
    <div className="w-full max-w-3xl mx-auto">
      <div className="grid grid-cols-7 gap-1 mb-1">
        {WEEKDAY_LABELS.map((label) => (
          <div
            key={label}
            className="text-center text-xs font-medium text-muted-foreground py-1"
          >
            {label}
          </div>
        ))}
      </div>
      <div className="grid grid-cols-7 gap-1">
        {cells.map((cell) => {
          const entries = byDate.get(cell.dateKey) ?? [];
          const count = entries.length;
          const isToday = cell.dateKey === today;
          const isSelected = cell.dateKey === selectedDateKey;

          return (
            <button
              key={cell.dateKey}
              type="button"
              onClick={() => onSelectDate(cell.dateKey)}
              className={cn(
                "relative min-h-[4.5rem] rounded-md border border-transparent p-1.5 text-left transition-colors",
                "hover:border-border hover:bg-accent/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                !cell.inMonth && !weekOnly && "opacity-40",
                isSelected && "border-primary bg-primary/10",
                isToday && "ring-1 ring-primary/60",
              )}
              aria-label={`${format(cell.date, "MMMM d, yyyy")}${count ? `, ${count} pages` : ""}`}
            >
              <div className="flex items-start justify-between gap-1">
                <span
                  className={cn(
                    "text-xs font-medium",
                    isToday && "text-primary",
                  )}
                >
                  {cell.date.getDate()}
                </span>
                {count > MAX_VISIBLE_DOTS && (
                  <Badge variant="secondary" className="h-4 px-1 text-[10px]">
                    {count}
                  </Badge>
                )}
              </div>
              {count > 0 && (
                <div className="mt-1 flex flex-wrap gap-0.5">
                  {entries.slice(0, MAX_VISIBLE_DOTS).map((entry) => (
                    <span
                      key={entry.path}
                      className="inline-block h-1.5 w-1.5 rounded-full"
                      style={{ backgroundColor: pageDotColor(entry) }}
                      title={entry.title}
                    />
                  ))}
                </div>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}

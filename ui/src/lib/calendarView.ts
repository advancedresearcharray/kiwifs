import {
  addMonths,
  endOfMonth,
  endOfWeek,
  format,
  parseISO,
  startOfMonth,
  startOfWeek,
  subMonths,
} from "date-fns";
import type { MetaResult, QueryResponse } from "@kw/lib/api";
import { tagColor } from "@kw/lib/kanbanUi";
import { titleize } from "@kw/lib/paths";

export type CalendarPageEntry = {
  path: string;
  date: string;
  title: string;
  tags: string[];
  state?: string;
};

export const DEFAULT_DATE_FIELDS = [
  "date",
  "due",
  "created",
  "last_executed",
  "published_at",
  "updated",
] as const;

const STATE_DOT_COLORS: Record<string, string> = {
  accepted: "#16A34A",
  approved: "#16A34A",
  done: "#16A34A",
  completed: "#16A34A",
  proposed: "#CA8A04",
  draft: "#9CA3AF",
  rejected: "#DC2626",
  blocked: "#DC2626",
  review: "#7C3AED",
  "in-progress": "#2563EB",
  todo: "#6B7280",
};

const ISO_DATE = /^\d{4}-\d{2}-\d{2}$/;

export function normalizeDateValue(raw: unknown): string | null {
  if (raw == null) return null;
  const text = String(raw).trim();
  if (!text) return null;
  const datePart = text.slice(0, 10);
  return ISO_DATE.test(datePart) ? datePart : null;
}

export function isLikelyDateField(field: string, value: unknown): boolean {
  if (normalizeDateValue(value)) return true;
  const lower = field.toLowerCase();
  return (
    lower.includes("date") ||
    lower.endsWith("_at") ||
    lower === "due" ||
    lower === "created" ||
    lower === "updated"
  );
}

export function discoverDateFields(results: MetaResult[]): string[] {
  const counts = new Map<string, number>();

  for (const field of DEFAULT_DATE_FIELDS) {
    counts.set(field, 0);
  }

  for (const result of results) {
    for (const [field, value] of Object.entries(result.frontmatter ?? {})) {
      if (!isLikelyDateField(field, value)) continue;
      counts.set(field, (counts.get(field) ?? 0) + 1);
    }
  }

  const discovered = [...counts.entries()]
    .filter(([, count]) => count > 0)
    .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
    .map(([field]) => field);

  if (discovered.length > 0) return discovered;
  return [...DEFAULT_DATE_FIELDS];
}

export function buildMonthRangeQuery(dateField: string, month: Date): string {
  const start = format(startOfMonth(month), "yyyy-MM-dd");
  const end = format(addMonths(startOfMonth(month), 1), "yyyy-MM-dd");
  return `TABLE _path, ${quoteField(dateField)}, tags, state WHERE ${quoteField(dateField)} >= DATE("${start}") AND ${quoteField(dateField)} < DATE("${end}")`;
}

function quoteField(field: string): string {
  return /^[a-zA-Z_][a-zA-Z0-9_-]*$/.test(field) ? field : `\`${field}\``;
}

function pageTitle(path: string, row: Record<string, unknown>): string {
  const explicit = row.title ?? row.name;
  if (typeof explicit === "string" && explicit.trim()) return explicit.trim();
  const base = path.split("/").pop() ?? path;
  return titleize(base.replace(/\.md$/i, ""));
}

function parseTags(raw: unknown): string[] {
  if (Array.isArray(raw)) {
    return raw.map(String).filter(Boolean);
  }
  if (typeof raw === "string" && raw.trim()) {
    return [raw.trim()];
  }
  return [];
}

export function parseCalendarQueryRows(data: QueryResponse): CalendarPageEntry[] {
  const cols = data.columns ?? [];
  const dateField =
    cols.find((c) => c !== "_path" && c !== "path" && c !== "tags" && c !== "state") ??
    "date";

  const entries: CalendarPageEntry[] = [];
  for (const row of data.rows ?? []) {
    const path = String(row._path ?? row.path ?? "");
    const date = normalizeDateValue(row[dateField]);
    if (!path || !date) continue;
    entries.push({
      path,
      date,
      title: pageTitle(path, row),
      tags: parseTags(row.tags),
      state: typeof row.state === "string" ? row.state : undefined,
    });
  }
  return entries;
}

export function groupEntriesByDate(
  entries: CalendarPageEntry[],
): Map<string, CalendarPageEntry[]> {
  const byDate = new Map<string, CalendarPageEntry[]>();
  for (const entry of entries) {
    const list = byDate.get(entry.date);
    if (list) list.push(entry);
    else byDate.set(entry.date, [entry]);
  }
  return byDate;
}

export function pageDotColor(entry: CalendarPageEntry): string {
  if (entry.state) {
    const key = entry.state.toLowerCase().trim();
    if (STATE_DOT_COLORS[key]) return STATE_DOT_COLORS[key];
  }
  if (entry.tags.length > 0) {
    return tagColor(entry.tags[0]).fg;
  }
  return "hsl(var(--primary))";
}

export type CalendarCell = {
  date: Date;
  dateKey: string;
  inMonth: boolean;
};

export function buildMonthGrid(month: Date): CalendarCell[] {
  const monthStart = startOfMonth(month);
  const monthEnd = endOfMonth(month);
  const gridStart = startOfWeek(monthStart, { weekStartsOn: 1 });
  const gridEnd = endOfWeek(monthEnd, { weekStartsOn: 1 });

  const cells: CalendarCell[] = [];
  let cursor = gridStart;
  while (cursor <= gridEnd) {
    cells.push({
      date: cursor,
      dateKey: format(cursor, "yyyy-MM-dd"),
      inMonth: cursor.getMonth() === month.getMonth(),
    });
    cursor = new Date(cursor.getFullYear(), cursor.getMonth(), cursor.getDate() + 1);
  }
  return cells;
}

/** Anchor date for mobile week-only grid: selected day in month, else today, else month start. */
export function weekGridAnchor(selectedDateKey: string | null, month: Date): Date {
  if (selectedDateKey) {
    try {
      const parsed = parseISO(selectedDateKey);
      if (
        !Number.isNaN(parsed.getTime()) &&
        parsed.getFullYear() === month.getFullYear() &&
        parsed.getMonth() === month.getMonth()
      ) {
        return parsed;
      }
    } catch {
      /* fall through */
    }
  }

  const today = new Date();
  if (
    today.getFullYear() === month.getFullYear() &&
    today.getMonth() === month.getMonth()
  ) {
    return today;
  }

  return startOfMonth(month);
}

export function defaultSelectedDateKey(month: Date): string {
  const today = todayKey();
  const parsed = parseISO(today);
  if (
    !Number.isNaN(parsed.getTime()) &&
    parsed.getFullYear() === month.getFullYear() &&
    parsed.getMonth() === month.getMonth()
  ) {
    return today;
  }
  return format(startOfMonth(month), "yyyy-MM-dd");
}

export function buildWeekGrid(anchor: Date): CalendarCell[] {
  const weekStart = startOfWeek(anchor, { weekStartsOn: 1 });
  const cells: CalendarCell[] = [];
  for (let i = 0; i < 7; i++) {
    const date = new Date(
      weekStart.getFullYear(),
      weekStart.getMonth(),
      weekStart.getDate() + i,
    );
    cells.push({
      date,
      dateKey: format(date, "yyyy-MM-dd"),
      inMonth: true,
    });
  }
  return cells;
}

export function monthLabel(month: Date): string {
  return format(month, "MMMM yyyy");
}

export function todayKey(): string {
  return format(new Date(), "yyyy-MM-dd");
}

export function shiftMonth(month: Date, delta: number): Date {
  return delta >= 0 ? addMonths(month, delta) : subMonths(month, Math.abs(delta));
}

export function parseMonthInput(value: string): Date | null {
  if (!/^\d{4}-\d{2}$/.test(value)) return null;
  try {
    const parsed = parseISO(`${value}-01`);
    return Number.isNaN(parsed.getTime()) ? null : startOfMonth(parsed);
  } catch {
    return null;
  }
}

export const WEEKDAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"] as const;

export const MAX_VISIBLE_DOTS = 3;

export function isCalendarViewPath(pathname: string): boolean {
  return pathname === "/view/calendar" || pathname.endsWith("/view/calendar");
}

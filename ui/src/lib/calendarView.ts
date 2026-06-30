import type { QueryResponse } from "@kw/lib/api";
import { tagColor } from "@kw/lib/kanbanUi";

export type CalendarEntry = {
  path: string;
  date: string;
  title?: string;
  tags?: string[];
  state?: string;
};

export const DEFAULT_DATE_FIELDS = ["date", "due", "created", "last_executed"];

export const WEEKDAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];

export const MONTH_NAMES = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
];

const STATE_DOT_COLORS: Record<string, string> = {
  accepted: "#22c55e",
  proposed: "#eab308",
  deprecated: "#9ca3af",
  rejected: "#ef4444",
  superseded: "#6b7280",
  draft: "#3b82f6",
  review: "#a855f7",
  published: "#14b8a6",
};

const DATE_LIKE_RE = /^\d{4}-\d{2}-\d{2}(?:[ T]\d{2}:\d{2})?/;

function localDateISO(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

export function todayISO(): string {
  return localDateISO(new Date());
}

export function formatYearMonth(year: number, month: number): string {
  return `${year}-${String(month).padStart(2, "0")}`;
}

export function parseYearMonth(yearMonth: string): { year: number; month: number } {
  const [year, month] = yearMonth.split("-").map(Number);
  return { year, month };
}

export function addMonths(yearMonth: string, delta: number): string {
  const { year, month } = parseYearMonth(yearMonth);
  const d = new Date(year, month - 1 + delta, 1);
  return formatYearMonth(d.getFullYear(), d.getMonth() + 1);
}

export function monthStartEnd(yearMonth: string): { start: string; end: string } {
  const { year, month } = parseYearMonth(yearMonth);
  const start = `${yearMonth}-01`;
  const next =
    month === 12
      ? formatYearMonth(year + 1, 1)
      : formatYearMonth(year, month + 1);
  return { start, end: `${next}-01` };
}

export function buildCalendarQuery(dateField: string, yearMonth: string): string {
  const { start, end } = monthStartEnd(yearMonth);
  return buildCalendarQueryRange(dateField, start, end);
}

export function buildCalendarQueryRange(dateField: string, start: string, end: string): string {
  return `TABLE _path, ${dateField}, tags, state, title WHERE striptime(${dateField}) >= DATE("${start}") AND striptime(${dateField}) < DATE("${end}")`;
}

/** True when DQL is a KiwiCalendar month/range TABLE query (not legacy CALENDAR). */
export function isCalendarTableQuery(dql: string): boolean {
  return /\bTABLE\b/i.test(dql) && /\bstriptime\s*\(/i.test(dql) && /\bDATE\s*\(/i.test(dql);
}

export function dayAfter(dateStr: string): string {
  const d = new Date(`${dateStr}T12:00:00`);
  d.setDate(d.getDate() + 1);
  return localDateISO(d);
}

export function isDateLikeValue(value: unknown): boolean {
  if (value instanceof Date) return !isNaN(value.getTime());
  if (typeof value === "string") return DATE_LIKE_RE.test(value);
  return false;
}

export function toDateKey(value: unknown): string | null {
  if (value == null) return null;
  if (value instanceof Date) {
    if (isNaN(value.getTime())) return null;
    return localDateISO(value);
  }
  const raw = String(value);
  const match = raw.match(/^(\d{4}-\d{2}-\d{2})/);
  return match ? match[1]! : null;
}

function normalizeTags(raw: unknown): string[] | undefined {
  if (raw == null) return undefined;
  if (Array.isArray(raw)) {
    const tags = raw.map(String).filter(Boolean);
    return tags.length > 0 ? tags : undefined;
  }
  if (typeof raw === "string" && raw.trim()) return [raw.trim()];
  return undefined;
}

export function parseCalendarResponse(
  data: QueryResponse,
  dateField: string,
): CalendarEntry[] {
  const entries: CalendarEntry[] = [];
  for (const row of data.rows) {
    const date = toDateKey(row[dateField]);
    if (!date) continue;
    const path = String(row["_path"] ?? row["path"] ?? "");
    if (!path) continue;
    entries.push({
      path,
      date,
      title: row.title != null ? String(row.title) : undefined,
      tags: normalizeTags(row.tags),
      state: row.state != null ? String(row.state) : undefined,
    });
  }
  return entries;
}

export function groupByDate(entries: CalendarEntry[]): Map<string, CalendarEntry[]> {
  const map = new Map<string, CalendarEntry[]>();
  for (const entry of entries) {
    const list = map.get(entry.date);
    if (list) list.push(entry);
    else map.set(entry.date, [entry]);
  }
  return map;
}

export function detectDateFields(samples: Record<string, unknown>[]): string[] {
  const counts = new Map<string, number>();
  for (const fm of samples) {
    for (const [key, value] of Object.entries(fm)) {
      if (key.startsWith("_")) continue;
      if (isDateLikeValue(value)) {
        counts.set(key, (counts.get(key) ?? 0) + 1);
      }
    }
  }
  const discovered = [...counts.entries()]
    .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]))
    .map(([key]) => key);

  const merged: string[] = [];
  for (const field of DEFAULT_DATE_FIELDS) {
    if (discovered.includes(field) || counts.has(field)) merged.push(field);
  }
  for (const field of discovered) {
    if (!merged.includes(field)) merged.push(field);
  }
  return merged.length > 0 ? merged : [...DEFAULT_DATE_FIELDS];
}

export function entryDotColor(entry: CalendarEntry): string {
  if (entry.state) {
    const key = entry.state.toLowerCase().trim();
    if (STATE_DOT_COLORS[key]) return STATE_DOT_COLORS[key]!;
  }
  if (entry.tags && entry.tags.length > 0) {
    return tagColor(entry.tags[0]!).bg;
  }
  return "hsl(var(--primary))";
}

export function buildMonthGrid(yearMonth: string): (number | null)[] {
  const { year, month } = parseYearMonth(yearMonth);
  const firstDay = new Date(year, month - 1, 1);
  const daysInMonth = new Date(year, month, 0).getDate();
  let startDow = firstDay.getDay() - 1;
  if (startDow < 0) startDow = 6;

  const cells: (number | null)[] = [];
  for (let i = 0; i < startDow; i++) cells.push(null);
  for (let d = 1; d <= daysInMonth; d++) cells.push(d);
  while (cells.length % 7 !== 0) cells.push(null);
  return cells;
}

/** Returns ISO date keys (YYYY-MM-DD) for the week containing `anchor`. Monday-start. */
export function weekDateKeys(anchor: Date): string[] {
  const d = new Date(anchor);
  const dow = d.getDay();
  const mondayOffset = dow === 0 ? -6 : 1 - dow;
  d.setDate(d.getDate() + mondayOffset);

  const keys: string[] = [];
  for (let i = 0; i < 7; i++) {
    keys.push(localDateISO(d));
    d.setDate(d.getDate() + 1);
  }
  return keys;
}

export function pageTitleFromPath(path: string, title?: string): string {
  if (title?.trim()) return title.trim();
  const base = path.split("/").pop() ?? path;
  return base.replace(/\.md$/i, "").replace(/[-_]/g, " ");
}

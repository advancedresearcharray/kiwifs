import {
  addDays,
  addMonths,
  endOfMonth,
  endOfWeek,
  format,
  isSameDay,
  isSameMonth,
  parseISO,
  startOfMonth,
  startOfWeek,
} from "date-fns";

export const CALENDAR_DATE_FIELD_STORAGE_KEY = "kiwifs-calendar-date-field";

/** Common frontmatter fields that may hold ISO dates. */
export const DATE_FIELD_CANDIDATES = [
  "date",
  "due",
  "due_date",
  "created",
  "last_executed",
  "reviewed",
  "next-review",
  "published_at",
  "scheduled",
  "deadline",
] as const;

const DATE_FIELD_PATTERN = /^[a-zA-Z_][a-zA-Z0-9_.-]*$/;

export type CalendarPage = {
  path: string;
  title?: string;
  tags?: string[];
  state?: string;
  status?: string;
  dateKey: string;
};

export type CalendarDayCell = {
  date: Date;
  inMonth: boolean;
  iso: string;
};

/** Sanitize a user-selected date field before DQL interpolation. */
export function sanitizeDateField(field: string): string | null {
  const trimmed = field.trim();
  if (!DATE_FIELD_PATTERN.test(trimmed)) return null;
  return trimmed;
}

function monthBounds(year: number, monthIndex: number): { start: string; end: string } {
  const start = new Date(year, monthIndex, 1);
  const next = addMonths(start, 1);
  return {
    start: format(start, "yyyy-MM-dd"),
    end: format(next, "yyyy-MM-dd"),
  };
}

/** DQL for pages within a calendar month (exclusive upper bound). */
export function buildMonthQuery(dateField: string, year: number, monthIndex: number): string {
  const field = sanitizeDateField(dateField) ?? "date";
  const { start, end } = monthBounds(year, monthIndex);
  return [
    `TABLE _path, title, tags, state, status, ${field}`,
    `WHERE striptime(${field}) >= DATE("${start}") AND striptime(${field}) < DATE("${end}")`,
  ].join("\n");
}

/** DQL for a Monday–Sunday week containing anchorDate. */
export function buildWeekQuery(dateField: string, anchorDate: Date): string {
  const field = sanitizeDateField(dateField) ?? "date";
  const weekStart = startOfWeek(anchorDate, { weekStartsOn: 1 });
  const weekEnd = addDays(weekStart, 7);
  return [
    `TABLE _path, title, tags, state, status, ${field}`,
    `WHERE striptime(${field}) >= DATE("${format(weekStart, "yyyy-MM-dd")}")`,
    `AND striptime(${field}) < DATE("${format(weekEnd, "yyyy-MM-dd")}")`,
  ].join("\n");
}

/** Discover pages that populate any known date column. */
export function buildFieldDiscoveryQuery(): string {
  const columns = DATE_FIELD_CANDIDATES.join(", ");
  const clauses = DATE_FIELD_CANDIDATES.map((f) => `${f} != null`).join(" OR ");
  return `TABLE _path, ${columns}\nWHERE ${clauses}\nLIMIT 200`;
}

/** Parse an ISO date (YYYY-MM-DD or datetime) to a calendar day key. */
export function parseCalendarDateValue(raw: unknown): string | null {
  if (raw == null) return null;
  const text = String(raw).trim();
  if (!text) return null;
  const dayMatch = text.match(/^(\d{4}-\d{2}-\d{2})/);
  if (dayMatch) return dayMatch[1]!;
  try {
    const parsed = parseISO(text);
    if (Number.isNaN(parsed.getTime())) return null;
    return format(parsed, "yyyy-MM-dd");
  } catch {
    return null;
  }
}

export function parseCalendarRows(
  rows: Record<string, unknown>[],
  dateField: string,
): CalendarPage[] {
  const field = sanitizeDateField(dateField) ?? "date";
  const pages: CalendarPage[] = [];
  for (const row of rows) {
    const path = String(row._path ?? row.path ?? "").trim();
    if (!path) continue;
    const dateKey = parseCalendarDateValue(row[field]);
    if (!dateKey) continue;
    pages.push({
      path,
      title: row.title != null ? String(row.title) : undefined,
      tags: normalizeTags(row.tags),
      state: row.state != null ? String(row.state) : undefined,
      status: row.status != null ? String(row.status) : undefined,
      dateKey,
    });
  }
  return pages;
}

function normalizeTags(raw: unknown): string[] | undefined {
  if (raw == null) return undefined;
  if (Array.isArray(raw)) {
    const tags = raw.map((t) => String(t).trim()).filter(Boolean);
    return tags.length > 0 ? tags : undefined;
  }
  const single = String(raw).trim();
  return single ? [single] : undefined;
}

/** Group parsed pages by YYYY-MM-DD. */
export function groupPagesByDate(pages: CalendarPage[]): Map<string, CalendarPage[]> {
  const map = new Map<string, CalendarPage[]>();
  for (const page of pages) {
    const list = map.get(page.dateKey);
    if (list) list.push(page);
    else map.set(page.dateKey, [page]);
  }
  return map;
}

/** Infer date fields present in discovery query rows, preserving candidate order. */
export function discoverDateFields(rows: Record<string, unknown>[]): string[] {
  const found = new Set<string>();
  for (const row of rows) {
    for (const field of DATE_FIELD_CANDIDATES) {
      if (parseCalendarDateValue(row[field]) != null) {
        found.add(field);
      }
    }
  }
  const ordered = DATE_FIELD_CANDIDATES.filter((f) => found.has(f));
  return ordered.length > 0 ? ordered : ["date"];
}

const WORKFLOW_DOT_CLASSES: Record<string, string> = {
  accepted: "bg-emerald-500",
  approved: "bg-emerald-500",
  done: "bg-emerald-500",
  completed: "bg-emerald-500",
  published: "bg-emerald-500",
  proposed: "bg-amber-400",
  draft: "bg-amber-400",
  review: "bg-violet-500",
  "in-progress": "bg-sky-500",
  blocked: "bg-red-500",
  rejected: "bg-red-500",
  deprecated: "bg-muted-foreground",
  superseded: "bg-muted-foreground",
};

const TAG_DOT_CLASSES = [
  "bg-blue-500",
  "bg-emerald-500",
  "bg-violet-500",
  "bg-amber-500",
  "bg-pink-500",
  "bg-teal-500",
  "bg-rose-500",
  "bg-indigo-500",
  "bg-cyan-500",
  "bg-lime-500",
];

function hashString(value: string): number {
  let hash = 0;
  for (let i = 0; i < value.length; i++) {
    hash = (hash * 31 + value.charCodeAt(i)) | 0;
  }
  return Math.abs(hash);
}

/** Tailwind class for a page dot (workflow state, then tag, then primary). */
export function pageDotClass(page: CalendarPage): string {
  const workflow = (page.state ?? page.status ?? "").toLowerCase().trim();
  if (workflow && WORKFLOW_DOT_CLASSES[workflow]) {
    return WORKFLOW_DOT_CLASSES[workflow]!;
  }
  const tag = page.tags?.[0]?.toLowerCase().trim();
  if (tag) {
    return TAG_DOT_CLASSES[hashString(tag) % TAG_DOT_CLASSES.length]!;
  }
  return "bg-primary";
}

/** Monday-first month grid including leading/trailing adjacent-month days. */
export function buildMonthGrid(viewDate: Date): CalendarDayCell[] {
  const monthStart = startOfMonth(viewDate);
  const gridStart = startOfWeek(monthStart, { weekStartsOn: 1 });
  const monthEnd = endOfMonth(viewDate);
  const gridEnd = endOfWeek(monthEnd, { weekStartsOn: 1 });

  const cells: CalendarDayCell[] = [];
  let cursor = gridStart;
  while (cursor <= gridEnd) {
    cells.push({
      date: cursor,
      inMonth: isSameMonth(cursor, viewDate),
      iso: format(cursor, "yyyy-MM-dd"),
    });
    cursor = addDays(cursor, 1);
  }
  return cells;
}

/** Mobile week row: Mon–Sun for the week containing focusDate within viewMonth. */
export function buildMobileWeekCells(viewDate: Date, focusDate: Date): CalendarDayCell[] {
  const weekStart = startOfWeek(focusDate, { weekStartsOn: 1 });
  const cells: CalendarDayCell[] = [];
  for (let i = 0; i < 7; i++) {
    const date = addDays(weekStart, i);
    cells.push({
      date,
      inMonth: isSameMonth(date, viewDate),
      iso: format(date, "yyyy-MM-dd"),
    });
  }
  return cells;
}

export function formatMonthLabel(viewDate: Date): string {
  return format(viewDate, "MMMM yyyy");
}

export function isTodayDate(date: Date, now = new Date()): boolean {
  return isSameDay(date, now);
}

export { addMonths, addDays, startOfMonth };

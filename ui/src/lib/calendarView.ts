// Calendar view helpers — date parsing, DQL builders, and page dot colors.

import type { QueryResponse } from "@kw/lib/api";
import { tagColor } from "@kw/lib/kanbanUi";

/** Frontmatter fields commonly used for dates. */
export const CANDIDATE_DATE_FIELDS = [
  "date",
  "due",
  "due-date",
  "created",
  "last_executed",
  "last-reviewed",
  "review-by",
  "next-review",
  "last-tested",
] as const;

export type CalendarPage = {
  path: string;
  date: string;
  title?: string;
  status?: string;
  tags?: string[];
};

export type CalendarPagesByDate = Map<string, CalendarPage[]>;

const DATE_RE = /^\d{4}-\d{2}-\d{2}(?:[ T]\d{2}:\d{2})?/;

export function isDateLikeString(value: unknown): boolean {
  return typeof value === "string" && DATE_RE.test(value);
}

export function normalizeDateKey(value: unknown): string | null {
  if (value == null) return null;
  const raw = String(value).trim();
  if (!DATE_RE.test(raw)) return null;
  return raw.slice(0, 10);
}

export function monthKey(year: number, month: number): string {
  return `${year}-${String(month).padStart(2, "0")}`;
}

export function monthRange(year: number, month: number): { start: string; end: string } {
  const start = `${monthKey(year, month)}-01`;
  const nextMonth = month === 12 ? { y: year + 1, m: 1 } : { y: year, m: month + 1 };
  const end = `${monthKey(nextMonth.y, nextMonth.m)}-01`;
  return { start, end };
}

export function quoteDateField(field: string): string {
  const safe = field.replace(/"/g, "");
  return /[^a-zA-Z0-9_.]/.test(safe) ? `\`${safe}\`` : safe;
}

/** DQL to fetch pages for one calendar month on a given date field. */
export function buildMonthQuery(
  dateField: string,
  year: number,
  month: number,
): string {
  const ym = monthKey(year, month);
  const field = quoteDateField(dateField);
  return [
    `TABLE _path, title, status, tags, ${field}`,
    `WHERE ${field} IS NOT NULL`,
    `AND dateformat(${field}, "%Y-%m") = "${ym}"`,
    `SORT ${field} ASC`,
  ].join(" ");
}

/** DQL COUNT probe — returns whether a field has any dated pages. */
export function buildFieldProbeQuery(dateField: string): string {
  const field = quoteDateField(dateField);
  return `COUNT WHERE ${field} IS NOT NULL`;
}

export function parseCalendarRows(
  data: QueryResponse,
  dateField: string,
): CalendarPage[] {
  const pages: CalendarPage[] = [];
  for (const row of data.rows) {
    const date = normalizeDateKey(row[dateField]);
    if (!date) continue;
    const path = String(row._path ?? row.path ?? "");
    if (!path) continue;
    const tagsRaw = row.tags;
    let tags: string[] | undefined;
    if (Array.isArray(tagsRaw)) {
      tags = tagsRaw.map(String);
    } else if (typeof tagsRaw === "string" && tagsRaw.trim()) {
      tags = tagsRaw.split(",").map((t) => t.trim()).filter(Boolean);
    }
    pages.push({
      path,
      date,
      title: row.title != null ? String(row.title) : undefined,
      status: row.status != null ? String(row.status) : undefined,
      tags,
    });
  }
  return pages;
}

export function groupPagesByDate(pages: CalendarPage[]): CalendarPagesByDate {
  const byDate = new Map<string, CalendarPage[]>();
  for (const page of pages) {
    const list = byDate.get(page.date);
    if (list) list.push(page);
    else byDate.set(page.date, [page]);
  }
  return byDate;
}

/** Workflow / status colors for calendar dots (ADR states, kanban, etc.). */
const STATE_COLORS: Record<string, string> = {
  accepted: "#22C55E",
  approved: "#22C55E",
  done: "#22C55E",
  completed: "#22C55E",
  published: "#22C55E",
  proposed: "#EAB308",
  draft: "#9CA3AF",
  pending: "#F97316",
  review: "#8B5CF6",
  rejected: "#EF4444",
  deprecated: "#6B7280",
  blocked: "#EF4444",
  active: "#3B82F6",
};

export function pageDotColor(page: CalendarPage): string {
  if (page.status) {
    const key = page.status.toLowerCase().trim();
    if (STATE_COLORS[key]) return STATE_COLORS[key]!;
  }
  if (page.tags?.length) {
    return tagColor(page.tags[0]!).bg;
  }
  return "#6366F1";
}

export type MonthCell = number | null;

/** Build a Monday-first month grid (6 rows max). */
export function buildMonthGrid(year: number, month: number): MonthCell[] {
  const firstDay = new Date(year, month - 1, 1);
  const daysInMonth = new Date(year, month, 0).getDate();
  let startDow = firstDay.getDay() - 1;
  if (startDow < 0) startDow = 6;

  const cells: MonthCell[] = [];
  for (let i = 0; i < startDow; i++) cells.push(null);
  for (let d = 1; d <= daysInMonth; d++) cells.push(d);
  while (cells.length % 7 !== 0) cells.push(null);
  return cells;
}

export function dateKeyFromCell(year: number, month: number, day: number): string {
  return `${monthKey(year, month)}-${String(day).padStart(2, "0")}`;
}

export const WEEKDAY_LABELS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"] as const;

export const MONTH_NAMES = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
] as const;

/** ISO week containing `anchor` — Monday through Sunday date keys. */
export function weekDateKeys(anchor: Date): string[] {
  const d = new Date(anchor.getFullYear(), anchor.getMonth(), anchor.getDate());
  const dow = d.getDay();
  const mondayOffset = dow === 0 ? -6 : 1 - dow;
  d.setDate(d.getDate() + mondayOffset);

  const keys: string[] = [];
  for (let i = 0; i < 7; i++) {
    const y = d.getFullYear();
    const m = d.getMonth() + 1;
    const day = d.getDate();
    keys.push(dateKeyFromCell(y, m, day));
    d.setDate(d.getDate() + 1);
  }
  return keys;
}

export async function discoverDateFields(
  probe: (dql: string) => Promise<{ total?: number; rows?: unknown[] }>,
  candidates: readonly string[] = CANDIDATE_DATE_FIELDS,
): Promise<string[]> {
  const found: string[] = [];
  await Promise.all(
    candidates.map(async (field) => {
      try {
        const result = await probe(buildFieldProbeQuery(field));
        const count =
          typeof result.total === "number"
            ? result.total
            : Array.isArray(result.rows) && result.rows[0] != null
              ? Number((result.rows[0] as Record<string, unknown>).count ?? 0)
              : 0;
        if (count > 0) found.push(field);
      } catch {
        // field absent or query unsupported — skip
      }
    }),
  );
  if (found.length === 0) return ["date"];
  return found.sort((a, b) => {
    if (a === "date") return -1;
    if (b === "date") return 1;
    return a.localeCompare(b);
  });
}

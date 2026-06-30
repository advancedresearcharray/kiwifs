import { describe, expect, it } from "vitest";
import {
  addMonths,
  buildCalendarQuery,
  buildCalendarQueryRange,
  buildMonthGrid,
  dayAfter,
  detectDateFields,
  entryDotColor,
  groupByDate,
  isCalendarTableQuery,
  monthStartEnd,
  parseCalendarResponse,
  toDateKey,
  weekDateKeys,
} from "./calendarView";

describe("calendarView", () => {
  it("builds month-scoped TABLE query", () => {
    const dql = buildCalendarQuery("date", "2026-06");
    expect(dql).toBe(
      'TABLE _path, date, tags, state, title WHERE striptime(date) >= DATE("2026-06-01") AND striptime(date) < DATE("2026-07-01")',
    );
    expect(isCalendarTableQuery(dql)).toBe(true);
  });

  it("does not treat legacy CALENDAR queries as calendar TABLE queries", () => {
    expect(isCalendarTableQuery("CALENDAR date FROM events/")).toBe(false);
  });

  it("builds arbitrary date-range queries for mobile week spans", () => {
    expect(buildCalendarQueryRange("due", "2026-05-26", "2026-06-02")).toBe(
      'TABLE _path, due, tags, state, title WHERE striptime(due) >= DATE("2026-05-26") AND striptime(due) < DATE("2026-06-02")',
    );
  });

  it("computes exclusive end date for week queries", () => {
    expect(dayAfter("2026-06-21")).toBe("2026-06-22");
  });

  it("computes month boundaries including year rollover", () => {
    expect(monthStartEnd("2026-12")).toEqual({ start: "2026-12-01", end: "2027-01-01" });
  });

  it("addMonths navigates across year boundary", () => {
    expect(addMonths("2026-12", 1)).toBe("2027-01");
    expect(addMonths("2026-01", -1)).toBe("2025-12");
  });

  it("parses query rows into calendar entries", () => {
    const entries = parseCalendarResponse(
      {
        columns: ["_path", "date", "tags", "state", "title"],
        rows: [
          { _path: "notes/a.md", date: "2026-06-15", tags: ["bug"], state: "proposed", title: "Bug report" },
          { _path: "notes/b.md", date: "2026-06-15T10:00:00Z", title: "Morning note" },
          { _path: "notes/c.md", date: "invalid" },
        ],
        total: 3,
        has_more: false,
      },
      "date",
    );
    expect(entries).toHaveLength(2);
    expect(entries[0]).toMatchObject({
      path: "notes/a.md",
      date: "2026-06-15",
      state: "proposed",
      title: "Bug report",
    });
    expect(entries[1]).toMatchObject({ date: "2026-06-15", title: "Morning note" });
  });

  it("groups entries by date", () => {
    const grouped = groupByDate([
      { path: "a.md", date: "2026-06-01" },
      { path: "b.md", date: "2026-06-01" },
      { path: "c.md", date: "2026-06-02" },
    ]);
    expect(grouped.get("2026-06-01")).toHaveLength(2);
    expect(grouped.get("2026-06-02")).toHaveLength(1);
  });

  it("always includes default date fields even when meta samples omit them", () => {
    const fields = detectDateFields([{ due: "2026-06-01", title: "x" }]);
    expect(fields[0]).toBe("date");
    expect(fields).toContain("due");
    expect(fields).toContain("created");
    expect(fields).toContain("last_executed");
  });

  it("detects date-like frontmatter fields with defaults first", () => {
    const fields = detectDateFields([
      { due: "2026-06-01", title: "x", custom_at: "2026-01-01" },
      { date: "2026-06-02", custom_at: "2026-02-02" },
    ]);
    expect(fields[0]).toBe("date");
    expect(fields).toContain("due");
    expect(fields).toContain("custom_at");
  });

  it("colors dots by workflow state then tag", () => {
    expect(entryDotColor({ path: "a.md", date: "2026-06-01", state: "accepted" })).toBe("#22c55e");
    expect(entryDotColor({ path: "b.md", date: "2026-06-01", tags: ["bug"] })).toBe("#FECACA");
    expect(entryDotColor({ path: "c.md", date: "2026-06-01" })).toContain("primary");
  });

  it("builds a Monday-start month grid", () => {
    // June 2026 starts on Monday
    const cells = buildMonthGrid("2026-06");
    expect(cells[0]).toBe(1);
    expect(cells.filter((c) => c != null)).toHaveLength(30);
  });

  it("normalizes date keys", () => {
    expect(toDateKey("2026-06-15T12:00:00Z")).toBe("2026-06-15");
    expect(toDateKey(new Date("2026-06-15"))).toBe("2026-06-15");
    expect(toDateKey("nope")).toBeNull();
  });

  it("returns seven ISO keys for a week", () => {
    const keys = weekDateKeys(new Date(2026, 5, 18)); // Wednesday, local time
    expect(keys).toHaveLength(7);
    expect(keys[0]).toBe("2026-06-15"); // Monday
    expect(keys[6]).toBe("2026-06-21"); // Sunday
  });
});

import { describe, expect, it } from "vitest";
import {
  buildMonthGrid,
  buildMonthRangeQuery,
  buildWeekGrid,
  discoverDateFields,
  groupEntriesByDate,
  isCalendarViewPath,
  normalizeDateValue,
  pageDotColor,
  parseCalendarQueryRows,
  parseMonthInput,
  shiftMonth,
} from "./calendarView";

describe("calendarView helpers", () => {
  it("normalizes ISO date strings from frontmatter values", () => {
    expect(normalizeDateValue("2026-06-15T10:00:00Z")).toBe("2026-06-15");
    expect(normalizeDateValue("not-a-date")).toBeNull();
  });

  it("discovers date fields from meta results with defaults as fallback", () => {
    const fields = discoverDateFields([
      {
        path: "pages/a.md",
        frontmatter: { due: "2026-06-01", title: "A" },
      },
      {
        path: "pages/b.md",
        frontmatter: { date: "2026-06-02", last_executed: "2026-06-03" },
      },
    ]);

    expect(fields).toContain("date");
    expect(fields).toContain("due");
    expect(fields[0]).toBe("date");
  });

  it("builds a month-scoped DQL query", () => {
    const q = buildMonthRangeQuery("due", new Date("2026-06-15T12:00:00Z"));
    expect(q).toContain('due >= DATE("2026-06-01")');
    expect(q).toContain('due < DATE("2026-07-01")');
    expect(q).toContain("tags, state");
  });

  it("parses query rows into calendar entries", () => {
    const entries = parseCalendarQueryRows({
      columns: ["_path", "date", "tags", "state"],
      rows: [
        {
          _path: "pages/a.md",
          date: "2026-06-03",
          tags: ["docs"],
          state: "accepted",
          title: "Alpha",
        },
        {
          _path: "pages/b.md",
          date: "invalid",
        },
      ],
      total: 2,
      has_more: false,
    });

    expect(entries).toHaveLength(1);
    expect(entries[0]).toMatchObject({
      path: "pages/a.md",
      date: "2026-06-03",
      title: "Alpha",
      tags: ["docs"],
      state: "accepted",
    });
  });

  it("groups entries by date and colors by workflow state", () => {
    const grouped = groupEntriesByDate([
      { path: "a.md", date: "2026-06-01", title: "A", tags: [], state: "accepted" },
      { path: "b.md", date: "2026-06-01", title: "B", tags: ["bug"], state: "proposed" },
    ]);

    expect(grouped.get("2026-06-01")).toHaveLength(2);
    expect(pageDotColor(grouped.get("2026-06-01")![0])).toBe("#16A34A");
    expect(pageDotColor(grouped.get("2026-06-01")![1])).toBe("#CA8A04");
  });

  it("builds month and week grids with Monday as first day", () => {
    const monthCells = buildMonthGrid(new Date("2026-06-15T12:00:00Z"));
    expect(monthCells.length % 7).toBe(0);
    expect(monthCells.some((c) => c.dateKey === "2026-06-01" && c.inMonth)).toBe(true);
    expect(monthCells.some((c) => !c.inMonth)).toBe(true);

    const weekCells = buildWeekGrid(new Date("2026-06-15T12:00:00Z"));
    expect(weekCells).toHaveLength(7);
    expect(weekCells[0].dateKey).toBe("2026-06-15");
  });

  it("shifts months and parses month input", () => {
    const june = new Date("2026-06-15T12:00:00Z");
    expect(shiftMonth(june, 1).getMonth()).toBe(6);
    expect(shiftMonth(june, -1).getMonth()).toBe(4);
    expect(parseMonthInput("2026-06")?.getMonth()).toBe(5);
    expect(parseMonthInput("bad")).toBeNull();
  });

  it("detects calendar view route", () => {
    expect(isCalendarViewPath("/view/calendar")).toBe(true);
    expect(isCalendarViewPath("/page/foo")).toBe(false);
  });
});

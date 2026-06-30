import { describe, expect, it } from "vitest";
import {
  buildFieldDiscoveryQuery,
  buildMobileWeekCells,
  buildMonthGrid,
  buildMonthQuery,
  buildWeekQuery,
  discoverDateFields,
  groupPagesByDate,
  pageDotClass,
  parseCalendarDateValue,
  parseCalendarRows,
  sanitizeDateField,
  type CalendarPage,
} from "./calendarView";

describe("calendarView DQL builders", () => {
  it("buildMonthQuery scopes striptime bounds", () => {
    expect(buildMonthQuery("date", 2026, 5)).toBe(
      'TABLE _path, title, tags, state, status, date\nWHERE striptime(date) >= DATE("2026-06-01") AND striptime(date) < DATE("2026-07-01")',
    );
  });

  it("buildWeekQuery covers Monday through Sunday", () => {
    expect(buildWeekQuery("due", new Date(2026, 5, 18))).toContain(
      'WHERE striptime(due) >= DATE("2026-06-15")',
    );
    expect(buildWeekQuery("due", new Date(2026, 5, 18))).toContain(
      'AND striptime(due) < DATE("2026-06-22")',
    );
  });

  it("buildFieldDiscoveryQuery lists candidate columns", () => {
    const q = buildFieldDiscoveryQuery();
    expect(q).toContain("TABLE _path, date, due");
    expect(q).toContain("date != null OR due != null");
    expect(q).toContain("LIMIT 200");
  });
});

describe("sanitizeDateField", () => {
  it("accepts valid field names", () => {
    expect(sanitizeDateField("date")).toBe("date");
    expect(sanitizeDateField("next-review")).toBe("next-review");
    expect(sanitizeDateField("meta.due")).toBe("meta.due");
  });

  it("rejects invalid field names", () => {
    expect(sanitizeDateField("date; DROP")).toBeNull();
    expect(sanitizeDateField("")).toBeNull();
    expect(sanitizeDateField("9date")).toBeNull();
  });
});

describe("parseCalendarDateValue", () => {
  it("parses ISO date and datetime strings", () => {
    expect(parseCalendarDateValue("2026-06-20")).toBe("2026-06-20");
    expect(parseCalendarDateValue("2026-06-20T14:30:00Z")).toBe("2026-06-20");
  });

  it("returns null for empty or invalid values", () => {
    expect(parseCalendarDateValue(null)).toBeNull();
    expect(parseCalendarDateValue("")).toBeNull();
    expect(parseCalendarDateValue("not-a-date")).toBeNull();
  });
});

describe("parseCalendarRows and grouping", () => {
  it("maps query rows to calendar pages", () => {
    const pages = parseCalendarRows(
      [
        { _path: "notes/a.md", title: "A", date: "2026-06-01", tags: ["bug"] },
        { _path: "notes/b.md", date: "2026-06-01T10:00:00Z", state: "accepted" },
        { _path: "notes/c.md", date: null },
      ],
      "date",
    );
    expect(pages).toHaveLength(2);
    expect(pages[0]?.path).toBe("notes/a.md");
    expect(pages[1]?.dateKey).toBe("2026-06-01");
  });

  it("groups pages by date key", () => {
    const pages: CalendarPage[] = [
      { path: "a.md", dateKey: "2026-06-01" },
      { path: "b.md", dateKey: "2026-06-01" },
      { path: "c.md", dateKey: "2026-06-02" },
    ];
    const grouped = groupPagesByDate(pages);
    expect(grouped.get("2026-06-01")).toHaveLength(2);
    expect(grouped.get("2026-06-02")).toHaveLength(1);
  });
});

describe("discoverDateFields", () => {
  it("orders fields by candidate priority", () => {
    expect(
      discoverDateFields([
        { _path: "x.md", due: "2026-06-01", date: "2026-06-02" },
        { _path: "y.md", created: "2026-05-01" },
      ]),
    ).toEqual(["date", "due", "created"]);
  });

  it("falls back to date when discovery is empty", () => {
    expect(discoverDateFields([])).toEqual(["date"]);
  });
});

describe("pageDotClass", () => {
  it("prefers workflow state colors", () => {
    expect(pageDotClass({ path: "a.md", dateKey: "2026-06-01", state: "accepted" })).toBe(
      "bg-emerald-500",
    );
    expect(pageDotClass({ path: "b.md", dateKey: "2026-06-01", status: "proposed" })).toBe(
      "bg-amber-400",
    );
  });

  it("uses tag hash when no workflow state", () => {
    const cls = pageDotClass({ path: "c.md", dateKey: "2026-06-01", tags: ["feature"] });
    expect(cls.startsWith("bg-")).toBe(true);
  });

  it("falls back to primary", () => {
    expect(pageDotClass({ path: "d.md", dateKey: "2026-06-01" })).toBe("bg-primary");
  });
});

describe("calendar grids", () => {
  it("buildMonthGrid is Monday-first and includes adjacent month days", () => {
    const cells = buildMonthGrid(new Date(2026, 2, 15));
    expect(cells[0]?.iso).toBe("2026-02-23");
    expect(cells.some((c) => c.iso === "2026-03-01")).toBe(true);
    expect(cells.some((c) => c.iso === "2026-04-05")).toBe(true);
    expect(cells.filter((c) => c.inMonth)).toHaveLength(31);
  });

  it("buildMobileWeekCells renders seven days for the focus week", () => {
    const cells = buildMobileWeekCells(new Date(2026, 5, 15), new Date(2026, 5, 18));
    expect(cells).toHaveLength(7);
    expect(cells[0]?.iso).toBe("2026-06-15");
    expect(cells[6]?.iso).toBe("2026-06-21");
  });

  it("resolves cross-month week pages while viewing June", () => {
    const cells = buildMobileWeekCells(new Date(2026, 5, 1), new Date(2026, 5, 30));
    const julyCell = cells.find((c) => c.iso === "2026-07-01");
    expect(julyCell).toBeDefined();
    expect(julyCell?.inMonth).toBe(false);

    const pages = parseCalendarRows(
      [{ _path: "events/july.md", date: "2026-07-01", title: "July event" }],
      "date",
    );
    const grouped = groupPagesByDate(pages);
    expect(grouped.get("2026-07-01")).toHaveLength(1);
    expect(cells.some((c) => grouped.has(c.iso))).toBe(true);
  });
});

describe("invalid field sanitization in queries", () => {
  it("falls back to date when field is invalid", () => {
    expect(buildMonthQuery("bad;field", 2026, 0)).toContain("striptime(date)");
  });
});

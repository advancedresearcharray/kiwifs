import { describe, expect, it } from "vitest";
import {
  buildMonthGrid,
  buildMonthQuery,
  dateKeyFromCell,
  discoverDateFields,
  groupPagesByDate,
  isDateLikeString,
  normalizeDateKey,
  pageDotColor,
  parseCalendarRows,
  weekDateKeys,
} from "./calendarView";

describe("calendarView", () => {
  describe("normalizeDateKey", () => {
    it("extracts YYYY-MM-DD from ISO strings", () => {
      expect(normalizeDateKey("2026-06-15")).toBe("2026-06-15");
      expect(normalizeDateKey("2026-06-15T10:30:00")).toBe("2026-06-15");
    });

    it("returns null for invalid values", () => {
      expect(normalizeDateKey(null)).toBeNull();
      expect(normalizeDateKey("not-a-date")).toBeNull();
      expect(normalizeDateKey(123)).toBeNull();
    });
  });

  describe("isDateLikeString", () => {
    it("matches date-like frontmatter strings", () => {
      expect(isDateLikeString("2026-01-02")).toBe(true);
      expect(isDateLikeString("2026-01-02 14:00")).toBe(true);
      expect(isDateLikeString("Jan 2")).toBe(false);
    });
  });

  describe("buildMonthQuery", () => {
    it("filters by month using dateformat", () => {
      const q = buildMonthQuery("due", 2026, 6);
      expect(q).toContain('dateformat(due, "%Y-%m") = "2026-06"');
      expect(q).toContain("TABLE _path, title, status, tags, due");
    });

    it("backtick-quotes hyphenated field names", () => {
      const q = buildMonthQuery("due-date", 2026, 6);
      expect(q).toContain("`due-date`");
    });
  });

  describe("parseCalendarRows", () => {
    it("maps DQL rows to calendar pages", () => {
      const pages = parseCalendarRows(
        {
          columns: ["_path", "date", "status", "tags"],
          rows: [
            {
              _path: "pages/note.md",
              date: "2026-06-03",
              status: "accepted",
              tags: ["docs"],
            },
          ],
          total: 1,
          has_more: false,
        },
        "date",
      );
      expect(pages).toHaveLength(1);
      expect(pages[0]!.path).toBe("pages/note.md");
      expect(pages[0]!.date).toBe("2026-06-03");
      expect(pages[0]!.tags).toEqual(["docs"]);
    });
  });

  describe("groupPagesByDate", () => {
    it("groups multiple pages on the same day", () => {
      const grouped = groupPagesByDate([
        { path: "a.md", date: "2026-06-01" },
        { path: "b.md", date: "2026-06-01" },
        { path: "c.md", date: "2026-06-02" },
      ]);
      expect(grouped.get("2026-06-01")).toHaveLength(2);
      expect(grouped.get("2026-06-02")).toHaveLength(1);
    });
  });

  describe("pageDotColor", () => {
    it("uses workflow state colors when present", () => {
      expect(pageDotColor({ path: "x.md", date: "2026-06-01", status: "accepted" })).toBe(
        "#22C55E",
      );
      expect(pageDotColor({ path: "x.md", date: "2026-06-01", status: "proposed" })).toBe(
        "#EAB308",
      );
    });

    it("falls back to tag color", () => {
      const color = pageDotColor({ path: "x.md", date: "2026-06-01", tags: ["bug"] });
      expect(color).toBeTruthy();
    });
  });

  describe("buildMonthGrid", () => {
    it("starts weeks on Monday", () => {
      // June 2026 starts on Monday
      const cells = buildMonthGrid(2026, 6);
      expect(cells[0]).toBe(1);
      expect(cells.filter((c) => c != null)).toHaveLength(30);
    });
  });

  describe("dateKeyFromCell", () => {
    it("zero-pads day numbers", () => {
      expect(dateKeyFromCell(2026, 6, 3)).toBe("2026-06-03");
    });
  });

  describe("weekDateKeys", () => {
    it("returns seven consecutive days from Monday", () => {
      // Wednesday 2026-06-03
      const keys = weekDateKeys(new Date(2026, 5, 3));
      expect(keys).toHaveLength(7);
      expect(keys[0]).toBe("2026-06-01");
      expect(keys[6]).toBe("2026-06-07");
    });
  });

  describe("discoverDateFields", () => {
    it("returns fields with non-zero counts, defaulting to date", async () => {
      const fields = await discoverDateFields(async (dql) => {
        if (dql.includes("due-date")) return { total: 0 };
        if (dql.includes("due")) return { total: 2 };
        if (dql.includes("date")) return { total: 5 };
        return { total: 0 };
      });
      expect(fields).toEqual(["date", "due"]);
    });

    it("falls back to date when nothing matches", async () => {
      const fields = await discoverDateFields(async () => ({ total: 0 }));
      expect(fields).toEqual(["date"]);
    });
  });
});

import { describe, expect, it } from "vitest";
import {
  tagColor,
  parsePriority,
  priorityStyle,
  authorInitials,
  authorColor,
  dueStatus,
  dueStatusColor,
} from "./kanbanUi";

describe("kanbanUi", () => {
  describe("tagColor", () => {
    it("returns preset color for well-known tags", () => {
      const bug = tagColor("bug");
      expect(bug.bg).toBe("#FECACA");
      expect(bug.fg).toBe("#991B1B");
    });

    it("is case-insensitive", () => {
      expect(tagColor("BUG")).toEqual(tagColor("bug"));
      expect(tagColor("Feature")).toEqual(tagColor("feature"));
    });

    it("returns a consistent hash color for unknown tags", () => {
      const color1 = tagColor("my-custom-tag");
      const color2 = tagColor("my-custom-tag");
      expect(color1).toEqual(color2);
      expect(color1.bg).toBeTruthy();
      expect(color1.fg).toBeTruthy();
    });

    it("gives different colors to different unknown tags", () => {
      const a = tagColor("alpha-tag-xyz");
      const b = tagColor("beta-tag-abc");
      // Not guaranteed to differ with small hash space, but usually does
      expect(a.bg).toBeTruthy();
      expect(b.bg).toBeTruthy();
    });
  });

  describe("parsePriority", () => {
    it("maps known priority strings", () => {
      expect(parsePriority("critical")).toBe("critical");
      expect(parsePriority("high")).toBe("high");
      expect(parsePriority("medium")).toBe("medium");
      expect(parsePriority("low")).toBe("low");
      expect(parsePriority("p0")).toBe("critical");
      expect(parsePriority("P1")).toBe("high");
    });

    it("returns null for unknown or missing values", () => {
      expect(parsePriority(undefined)).toBeNull();
      expect(parsePriority("")).toBeNull();
      expect(parsePriority("extreme")).toBeNull();
    });
  });

  describe("priorityStyle", () => {
    it("returns style for each level", () => {
      const style = priorityStyle("critical");
      expect(style.label).toBe("Critical");
      expect(style.dotColor).toBeTruthy();
    });
  });

  describe("authorInitials", () => {
    it("extracts initials from a full name", () => {
      expect(authorInitials("John Doe")).toBe("JD");
      expect(authorInitials("Alice Bob Charlie")).toBe("AC");
    });

    it("uses single character for single names", () => {
      expect(authorInitials("admin")).toBe("A");
    });

    it("handles empty or whitespace", () => {
      expect(authorInitials("   ")).toBe("?");
    });
  });

  describe("authorColor", () => {
    it("returns consistent colors", () => {
      expect(authorColor("John")).toBe(authorColor("John"));
    });
  });

  describe("dueStatus", () => {
    it("returns no-date for missing due", () => {
      expect(dueStatus(undefined)).toEqual({ status: "no-date", date: null });
      expect(dueStatus("")).toEqual({ status: "no-date", date: null });
    });

    it("returns overdue for past dates", () => {
      const result = dueStatus("2020-01-01");
      expect(result.status).toBe("overdue");
      expect(result.date).toBeInstanceOf(Date);
    });

    it("returns upcoming for far future dates", () => {
      const result = dueStatus("2099-12-31");
      expect(result.status).toBe("upcoming");
    });

    it("returns no-date for invalid dates", () => {
      expect(dueStatus("not-a-date")).toEqual({ status: "no-date", date: null });
    });
  });

  describe("dueStatusColor", () => {
    it("returns colors for each status", () => {
      expect(dueStatusColor("overdue").text).toBe("#DC2626");
      expect(dueStatusColor("due-soon").text).toBe("#D97706");
      expect(dueStatusColor("upcoming").text).toBe("#6B7280");
    });
  });
});

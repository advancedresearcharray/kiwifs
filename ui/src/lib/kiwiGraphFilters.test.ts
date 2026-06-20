import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import {
  collectRelationTypes,
  edgeMatchesRelationFilter,
  loadRelationFilterFromSession,
  nodeMatchesRelationFilter,
  reconcileRelationFilter,
  relationLabel,
  RELATION_FILTER_SESSION_KEY,
  resolveGraphLinks,
  saveRelationFilterToSession,
  shouldShowRelationFilters,
} from "./kiwiGraphFilters";

describe("kiwiGraphFilters", () => {
  describe("relationLabel", () => {
    it("labels empty relation as wiki-link", () => {
      expect(relationLabel("")).toBe("wiki-link");
    });

    it("passes through typed relation names", () => {
      expect(relationLabel("contradicts")).toBe("contradicts");
      expect(relationLabel("cites")).toBe("cites");
    });
  });

  describe("collectRelationTypes", () => {
    it("returns unique sorted relation types with wiki-link first", () => {
      expect(
        collectRelationTypes([
          { relation: "cites" },
          { relation: "" },
          { relation: "contradicts" },
          { relation: "cites" },
        ]),
      ).toEqual(["", "cites", "contradicts"]);
    });
  });

  describe("edgeMatchesRelationFilter", () => {
    it("matches all edges when filter is empty", () => {
      const all = new Set<string>();
      expect(edgeMatchesRelationFilter("", all)).toBe(true);
      expect(edgeMatchesRelationFilter("cites", all)).toBe(true);
    });

    it("matches only selected relation types", () => {
      const selected = new Set(["cites", "contradicts"]);
      expect(edgeMatchesRelationFilter("cites", selected)).toBe(true);
      expect(edgeMatchesRelationFilter("contradicts", selected)).toBe(true);
      expect(edgeMatchesRelationFilter("", selected)).toBe(false);
      expect(edgeMatchesRelationFilter("supersedes", selected)).toBe(false);
    });
  });

  describe("nodeMatchesRelationFilter", () => {
    const links = [
      { source: "a.md", target: "b.md", relation: "cites" },
      { source: "a.md", target: "c.md", relation: "" },
      { source: "b.md", target: "c.md", relation: "contradicts" },
    ];

    it("matches all nodes when filter is empty", () => {
      expect(nodeMatchesRelationFilter("a.md", links, new Set())).toBe(true);
      expect(nodeMatchesRelationFilter("z.md", links, new Set())).toBe(true);
    });

    it("matches nodes on filtered edges only", () => {
      const citesOnly = new Set(["cites"]);
      expect(nodeMatchesRelationFilter("a.md", links, citesOnly)).toBe(true);
      expect(nodeMatchesRelationFilter("b.md", links, citesOnly)).toBe(true);
      expect(nodeMatchesRelationFilter("c.md", links, citesOnly)).toBe(false);
    });

    it("supports multi-select relation filters", () => {
      const selected = new Set(["cites", "contradicts"]);
      expect(nodeMatchesRelationFilter("c.md", links, selected)).toBe(true);
      expect(nodeMatchesRelationFilter("a.md", links, selected)).toBe(true);
    });
  });

  describe("resolveGraphLinks", () => {
    it("keeps separate edges per relation between the same nodes", () => {
      const nodeIds = new Set(["pages/a.md", "pages/b.md"]);
      const resolver = (target: string) =>
        target === "pages/b.md" ? "pages/b.md" : null;
      const links = resolveGraphLinks(
        [
          { source: "pages/a.md", target: "pages/b.md", relation: "cites" },
          { source: "pages/a.md", target: "pages/b.md", relation: "contradicts" },
          { source: "pages/a.md", target: "pages/b.md" },
        ],
        resolver,
        nodeIds,
      );
      expect(links).toHaveLength(3);
      expect(links.map((l) => l.relation).sort()).toEqual(["", "cites", "contradicts"]);
    });
  });

  describe("shouldShowRelationFilters", () => {
    it("hides controls when only wiki-links exist", () => {
      expect(shouldShowRelationFilters([""])).toBe(false);
    });

    it("shows controls for typed links even if only one relation bucket", () => {
      expect(shouldShowRelationFilters(["cites"])).toBe(true);
    });

    it("shows controls when multiple relation types exist", () => {
      expect(shouldShowRelationFilters(["", "cites"])).toBe(true);
    });
  });

  describe("reconcileRelationFilter", () => {
    it("returns empty set when filter is empty", () => {
      expect(reconcileRelationFilter(new Set(), ["", "cites"])).toEqual(new Set());
    });

    it("keeps only relations present in the graph", () => {
      expect(
        reconcileRelationFilter(new Set(["cites", "contradicts"]), ["", "cites"]),
      ).toEqual(new Set(["cites"]));
    });

    it("resets to All when no selected relations remain valid", () => {
      expect(
        reconcileRelationFilter(new Set(["cites"]), ["", "contradicts"]),
      ).toEqual(new Set());
    });
  });

  describe("session persistence", () => {
    const storage = new Map<string, string>();

    beforeEach(() => {
      storage.clear();
      vi.stubGlobal("sessionStorage", {
        getItem: (key: string) => storage.get(key) ?? null,
        setItem: (key: string, value: string) => {
          storage.set(key, value);
        },
        removeItem: (key: string) => {
          storage.delete(key);
        },
        clear: () => storage.clear(),
      });
    });

    afterEach(() => {
      vi.unstubAllGlobals();
    });

    it("round-trips selected relation types", () => {
      saveRelationFilterToSession(new Set(["cites", "contradicts"]));
      expect(loadRelationFilterFromSession()).toEqual(new Set(["cites", "contradicts"]));
    });

    it("clears storage when all relations are selected", () => {
      saveRelationFilterToSession(new Set(["cites"]));
      saveRelationFilterToSession(new Set());
      expect(sessionStorage.getItem(RELATION_FILTER_SESSION_KEY)).toBeNull();
      expect(loadRelationFilterFromSession()).toEqual(new Set());
    });
  });
});

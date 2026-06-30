import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import {
  SPLIT_VIEW_STORAGE_KEY,
  closeSplitState,
  compareVersionSplitState,
  createSplitState,
  isMobileViewport,
  loadSplitViewFromSession,
  navigateSplitPane,
  openInSplitState,
  saveSplitViewToSession,
  toggleSplitState,
} from "./splitView";

describe("splitView", () => {
  describe("toggleSplitState", () => {
    it("opens split with active path on both panes", () => {
      const next = toggleSplitState(closeSplitState(), "docs/readme.md");
      expect(next.enabled).toBe(true);
      expect(next.leftPath).toBe("docs/readme.md");
      expect(next.rightPath).toBe("docs/readme.md");
    });

    it("closes an active split", () => {
      const open = createSplitState("a.md", "b.md");
      const next = toggleSplitState(open, "a.md");
      expect(next.enabled).toBe(false);
    });

    it("no-ops when no active path and split is closed", () => {
      const next = toggleSplitState(closeSplitState(), null);
      expect(next.enabled).toBe(false);
    });
  });

  describe("openInSplitState", () => {
    it("uses active path as left pane when opening from tree", () => {
      const next = openInSplitState(closeSplitState(), "left.md", "right.md");
      expect(next).toEqual(createSplitState("left.md", "right.md"));
    });

    it("preserves left pane when already split", () => {
      const current = createSplitState("keep.md", "old-right.md");
      const next = openInSplitState(current, "ignored.md", "new-right.md");
      expect(next.leftPath).toBe("keep.md");
      expect(next.rightPath).toBe("new-right.md");
    });
  });

  describe("compareVersionSplitState", () => {
    it("pins historical version on the right pane", () => {
      const next = compareVersionSplitState("page.md", "abc1234");
      expect(next.leftPath).toBe("page.md");
      expect(next.rightPath).toBe("page.md");
      expect(next.rightVersionHash).toBe("abc1234");
    });
  });

  describe("navigateSplitPane", () => {
    it("clears version hash when navigating the right pane", () => {
      const state = compareVersionSplitState("page.md", "deadbeef");
      const next = navigateSplitPane(state, "right", "other.md");
      expect(next.rightPath).toBe("other.md");
      expect(next.rightVersionHash).toBeNull();
    });
  });

  describe("session persistence", () => {
    const store = new Map<string, string>();

    beforeEach(() => {
      store.clear();
      vi.stubGlobal("sessionStorage", {
        getItem: (key: string) => store.get(key) ?? null,
        setItem: (key: string, value: string) => {
          store.set(key, value);
        },
        removeItem: (key: string) => {
          store.delete(key);
        },
        clear: () => store.clear(),
      });
    });

    afterEach(() => {
      vi.unstubAllGlobals();
    });

    it("round-trips enabled split state", () => {
      const state = createSplitState("a.md", "b.md", { sizes: [60, 40] });
      saveSplitViewToSession(state);
      const loaded = loadSplitViewFromSession();
      expect(loaded?.enabled).toBe(true);
      expect(loaded?.leftPath).toBe("a.md");
      expect(loaded?.rightPath).toBe("b.md");
      expect(loaded?.sizes).toEqual([60, 40]);
    });

    it("removes storage when split is closed", () => {
      saveSplitViewToSession(createSplitState("a.md", "b.md"));
      saveSplitViewToSession(closeSplitState());
      expect(sessionStorage.getItem(SPLIT_VIEW_STORAGE_KEY)).toBeNull();
    });
  });

  describe("isMobileViewport", () => {
    it("treats <=767px as mobile", () => {
      expect(isMobileViewport(767)).toBe(true);
      expect(isMobileViewport(768)).toBe(false);
    });
  });
});

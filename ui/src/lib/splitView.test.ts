import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import {
  SPLIT_VIEW_STORAGE_KEY,
  clampPaneSize,
  defaultSplitViewState,
  loadSplitViewState,
  openSplitView,
  parseSplitViewState,
  saveSplitViewState,
  toggleSplitView,
  updateSplitPanePath,
} from "./splitView";

describe("splitView", () => {
  beforeEach(() => {
    vi.stubGlobal("sessionStorage", {
      store: {} as Record<string, string>,
      getItem(key: string) {
        return this.store[key] ?? null;
      },
      setItem(key: string, value: string) {
        this.store[key] = value;
      },
      removeItem(key: string) {
        delete this.store[key];
      },
      clear() {
        this.store = {};
      },
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("clamps pane sizes to safe bounds", () => {
    expect(clampPaneSize(5)).toBe(20);
    expect(clampPaneSize(95)).toBe(80);
    expect(clampPaneSize(50)).toBe(50);
    expect(clampPaneSize(Number.NaN)).toBe(50);
  });

  it("creates default split state with both panes", () => {
    const state = defaultSplitViewState("pages/a.md", "pages/b.md");
    expect(state.enabled).toBe(true);
    expect(state.left.path).toBe("pages/a.md");
    expect(state.right.path).toBe("pages/b.md");
    expect(state.leftSize).toBe(50);
  });

  it("opens split with optional version on the right pane", () => {
    const state = openSplitView(null, "pages/a.md", "pages/a.md", "abc123");
    expect(state.right.versionHash).toBe("abc123");
    expect(state.left.versionHash).toBeNull();
  });

  it("toggles split off when already enabled", () => {
    const open = defaultSplitViewState("a.md");
    saveSplitViewState(open);
    expect(toggleSplitView(open, "a.md")).toBeNull();
    expect(loadSplitViewState()).toBeNull();
  });

  it("toggles split on from closed state", () => {
    const next = toggleSplitView(null, "pages/x.md");
    expect(next?.enabled).toBe(true);
    expect(next?.left.path).toBe("pages/x.md");
    expect(loadSplitViewState()?.left.path).toBe("pages/x.md");
  });

  it("persists and restores from sessionStorage", () => {
    const state = openSplitView(null, "left.md", "right.md");
    saveSplitViewState(state);
    const loaded = loadSplitViewState();
    expect(loaded).toEqual(state);
  });

  it("updates pane paths independently", () => {
    const base = defaultSplitViewState("a.md", "b.md");
    const updated = updateSplitPanePath(base, "right", "c.md");
    expect(updated.right.path).toBe("c.md");
    expect(updated.left.path).toBe("a.md");
    expect(loadSplitViewState()?.right.path).toBe("c.md");
  });

  it("rejects invalid persisted payloads", () => {
    expect(parseSplitViewState(null)).toBeNull();
    expect(parseSplitViewState({ enabled: true })).toBeNull();
    sessionStorage.setItem(SPLIT_VIEW_STORAGE_KEY, "{not json");
    expect(loadSplitViewState()).toBeNull();
  });
});

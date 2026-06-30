import { describe, expect, it, beforeEach, afterEach, vi } from "vitest";
import {
  SPLIT_VIEW_SESSION_KEY,
  closeSecondaryPane,
  createSplitViewState,
  loadSplitViewState,
  openPathInSplit,
  openVersionInSplit,
  saveSplitViewState,
  splitViewHasSecondary,
  syncSplitViewWithActivePath,
  toggleSplitView,
  clampSplitSizes,
} from "./splitView";

describe("splitView", () => {
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

  it("persists and restores split state", () => {
    const state = openPathInSplit(createSplitViewState(), "notes/b.md");
    saveSplitViewState(state);
    expect(loadSplitViewState()).toEqual(state);
  });

  it("ignores invalid session payloads", () => {
    storage.set(SPLIT_VIEW_SESSION_KEY, "{not json");
    expect(loadSplitViewState()).toBeNull();
    storage.set(SPLIT_VIEW_SESSION_KEY, JSON.stringify({ enabled: true, sizes: [0, 0] }));
    expect(loadSplitViewState()?.sizes).toEqual([50, 50]);
  });

  it("opens a page in the secondary pane", () => {
    const next = openPathInSplit(createSplitViewState(), "a.md");
    expect(next.enabled).toBe(true);
    expect(next.rightPath).toBe("a.md");
    expect(next.rightVersion).toBeNull();
    expect(splitViewHasSecondary(next)).toBe(true);
  });

  it("opens a historical version in the secondary pane", () => {
    const next = openVersionInSplit(createSplitViewState(), { path: "a.md", hash: "abc123" });
    expect(next.rightPath).toBeNull();
    expect(next.rightVersion).toEqual({ path: "a.md", hash: "abc123" });
  });

  it("toggles split view on and off", () => {
    const off = createSplitViewState();
    const on = toggleSplitView(off, "page.md");
    expect(on.enabled).toBe(true);
    expect(on.rightPath).toBe("page.md");
    const offAgain = toggleSplitView(on, "page.md");
    expect(offAgain.enabled).toBe(false);
  });

  it("toggle preserves custom pane sizes across on/off cycles", () => {
    const custom: [number, number] = [35, 65];
    const base = createSplitViewState({ sizes: custom });
    const on = toggleSplitView(base, "a.md");
    expect(on.sizes).toEqual(custom);
    const off = toggleSplitView(on, "a.md");
    expect(off.sizes).toEqual(custom);
    expect(off.enabled).toBe(false);
  });

  it("double toggle returns to disabled state without side effects", () => {
    const initial = createSplitViewState();
    const once = toggleSplitView(initial, "notes/x.md");
    const twice = toggleSplitView(once, "notes/x.md");
    expect(twice).toEqual(initial);
  });

  it("toggle off clears secondary content including version compare", () => {
    const withVersion = openVersionInSplit(createSplitViewState(), { path: "a.md", hash: "abc" });
    const cleared = toggleSplitView(withVersion, "a.md");
    expect(cleared.enabled).toBe(false);
    expect(splitViewHasSecondary(cleared)).toBe(false);
  });

  it("toggle on when already open via openPathInSplit turns off", () => {
    const open = openPathInSplit(createSplitViewState(), "left.md");
    expect(open.enabled).toBe(true);
    const closed = toggleSplitView(open, "left.md");
    expect(closed.enabled).toBe(false);
  });

  it("does not enable split without an active page", () => {
    const state = toggleSplitView(createSplitViewState(), null);
    expect(state.enabled).toBe(false);
  });

  it("closes split view from the secondary pane", () => {
    const state = openVersionInSplit(createSplitViewState(), { path: "a.md", hash: "deadbeef" });
    const cleared = closeSecondaryPane(state);
    expect(cleared.enabled).toBe(false);
    expect(splitViewHasSecondary(cleared)).toBe(false);
    expect(cleared.sizes).toEqual(state.sizes);
  });

  it("clamps resize percentages", () => {
    expect(clampSplitSizes(10)).toEqual([20, 80]);
    expect(clampSplitSizes(90)).toEqual([80, 20]);
    expect(clampSplitSizes(55)).toEqual([55, 45]);
  });

  describe("syncSplitViewWithActivePath", () => {
    it("disables split view when primary page is cleared", () => {
      const open = openPathInSplit(createSplitViewState({ sizes: [40, 60] }), "a.md");
      const synced = syncSplitViewWithActivePath(open, null);
      expect(synced.enabled).toBe(false);
      expect(synced.sizes).toEqual([40, 60]);
    });

    it("leaves split view unchanged when primary page exists", () => {
      const open = openPathInSplit(createSplitViewState(), "a.md");
      expect(syncSplitViewWithActivePath(open, "a.md")).toBe(open);
    });
  });
});

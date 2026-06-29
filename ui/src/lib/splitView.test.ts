import { describe, expect, it, beforeEach, vi } from "vitest";
import {
  clampPaneSize,
  defaultSplitViewState,
  loadSplitViewState,
  saveSplitViewState,
  splitViewStorageKey,
  SPLIT_VIEW_MIN_PANE,
} from "./splitView";

function mockSessionStorage() {
  const store = new Map<string, string>();
  const sessionStorage = {
    get length() {
      return store.size;
    },
    key(index: number) {
      return [...store.keys()][index] ?? null;
    },
    getItem(key: string) {
      return store.get(key) ?? null;
    },
    setItem(key: string, value: string) {
      store.set(key, value);
    },
    removeItem(key: string) {
      store.delete(key);
    },
    clear() {
      store.clear();
    },
  };
  vi.stubGlobal("sessionStorage", sessionStorage);
  vi.stubGlobal("window", { sessionStorage });
  return store;
}

describe("splitView storage", () => {
  beforeEach(() => {
    mockSessionStorage().clear();
  });

  it("uses space-scoped sessionStorage keys", () => {
    expect(splitViewStorageKey("default")).toBe("kiwifs-split-view:default");
    expect(splitViewStorageKey("notes")).toBe("kiwifs-split-view:notes");
  });

  it("returns defaults when storage is empty", () => {
    const state = loadSplitViewState("default", "docs/readme.md");
    expect(state).toEqual(defaultSplitViewState("docs/readme.md"));
  });

  it("persists and restores enabled split state", () => {
    const initial = {
      ...defaultSplitViewState("a.md"),
      enabled: true,
      rightPath: "b.md",
      leftSize: 60,
      rightSize: 40,
      leftVersionHash: "abc1234",
    };
    saveSplitViewState("default", initial);
    const loaded = loadSplitViewState("default", "fallback.md");
    expect(loaded.enabled).toBe(true);
    expect(loaded.rightPath).toBe("b.md");
    expect(loaded.leftSize).toBe(60);
    expect(loaded.leftVersionHash).toBe("abc1234");
  });

  it("clamps invalid pane sizes", () => {
    expect(clampPaneSize(5)).toBe(SPLIT_VIEW_MIN_PANE);
    expect(clampPaneSize(95)).toBe(100 - SPLIT_VIEW_MIN_PANE);
    expect(clampPaneSize(Number.NaN)).toBe(50);
  });
});

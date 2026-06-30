import { describe, expect, it, beforeEach } from "vitest";
import {
  SPLIT_VIEW_STORAGE_KEY,
  clampPaneSize,
  closeSplitView,
  loadSplitViewState,
  openSplitView,
  parseSplitViewState,
  saveSplitViewState,
  toggleSplitView,
} from "./splitView";

function mockSessionStorage() {
  const store = new Map<string, string>();
  const storage = {
    getItem: (key: string) => store.get(key) ?? null,
    setItem: (key: string, value: string) => {
      store.set(key, value);
    },
    removeItem: (key: string) => {
      store.delete(key);
    },
    clear: () => {
      store.clear();
    },
  };
  Object.defineProperty(globalThis, "sessionStorage", {
    value: storage,
    configurable: true,
  });
  return store;
}

describe("splitView", () => {
  beforeEach(() => {
    mockSessionStorage();
  });

  it("clamps pane sizes to allowed range", () => {
    expect(clampPaneSize(10)).toBe(25);
    expect(clampPaneSize(90)).toBe(75);
    expect(clampPaneSize(50)).toBe(50);
  });

  it("opens and closes split view state", () => {
    const base = loadSplitViewState();
    const opened = openSplitView(base, { path: "a.md" }, { path: "b.md" });
    expect(opened.enabled).toBe(true);
    expect(opened.left?.path).toBe("a.md");
    expect(opened.right?.path).toBe("b.md");

    const closed = closeSplitView(opened);
    expect(closed.enabled).toBe(false);
    expect(closed.right).toBeNull();
  });

  it("toggles split view from primary path", () => {
    const base = loadSplitViewState();
    const opened = toggleSplitView(base, "notes/index.md");
    expect(opened.enabled).toBe(true);
    expect(opened.left?.path).toBe("notes/index.md");
    expect(opened.right?.path).toBe("notes/index.md");

    const closed = toggleSplitView(opened, "notes/index.md");
    expect(closed.enabled).toBe(false);
  });

  it("persists enabled split state in sessionStorage", () => {
    const state = openSplitView(loadSplitViewState(), { path: "left.md" }, { path: "right.md", versionHash: "abc1234" });
    saveSplitViewState(state);

    const raw = sessionStorage.getItem(SPLIT_VIEW_STORAGE_KEY);
    expect(raw).toBeTruthy();
    const restored = parseSplitViewState(raw);
    expect(restored?.enabled).toBe(true);
    expect(restored?.right?.versionHash).toBe("abc1234");
  });

  it("clears storage when split is disabled", () => {
    saveSplitViewState(openSplitView(loadSplitViewState(), { path: "a.md" }, { path: "b.md" }));
    saveSplitViewState(closeSplitView(loadSplitViewState()));
    expect(sessionStorage.getItem(SPLIT_VIEW_STORAGE_KEY)).toBeNull();
  });

  it("ignores invalid persisted payloads", () => {
    sessionStorage.setItem(SPLIT_VIEW_STORAGE_KEY, "{not-json");
    expect(loadSplitViewState().enabled).toBe(false);
  });
});

describe("toggle_split_view keybinding chord", () => {
  it("matches backslash with mod", async () => {
    const { eventMatchesChord } = await import("./kiwiKeybindings");
    const e = {
      key: "\\",
      ctrlKey: true,
      metaKey: false,
      shiftKey: false,
      altKey: false,
    } as KeyboardEvent;
    expect(eventMatchesChord(e, "mod+\\")).toBe(true);
  });
});

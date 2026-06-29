import { describe, expect, it, vi } from "vitest";
import {
  DEFAULT_KEYBINDINGS,
  buildChordIndex,
  eventMatchesChord,
  formatChordDisplay,
  formatChordSegments,
  getCustomShortcutItems,
  isTextInputTarget,
  matchBoundAction,
  mergeKeybindings,
  normalizeChord,
} from "./kiwiKeybindings";

describe("normalizeChord", () => {
  it("canonicalizes modifier order and aliases", () => {
    expect(normalizeChord("Ctrl+Shift+B")).toBe("mod+shift+b");
    expect(normalizeChord("Mod+K")).toBe("mod+k");
    expect(normalizeChord("Escape")).toBe("escape");
  });
});

describe("eventMatchesChord", () => {
  it("matches mod shortcuts cross-browser", () => {
    const e = {
      key: "k",
      ctrlKey: true,
      metaKey: false,
      shiftKey: false,
      altKey: false,
    } as KeyboardEvent;
    expect(eventMatchesChord(e, "Mod+K")).toBe(true);
  });

  it("matches help shortcut on slash and question mark", () => {
    const slash = {
      key: "/",
      ctrlKey: true,
      metaKey: false,
      shiftKey: false,
      altKey: false,
    } as KeyboardEvent;
    const question = {
      key: "?",
      ctrlKey: true,
      metaKey: false,
      shiftKey: true,
      altKey: false,
    } as KeyboardEvent;
    expect(eventMatchesChord(slash, "Mod+/")).toBe(true);
    expect(eventMatchesChord(question, "Mod+/")).toBe(true);
  });
});

describe("mergeKeybindings", () => {
  it("keeps defaults when config is empty", () => {
    const merged = mergeKeybindings(null);
    expect(merged.search).toBe(DEFAULT_KEYBINDINGS.search);
  });

  it("applies server overrides", () => {
    const merged = mergeKeybindings({
      bindings: { search: "mod+j" },
      defaults: DEFAULT_KEYBINDINGS,
      conflicts: [],
    });
    expect(merged.search).toBe("mod+j");
    expect(merged.save).toBe(DEFAULT_KEYBINDINGS.save);
  });
});

describe("matchBoundAction", () => {
  it("returns the first matching action", () => {
    const bindings = mergeKeybindings({
      bindings: { graph: "mod+g" },
      defaults: DEFAULT_KEYBINDINGS,
      conflicts: [],
    });
    const e = {
      key: "g",
      ctrlKey: true,
      metaKey: false,
      shiftKey: false,
      altKey: false,
    } as KeyboardEvent;
    expect(matchBoundAction(e, bindings)).toBe("graph");
  });
});

describe("buildChordIndex", () => {
  it("groups actions by normalized chord", () => {
    const bindings = mergeKeybindings({
      bindings: { search: "mod+k", new_page: "mod+k" },
      defaults: DEFAULT_KEYBINDINGS,
      conflicts: [],
    });
    const index = buildChordIndex(bindings);
    expect(index.get("mod+k")?.sort()).toEqual(["new_page", "search"]);
  });
});

describe("formatChordDisplay", () => {
  it("formats mod shortcuts for display", () => {
    expect(formatChordDisplay("mod+k")).toMatch(/K/i);
  });
});

describe("formatChordSegments", () => {
  it("shows ⌘ on macOS and Ctrl elsewhere", () => {
    vi.stubGlobal("navigator", { platform: "MacIntel" });
    expect(formatChordSegments("mod+k")).toEqual(["⌘", "K"]);
    vi.stubGlobal("navigator", { platform: "Win32" });
    expect(formatChordSegments("mod+k")).toEqual(["Ctrl", "K"]);
    vi.unstubAllGlobals();
  });
});

describe("getCustomShortcutItems", () => {
  it("lists only bindings that differ from defaults", () => {
    const bindings = mergeKeybindings({
      bindings: { search: "mod+j", save: "mod+s" },
      defaults: DEFAULT_KEYBINDINGS,
      conflicts: [],
    });
    const custom = getCustomShortcutItems(bindings, DEFAULT_KEYBINDINGS);
    expect(custom.map((item) => item.action)).toEqual(["search"]);
  });
});

describe("isTextInputTarget", () => {
  it("detects native inputs and editor surfaces", () => {
    class MockHTMLElement {
      tagName = "";
      isContentEditable = false;
      closest(_selector: string): MockHTMLElement | null {
        return null;
      }
    }
    vi.stubGlobal("HTMLElement", MockHTMLElement);

    const input = new MockHTMLElement();
    input.tagName = "INPUT";
    expect(
      isTextInputTarget({ target: input } as unknown as KeyboardEvent),
    ).toBe(true);

    const editorHost = new MockHTMLElement();
    editorHost.closest = (selector: string) =>
      selector.includes("cm-editor") ? editorHost : null;
    expect(
      isTextInputTarget({ target: editorHost } as unknown as KeyboardEvent),
    ).toBe(true);

    const div = new MockHTMLElement();
    expect(
      isTextInputTarget({ target: div } as unknown as KeyboardEvent),
    ).toBe(false);

    vi.unstubAllGlobals();
  });
});

describe("plain question mark shortcut", () => {
  it("does not match mod+/ binding without a modifier", () => {
    const e = {
      key: "?",
      ctrlKey: false,
      metaKey: false,
      shiftKey: true,
      altKey: false,
    } as KeyboardEvent;
    expect(matchBoundAction(e, DEFAULT_KEYBINDINGS)).toBeNull();
  });
});

import { describe, expect, it } from "vitest";
import {
  DEFAULT_KEYBINDINGS,
  buildChordIndex,
  buildShortcutSectionsForDisplay,
  eventMatchesChord,
  formatChordDisplay,
  getCustomKeybindingItems,
  isKeyboardShortcutIgnoredTarget,
  isQuestionMarkShortcut,
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

describe("isKeyboardShortcutIgnoredTarget", () => {
  it("ignores native text inputs", () => {
    const input = { tagName: "INPUT", isContentEditable: false, closest: () => null } as unknown as HTMLElement;
    expect(isKeyboardShortcutIgnoredTarget(input)).toBe(true);
  });

  it("ignores contenteditable regions", () => {
    const div = {
      tagName: "DIV",
      isContentEditable: true,
      closest: () => null,
    } as unknown as HTMLElement;
    expect(isKeyboardShortcutIgnoredTarget(div)).toBe(true);
  });

  it("allows non-editable elements", () => {
    const button = {
      tagName: "BUTTON",
      isContentEditable: false,
      closest: () => null,
    } as unknown as HTMLElement;
    expect(isKeyboardShortcutIgnoredTarget(button)).toBe(false);
  });
});

describe("isQuestionMarkShortcut", () => {
  it("matches standalone question mark", () => {
    expect(
      isQuestionMarkShortcut({
        key: "?",
        shiftKey: true,
        metaKey: false,
        ctrlKey: false,
        altKey: false,
      }),
    ).toBe(true);
  });

  it("rejects mod+/ combo", () => {
    expect(
      isQuestionMarkShortcut({
        key: "/",
        shiftKey: false,
        metaKey: true,
        ctrlKey: false,
        altKey: false,
      }),
    ).toBe(false);
  });
});

describe("buildShortcutSectionsForDisplay", () => {
  it("adds Custom section for config overrides", () => {
    const bindings = mergeKeybindings({
      bindings: { search: "mod+j" },
      defaults: DEFAULT_KEYBINDINGS,
      conflicts: [],
    });
    const custom = getCustomKeybindingItems(bindings, DEFAULT_KEYBINDINGS);
    expect(custom).toEqual([{ action: "search", label: "Search" }]);

    const sections = buildShortcutSectionsForDisplay(bindings, DEFAULT_KEYBINDINGS);
    expect(sections.some((s) => s.section === "Custom")).toBe(true);
    expect(sections.find((s) => s.section === "Custom")?.items).toEqual(custom);
  });

  it("omits Custom section when bindings match defaults", () => {
    const bindings = mergeKeybindings(null);
    const sections = buildShortcutSectionsForDisplay(bindings, DEFAULT_KEYBINDINGS);
    expect(sections.some((s) => s.section === "Custom")).toBe(false);
  });
});

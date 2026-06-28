import { describe, expect, it } from "vitest";
import {
  DEFAULT_KEYBINDINGS,
  buildChordIndex,
  buildShortcutDisplaySections,
  eventMatchesChord,
  formatChordDisplay,
  formatChordParts,
  isBareQuestionMark,
  isTypingTarget,
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

describe("formatChordParts", () => {
  it("returns separate kbd parts for mac and non-mac", () => {
    expect(formatChordParts("mod+shift+b", true)).toEqual(["⌘", "⇧", "B"]);
    expect(formatChordParts("mod+k", false)).toEqual(["Ctrl", "K"]);
  });
});

describe("isBareQuestionMark", () => {
  it("detects unmodified question mark", () => {
    const bare = { key: "?", metaKey: false, ctrlKey: false, shiftKey: false, altKey: false } as KeyboardEvent;
    const shifted = { key: "?", metaKey: false, ctrlKey: false, shiftKey: true, altKey: false } as KeyboardEvent;
    expect(isBareQuestionMark(bare)).toBe(true);
    expect(isBareQuestionMark(shifted)).toBe(false);
  });
});

describe("isTypingTarget", () => {
  it("detects inputs and editor surfaces", () => {
    const input = { tagName: "INPUT", isContentEditable: false, closest: () => null } as unknown as EventTarget;
    expect(isTypingTarget(input)).toBe(true);
  });
});

describe("buildShortcutDisplaySections", () => {
  it("includes a custom section for overridden bindings", () => {
    const bindings = mergeKeybindings({
      bindings: { search: "mod+j" },
      defaults: DEFAULT_KEYBINDINGS,
      conflicts: [],
    });
    const sections = buildShortcutDisplaySections(bindings, DEFAULT_KEYBINDINGS, false);
    expect(sections.some((s) => s.name === "Navigation")).toBe(true);
    expect(sections.find((s) => s.name === "Custom")?.items).toEqual([
      expect.objectContaining({ action: "search", custom: true }),
    ]);
  });
});

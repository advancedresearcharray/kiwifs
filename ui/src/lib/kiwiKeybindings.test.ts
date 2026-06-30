import { describe, expect, it } from "vitest";
import {
  DEFAULT_KEYBINDINGS,
  buildChordIndex,
  eventMatchesChord,
  formatChordDisplay,
  isKeyboardShortcutTargetIgnored,
  matchBoundAction,
  mergeKeybindings,
  normalizeChord,
  shouldTriggerBareShortcutsHelp,
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

describe("isKeyboardShortcutTargetIgnored", () => {
  function mockTarget(match: boolean): EventTarget {
    return {
      closest: () => (match ? {} : null),
    } as unknown as EventTarget;
  }

  it("ignores native text inputs", () => {
    expect(isKeyboardShortcutTargetIgnored(mockTarget(true))).toBe(true);
  });

  it("ignores CodeMirror editor surfaces", () => {
    expect(isKeyboardShortcutTargetIgnored(mockTarget(true))).toBe(true);
  });

  it("allows shortcuts from non-editable targets", () => {
    expect(isKeyboardShortcutTargetIgnored(mockTarget(false))).toBe(false);
  });

  it("allows shortcuts when target lacks closest", () => {
    expect(isKeyboardShortcutTargetIgnored({} as EventTarget)).toBe(false);
  });
});

describe("shouldTriggerBareShortcutsHelp", () => {
  function mockTarget(ignored: boolean): EventTarget {
    return {
      closest: () => (ignored ? {} : null),
    } as unknown as EventTarget;
  }

  it("opens on bare question mark outside inputs", () => {
    const e = {
      key: "?",
      shiftKey: true,
      metaKey: false,
      ctrlKey: false,
      altKey: false,
      target: mockTarget(false),
    } as KeyboardEvent;
    expect(shouldTriggerBareShortcutsHelp(e)).toBe(true);
  });

  it("does not open when typing in an input", () => {
    const e = {
      key: "?",
      shiftKey: true,
      metaKey: false,
      ctrlKey: false,
      altKey: false,
      target: mockTarget(true),
    } as KeyboardEvent;
    expect(shouldTriggerBareShortcutsHelp(e)).toBe(false);
  });

  it("does not open when mod is held (handled by shortcuts_help binding)", () => {
    const e = {
      key: "?",
      shiftKey: true,
      metaKey: true,
      ctrlKey: false,
      altKey: false,
      target: mockTarget(false),
    } as KeyboardEvent;
    expect(shouldTriggerBareShortcutsHelp(e)).toBe(false);
  });
});

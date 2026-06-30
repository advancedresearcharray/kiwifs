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
  normalizeKeyPart,
  shouldTriggerBareShortcutsHelp,
} from "./kiwiKeybindings";

describe("normalizeKeyPart", () => {
  it("canonicalizes single- and multi-character key aliases uniformly", () => {
    expect(normalizeKeyPart("Escape")).toBe("escape");
    expect(normalizeKeyPart("Backslash")).toBe("\\");
    expect(normalizeKeyPart("/")).toBe("/");
    expect(normalizeKeyPart("Enter")).toBe("enter");
    expect(normalizeKeyPart("Return")).toBe("enter");
    expect(normalizeKeyPart("Tab")).toBe("tab");
    expect(normalizeKeyPart("F1")).toBe("f1");
    expect(normalizeKeyPart("k")).toBe("k");
  });
});

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

  it("matches backslash split-view shortcut", () => {
    const e = {
      key: "\\",
      ctrlKey: true,
      metaKey: false,
      shiftKey: false,
      altKey: false,
    } as KeyboardEvent;
    expect(eventMatchesChord(e, "mod+\\")).toBe(true);
    expect(normalizeChord("Mod+\\")).toBe("mod+\\");
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
    expect(merged.toggle_split_view).toBe("mod+\\");
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

  it("matches toggle_split_view via mod+\\ without affecting other bindings", () => {
    const bindings = mergeKeybindings(null);
    const splitToggle = {
      key: "\\",
      ctrlKey: true,
      metaKey: false,
      shiftKey: false,
      altKey: false,
    } as KeyboardEvent;
    const unrelated = {
      key: "s",
      ctrlKey: true,
      metaKey: false,
      shiftKey: false,
      altKey: false,
    } as KeyboardEvent;
    expect(matchBoundAction(splitToggle, bindings)).toBe("toggle_split_view");
    expect(matchBoundAction(unrelated, bindings)).toBe("save");
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

  it("does not open when alt is held", () => {
    const e = {
      key: "?",
      shiftKey: true,
      metaKey: false,
      ctrlKey: false,
      altKey: true,
      target: mockTarget(false),
    } as KeyboardEvent;
    expect(shouldTriggerBareShortcutsHelp(e)).toBe(false);
  });

  it("opens on shift+/ without question key code", () => {
    const e = {
      key: "/",
      shiftKey: true,
      metaKey: false,
      ctrlKey: false,
      altKey: false,
      target: mockTarget(false),
    } as KeyboardEvent;
    expect(shouldTriggerBareShortcutsHelp(e)).toBe(true);
  });

  it("does not open inside cmdk filter input", () => {
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
});

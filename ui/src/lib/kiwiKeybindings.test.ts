import { describe, expect, it } from "vitest";
import {
  DEFAULT_KEYBINDINGS,
  buildChordIndex,
  eventMatchesChord,
  formatChordDisplay,
  isMacPlatform,
  isShortcutBlockedTarget,
  isStandaloneHelpKey,
  isTypingTarget,
  matchBoundAction,
  mergeKeybindings,
  normalizeChord,
  shouldDispatchKeybinding,
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

describe("isTypingTarget", () => {
  it("detects input elements", () => {
    const input = { tagName: "INPUT", isContentEditable: false } as HTMLElement;
    const button = { tagName: "BUTTON", isContentEditable: false } as HTMLElement;
    expect(isTypingTarget(input)).toBe(true);
    expect(isTypingTarget(button)).toBe(false);
  });
});

describe("isShortcutBlockedTarget", () => {
  it("blocks CodeMirror and ProseMirror editor surfaces", () => {
    const cmChild = {
      tagName: "DIV",
      isContentEditable: false,
      closest: (sel: string) => (sel.includes("cm-editor") ? {} : null),
    } as unknown as HTMLElement;
    const proseChild = {
      tagName: "P",
      isContentEditable: false,
      closest: (sel: string) => (sel.includes("ProseMirror") ? {} : null),
    } as unknown as HTMLElement;
    expect(isShortcutBlockedTarget(cmChild)).toBe(true);
    expect(isShortcutBlockedTarget(proseChild)).toBe(true);
  });
});

describe("isStandaloneHelpKey", () => {
  it("matches bare question mark without modifiers", () => {
    const help = {
      key: "?",
      metaKey: false,
      ctrlKey: false,
      altKey: false,
      shiftKey: true,
    } as KeyboardEvent;
    const modHelp = { ...help, metaKey: true } as KeyboardEvent;
    expect(isStandaloneHelpKey(help)).toBe(true);
    expect(isStandaloneHelpKey(modHelp)).toBe(false);
  });
});

describe("shouldDispatchKeybinding", () => {
  it("allows help and close overlay while typing", () => {
    const input = { tagName: "INPUT", isContentEditable: false } as HTMLElement;
    expect(shouldDispatchKeybinding("shortcuts_help", input, { editing: true, splitEditing: false })).toBe(true);
    expect(shouldDispatchKeybinding("close_overlay", input, { editing: true, splitEditing: false })).toBe(true);
  });

  it("blocks navigation shortcuts in inputs and while editing", () => {
    const input = { tagName: "INPUT", isContentEditable: false } as HTMLElement;
    const body = { tagName: "BODY", isContentEditable: false, closest: () => null } as unknown as HTMLElement;
    expect(shouldDispatchKeybinding("search", input, { editing: false, splitEditing: false })).toBe(false);
    expect(shouldDispatchKeybinding("search", body, { editing: true, splitEditing: false })).toBe(false);
    expect(shouldDispatchKeybinding("search", body, { editing: false, splitEditing: false })).toBe(true);
  });
});

describe("isMacPlatform", () => {
  it("detects macOS from platform string", () => {
    const original = navigator.platform;
    Object.defineProperty(navigator, "platform", { value: "MacIntel", configurable: true });
    expect(isMacPlatform()).toBe(true);
    Object.defineProperty(navigator, "platform", { value: original, configurable: true });
  });
});

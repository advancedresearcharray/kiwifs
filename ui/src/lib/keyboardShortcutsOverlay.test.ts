import { describe, expect, it } from "vitest";
import {
  DEFAULT_KEYBINDINGS,
  SHORTCUT_SECTIONS,
  eventMatchesChord,
  type KeybindingAction,
} from "./kiwiKeybindings";
import {
  allDefaultBindingsDisplayable,
  buildShortcutRows,
  formatConflictSummary,
  resolveShortcutsOverlayKey,
  sanitizeOverlayText,
  shortcutSearchValue,
} from "./keyboardShortcutsOverlay";

function mockTarget(ignored: boolean): EventTarget {
  return {
    closest: () => (ignored ? {} : null),
  } as unknown as EventTarget;
}

function keyEvent(
  overrides: Partial<KeyboardEvent> & { key: string },
): KeyboardEvent {
  return {
    shiftKey: false,
    metaKey: false,
    ctrlKey: false,
    altKey: false,
    target: mockTarget(false),
    ...overrides,
  } as KeyboardEvent;
}

describe("sanitizeOverlayText", () => {
  it("strips control characters", () => {
    expect(sanitizeOverlayText("hello\x00world")).toBe("helloworld");
    expect(sanitizeOverlayText("tab\there")).toBe("tabhere");
  });

  it("caps length", () => {
    expect(sanitizeOverlayText("a".repeat(300), 50)).toHaveLength(50);
  });
});

describe("shortcutSearchValue", () => {
  it("includes section, label, and chord for fuzzy filter", () => {
    const value = shortcutSearchValue("Navigation", "Search", "mod+k");
    expect(value).toContain("navigation");
    expect(value).toContain("search");
    expect(value).toContain("mod+k");
  });

  it("sanitizes injected control chars in labels", () => {
    const value = shortcutSearchValue("Nav", "Save\x07", "mod+s");
    expect(value).not.toContain("\x07");
    expect(value).toContain("save");
  });
});

describe("buildShortcutRows", () => {
  it("includes every shortcut from SHORTCUT_SECTIONS", () => {
    const rows = buildShortcutRows(DEFAULT_KEYBINDINGS);
    const expectedCount = SHORTCUT_SECTIONS.reduce((n, s) => n + s.items.length, 0);
    expect(rows).toHaveLength(expectedCount);
  });

  it("groups rows by Navigation, Views, and Editor sections", () => {
    const rows = buildShortcutRows(DEFAULT_KEYBINDINGS);
    const sections = [...new Set(rows.map((r) => r.section))];
    expect(sections).toEqual(["Navigation", "Views", "Editor"]);
  });

  it("uses custom bindings for chord display", () => {
    const custom = { ...DEFAULT_KEYBINDINGS, search: "mod+j" };
    const searchRow = buildShortcutRows(custom).find((r) => r.action === "search");
    expect(searchRow?.chordDisplay.toLowerCase()).toMatch(/j/);
  });

  it("does not expose raw HTML in labels", () => {
    const rows = buildShortcutRows(DEFAULT_KEYBINDINGS);
    for (const row of rows) {
      expect(row.label).not.toMatch(/[<>]/);
      expect(row.searchValue).not.toMatch(/[\x00-\x1f\x7f]/);
    }
  });
});

describe("formatConflictSummary", () => {
  it("formats conflicts with sanitized action names", () => {
    const summary = formatConflictSummary([
      { chord: "mod+k", actions: ["search", "new_page"] },
    ]);
    expect(summary).toContain("search");
    expect(summary).toContain("new_page");
  });

  it("strips control chars from action names", () => {
    const summary = formatConflictSummary([
      { chord: "mod+k", actions: ["search\x01"] },
    ]);
    expect(summary).not.toContain("\x01");
  });
});

describe("resolveShortcutsOverlayKey", () => {
  it("toggles on bare ? outside inputs", () => {
    expect(
      resolveShortcutsOverlayKey(
        keyEvent({ key: "?", shiftKey: true }),
        DEFAULT_KEYBINDINGS,
      ),
    ).toBe("toggle");
  });

  it("toggles on shift+/ outside inputs", () => {
    expect(
      resolveShortcutsOverlayKey(
        keyEvent({ key: "/", shiftKey: true }),
        DEFAULT_KEYBINDINGS,
      ),
    ).toBe("toggle");
  });

  it("does not toggle bare ? inside inputs", () => {
    expect(
      resolveShortcutsOverlayKey(
        keyEvent({ key: "?", shiftKey: true, target: mockTarget(true) }),
        DEFAULT_KEYBINDINGS,
      ),
    ).toBeNull();
  });

  it("toggles on mod+/ even inside inputs", () => {
    expect(
      resolveShortcutsOverlayKey(
        keyEvent({
          key: "/",
          ctrlKey: true,
          target: mockTarget(true),
        }),
        DEFAULT_KEYBINDINGS,
      ),
    ).toBe("toggle");
  });

  it("does not toggle on unrelated keys", () => {
    expect(
      resolveShortcutsOverlayKey(
        keyEvent({ key: "k", ctrlKey: true }),
        DEFAULT_KEYBINDINGS,
      ),
    ).toBeNull();
  });

  it("does not toggle when alt is held with bare ?", () => {
    expect(
      resolveShortcutsOverlayKey(
        keyEvent({ key: "?", shiftKey: true, altKey: true }),
        DEFAULT_KEYBINDINGS,
      ),
    ).toBeNull();
  });

  it("does not toggle bare ? while filtering inside the overlay cmdk input", () => {
    expect(
      resolveShortcutsOverlayKey(
        keyEvent({ key: "?", shiftKey: true, target: mockTarget(true) }),
        DEFAULT_KEYBINDINGS,
      ),
    ).toBeNull();
  });

  it("toggles on custom shortcuts_help binding", () => {
    const custom = { ...DEFAULT_KEYBINDINGS, shortcuts_help: "mod+h" };
    expect(
      resolveShortcutsOverlayKey(
        keyEvent({ key: "h", ctrlKey: true }),
        custom,
      ),
    ).toBe("toggle");
  });
});

describe("default keybinding chords", () => {
  const bindingEvents: [KeybindingAction, Partial<KeyboardEvent>][] = [
    ["search", { key: "k", ctrlKey: true }],
    ["new_page", { key: "n", ctrlKey: true }],
    ["toggle_editor", { key: "e", ctrlKey: true }],
    ["save", { key: "s", ctrlKey: true }],
    ["toggle_sidebar", { key: "b", ctrlKey: true }],
    ["graph", { key: "g", ctrlKey: true }],
    ["toggle_bases", { key: "b", ctrlKey: true, shiftKey: true }],
    ["toggle_timeline", { key: "t", ctrlKey: true, shiftKey: true }],
    ["toggle_kanban", { key: "w", ctrlKey: true, shiftKey: true }],
    ["toggle_mode", { key: "e", ctrlKey: true, shiftKey: true }],
    ["shortcuts_help", { key: "/", ctrlKey: true }],
    ["undo", { key: "z", ctrlKey: true }],
    ["focus_tree_filter", { key: "f", ctrlKey: true, altKey: true }],
    ["close_overlay", { key: "Escape" }],
  ];

  it.each(bindingEvents)("matches default chord for %s", (action, partial) => {
    const chord = DEFAULT_KEYBINDINGS[action];
    const e = keyEvent({ key: partial.key ?? "", ...partial });
    expect(eventMatchesChord(e, chord)).toBe(true);
  });

  it("all default bindings are displayable", () => {
    expect(allDefaultBindingsDisplayable()).toBe(true);
  });
});

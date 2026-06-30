/**
 * Central keyboard shortcut manager for KiwiFS.
 *
 * Bindings are loaded from GET /api/kiwi/keybindings (defaults merged with
 * .kiwi/keybindings.json and [ui.keybindings] in config.toml).
 */

export type KeybindingAction =
  | "search"
  | "new_page"
  | "toggle_editor"
  | "save"
  | "toggle_sidebar"
  | "graph"
  | "toggle_bases"
  | "toggle_timeline"
  | "toggle_kanban"
  | "toggle_mode"
  | "shortcuts_help"
  | "undo"
  | "focus_tree_filter"
  | "close_overlay";

export type ParsedChord = {
  mod: boolean;
  shift: boolean;
  alt: boolean;
  key: string;
};

export type KeybindingConflict = {
  chord: string;
  actions: string[];
};

export type KeybindingsConfig = {
  bindings: Partial<Record<KeybindingAction, string>>;
  defaults: Partial<Record<KeybindingAction, string>>;
  conflicts: KeybindingConflict[];
};

export const DEFAULT_KEYBINDINGS: Record<KeybindingAction, string> = {
  search: "mod+k",
  new_page: "mod+n",
  toggle_editor: "mod+e",
  save: "mod+s",
  toggle_sidebar: "mod+b",
  graph: "mod+g",
  toggle_bases: "mod+shift+b",
  toggle_timeline: "mod+shift+t",
  toggle_kanban: "mod+shift+w",
  toggle_mode: "mod+shift+e",
  shortcuts_help: "mod+/",
  undo: "mod+z",
  focus_tree_filter: "mod+alt+f",
  close_overlay: "escape",
};

export function normalizeChord(chord: string): string {
  const parts = chord.trim().split("+").map((p) => p.trim().toLowerCase()).filter(Boolean);
  const mods: string[] = [];
  let key = "";
  for (const part of parts) {
    switch (part) {
      case "ctrl":
      case "control":
        if (!mods.includes("mod")) mods.push("mod");
        break;
      case "cmd":
      case "command":
      case "meta":
      case "mod":
        if (!mods.includes("mod")) mods.push("mod");
        break;
      case "shift":
        if (!mods.includes("shift")) mods.push("shift");
        break;
      case "alt":
      case "option":
        if (!mods.includes("alt")) mods.push("alt");
        break;
      case "esc":
      case "escape":
        key = "escape";
        break;
      case "slash":
        key = "/";
        break;
      case "question":
        key = "?";
        break;
      default:
        key = part.length === 1 ? part : part;
    }
  }
  mods.sort();
  if (!key) throw new Error(`invalid chord: ${chord}`);
  return [...mods, key].join("+");
}

export function parseChord(chord: string): ParsedChord {
  const normalized = normalizeChord(chord);
  const parts = normalized.split("+");
  const key = parts[parts.length - 1] ?? "";
  return {
    mod: parts.includes("mod"),
    shift: parts.includes("shift"),
    alt: parts.includes("alt"),
    key,
  };
}

export function eventMatchesChord(e: KeyboardEvent, chord: string): boolean {
  const parsed = parseChord(chord);
  const mod = e.metaKey || e.ctrlKey;
  if (parsed.mod !== mod) return false;
  if (parsed.alt !== e.altKey) return false;

  const eventKey = e.key.length === 1 ? e.key.toLowerCase() : e.key.toLowerCase();
  const isHelpSlash = parsed.key === "/" && parsed.mod && !parsed.shift && !parsed.alt;

  if (!isHelpSlash && parsed.shift !== e.shiftKey) return false;

  if (parsed.key === "escape") return eventKey === "escape";
  if (isHelpSlash) {
    return eventKey === "/" || eventKey === "slash" || eventKey === "?";
  }
  if (parsed.key === "?") return eventKey === "?" || (e.shiftKey && eventKey === "/");
  return eventKey === parsed.key;
}

export function formatChordSegments(chord: string): string[] {
  const isMac = typeof navigator !== "undefined" && navigator.platform.includes("Mac");
  const parsed = parseChord(chord);
  const parts: string[] = [];
  if (parsed.mod) parts.push(isMac ? "⌘" : "Ctrl");
  if (parsed.shift) parts.push("Shift");
  if (parsed.alt) parts.push(isMac ? "⌥" : "Alt");

  let keyLabel = parsed.key.toUpperCase();
  if (parsed.key === "/") keyLabel = "/";
  if (parsed.key === "?") keyLabel = "?";
  if (parsed.key === "escape") keyLabel = "Esc";
  parts.push(keyLabel);
  return parts;
}

export function formatChordDisplay(chord: string): string {
  const segments = formatChordSegments(chord);
  const isMac = typeof navigator !== "undefined" && navigator.platform.includes("Mac");
  const parsed = parseChord(chord);

  if (isMac && parsed.mod && !parsed.shift && !parsed.alt && segments.length === 2) {
    return `${segments[0]}${segments[1]}`;
  }
  if (segments.length <= 1) return segments[0] ?? "";
  return segments.join("+");
}

/** True when the event target is an editable field (skip global shortcuts). */
export function isTextInputTarget(e: KeyboardEvent): boolean {
  const target = e.target;
  if (!(target instanceof HTMLElement)) return false;
  const tag = target.tagName;
  if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return true;
  if (target.isContentEditable) return true;
  return Boolean(
    target.closest(".cm-editor, [contenteditable='true'], [role='textbox'], [cmdk-input], [cmdk-input-wrapper]"),
  );
}

/** Plain `?` (no ctrl/meta/alt) opens the shortcuts overlay outside editable targets. */
export function shouldOpenShortcutsHelp(
  e: Pick<KeyboardEvent, "key" | "shiftKey" | "ctrlKey" | "metaKey" | "altKey">,
): boolean {
  if (e.ctrlKey || e.metaKey || e.altKey) return false;
  return e.key === "?" || (e.key === "/" && e.shiftKey);
}

export function mergeKeybindings(config: KeybindingsConfig | null | undefined): Record<KeybindingAction, string> {
  const merged = { ...DEFAULT_KEYBINDINGS };
  const source = config?.bindings ?? {};
  for (const [action, chord] of Object.entries(source)) {
    if (!chord || !(action in DEFAULT_KEYBINDINGS)) continue;
    try {
      merged[action as KeybindingAction] = normalizeChord(chord);
    } catch {
      // ignore invalid override; default remains
    }
  }
  return merged;
}

export type ShortcutSection = {
  section: string;
  items: { action: KeybindingAction; label: string }[];
};

export const SHORTCUT_SECTIONS: ShortcutSection[] = [
  {
    section: "Navigation",
    items: [
      { action: "search", label: "Search" },
      { action: "new_page", label: "New page" },
      { action: "toggle_editor", label: "Toggle editor" },
      { action: "toggle_sidebar", label: "Toggle sidebar" },
      { action: "shortcuts_help", label: "Keyboard shortcuts" },
    ],
  },
  {
    section: "Views",
    items: [
      { action: "graph", label: "Knowledge graph" },
      { action: "toggle_bases", label: "Toggle Bases" },
      { action: "toggle_timeline", label: "Toggle Timeline" },
      { action: "toggle_kanban", label: "Toggle Kanban" },
    ],
  },
  {
    section: "Editor",
    items: [
      { action: "save", label: "Save (also auto-saves after 2s)" },
      { action: "toggle_mode", label: "Toggle Visual / Source (while editing)" },
      { action: "focus_tree_filter", label: "Focus tree filter" },
      { action: "undo", label: "Undo last file operation" },
      { action: "close_overlay", label: "Close overlay / cancel" },
    ],
  },
];

const ACTION_LABELS: Record<KeybindingAction, string> = Object.fromEntries(
  SHORTCUT_SECTIONS.flatMap((section) =>
    section.items.map((item) => [item.action, item.label] as const),
  ),
) as Record<KeybindingAction, string>;

export function actionLabel(action: KeybindingAction): string {
  return ACTION_LABELS[action] ?? action.replace(/_/g, " ");
}

export function getCustomShortcutItems(
  bindings: Record<KeybindingAction, string>,
  defaults: Partial<Record<KeybindingAction, string>>,
): ShortcutSection["items"] {
  const items: ShortcutSection["items"] = [];
  for (const action of Object.keys(DEFAULT_KEYBINDINGS) as KeybindingAction[]) {
    const chord = bindings[action];
    const def = defaults[action] ?? DEFAULT_KEYBINDINGS[action];
    if (!chord || normalizeChord(def) === normalizeChord(chord)) continue;
    items.push({ action, label: actionLabel(action) });
  }
  return items.sort((a, b) => a.label.localeCompare(b.label));
}

export function buildChordIndex(bindings: Record<KeybindingAction, string>): Map<string, KeybindingAction[]> {
  const index = new Map<string, KeybindingAction[]>();
  for (const [action, chord] of Object.entries(bindings) as [KeybindingAction, string][]) {
    const list = index.get(chord) ?? [];
    list.push(action);
    index.set(chord, list);
  }
  return index;
}

export function matchBoundAction(
  e: KeyboardEvent,
  bindings: Record<KeybindingAction, string>,
): KeybindingAction | null {
  for (const [action, chord] of Object.entries(bindings) as [KeybindingAction, string][]) {
    if (eventMatchesChord(e, chord)) return action;
  }
  return null;
}

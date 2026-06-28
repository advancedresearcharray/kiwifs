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

export function isMacPlatform(): boolean {
  if (typeof navigator === "undefined") return false;
  const platform = (navigator as Navigator & { userAgentData?: { platform?: string } }).userAgentData?.platform;
  if (platform) return /mac/i.test(platform);
  return /Mac|iPhone|iPad|iPod/.test(navigator.platform);
}

export function formatChordParts(chord: string, mac = isMacPlatform()): string[] {
  const parsed = parseChord(chord);
  const parts: string[] = [];
  if (parsed.mod) parts.push(mac ? "⌘" : "Ctrl");
  if (parsed.alt) parts.push(mac ? "⌥" : "Alt");
  if (parsed.shift) parts.push(mac ? "⇧" : "Shift");

  switch (parsed.key) {
    case "escape":
      parts.push("Esc");
      break;
    case "/":
      parts.push("/");
      break;
    case "?":
      parts.push("?");
      break;
    default:
      parts.push(parsed.key.length === 1 ? parsed.key.toUpperCase() : parsed.key);
  }

  return parts;
}

export function formatChordDisplay(chord: string): string {
  const isMac = isMacPlatform();
  const parsed = parseChord(chord);
  const parts: string[] = [];
  if (parsed.mod) parts.push(isMac ? "⌘" : "Ctrl");
  if (parsed.shift) parts.push("Shift");
  if (parsed.alt) parts.push(isMac ? "⌥" : "Alt");

  let keyLabel = parsed.key.toUpperCase();
  if (parsed.key === "/") keyLabel = "/";
  if (parsed.key === "?") keyLabel = "?";
  if (parsed.key === "escape") keyLabel = "Esc";

  if (isMac && parsed.mod && !parsed.shift && !parsed.alt) {
    return `${parts[0]}${keyLabel}`;
  }
  if (parts.length === 0) return keyLabel;
  return `${parts.join("+")}+${keyLabel}`;
}

export function isTypingTarget(target: EventTarget | null): boolean {
  if (!target || typeof target !== "object" || !("tagName" in target)) return false;
  const el = target as HTMLElement;
  const tag = el.tagName;
  if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return true;
  if (el.isContentEditable) return true;
  if (el.closest(".cm-editor, [contenteditable='true'], [role='textbox']")) return true;
  return false;
}

export function isBareQuestionMark(event: KeyboardEvent): boolean {
  return (
    event.key === "?" &&
    !event.metaKey &&
    !event.ctrlKey &&
    !event.altKey &&
    !event.shiftKey
  );
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

export type ShortcutDisplayItem = {
  action: KeybindingAction;
  label: string;
  chord: string;
  keys: string[];
  custom: boolean;
};

export type ShortcutDisplaySection = {
  name: string;
  items: ShortcutDisplayItem[];
};

const ACTION_LABELS: Record<KeybindingAction, string> = Object.fromEntries(
  SHORTCUT_SECTIONS.flatMap((section) => section.items.map((item) => [item.action, item.label])),
) as Record<KeybindingAction, string>;

export function buildShortcutDisplaySections(
  bindings: Record<KeybindingAction, string>,
  defaults: Partial<Record<KeybindingAction, string>> = DEFAULT_KEYBINDINGS,
  mac = isMacPlatform(),
): ShortcutDisplaySection[] {
  const sections: ShortcutDisplaySection[] = SHORTCUT_SECTIONS.map(({ section, items }) => ({
    name: section,
    items: items.map(({ action, label }) => ({
      action,
      label,
      chord: bindings[action],
      keys: formatChordParts(bindings[action], mac),
      custom: false,
    })),
  }));

  const customItems = (Object.keys(DEFAULT_KEYBINDINGS) as KeybindingAction[])
    .filter((action) => {
      const def = defaults[action] ?? DEFAULT_KEYBINDINGS[action];
      try {
        return normalizeChord(bindings[action]) !== normalizeChord(def);
      } catch {
        return false;
      }
    })
    .map((action) => ({
      action,
      label: ACTION_LABELS[action] ?? action,
      chord: bindings[action],
      keys: formatChordParts(bindings[action], mac),
      custom: true,
    }));

  if (customItems.length > 0) {
    sections.push({ name: "Custom", items: customItems });
  }

  return sections;
}

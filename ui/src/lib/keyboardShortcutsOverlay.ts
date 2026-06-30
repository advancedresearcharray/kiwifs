import {
  DEFAULT_KEYBINDINGS,
  formatChordDisplay,
  matchBoundAction,
  SHORTCUT_SECTIONS,
  shouldTriggerBareShortcutsHelp,
  type KeybindingAction,
} from "./kiwiKeybindings";

/** Strip control characters and cap length for safe overlay display. */
export function sanitizeOverlayText(text: string, maxLen = 200): string {
  return text.replace(/[\x00-\x1f\x7f]/g, "").slice(0, maxLen);
}

/** Build cmdk search value from section, label, and binding chord. */
export function shortcutSearchValue(section: string, label: string, chord: string): string {
  const safeLabel = sanitizeOverlayText(`${section} ${label}`);
  return `${safeLabel} ${formatChordDisplay(chord)} ${chord}`.toLowerCase();
}

export type ShortcutRow = {
  action: KeybindingAction;
  section: string;
  label: string;
  chordDisplay: string;
  searchValue: string;
};

/** Flatten grouped shortcut sections into display rows using live bindings. */
export function buildShortcutRows(bindings: Record<KeybindingAction, string>): ShortcutRow[] {
  const rows: ShortcutRow[] = [];
  for (const section of SHORTCUT_SECTIONS) {
    for (const item of section.items) {
      const chord = bindings[item.action];
      rows.push({
        action: item.action,
        section: section.section,
        label: sanitizeOverlayText(item.label),
        chordDisplay: formatChordDisplay(chord),
        searchValue: shortcutSearchValue(section.section, item.label, chord),
      });
    }
  }
  return rows;
}

export type BindingConflict = { chord: string; actions: string[] };

/** Human-readable conflict summary for the overlay banner (display-only). */
export function formatConflictSummary(conflicts: BindingConflict[]): string {
  return conflicts
    .map((c) => {
      const actions = c.actions.map((a) => sanitizeOverlayText(a)).join(" / ");
      return `${actions} (${formatChordDisplay(c.chord)})`;
    })
    .join("; ");
}

/**
 * Returns "toggle" when the user pressed a key that opens/closes the shortcuts overlay.
 * Bare "?" is suppressed inside editable targets; mod+/ (shortcuts_help) works globally.
 */
export function resolveShortcutsOverlayKey(
  e: KeyboardEvent,
  bindings: Record<KeybindingAction, string>,
): "toggle" | null {
  if (shouldTriggerBareShortcutsHelp(e)) return "toggle";
  if (matchBoundAction(e, bindings) === "shortcuts_help") return "toggle";
  return null;
}

/** Every default binding chord round-trips through formatChordDisplay without throwing. */
export function allDefaultBindingsDisplayable(): boolean {
  for (const chord of Object.values(DEFAULT_KEYBINDINGS)) {
    formatChordDisplay(chord);
  }
  return true;
}

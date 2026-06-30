import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@kw/components/ui/command";
import {
  buildShortcutRows,
  formatConflictSummary,
  type BindingConflict,
} from "../lib/keyboardShortcutsOverlay";
import { SHORTCUT_SECTIONS, type KeybindingAction } from "../lib/kiwiKeybindings";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  bindings: Record<KeybindingAction, string>;
  conflicts?: BindingConflict[];
};

export function KeyboardShortcuts({ open, onOpenChange, bindings, conflicts = [] }: Props) {
  const rows = buildShortcutRows(bindings);
  const rowsBySection = new Map<string, typeof rows>();
  for (const row of rows) {
    const list = rowsBySection.get(row.section) ?? [];
    list.push(row);
    rowsBySection.set(row.section, list);
  }

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Keyboard shortcuts"
      contentClassName="sm:max-w-md"
    >
      <CommandInput placeholder="Filter shortcuts…" />
      {conflicts.length > 0 && (
        <div className="rounded-none border-b border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
          Conflicting bindings detected: {formatConflictSummary(conflicts)}
        </div>
      )}
      <CommandList>
        <CommandEmpty>No shortcuts found.</CommandEmpty>
        {SHORTCUT_SECTIONS.map((s) => (
          <CommandGroup key={s.section} heading={s.section}>
            {(rowsBySection.get(s.section) ?? []).map((row) => (
              <CommandItem
                key={row.action}
                value={row.searchValue}
                onSelect={() => {}}
                className="flex items-center justify-between aria-selected:bg-accent"
              >
                <span>{row.label}</span>
                <kbd className="ml-4 shrink-0 px-2 py-0.5 rounded border border-border bg-muted font-mono text-xs text-muted-foreground">
                  {row.chordDisplay}
                </kbd>
              </CommandItem>
            ))}
          </CommandGroup>
        ))}
      </CommandList>
    </CommandDialog>
  );
}

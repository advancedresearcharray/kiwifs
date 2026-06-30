import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@kw/components/ui/command";
import {
  formatChordDisplay,
  SHORTCUT_SECTIONS,
  type KeybindingAction,
} from "../lib/kiwiKeybindings";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  bindings: Record<KeybindingAction, string>;
  conflicts?: { chord: string; actions: string[] }[];
};

function shortcutSearchValue(section: string, label: string, chord: string): string {
  return `${section} ${label} ${formatChordDisplay(chord)} ${chord}`.toLowerCase();
}

export function KeyboardShortcuts({ open, onOpenChange, bindings, conflicts = [] }: Props) {
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
          Conflicting bindings detected:{" "}
          {conflicts.map((c) => `${c.actions.join(" / ")} (${formatChordDisplay(c.chord)})`).join("; ")}
        </div>
      )}
      <CommandList>
        <CommandEmpty>No shortcuts found.</CommandEmpty>
        {SHORTCUT_SECTIONS.map((s) => (
          <CommandGroup key={s.section} heading={s.section}>
            {s.items.map((item) => (
              <CommandItem
                key={item.action}
                value={shortcutSearchValue(s.section, item.label, bindings[item.action])}
                onSelect={() => {}}
                className="flex items-center justify-between aria-selected:bg-accent"
              >
                <span>{item.label}</span>
                <kbd className="ml-4 shrink-0 px-2 py-0.5 rounded border border-border bg-muted font-mono text-xs text-muted-foreground">
                  {formatChordDisplay(bindings[item.action])}
                </kbd>
              </CommandItem>
            ))}
          </CommandGroup>
        ))}
      </CommandList>
    </CommandDialog>
  );
}

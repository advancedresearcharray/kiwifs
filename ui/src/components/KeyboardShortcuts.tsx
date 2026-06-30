import { useMemo } from "react";
import { Keyboard } from "lucide-react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandShortcut,
} from "@kw/components/ui/command";
import {
  buildShortcutSectionsForDisplay,
  DEFAULT_KEYBINDINGS,
  formatChordDisplay,
  type KeybindingAction,
} from "../lib/kiwiKeybindings";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  bindings: Record<KeybindingAction, string>;
  defaults?: Partial<Record<KeybindingAction, string>>;
  conflicts?: { chord: string; actions: string[] }[];
};

export function KeyboardShortcuts({
  open,
  onOpenChange,
  bindings,
  defaults,
  conflicts = [],
}: Props) {
  const resolvedDefaults = { ...DEFAULT_KEYBINDINGS, ...defaults };
  const sections = useMemo(
    () => buildShortcutSectionsForDisplay(bindings, resolvedDefaults),
    [bindings, resolvedDefaults],
  );

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Keyboard shortcuts"
      contentClassName="sm:max-w-md"
    >
      <CommandInput placeholder="Filter shortcuts…" />
      {conflicts.length > 0 && (
        <div className="border-b border-border px-3 py-2 text-xs text-destructive">
          Conflicting bindings:{" "}
          {conflicts
            .map((c) => `${c.actions.join(" / ")} (${formatChordDisplay(c.chord)})`)
            .join("; ")}
        </div>
      )}
      <CommandList>
        <CommandEmpty>No shortcuts found.</CommandEmpty>
        {sections.map((s) => (
          <CommandGroup key={s.section} heading={s.section}>
            {s.items.map((item) => {
              const chord = formatChordDisplay(bindings[item.action]);
              return (
                <CommandItem
                  key={`${s.section}-${item.action}`}
                  value={`${item.label} ${chord} ${s.section}`}
                  onSelect={() => onOpenChange(false)}
                >
                  <Keyboard className="h-4 w-4 shrink-0 opacity-50" />
                  <span>{item.label}</span>
                  <CommandShortcut>
                    <kbd className="px-2 py-0.5 rounded border border-border bg-muted font-mono text-xs text-muted-foreground">
                      {chord}
                    </kbd>
                  </CommandShortcut>
                </CommandItem>
              );
            })}
          </CommandGroup>
        ))}
      </CommandList>
    </CommandDialog>
  );
}

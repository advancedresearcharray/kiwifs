import { Keyboard } from "lucide-react";
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
  isCustomBinding,
  SHORTCUT_SECTIONS,
  type KeybindingAction,
} from "../lib/kiwiKeybindings";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  bindings: Record<KeybindingAction, string>;
  defaults: Record<KeybindingAction, string>;
  conflicts?: { chord: string; actions: string[] }[];
};

export function KeyboardShortcuts({
  open,
  onOpenChange,
  bindings,
  defaults,
  conflicts = [],
}: Props) {
  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Keyboard shortcuts"
      contentClassName="sm:max-w-md"
    >
      <div className="flex items-center gap-2 border-b border-border px-3 py-2">
        <Keyboard className="h-4 w-4 shrink-0 text-muted-foreground" />
        <span className="text-sm font-medium">Keyboard shortcuts</span>
      </div>
      {conflicts.length > 0 && (
        <div className="mx-3 mt-2 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
          Conflicting bindings detected:{" "}
          {conflicts
            .map((c) => `${c.actions.join(" / ")} (${formatChordDisplay(c.chord)})`)
            .join("; ")}
        </div>
      )}
      <CommandInput placeholder="Filter shortcuts…" />
      <CommandList>
        <CommandEmpty>No shortcuts found.</CommandEmpty>
        {SHORTCUT_SECTIONS.map((s) => (
          <CommandGroup key={s.section} heading={s.section}>
            {s.items.map((item) => (
              <CommandItem
                key={item.action}
                value={`${item.label} ${s.section} ${formatChordDisplay(bindings[item.action])}`}
                className="flex items-center justify-between gap-4 aria-selected:bg-transparent data-[selected=true]:bg-transparent"
              >
                <span>
                  {item.label}
                  {isCustomBinding(item.action, bindings, defaults) && (
                    <span className="ml-1 text-muted-foreground">(custom)</span>
                  )}
                </span>
                <kbd className="shrink-0 rounded border border-border bg-muted px-2 py-0.5 font-mono text-xs text-muted-foreground">
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

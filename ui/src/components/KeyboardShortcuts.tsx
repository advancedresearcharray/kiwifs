import { Keyboard } from "lucide-react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@kw/components/ui/command";
import { useKeybindings } from "../hooks/useKeybindings";
import {
  SHORTCUT_SECTIONS,
  formatChordDisplay,
  formatChordSegments,
  getCustomShortcutItems,
  type KeybindingAction,
} from "../lib/kiwiKeybindings";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

function ShortcutKeys({ action, bindings }: { action: KeybindingAction; bindings: Record<KeybindingAction, string> }) {
  const segments = formatChordSegments(bindings[action]);
  return (
    <span className="ml-auto flex items-center gap-1">
      {segments.map((segment) => (
        <kbd
          key={segment}
          className="px-1.5 py-0.5 rounded border border-border bg-muted font-mono text-[10px] text-muted-foreground"
        >
          {segment}
        </kbd>
      ))}
    </span>
  );
}

export function KeyboardShortcuts({ open, onOpenChange }: Props) {
  const { bindings, defaults, conflicts } = useKeybindings();
  const customItems = getCustomShortcutItems(bindings, defaults);

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Keyboard shortcuts"
      contentClassName="sm:max-w-lg"
      commandProps={{ label: "Keyboard shortcuts" }}
    >
      <div className="flex items-center gap-2 border-b border-border px-3 py-2">
        <Keyboard className="h-4 w-4 shrink-0 text-muted-foreground" />
        <span className="text-sm font-medium">Keyboard shortcuts</span>
      </div>
      {conflicts.length > 0 && (
        <div className="border-b border-border px-3 py-2 text-xs text-destructive">
          Conflicting bindings:{" "}
          {conflicts.map((c) => `${c.actions.join(" / ")} (${formatChordDisplay(c.chord)})`).join("; ")}
        </div>
      )}
      <CommandInput placeholder="Filter shortcuts…" />
      <CommandList>
        <CommandEmpty>No matching shortcuts.</CommandEmpty>
        {SHORTCUT_SECTIONS.map((section) => (
          <CommandGroup key={section.section} heading={section.section}>
            {section.items.map((item) => (
              <CommandItem
                key={item.action}
                value={`${section.section} ${item.label} ${formatChordSegments(bindings[item.action]).join(" ")}`}
                className="flex items-center justify-between"
                onSelect={() => {}}
              >
                <span>{item.label}</span>
                <ShortcutKeys action={item.action} bindings={bindings} />
              </CommandItem>
            ))}
          </CommandGroup>
        ))}
        {customItems.length > 0 && (
          <CommandGroup heading="Custom">
            {customItems.map((item) => (
              <CommandItem
                key={item.action}
                value={`Custom ${item.label} ${formatChordSegments(bindings[item.action]).join(" ")}`}
                className="flex items-center justify-between"
                onSelect={() => {}}
              >
                <span>{item.label}</span>
                <ShortcutKeys action={item.action} bindings={bindings} />
              </CommandItem>
            ))}
          </CommandGroup>
        )}
      </CommandList>
    </CommandDialog>
  );
}

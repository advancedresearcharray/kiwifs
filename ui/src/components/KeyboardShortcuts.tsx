import { HelpCircle } from "lucide-react";
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@kw/components/ui/command";
import { cn } from "@kw/lib/cn";
import {
  formatChordParts,
  isMacPlatform,
  type ShortcutDisplaySection,
} from "../lib/kiwiKeybindings";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  sections: ShortcutDisplaySection[];
  conflicts?: { chord: string; actions: string[] }[];
};

function KbdCombo({ keys }: { keys: string[] }) {
  return (
    <span className="ml-auto flex items-center gap-1">
      {keys.map((key, index) => (
        <kbd
          key={`${key}-${index}`}
          className="px-2 py-0.5 rounded border border-border bg-muted font-mono text-xs text-muted-foreground"
        >
          {key}
        </kbd>
      ))}
    </span>
  );
}

export function KeyboardShortcutsHelpButton({
  onClick,
  className,
}: {
  onClick: () => void;
  className?: string;
}) {
  const mac = isMacPlatform();
  const mod = mac ? "⌘" : "Ctrl+";
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label="Keyboard shortcuts"
      title={`Keyboard shortcuts (${mod}/ or ?)`}
      className={cn(
        "inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-accent-foreground transition-colors",
        className,
      )}
    >
      <HelpCircle className="h-4 w-4" />
    </button>
  );
}

export function KeyboardShortcuts({ open, onOpenChange, sections, conflicts = [] }: Props) {
  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Keyboard shortcuts"
      contentClassName="sm:max-w-lg"
      commandProps={{ shouldFilter: true }}
    >
      <CommandInput placeholder="Search shortcuts…" />
      <CommandList>
        {conflicts.length > 0 && (
          <div className="mx-2 mt-2 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
            Conflicting bindings:{" "}
            {conflicts.map((c) => `${c.actions.join(" / ")} (${c.chord})`).join("; ")}
          </div>
        )}
        <CommandEmpty>No shortcuts found.</CommandEmpty>
        {sections.map((section) => (
          <CommandGroup key={section.name} heading={section.name}>
            {section.items.map((item) => (
              <CommandItem
                key={`${section.name}-${item.action}`}
                value={`${section.name} ${item.label} ${item.keys.join(" ")}`}
                className="flex items-center justify-between gap-3"
              >
                <span className="flex items-center gap-2">
                  {item.label}
                  {item.custom && (
                    <span className="rounded bg-primary/10 px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-wide text-primary">
                      Custom
                    </span>
                  )}
                </span>
                <KbdCombo keys={item.keys.length > 0 ? item.keys : formatChordParts(item.chord)} />
              </CommandItem>
            ))}
          </CommandGroup>
        ))}
      </CommandList>
    </CommandDialog>
  );
}

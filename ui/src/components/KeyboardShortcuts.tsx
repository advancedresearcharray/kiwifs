import { Keyboard } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@kw/components/ui/dialog";
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

export function KeyboardShortcuts({ open, onOpenChange, bindings, conflicts = [] }: Props) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Keyboard className="h-4 w-4" />
            Keyboard shortcuts
          </DialogTitle>
        </DialogHeader>
        {conflicts.length > 0 && (
          <div className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
            Conflicting bindings detected:{" "}
            {conflicts.map((c) => `${c.actions.join(" / ")} (${formatChordDisplay(c.chord)})`).join("; ")}
          </div>
        )}
        <div className="space-y-4">
          {SHORTCUT_SECTIONS.map((s) => (
            <div key={s.section}>
              <div className="text-xs uppercase tracking-wider text-muted-foreground mb-2">
                {s.section}
              </div>
              <div className="space-y-1.5">
                {s.items.map((item) => (
                  <div
                    key={item.action}
                    className="flex items-center justify-between text-sm"
                  >
                    <span>{item.label}</span>
                    <kbd className="px-2 py-0.5 rounded border border-border bg-muted font-mono text-xs text-muted-foreground">
                      {formatChordDisplay(bindings[item.action])}
                    </kbd>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </DialogContent>
    </Dialog>
  );
}

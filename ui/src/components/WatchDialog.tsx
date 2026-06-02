import { useState } from "react";
import { Eye, EyeOff, Mail, MessageSquare, AlertCircle } from "lucide-react";
import { Button } from "@kw/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@kw/components/ui/dialog";

type Channel = "email" | "slack" | "discord";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  path: string;
  isWatched: boolean;
  isCloud: boolean;
  onWatch: (path: string, channel?: string) => void;
  onUnwatch: (path: string) => void;
};

const CHANNELS: { id: Channel; label: string; icon: typeof Mail; description: string }[] = [
  { id: "email", label: "Email", icon: Mail, description: "Get notified via your account email" },
  { id: "slack", label: "Slack", icon: MessageSquare, description: "Post to a Slack channel via webhook" },
  { id: "discord", label: "Discord", icon: MessageSquare, description: "Post to a Discord channel via webhook" },
];

export function WatchDialog({ open, onOpenChange, path, isWatched, isCloud, onWatch, onUnwatch }: Props) {
  const [selected, setSelected] = useState<Channel>("email");
  const [busy, setBusy] = useState(false);

  if (!isCloud) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-muted-foreground" />
              Notifications not available
            </DialogTitle>
            <DialogDescription>
              Page watch notifications require KiwiFS Cloud. In standalone mode,
              you can configure webhooks manually via the API to receive change events.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  if (isWatched) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <EyeOff className="h-5 w-5" />
              Unwatch this page?
            </DialogTitle>
            <DialogDescription>
              You'll stop receiving notifications when <code className="font-mono text-xs bg-muted px-1 py-0.5 rounded">{path}</code> is modified.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              disabled={busy}
              onClick={async () => {
                setBusy(true);
                onUnwatch(path);
                setBusy(false);
                onOpenChange(false);
              }}
            >
              {busy ? "Unwatching..." : "Unwatch"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Eye className="h-5 w-5" />
            Watch this page
          </DialogTitle>
          <DialogDescription>
            Get notified when <code className="font-mono text-xs bg-muted px-1 py-0.5 rounded">{path}</code> is modified.
            Choose how you'd like to be notified:
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-2 py-2">
          {CHANNELS.map((ch) => {
            const Icon = ch.icon;
            const active = selected === ch.id;
            return (
              <button
                key={ch.id}
                type="button"
                onClick={() => setSelected(ch.id)}
                className={
                  "flex items-center gap-3 w-full rounded-md border px-3 py-2.5 text-left text-sm transition-colors " +
                  (active
                    ? "border-primary bg-primary/5 ring-1 ring-primary/20"
                    : "border-border hover:bg-accent")
                }
              >
                <Icon className={"h-4 w-4 shrink-0 " + (active ? "text-primary" : "text-muted-foreground")} />
                <div className="min-w-0">
                  <div className={"font-medium " + (active ? "text-primary" : "")}>{ch.label}</div>
                  <div className="text-xs text-muted-foreground">{ch.description}</div>
                </div>
              </button>
            );
          })}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            disabled={busy}
            onClick={async () => {
              setBusy(true);
              onWatch(path, selected);
              setBusy(false);
              onOpenChange(false);
            }}
          >
            {busy ? "Watching..." : "Watch page"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

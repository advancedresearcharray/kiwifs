import { useState } from "react";
import { Eye, EyeOff, Bell, AlertCircle } from "lucide-react";
import { Button } from "@kw/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@kw/components/ui/dialog";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  path: string;
  isWatched: boolean;
  isCloud: boolean;
  onWatch: (path: string) => void;
  onUnwatch: (path: string) => void;
};

export function WatchDialog({ open, onOpenChange, path, isWatched, isCloud, onWatch, onUnwatch }: Props) {
  const [busy, setBusy] = useState(false);

  if (!isCloud) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-muted-foreground" />
              Notifications not available
            </DialogTitle>
            <DialogDescription>
              Page watch notifications require Kiwi Cloud. In standalone mode,
              configure webhooks via the API for change events.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => onOpenChange(false)}>
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
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <EyeOff className="h-5 w-5" />
              Unwatch page
            </DialogTitle>
            <DialogDescription>
              Stop receiving notifications for changes to{" "}
              <code className="font-mono text-xs bg-muted px-1 py-0.5 rounded">{path}</code>
            </DialogDescription>
          </DialogHeader>
          <DialogFooter className="gap-2 sm:gap-0">
            <Button variant="outline" size="sm" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              size="sm"
              disabled={busy}
              onClick={async () => {
                setBusy(true);
                onUnwatch(path);
                setBusy(false);
                onOpenChange(false);
              }}
            >
              Unwatch
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Eye className="h-5 w-5" />
            Watch page
          </DialogTitle>
          <DialogDescription>
            Get notified when{" "}
            <code className="font-mono text-xs bg-muted px-1 py-0.5 rounded">{path}</code>{" "}
            is modified. Notifications are sent to your connected channels.
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-2 text-xs text-muted-foreground bg-muted/50 rounded-md px-3 py-2">
          <Bell className="h-3.5 w-3.5 shrink-0" />
          <span>Notifications go to your email and connected integrations (Slack, Discord).</span>
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="outline" size="sm" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            size="sm"
            disabled={busy}
            onClick={async () => {
              setBusy(true);
              onWatch(path);
              setBusy(false);
              onOpenChange(false);
            }}
          >
            Watch page
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

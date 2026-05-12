// KiwiClipDialog — Dialog to clip a URL into the knowledge base.

import { useEffect, useState } from "react";
import { Check, Clipboard, Loader2 } from "lucide-react";
import { api } from "@kw/lib/api";
import { Button } from "@kw/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@kw/components/ui/dialog";
import { Input } from "@kw/components/ui/input";
import { Label } from "@kw/components/ui/label";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onClipped: (path: string) => void;
};

export function KiwiClipDialog({ open, onOpenChange, onClipped }: Props) {
  const [url, setUrl] = useState("");
  const [title, setTitle] = useState("");
  const [folder, setFolder] = useState("clips/");
  const [tags, setTags] = useState("");
  const [clipping, setClipping] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<{ path: string; title: string } | null>(
    null,
  );

  useEffect(() => {
    if (!open) {
      setUrl("");
      setTitle("");
      setFolder("clips/");
      setTags("");
      setError(null);
      setResult(null);
    }
  }, [open]);

  async function handleClip() {
    if (!url.trim()) {
      setError("URL is required.");
      return;
    }
    setError(null);
    setClipping(true);
    try {
      const tagList = tags
        .split(",")
        .map((t) => t.trim())
        .filter(Boolean);
      const r = await api.clipUrl({
        url: url.trim(),
        title: title.trim() || undefined,
        tags: tagList.length > 0 ? tagList : undefined,
        folder: folder.trim() || undefined,
      });
      setResult({ path: r.path, title: r.title });
    } catch (e) {
      setError(String(e));
    } finally {
      setClipping(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Clipboard className="h-4 w-4 text-primary" />
            Clip URL
          </DialogTitle>
        </DialogHeader>

        {result ? (
          <div className="py-4 text-center space-y-3">
            <div className="h-10 w-10 mx-auto rounded-full bg-green-500/10 text-green-600 grid place-items-center">
              <Check className="h-5 w-5" />
            </div>
            <div className="text-sm font-medium">Clipped successfully!</div>
            <div className="text-xs text-muted-foreground font-mono">
              {result.path}
            </div>
            <Button
              variant="outline"
              onClick={() => {
                onClipped(result.path);
              }}
            >
              Open page
            </Button>
          </div>
        ) : (
          <>
            <div className="grid gap-4">
              <div className="grid gap-1.5">
                <Label htmlFor="clip-url">URL</Label>
                <Input
                  id="clip-url"
                  type="url"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://example.com/article"
                  autoFocus
                  onKeyDown={(e) => {
                    if (e.key === "Enter") handleClip();
                  }}
                />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="clip-title">
                  Title override{" "}
                  <span className="text-muted-foreground">(optional)</span>
                </Label>
                <Input
                  id="clip-title"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="Auto-detected from page"
                />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="clip-folder">Folder</Label>
                <Input
                  id="clip-folder"
                  value={folder}
                  onChange={(e) => setFolder(e.target.value)}
                  placeholder="clips/"
                  className="font-mono"
                />
              </div>
              <div className="grid gap-1.5">
                <Label htmlFor="clip-tags">
                  Tags{" "}
                  <span className="text-muted-foreground">(comma-separated)</span>
                </Label>
                <Input
                  id="clip-tags"
                  value={tags}
                  onChange={(e) => setTags(e.target.value)}
                  placeholder="web, article, research"
                />
              </div>
              {error && (
                <div className="text-sm text-destructive font-mono">{error}</div>
              )}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <Button onClick={handleClip} disabled={clipping}>
                {clipping ? (
                  <>
                    <Loader2 className="h-4 w-4 animate-spin mr-1" />
                    Clipping...
                  </>
                ) : (
                  "Clip"
                )}
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}

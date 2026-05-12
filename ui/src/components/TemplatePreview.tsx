// TemplatePreview — Renders resolved template content in a read-only preview.
// Variables that were resolved are shown inline. The content comes from the
// server's `POST /api/kiwi/templates/preview` endpoint.

import { Loader2 } from "lucide-react";
import { ScrollArea } from "@kw/components/ui/scroll-area";

type Props = {
  content: string | null;
  loading: boolean;
  error: string | null;
};

export function TemplatePreview({ content, loading, error }: Props) {
  if (loading) {
    return (
      <div className="flex items-center gap-2 py-8 justify-center text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        Resolving template...
      </div>
    );
  }

  if (error) {
    return (
      <div className="py-4 px-3 text-sm text-destructive font-mono">
        {error}
      </div>
    );
  }

  if (!content) {
    return (
      <div className="py-8 text-center text-sm text-muted-foreground">
        Select a template to see a preview.
      </div>
    );
  }

  return (
    <ScrollArea className="max-h-64">
      <pre className="text-xs font-mono p-3 whitespace-pre-wrap text-muted-foreground bg-muted/30 rounded-md">
        {content}
      </pre>
    </ScrollArea>
  );
}

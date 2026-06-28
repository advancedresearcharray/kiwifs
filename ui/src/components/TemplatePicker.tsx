import { useEffect, useState } from "react";
import { File, FileText, Info, Loader2 } from "lucide-react";
import { api } from "@kw/lib/api";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@kw/components/ui/dialog";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@kw/components/ui/tooltip";
import { TemplatePreview } from "./TemplatePreview";

type Template = { name: string; path: string };

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (content: string) => void;
};

const VARIABLE_HELP = [
  { var: '{{date "format"}}', desc: "Current date in the given format" },
  { var: '{{time "format"}}', desc: "Current time in the given format" },
  { var: "{{related_pages path N}}", desc: "N pages related to path" },
  { var: "{{recent_pages N}}", desc: "N most recently modified pages" },
  { var: '{{schema_fields "type"}}', desc: "Frontmatter fields for a schema type" },
];

export function TemplatePicker({ open, onOpenChange, onSelect }: Props) {
  const [templates, setTemplates] = useState<Template[]>([]);
  const [loading, setLoading] = useState(true);
  const [fetching, setFetching] = useState<string | null>(null);
  const [hoveredTemplate, setHoveredTemplate] = useState<string | null>(null);
  const [previewContent, setPreviewContent] = useState<string | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setLoading(true);
    setHoveredTemplate(null);
    setPreviewContent(null);
    api
      .listTemplates()
      .then((r) => setTemplates(r.templates ?? []))
      .catch(() => setTemplates([]))
      .finally(() => setLoading(false));
  }, [open]);

  // Load preview when hovering a template
  useEffect(() => {
    if (!hoveredTemplate) {
      setPreviewContent(null);
      setPreviewError(null);
      return;
    }
    setPreviewLoading(true);
    setPreviewError(null);
    api
      .previewTemplate(hoveredTemplate)
      .then((r) => setPreviewContent(r.content))
      .catch((e) => {
        setPreviewError(String(e));
        setPreviewContent(null);
      })
      .finally(() => setPreviewLoading(false));
  }, [hoveredTemplate]);

  function handleSelect(name: string) {
    if (name === "__blank__") {
      onSelect("");
      onOpenChange(false);
      return;
    }
    setFetching(name);
    api
      .readTemplate(name)
      .then((r) => {
        onSelect(r.content);
        onOpenChange(false);
      })
      .catch(() => {})
      .finally(() => setFetching(null));
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            Choose a template
            <Tooltip>
              <TooltipTrigger asChild>
                <Info className="h-3.5 w-3.5 text-muted-foreground cursor-help" />
              </TooltipTrigger>
              <TooltipContent side="right" className="max-w-xs">
                <div className="text-xs space-y-1">
                  <div className="font-medium mb-1">Template variables:</div>
                  {VARIABLE_HELP.map((v) => (
                    <div key={v.var}>
                      <code className="bg-muted px-1 rounded text-[10px]">
                        {v.var}
                      </code>
                      <span className="text-muted-foreground ml-1">{v.desc}</span>
                    </div>
                  ))}
                </div>
              </TooltipContent>
            </Tooltip>
          </DialogTitle>
        </DialogHeader>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {/* Template list */}
          <div className="space-y-1">
            <button
              type="button"
              onClick={() => handleSelect("__blank__")}
              onMouseEnter={() => setHoveredTemplate(null)}
              className="w-full flex items-center gap-3 px-3 py-2.5 rounded-md text-left hover:bg-accent hover:text-accent-foreground transition-colors"
            >
              <File className="h-5 w-5 text-muted-foreground shrink-0" />
              <div>
                <div className="text-sm font-medium">Blank page</div>
                <div className="text-xs text-muted-foreground">
                  Start from scratch
                </div>
              </div>
            </button>
            {loading ? (
              <div className="flex items-center gap-2 px-3 py-4 text-sm text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                Loading templates...
              </div>
            ) : (
              templates.map((t) => (
                <button
                  key={t.name}
                  type="button"
                  onClick={() => handleSelect(t.name)}
                  onMouseEnter={() => setHoveredTemplate(t.name)}
                  disabled={fetching === t.name}
                  className="w-full flex items-center gap-3 px-3 py-2.5 rounded-md text-left hover:bg-accent hover:text-accent-foreground transition-colors disabled:opacity-50"
                >
                  <FileText className="h-5 w-5 text-primary shrink-0" />
                  <div>
                    <div className="text-sm font-medium">{t.name}</div>
                    <div className="text-xs text-muted-foreground truncate">
                      {t.path}
                    </div>
                  </div>
                  {fetching === t.name && (
                    <Loader2 className="h-4 w-4 animate-spin ml-auto" />
                  )}
                </button>
              ))
            )}
          </div>

          {/* Preview panel */}
          <div className="border border-border rounded-md overflow-hidden hidden sm:block">
            <div className="text-xs font-medium px-3 py-2 border-b border-border bg-muted/30">
              Preview
            </div>
            <TemplatePreview
              content={previewContent}
              loading={previewLoading}
              error={previewError}
            />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

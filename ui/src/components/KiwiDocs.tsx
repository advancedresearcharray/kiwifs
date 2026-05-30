import { useEffect, useMemo, useRef, useState, useCallback } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import remarkMath from "remark-math";
import rehypeSlug from "rehype-slug";
import rehypeRaw from "rehype-raw";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import rehypeKatex from "rehype-katex";
import matter from "gray-matter";
import {
  ArrowLeft,
  FileText,
  Maximize2,
  Minimize2,
  ZoomIn,
  ZoomOut,
} from "lucide-react";
import { api, type TreeEntry } from "@kw/lib/api";
import { titleize } from "@kw/lib/paths";
import { buildResolver, remarkWikiLinks } from "@kw/lib/wikiLinks";
import { remarkMark, stripObsidianComments, remarkInlineTags } from "@kw/lib/remarkPlugins";
import remarkEmoji from "remark-emoji";
import remarkSupersub from "remark-supersub";
import remarkDefinitionList from "remark-definition-list";
import remarkDirective from "remark-directive";
import { remarkKiwiDirectives } from "@kw/lib/remarkDirectives";
import { ShikiCode } from "./ShikiCode";
import { MermaidDiagram } from "./MermaidDiagram";
import { ErrorBoundary } from "./ErrorBoundary";
import { Button } from "@kw/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@kw/components/ui/tooltip";

const sanitizeSchema = {
  ...defaultSchema,
  tagNames: [
    ...(defaultSchema.tagNames || []),
    "details", "summary", "kbd", "mark", "span", "div", "figure", "figcaption",
    "section", "sup", "sub", "dl", "dt", "dd", "abbr",
  ],
  protocols: {
    ...defaultSchema.protocols,
    href: ["http", "https", "mailto"],
  },
  attributes: {
    ...defaultSchema.attributes,
    "*": [...(defaultSchema.attributes?.["*"] || []), "className", "style", "role", "id",
      "data-footnotes", "data-footnote-ref", "data-footnote-backref",
      "aria-describedby", "aria-label"],
    a: [...(defaultSchema.attributes?.a || []), "className"],
    img: [...(defaultSchema.attributes?.img || []), "width", "height"],
    abbr: ["title"],
  },
};

type Props = {
  path: string;
  tree: TreeEntry | null;
  onClose: () => void;
  onNavigate: (path: string) => void;
};

function parseDoc(content: string): { body: string; meta: Record<string, unknown> } {
  try {
    const m = matter(content);
    return { body: m.content, meta: m.data || {} };
  } catch {
    return { body: content, meta: {} };
  }
}

export function KiwiDocs({ path, tree, onClose }: Props) {
  const [content, setContent] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [zoom, setZoom] = useState(80);
  const [fullscreen, setFullscreen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    let cancelled = false;
    setContent(null);
    setError(null);
    api
      .readFile(path)
      .then((r) => {
        if (cancelled) return;
        let text = r.content;
        // readFile returns raw response text; if it's JSON-wrapped, extract .content
        if (text.startsWith("{")) {
          try { text = JSON.parse(text).content ?? text; } catch {}
        }
        setContent(text);
      })
      .catch((e) => {
        if (!cancelled) setError(String(e));
      });
    return () => { cancelled = true; };
  }, [path]);

  const resolver = useMemo(() => buildResolver(tree), [tree]);
  const parsed = useMemo(() => {
    if (content == null) return { body: "", meta: {} as Record<string, unknown> };
    return parseDoc(content);
  }, [content]);

  const docTitle = typeof parsed.meta.title === "string"
    ? parsed.meta.title
    : titleize(path);

  const toggleFullscreen = useCallback(() => setFullscreen((v) => !v), []);

  const toolbarProps = {
    title: docTitle,
    zoom,
    fullscreen,
    onClose,
    onZoomIn: () => setZoom((z) => Math.min(200, z + 10)),
    onZoomOut: () => setZoom((z) => Math.max(40, z - 10)),
    onToggleFullscreen: toggleFullscreen,
  };

  if (error) {
    return (
      <div className="flex flex-col h-full">
        <DocsToolbar {...toolbarProps} />
        <div className="flex-1 grid place-items-center">
          <p className="text-sm text-destructive">{error}</p>
        </div>
      </div>
    );
  }

  if (content === null) {
    return (
      <div className="flex flex-col h-full">
        <DocsToolbar {...toolbarProps} title="Loading..." />
        <div className="flex-1 grid place-items-center">
          <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
        </div>
      </div>
    );
  }

  return (
    <div className={`flex flex-col h-full ${fullscreen ? "fixed inset-0 z-50 bg-background" : ""}`}>
      <DocsToolbar {...toolbarProps} />

      <div
        ref={containerRef}
        className="flex-1 overflow-auto"
        style={{
          background: "var(--kiwi-docs-bg, #e8e8e8)",
        }}
      >
        <div
          className="kiwi-docs-page-container mx-auto my-8"
          style={{
            transform: `scale(${zoom / 100})`,
            transformOrigin: "top center",
          }}
        >
          {/* The "page" — a letter-sized white surface */}
          <div className="kiwi-docs-page">
            <ErrorBoundary>
              <ReactMarkdown
                remarkPlugins={[
                  remarkGfm,
                  remarkMath,
                  remarkMark,
                  remarkInlineTags,
                  remarkEmoji,
                  remarkSupersub,
                  remarkDefinitionList,
                  remarkDirective,
                  remarkKiwiDirectives,
                  [remarkWikiLinks, { resolver }],
                ]}
                rehypePlugins={[
                  rehypeRaw,
                  [rehypeSanitize, sanitizeSchema],
                  rehypeKatex,
                  rehypeSlug,
                ]}
                components={{
                  code: ({ className, children, ...rest }: any) => {
                    const match = /language-([A-Za-z0-9_-]+)/.exec(className || "");
                    const lang = match ? match[1] : undefined;
                    const raw = String(children).replace(/\n$/, "");
                    if (lang === "mermaid") return <MermaidDiagram chart={raw} />;
                    if (!lang || !raw.includes("\n")) {
                      return <code className={className} {...rest}>{children}</code>;
                    }
                    return <ShikiCode code={raw} lang={lang} />;
                  },
                  pre: ({ children }) => <>{children}</>,
                  table: ({ children, node: _node, ...rest }: any) => (
                    <table {...rest}>{children}</table>
                  ),
                  section: ({ children, node: _node, ...rest }: any) => {
                    const props = rest as Record<string, unknown>;
                    if (props["data-footnotes"] !== undefined || props.className === "footnotes") {
                      return (
                        <section className="kiwi-docs-footnotes" role="doc-endnotes" {...rest}>
                          <hr />
                          <h2>Footnotes</h2>
                          {children}
                        </section>
                      );
                    }
                    return <section {...rest}>{children}</section>;
                  },
                }}
              >
                {stripObsidianComments(parsed.body)}
              </ReactMarkdown>
            </ErrorBoundary>
          </div>
        </div>
      </div>

      {/* Scoped styles for the document view */}
      <style>{DOCS_STYLES}</style>
    </div>
  );
}

const DOCS_STYLES = `
  :root {
    --kiwi-docs-bg: #e0e0e0;
  }
  .dark {
    --kiwi-docs-bg: #1a1a1a;
  }

  .kiwi-docs-page-container {
    width: 8.5in;
  }

  .kiwi-docs-page {
    background: white;
    color: #1a1a1a;
    width: 8.5in;
    min-height: 11in;
    padding: 1in 1.25in;
    box-shadow: 0 2px 8px rgba(0,0,0,0.15), 0 0 1px rgba(0,0,0,0.1);
    font-family: "Georgia", "Times New Roman", "Noto Serif", serif;
    font-size: 12pt;
    line-height: 1.6;
    box-sizing: border-box;
  }

  .kiwi-docs-page h1 {
    font-size: 20pt;
    font-weight: 700;
    margin: 0 0 0.5em 0;
    line-height: 1.3;
    color: #111;
  }
  .kiwi-docs-page h2 {
    font-size: 16pt;
    font-weight: 700;
    margin: 1.5em 0 0.5em 0;
    color: #222;
    break-after: avoid;
  }
  .kiwi-docs-page h3 {
    font-size: 13pt;
    font-weight: 700;
    margin: 1.2em 0 0.4em 0;
    color: #333;
    break-after: avoid;
  }
  .kiwi-docs-page h4,
  .kiwi-docs-page h5,
  .kiwi-docs-page h6 {
    font-size: 12pt;
    font-weight: 700;
    margin: 1em 0 0.3em 0;
    break-after: avoid;
  }

  .kiwi-docs-page p {
    margin: 0 0 0.8em 0;
    text-align: justify;
    orphans: 3;
    widows: 3;
  }

  .kiwi-docs-page table {
    width: 100%;
    border-collapse: collapse;
    margin: 1em 0;
    font-size: 10.5pt;
    break-inside: avoid;
  }
  .kiwi-docs-page th,
  .kiwi-docs-page td {
    border: 1px solid #ccc;
    padding: 6px 10px;
    text-align: left;
  }
  .kiwi-docs-page th {
    background: #f5f5f5;
    font-weight: 600;
  }

  .kiwi-docs-page blockquote {
    border-left: 3px solid #bbb;
    padding-left: 1em;
    margin: 1em 0;
    color: #555;
    font-style: italic;
  }

  .kiwi-docs-page code {
    font-family: "JetBrains Mono", "Fira Code", "Consolas", monospace;
    font-size: 0.85em;
    background: #f4f4f4;
    padding: 0.15em 0.3em;
    border-radius: 3px;
  }
  .kiwi-docs-page pre {
    background: #f8f8f8;
    border: 1px solid #e0e0e0;
    border-radius: 4px;
    padding: 0.8em 1em;
    overflow-x: auto;
    font-size: 9.5pt;
    margin: 1em 0;
    break-inside: avoid;
  }
  .kiwi-docs-page pre code {
    background: none;
    padding: 0;
  }

  .kiwi-docs-page ul,
  .kiwi-docs-page ol {
    margin: 0 0 0.8em 0;
    padding-left: 2em;
  }
  .kiwi-docs-page li {
    margin-bottom: 0.3em;
  }

  .kiwi-docs-page a {
    color: #1a56db;
    text-decoration: underline;
  }

  .kiwi-docs-page hr {
    border: none;
    border-top: 1px solid #ddd;
    margin: 1.5em 0;
  }

  .kiwi-docs-page img {
    max-width: 100%;
    height: auto;
  }

  .kiwi-docs-footnotes {
    border-top: 1px solid #ccc;
    margin-top: 2em;
    padding-top: 1em;
    font-size: 10pt;
  }
  .kiwi-docs-footnotes h2 {
    font-size: 11pt;
    color: #666;
    margin: 0 0 0.5em 0;
  }
  .kiwi-docs-page sup a[data-footnote-ref] {
    color: #1a56db;
    font-size: 0.8em;
    text-decoration: none;
  }
  .kiwi-docs-page section[data-footnotes] {
    border-top: 1px solid #ccc;
    margin-top: 2em;
    padding-top: 0.8em;
    font-size: 10pt;
  }

  .kiwi-docs-page figure {
    margin: 1em 0;
    text-align: center;
  }
  .kiwi-docs-page figcaption {
    font-size: 10pt;
    color: #666;
    margin-top: 0.5em;
  }

  .kiwi-docs-page .katex-display {
    margin: 1em 0;
    break-inside: avoid;
  }

  @media print {
    .kiwi-docs-page {
      box-shadow: none;
      margin: 0;
      padding: 0;
    }
    @page {
      size: letter;
      margin: 1in 1.25in;
    }
  }
`;

function DocsToolbar({
  title,
  zoom,
  fullscreen,
  onClose,
  onZoomIn,
  onZoomOut,
  onToggleFullscreen,
}: {
  title: string;
  zoom: number;
  fullscreen: boolean;
  onClose: () => void;
  onZoomIn: () => void;
  onZoomOut: () => void;
  onToggleFullscreen: () => void;
}) {
  return (
    <div className="h-12 shrink-0 border-b border-border bg-card flex items-center px-3 gap-2">
      <div className="flex items-center gap-2 min-w-0 flex-1">
        <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0" onClick={onClose}>
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <FileText className="h-4 w-4 text-primary shrink-0" />
        <span className="text-sm font-medium truncate">{title}</span>
      </div>

      <div className="flex items-center gap-0.5">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={onZoomOut}>
              <ZoomOut className="h-3.5 w-3.5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent side="bottom">Zoom out</TooltipContent>
        </Tooltip>
        <span className="text-xs text-muted-foreground min-w-[3rem] text-center tabular-nums">
          {zoom}%
        </span>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon" className="h-7 w-7" onClick={onZoomIn}>
              <ZoomIn className="h-3.5 w-3.5" />
            </Button>
          </TooltipTrigger>
          <TooltipContent side="bottom">Zoom in</TooltipContent>
        </Tooltip>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onToggleFullscreen}>
              {fullscreen ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
            </Button>
          </TooltipTrigger>
          <TooltipContent side="bottom">{fullscreen ? "Exit fullscreen" : "Fullscreen"}</TooltipContent>
        </Tooltip>
      </div>
    </div>
  );
}

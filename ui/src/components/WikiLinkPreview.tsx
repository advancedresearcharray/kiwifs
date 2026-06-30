import React, { useCallback, useRef, useState } from "react";
import { FileQuestion, Tag } from "lucide-react";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@kw/components/ui/hover-card";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@kw/components/ui/card";
import { Badge } from "@kw/components/ui/badge";
import {
  fetchWikiLinkPeek,
  getCachedWikiLinkPeek,
  type WikiLinkPeekData,
  type WikiLinkPeekResult,
} from "@kw/lib/wikiLinkPeek";
import { canOpenHoverPreview, parseWikiLinkHref } from "@kw/lib/wikiLinkAnchor";

function scrollToAnchor(anchor: string, delayMs = 100) {
  requestAnimationFrame(() => {
    setTimeout(() => {
      const el = document.getElementById(anchor.startsWith("#") ? anchor.slice(1) : anchor);
      el?.scrollIntoView({ behavior: "smooth", block: "start" });
    }, delayMs);
  });
}

type WikiLinkAnchorArgs = {
  href: string;
  children: React.ReactNode;
  onNavigate: (path: string) => void;
  rest?: Record<string, unknown>;
};

/** Renders wiki-link anchors with hover preview, or null for non-wiki hrefs. */
export function renderWikiLinkAnchor({
  href,
  children,
  onNavigate,
  rest,
}: WikiLinkAnchorArgs): React.ReactNode | null {
  const parsed = parseWikiLinkHref(href);
  if (parsed.kind === "resolved") {
    return (
      <WikiLinkPreview
        pagePath={parsed.pagePath}
        anchor={parsed.anchor}
        className="wiki-link"
        onNavigate={onNavigate}
        onAnchorScroll={(a) => scrollToAnchor(a)}
        {...rest}
      >
        {children}
      </WikiLinkPreview>
    );
  }

  if (parsed.kind === "missing") {
    return (
      <WikiLinkPreview
        pagePath={parsed.pagePath}
        missing
        className="wiki-link-missing"
        title={`Missing: ${parsed.pagePath} — click to create`}
        onNavigate={onNavigate}
        {...rest}
      >
        {children}
      </WikiLinkPreview>
    );
  }

  return null;
}

type WikiLinkPreviewProps = {
  pagePath: string;
  missing?: boolean;
  anchor?: string;
  className?: string;
  title?: string;
  onNavigate: (path: string) => void;
  onAnchorScroll?: (anchor: string) => void;
  children: React.ReactNode;
};

function PreviewSkeleton() {
  return (
    <div className="space-y-2 p-4" data-testid="wiki-link-preview-loading">
      <div className="h-4 w-3/4 animate-pulse rounded bg-muted" />
      <div className="h-3 w-1/2 animate-pulse rounded bg-muted" />
      <div className="h-3 w-full animate-pulse rounded bg-muted" />
      <div className="h-3 w-5/6 animate-pulse rounded bg-muted" />
    </div>
  );
}

function PreviewBody({
  result,
  missing,
  targetPath,
}: {
  result: WikiLinkPeekResult | null;
  missing?: boolean;
  targetPath: string;
}) {
  if (missing || result?.status === "not_found") {
    return (
      <CardContent className="p-4 pt-0">
        <div className="flex items-start gap-2 text-sm text-muted-foreground">
          <FileQuestion className="mt-0.5 h-4 w-4 shrink-0" />
          <div>
            <p className="font-medium text-foreground">Page not found</p>
            <p className="mt-1 break-all text-xs">{targetPath}</p>
          </div>
        </div>
      </CardContent>
    );
  }

  if (!result) {
    return <PreviewSkeleton />;
  }

  if (result.status === "error") {
    return (
      <CardContent className="p-4 pt-0 text-sm text-muted-foreground">
        {result.message}
      </CardContent>
    );
  }

  const data: WikiLinkPeekData = result.data;
  return (
    <CardContent className="space-y-2 p-4 pt-0">
      {data.snippet ? (
        <p className="text-sm text-muted-foreground line-clamp-4">{data.snippet}</p>
      ) : (
        <p className="text-sm italic text-muted-foreground">No preview available.</p>
      )}
      {data.tags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {data.tags.map((tag) => (
            <Badge key={tag} variant="secondary" className="gap-1 text-xs">
              <Tag className="h-3 w-3" />
              {tag}
            </Badge>
          ))}
        </div>
      )}
    </CardContent>
  );
}

export function WikiLinkPreview({
  pagePath,
  missing,
  anchor,
  className,
  title,
  onNavigate,
  onAnchorScroll,
  children,
}: WikiLinkPreviewProps) {
  const [open, setOpen] = useState(false);
  const [result, setResult] = useState<WikiLinkPeekResult | null>(null);
  const pointerInside = useRef(false);
  const fetchGeneration = useRef(0);

  const loadPreview = useCallback(async () => {
    if (missing) {
      setResult({ status: "not_found" });
      return;
    }
    const cached = getCachedWikiLinkPeek(pagePath);
    if (cached) {
      setResult(cached);
      return;
    }
    setResult(null);
    const generation = ++fetchGeneration.current;
    const peek = await fetchWikiLinkPeek(pagePath);
    if (generation !== fetchGeneration.current) return;
    setResult(peek);
  }, [missing, pagePath]);

  const handleOpenChange = useCallback(
    (nextOpen: boolean) => {
      if (!nextOpen) {
        setOpen(false);
        fetchGeneration.current += 1;
        return;
      }
      // Radix opens on keyboard focus too; only allow pointer-driven opens.
      if (!pointerInside.current) return;
      setOpen(true);
      void loadPreview();
    },
    [loadPreview],
  );

  const markPointerInside = () => {
    pointerInside.current = true;
  };

  const markPointerOutside = () => {
    pointerInside.current = false;
  };

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    e.preventDefault();
    if (missing) {
      onNavigate(`${pagePath}.md`);
      return;
    }
    onNavigate(pagePath);
    if (anchor) onAnchorScroll?.(anchor);
  };

  const displayTitle =
    result?.status === "ok"
      ? result.data.title
      : missing
        ? "Missing page"
        : pagePath.split("/").pop()?.replace(/\.md$/, "") ?? pagePath;

  const previewResult: WikiLinkPeekResult | null = result;

  if (!canOpenHoverPreview()) {
    return (
      <a
        href={missing ? "#" : anchor ? `#${pagePath}${anchor}` : `#${pagePath}`}
        onClick={handleClick}
        title={title}
        className={className}
      >
        {children}
      </a>
    );
  }

  return (
    <HoverCard open={open} openDelay={300} closeDelay={100} onOpenChange={handleOpenChange}>
      <HoverCardTrigger asChild>
        <a
          href={missing ? "#" : anchor ? `#${pagePath}${anchor}` : `#${pagePath}`}
          onClick={handleClick}
          onPointerEnter={markPointerInside}
          onPointerLeave={markPointerOutside}
          title={title}
          className={className}
        >
          {children}
        </a>
      </HoverCardTrigger>
      <HoverCardContent
        className="max-w-[320px]"
        onPointerEnter={markPointerInside}
        onPointerLeave={markPointerOutside}
        onPointerDown={(e: React.PointerEvent) => e.preventDefault()}
        onClick={(e: React.MouseEvent) => e.stopPropagation()}
      >
        <Card className="border-0 shadow-none">
          <CardHeader className="space-y-1 p-4 pb-2">
            <CardTitle className="text-sm leading-snug">{displayTitle}</CardTitle>
            <CardDescription className="break-all text-xs">{pagePath}</CardDescription>
          </CardHeader>
          <PreviewBody
            result={previewResult}
            missing={missing}
            targetPath={pagePath}
          />
        </Card>
      </HoverCardContent>
    </HoverCard>
  );
}

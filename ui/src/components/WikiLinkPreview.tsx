import React, { useEffect, useState } from "react";
import { FileQuestion } from "lucide-react";
import { api } from "@kw/lib/api";
import {
  fetchPeekData,
  supportsHoverPreview,
  type PeekData,
} from "@kw/lib/wikiLinkPreview";
import { Badge } from "@kw/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@kw/components/ui/card";
import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "@kw/components/ui/hover-card";

type WikiLinkPreviewProps = {
  pagePath: string;
  isMissing?: boolean;
  children: React.ReactElement;
};

function PreviewSkeleton() {
  return (
    <div className="space-y-2 p-4" data-testid="wiki-preview-skeleton">
      <div className="h-4 w-3/4 animate-pulse rounded bg-muted" />
      <div className="h-3 w-1/2 animate-pulse rounded bg-muted" />
      <div className="mt-3 space-y-1.5">
        <div className="h-3 w-full animate-pulse rounded bg-muted" />
        <div className="h-3 w-full animate-pulse rounded bg-muted" />
        <div className="h-3 w-2/3 animate-pulse rounded bg-muted" />
      </div>
    </div>
  );
}

function PreviewBody({
  pagePath,
  isMissing,
}: {
  pagePath: string;
  isMissing?: boolean;
}) {
  const [state, setState] = useState<
    "loading" | "loaded" | "not-found" | "error"
  >(isMissing ? "not-found" : "loading");
  const [data, setData] = useState<PeekData | null>(null);

  useEffect(() => {
    if (isMissing) {
      setState("not-found");
      return;
    }

    let cancelled = false;
    setState("loading");
    fetchPeekData(pagePath, (path) => api.peek(path))
      .then((result) => {
        if (cancelled) return;
        if ("notFound" in result) {
          setData(null);
          setState("not-found");
          return;
        }
        setData(result);
        setState("loaded");
      })
      .catch(() => {
        if (!cancelled) setState("error");
      });

    return () => {
      cancelled = true;
    };
  }, [pagePath, isMissing]);

  if (state === "loading") return <PreviewSkeleton />;

  if (state === "not-found") {
    return (
      <Card className="border-0 shadow-none">
        <CardHeader className="space-y-1 p-4 pb-2">
          <CardTitle className="flex items-center gap-2 text-sm">
            <FileQuestion className="h-4 w-4 text-muted-foreground" />
            Page not found
          </CardTitle>
          <CardDescription className="truncate font-mono text-xs">
            {pagePath}
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  if (state === "error" || !data) {
    return (
      <Card className="border-0 shadow-none">
        <CardContent className="p-4 text-sm text-muted-foreground">
          Preview unavailable
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-0 shadow-none">
      <CardHeader className="space-y-1 p-4 pb-2">
        <CardTitle className="text-sm leading-snug">{data.title}</CardTitle>
        <CardDescription className="truncate font-mono text-xs">
          {data.path}
        </CardDescription>
      </CardHeader>
      {data.snippet && (
        <CardContent className="px-4 pb-2 pt-0 text-sm text-muted-foreground leading-relaxed">
          {data.snippet}
        </CardContent>
      )}
      {data.tags.length > 0 && (
        <CardContent className="flex flex-wrap gap-1 px-4 pb-4 pt-0">
          {data.tags.map((tag) => (
            <Badge key={tag} variant="secondary" className="text-xs">
              {tag}
            </Badge>
          ))}
        </CardContent>
      )}
    </Card>
  );
}

export function WikiLinkPreview({
  pagePath,
  isMissing,
  children,
}: WikiLinkPreviewProps) {
  if (!supportsHoverPreview()) return children;

  return (
    <HoverCard openDelay={300} closeDelay={100}>
      <HoverCardTrigger asChild>{children}</HoverCardTrigger>
      <HoverCardContent
        side="top"
        align="start"
        onClick={(event) => event.stopPropagation()}
        onPointerDown={(event) => event.stopPropagation()}
      >
        <PreviewBody pagePath={pagePath} isMissing={isMissing} />
      </HoverCardContent>
    </HoverCard>
  );
}

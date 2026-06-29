import React, { useEffect, useState } from "react";
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
import {
  wikiLinkPreviewCache,
  type WikiLinkPreviewData,
} from "@kw/lib/wikiLinkPreviewCache";

type Props = {
  pagePath: string;
  isMissing?: boolean;
  children: React.ReactElement;
};

function PreviewSkeleton() {
  return (
    <div className="animate-pulse space-y-2 p-4">
      <div className="h-4 w-3/4 rounded bg-muted" />
      <div className="h-3 w-full rounded bg-muted" />
      <div className="h-3 w-5/6 rounded bg-muted" />
      <div className="flex gap-1 pt-1">
        <div className="h-5 w-12 rounded-full bg-muted" />
        <div className="h-5 w-14 rounded-full bg-muted" />
      </div>
    </div>
  );
}

function PreviewBody({
  data,
  missing,
}: {
  data: WikiLinkPreviewData | null;
  missing: boolean;
}) {
  if (missing) {
    return (
      <Card className="border-0 shadow-none">
        <CardHeader className="p-4 pb-2">
          <CardTitle className="text-sm">Page not found</CardTitle>
          <CardDescription className="text-xs">
            This wiki-link does not resolve to an existing page.
          </CardDescription>
        </CardHeader>
      </Card>
    );
  }

  if (!data) return <PreviewSkeleton />;

  return (
    <Card className="border-0 shadow-none">
      <CardHeader className="p-4 pb-2">
        <CardTitle className="text-sm leading-snug">{data.title}</CardTitle>
        <CardDescription className="truncate font-mono text-[11px]">
          {data.path}
        </CardDescription>
      </CardHeader>
      {data.snippet ? (
        <CardContent className="p-4 pt-0 text-xs leading-relaxed text-muted-foreground">
          {data.snippet}
        </CardContent>
      ) : null}
      {data.tags.length > 0 ? (
        <CardContent className="flex flex-wrap gap-1 p-4 pt-0">
          {data.tags.map((tag) => (
            <Badge key={tag} variant="secondary" className="text-[10px]">
              {tag}
            </Badge>
          ))}
        </CardContent>
      ) : null}
    </Card>
  );
}

export function WikiLinkPreview({ pagePath, isMissing, children }: Props) {
  const [open, setOpen] = useState(false);
  const [data, setData] = useState<WikiLinkPreviewData | null>(null);
  const [missing, setMissing] = useState(Boolean(isMissing));
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!open || isMissing) return;

    const cached = wikiLinkPreviewCache.get(pagePath);
    if (cached === "missing") {
      setMissing(true);
      setData(null);
      setLoading(false);
      return;
    }
    if (cached) {
      setMissing(false);
      setData(cached);
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    wikiLinkPreviewCache
      .load(pagePath)
      .then((result) => {
        if (cancelled) return;
        if (result === "missing") {
          setMissing(true);
          setData(null);
        } else {
          setMissing(false);
          setData(result);
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [open, pagePath, isMissing]);

  return (
    <HoverCard open={open} onOpenChange={setOpen} openDelay={300} closeDelay={100}>
      <HoverCardTrigger asChild>{children}</HoverCardTrigger>
      <HoverCardContent
        className="max-w-[320px] p-0"
        onClick={(e: React.MouseEvent) => e.preventDefault()}
      >
        {loading && !data && !missing ? (
          <PreviewSkeleton />
        ) : (
          <PreviewBody data={data} missing={missing} />
        )}
      </HoverCardContent>
    </HoverCard>
  );
}

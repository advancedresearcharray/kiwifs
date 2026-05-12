// Card grid view for the Bases component.

import { cn } from "@kw/lib/cn";
import { Badge } from "@kw/components/ui/badge";
import type { ViewColumn, ViewRow } from "./BasesTable";

type Props = {
  data: ViewRow[];
  columns: ViewColumn[];
  onNavigate: (path: string) => void;
  cardSize: "small" | "medium" | "large";
};

const GRID_CLASS: Record<Props["cardSize"], string> = {
  small: "grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5",
  medium: "grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4",
  large: "grid-cols-1 sm:grid-cols-2 lg:grid-cols-3",
};

export function BasesCards({ data, columns, onNavigate, cardSize }: Props) {
  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-muted-foreground text-sm">
        No results
      </div>
    );
  }

  // Determine which columns to show (skip path and title, they're rendered separately)
  const propCols = columns.filter((c) => c.key !== "path" && c.key !== "title");

  return (
    <div className={cn("grid gap-3 p-4", GRID_CLASS[cardSize])}>
      {data.map((row) => {
        const cover =
          (row.cover as string) || (row.image as string) || null;
        const tags = Array.isArray(row.tags)
          ? (row.tags as string[])
          : typeof row.tags === "string"
            ? (row.tags as string).split(",").map((s) => s.trim())
            : [];
        const excerpt =
          typeof row.excerpt === "string" ? row.excerpt : "";

        return (
          <button
            key={row.path}
            type="button"
            onClick={() => onNavigate(row.path)}
            className="text-left border border-border rounded-lg overflow-hidden bg-card hover:bg-accent/50 transition-colors focus-visible:ring-2 focus-visible:ring-ring"
          >
            {cover && (
              <div className="w-full h-32 bg-muted overflow-hidden">
                <img
                  src={cover}
                  alt=""
                  className="w-full h-full object-cover"
                  loading="lazy"
                />
              </div>
            )}
            <div className="p-3 space-y-1.5">
              <div className="font-medium text-sm truncate">
                {row.title}
              </div>
              {excerpt && (
                <div className="text-xs text-muted-foreground line-clamp-2">
                  {excerpt}
                </div>
              )}
              {tags.length > 0 && (
                <div className="flex flex-wrap gap-1">
                  {tags.slice(0, 5).map((t) => (
                    <Badge
                      key={t}
                      variant="secondary"
                      className="text-[10px] px-1 py-0"
                    >
                      {t}
                    </Badge>
                  ))}
                </div>
              )}
              {propCols.length > 0 && cardSize !== "small" && (
                <div className="text-xs text-muted-foreground space-y-0.5 pt-1 border-t border-border/50">
                  {propCols.slice(0, 3).map((col) => {
                    const v = row[col.key];
                    if (v == null) return null;
                    return (
                      <div key={col.key} className="flex justify-between">
                        <span>{col.label}</span>
                        <span className="truncate ml-2">
                          {Array.isArray(v) ? v.join(", ") : String(v)}
                        </span>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </button>
        );
      })}
    </div>
  );
}

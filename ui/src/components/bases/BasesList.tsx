// Simple bulleted list view for the Bases component.

import { File } from "lucide-react";
import type { ViewColumn, ViewRow } from "./BasesTable";

type Props = {
  data: ViewRow[];
  columns: ViewColumn[];
  onNavigate: (path: string) => void;
};

export function BasesList({ data, columns, onNavigate }: Props) {
  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-muted-foreground text-sm">
        No results
      </div>
    );
  }

  const propCols = columns.filter((c) => c.key !== "path" && c.key !== "title");

  return (
    <ul className="divide-y divide-border/50">
      {data.map((row) => (
        <li key={row.path}>
          <button
            type="button"
            onClick={() => onNavigate(row.path)}
            className="w-full text-left px-4 py-2.5 hover:bg-accent/50 transition-colors flex items-start gap-2"
          >
            <File className="h-4 w-4 text-muted-foreground shrink-0 mt-0.5" />
            <div className="min-w-0 flex-1">
              <div className="font-medium text-sm">{row.title}</div>
              {propCols.length > 0 && (
                <div className="text-xs text-muted-foreground flex flex-wrap gap-x-3 gap-y-0.5 mt-0.5">
                  {propCols.map((col) => {
                    const v = row[col.key];
                    if (v == null) return null;
                    return (
                      <span key={col.key}>
                        {col.label}: {Array.isArray(v) ? v.join(", ") : String(v)}
                      </span>
                    );
                  })}
                </div>
              )}
            </div>
          </button>
        </li>
      ))}
    </ul>
  );
}

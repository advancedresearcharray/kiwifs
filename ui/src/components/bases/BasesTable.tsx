// TanStack Table wrapper for the Bases "Table" layout.
// Handles column sorting, resizing, row selection, and footer summaries.

import { useMemo, useState } from "react";
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  flexRender,
  type ColumnDef,
  type SortingState,
} from "@tanstack/react-table";
import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
import { cn } from "@kw/lib/cn";
import { ScrollArea } from "@kw/components/ui/scroll-area";

export type ViewColumn = {
  key: string;
  label: string;
  summary?: "sum" | "avg" | "count" | "min" | "max";
};

export type ViewRow = Record<string, unknown> & { path: string; title: string };

type Props = {
  columns: ViewColumn[];
  data: ViewRow[];
  onNavigate: (path: string) => void;
};

function computeSummary(
  rows: ViewRow[],
  key: string,
  kind: ViewColumn["summary"],
): string {
  if (!kind) return "";
  const nums = rows
    .map((r) => Number(r[key]))
    .filter((n) => !isNaN(n));
  if (nums.length === 0) return "-";
  switch (kind) {
    case "count":
      return String(nums.length);
    case "sum":
      return nums.reduce((a, b) => a + b, 0).toLocaleString();
    case "avg":
      return (nums.reduce((a, b) => a + b, 0) / nums.length).toFixed(2);
    case "min":
      return Math.min(...nums).toLocaleString();
    case "max":
      return Math.max(...nums).toLocaleString();
    default:
      return "";
  }
}

export function BasesTable({ columns, data, onNavigate }: Props) {
  const [sorting, setSorting] = useState<SortingState>([]);

  const tableColumns = useMemo<ColumnDef<ViewRow>[]>(
    () =>
      columns.map((col) => ({
        accessorKey: col.key,
        header: col.label,
        cell: ({ getValue }) => {
          const v = getValue();
          if (v == null) return <span className="text-muted-foreground">-</span>;
          if (Array.isArray(v)) return v.join(", ");
          return String(v);
        },
        footer: col.summary
          ? () => (
              <span className="text-xs text-muted-foreground">
                {col.summary}: {computeSummary(data, col.key, col.summary)}
              </span>
            )
          : undefined,
        enableSorting: true,
        enableResizing: true,
        size: 160,
        minSize: 80,
      })),
    [columns, data],
  );

  const table = useReactTable({
    data,
    columns: tableColumns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    columnResizeMode: "onChange",
  });

  const hasSummary = columns.some((c) => c.summary);

  return (
    <ScrollArea className="h-full">
      <div className="min-w-full">
        <table className="w-full text-sm border-collapse">
          <thead className="sticky top-0 bg-card z-10">
            {table.getHeaderGroups().map((hg) => (
              <tr key={hg.id}>
                {hg.headers.map((header) => (
                  <th
                    key={header.id}
                    className="px-3 py-2 text-left text-xs font-medium text-muted-foreground border-b border-border select-none relative"
                    style={{ width: header.getSize() }}
                  >
                    <button
                      type="button"
                      className="flex items-center gap-1 hover:text-foreground transition-colors"
                      onClick={header.column.getToggleSortingHandler()}
                    >
                      {flexRender(header.column.columnDef.header, header.getContext())}
                      {header.column.getIsSorted() === "asc" ? (
                        <ArrowUp className="h-3 w-3" />
                      ) : header.column.getIsSorted() === "desc" ? (
                        <ArrowDown className="h-3 w-3" />
                      ) : (
                        <ArrowUpDown className="h-3 w-3 opacity-30" />
                      )}
                    </button>
                    {/* Resize handle */}
                    <div
                      onMouseDown={header.getResizeHandler()}
                      onTouchStart={header.getResizeHandler()}
                      className={cn(
                        "absolute right-0 top-0 h-full w-1 cursor-col-resize select-none touch-none",
                        header.column.getIsResizing()
                          ? "bg-primary"
                          : "hover:bg-border",
                      )}
                    />
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {table.getRowModel().rows.map((row) => (
              <tr
                key={row.id}
                className="hover:bg-accent/50 cursor-pointer transition-colors"
                onClick={() => onNavigate(row.original.path)}
              >
                {row.getVisibleCells().map((cell) => (
                  <td
                    key={cell.id}
                    className="px-3 py-2 border-b border-border/50"
                    style={{ width: cell.column.getSize() }}
                  >
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </td>
                ))}
              </tr>
            ))}
            {table.getRowModel().rows.length === 0 && (
              <tr>
                <td
                  colSpan={columns.length}
                  className="px-3 py-8 text-center text-muted-foreground text-sm"
                >
                  No results
                </td>
              </tr>
            )}
          </tbody>
          {hasSummary && (
            <tfoot>
              {table.getFooterGroups().map((fg) => (
                <tr key={fg.id} className="bg-muted/30">
                  {fg.headers.map((header) => (
                    <td
                      key={header.id}
                      className="px-3 py-1.5 border-t border-border"
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(header.column.columnDef.footer, header.getContext())}
                    </td>
                  ))}
                </tr>
              ))}
            </tfoot>
          )}
        </table>
      </div>
    </ScrollArea>
  );
}

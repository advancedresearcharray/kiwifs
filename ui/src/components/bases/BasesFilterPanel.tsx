// Filter editor panel for the Bases component.

import { Plus, Trash2, X } from "lucide-react";
import { Button } from "@kw/components/ui/button";
import { Input } from "@kw/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@kw/components/ui/select";

export type FilterOp = "equals" | "contains" | "gt" | "lt" | "is_empty";

export type ViewFilter = {
  property: string;
  op: FilterOp;
  value: string;
};

type Props = {
  filters: ViewFilter[];
  onChange: (filters: ViewFilter[]) => void;
  conjunction: "and" | "or";
  onConjunctionChange: (c: "and" | "or") => void;
  properties: string[];
  onClose: () => void;
};

const OP_LABELS: Record<FilterOp, string> = {
  equals: "equals",
  contains: "contains",
  gt: "greater than",
  lt: "less than",
  is_empty: "is empty",
};

export function BasesFilterPanel({
  filters,
  onChange,
  conjunction,
  onConjunctionChange,
  properties,
  onClose,
}: Props) {
  const addFilter = () => {
    onChange([
      ...filters,
      { property: properties[0] || "title", op: "contains", value: "" },
    ]);
  };

  const updateFilter = (idx: number, partial: Partial<ViewFilter>) => {
    const next = filters.map((f, i) => (i === idx ? { ...f, ...partial } : f));
    onChange(next);
  };

  const removeFilter = (idx: number) => {
    onChange(filters.filter((_, i) => i !== idx));
  };

  return (
    <div className="border border-border rounded-lg bg-card p-3 space-y-2 shadow-md">
      <div className="flex items-center justify-between">
        <div className="text-xs font-medium">Filters</div>
        <Button variant="ghost" size="icon" className="h-6 w-6" onClick={onClose}>
          <X className="h-3.5 w-3.5" />
        </Button>
      </div>

      {filters.length > 1 && (
        <div className="flex items-center gap-1 text-xs">
          <span className="text-muted-foreground">Match</span>
          <Button
            variant={conjunction === "and" ? "secondary" : "ghost"}
            size="sm"
            className="h-5 px-2 text-xs"
            onClick={() => onConjunctionChange("and")}
          >
            AND
          </Button>
          <Button
            variant={conjunction === "or" ? "secondary" : "ghost"}
            size="sm"
            className="h-5 px-2 text-xs"
            onClick={() => onConjunctionChange("or")}
          >
            OR
          </Button>
        </div>
      )}

      <div className="space-y-2">
        {filters.map((f, i) => (
          <div key={i} className="flex items-center gap-1.5">
            <Select
              value={f.property}
              onValueChange={(v) => updateFilter(i, { property: v })}
            >
              <SelectTrigger className="h-7 w-28 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {properties.map((p) => (
                  <SelectItem key={p} value={p} className="text-xs">
                    {p}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select
              value={f.op}
              onValueChange={(v) => updateFilter(i, { op: v as FilterOp })}
            >
              <SelectTrigger className="h-7 w-28 text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {(Object.keys(OP_LABELS) as FilterOp[]).map((op) => (
                  <SelectItem key={op} value={op} className="text-xs">
                    {OP_LABELS[op]}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {f.op !== "is_empty" && (
              <Input
                value={f.value}
                onChange={(e) => updateFilter(i, { value: e.target.value })}
                className="h-7 w-28 text-xs"
                placeholder="Value..."
              />
            )}
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 shrink-0"
              onClick={() => removeFilter(i)}
            >
              <Trash2 className="h-3 w-3" />
            </Button>
          </div>
        ))}
      </div>

      <Button
        variant="ghost"
        size="sm"
        className="h-7 text-xs gap-1"
        onClick={addFilter}
      >
        <Plus className="h-3 w-3" /> Add filter
      </Button>
    </div>
  );
}

import { Plus, X } from "lucide-react";
import type { WorkflowDef } from "@kw/lib/api";
import { Button } from "@kw/components/ui/button";
import { Input } from "@kw/components/ui/input";

export type EditStateRow = WorkflowDef["states"][number] & { id: string; wip_limit?: number };

type Props = {
  rows: EditStateRow[];
  disabled: boolean;
  onAdd: () => void;
  onRemove: (id: string) => void;
  onNameChange: (id: string, name: string) => void;
  onColorChange: (id: string, color: string) => void;
  onWipLimitChange?: (id: string, limit: number | undefined) => void;
};

function parseWipLimitInput(value: string): number | undefined {
  if (!value) {
    return undefined;
  }
  const parsed = parseInt(value, 10);
  if (parsed <= 0) {
    return undefined;
  }
  return parsed;
}

export function ColumnRowsEditor({
  rows,
  disabled,
  onAdd,
  onRemove,
  onNameChange,
  onColorChange,
  onWipLimitChange,
}: Props) {
  return (
    <div className="space-y-3">
      {rows.map((row, index) => (
        <div key={row.id} className="flex items-center gap-2">
          <Input
            type="color"
            value={row.color}
            onChange={(event) => onColorChange(row.id, event.target.value)}
            disabled={disabled}
            className="h-9 w-12 shrink-0 cursor-pointer p-1"
            aria-label={`Column ${index + 1} color`}
          />
          <Input
            value={row.name}
            onChange={(event) => onNameChange(row.id, event.target.value)}
            placeholder={`Column ${index + 1}`}
            disabled={disabled}
          />
          {onWipLimitChange && (
            <Input
              type="number"
              min={0}
              value={row.wip_limit ?? ""}
              onChange={(event) => {
                onWipLimitChange(row.id, parseWipLimitInput(event.target.value));
              }}
              placeholder="WIP"
              disabled={disabled}
              className="h-9 w-16 shrink-0 text-xs"
              title="Work-in-progress limit (0 = unlimited)"
              aria-label={`WIP limit for ${row.name || `column ${index + 1}`}`}
            />
          )}
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => onRemove(row.id)}
            disabled={disabled || rows.length <= 1}
            aria-label={`Remove column ${row.name || index + 1}`}
          >
            <X className="h-3.5 w-3.5" />
          </Button>
        </div>
      ))}

      <Button type="button" variant="outline" size="sm" onClick={onAdd} disabled={disabled}>
        <Plus className="h-3.5 w-3.5" />
        Add column
      </Button>
    </div>
  );
}

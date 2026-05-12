// Sort editor panel for the Bases component.
// Supports multiple sort keys with drag-to-reorder via @dnd-kit.

import {
  DndContext,
  closestCenter,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
  arrayMove,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical, Plus, Trash2, X } from "lucide-react";
import { Button } from "@kw/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@kw/components/ui/select";

export type SortKey = {
  id: string;
  property: string;
  direction: "asc" | "desc";
};

type Props = {
  sorts: SortKey[];
  onChange: (sorts: SortKey[]) => void;
  properties: string[];
  onClose: () => void;
};

let _sortCounter = 0;

function SortItem({
  sort,
  onUpdate,
  onRemove,
  properties,
}: {
  sort: SortKey;
  onUpdate: (partial: Partial<SortKey>) => void;
  onRemove: () => void;
  properties: string[];
}) {
  const { attributes, listeners, setNodeRef, transform, transition } =
    useSortable({ id: sort.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  return (
    <div ref={setNodeRef} style={style} className="flex items-center gap-1.5">
      <button
        type="button"
        {...attributes}
        {...listeners}
        className="cursor-grab active:cursor-grabbing text-muted-foreground"
      >
        <GripVertical className="h-3.5 w-3.5" />
      </button>
      <Select
        value={sort.property}
        onValueChange={(v) => onUpdate({ property: v })}
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
        value={sort.direction}
        onValueChange={(v) => onUpdate({ direction: v as "asc" | "desc" })}
      >
        <SelectTrigger className="h-7 w-20 text-xs">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="asc" className="text-xs">Asc</SelectItem>
          <SelectItem value="desc" className="text-xs">Desc</SelectItem>
        </SelectContent>
      </Select>
      <Button variant="ghost" size="icon" className="h-6 w-6 shrink-0" onClick={onRemove}>
        <Trash2 className="h-3 w-3" />
      </Button>
    </div>
  );
}

export function BasesSortPanel({ sorts, onChange, properties, onClose }: Props) {
  const addSort = () => {
    onChange([
      ...sorts,
      { id: `sort-${++_sortCounter}`, property: properties[0] || "title", direction: "asc" },
    ]);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const oldIdx = sorts.findIndex((s) => s.id === active.id);
    const newIdx = sorts.findIndex((s) => s.id === over.id);
    if (oldIdx === -1 || newIdx === -1) return;
    onChange(arrayMove(sorts, oldIdx, newIdx));
  };

  return (
    <div className="border border-border rounded-lg bg-card p-3 space-y-2 shadow-md">
      <div className="flex items-center justify-between">
        <div className="text-xs font-medium">Sort</div>
        <Button variant="ghost" size="icon" className="h-6 w-6" onClick={onClose}>
          <X className="h-3.5 w-3.5" />
        </Button>
      </div>

      <DndContext collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
        <SortableContext items={sorts} strategy={verticalListSortingStrategy}>
          <div className="space-y-1.5">
            {sorts.map((s, i) => (
              <SortItem
                key={s.id}
                sort={s}
                onUpdate={(partial) => {
                  const next = sorts.map((ss, j) =>
                    j === i ? { ...ss, ...partial } : ss,
                  );
                  onChange(next);
                }}
                onRemove={() => onChange(sorts.filter((_, j) => j !== i))}
                properties={properties}
              />
            ))}
          </div>
        </SortableContext>
      </DndContext>

      <Button
        variant="ghost"
        size="sm"
        className="h-7 text-xs gap-1"
        onClick={addSort}
      >
        <Plus className="h-3 w-3" /> Add sort
      </Button>
    </div>
  );
}

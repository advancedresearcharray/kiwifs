// Single Kanban column — wraps children in a droppable sortable context.

import { useDroppable } from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import type { ReactNode } from "react";

type Props = {
  id: string;
  state: string;
  color: string;
  count: number;
  children: ReactNode;
  items: string[];
};

export function KanbanColumn({ id, state, color, count, children, items }: Props) {
  const { setNodeRef, isOver } = useDroppable({ id });

  return (
    <div
      className={`flex flex-col w-72 shrink-0 rounded-lg border border-border bg-card/50 ${
        isOver ? "ring-2 ring-primary/50" : ""
      }`}
    >
      {/* Column header */}
      <div className="flex items-center gap-2 px-3 py-2.5 border-b border-border/50">
        <span
          className="h-2.5 w-2.5 rounded-full shrink-0"
          style={{ backgroundColor: color }}
        />
        <span className="font-medium text-sm">{state}</span>
        <span className="ml-auto text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded-full">
          {count}
        </span>
      </div>

      {/* Droppable area */}
      <div
        ref={setNodeRef}
        className="flex-1 overflow-auto kiwi-scroll p-2 space-y-1.5 min-h-[120px]"
      >
        <SortableContext items={items} strategy={verticalListSortingStrategy}>
          {children}
        </SortableContext>
      </div>
    </div>
  );
}

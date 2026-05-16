import { Plus } from "lucide-react";
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
  wipLimit?: number;
  onAdd?: (state: string) => void;
};

export function KanbanColumn({ id, state, color, count, children, items, wipLimit, onAdd }: Props) {
  const { setNodeRef, isOver } = useDroppable({ id });
  const isOverLimit = wipLimit != null && wipLimit > 0 && count > wipLimit;
  const isAtLimit = wipLimit != null && wipLimit > 0 && count === wipLimit;

  return (
    <div
      className={`flex flex-col min-w-[18rem] max-w-[18rem] h-fit rounded-lg border bg-muted py-2 pl-2 pr-1 transition-colors ${
        isOver
          ? "border-primary/40 bg-primary/[0.04]"
          : isOverLimit
            ? "border-destructive/40"
            : "border-border/70"
      }`}
    >
      <div className="mb-2 flex items-center justify-between px-2">
        <div className="flex items-center gap-2 min-w-0">
          <span
            className="h-2 w-2 rounded-full shrink-0"
            style={{ backgroundColor: color }}
          />
          <span className="font-medium text-sm truncate">{state}</span>
          <span className={`ml-0.5 rounded-full px-1.5 py-px text-[11px] tabular-nums ${
            isOverLimit
              ? "bg-destructive/10 text-destructive font-medium"
              : isAtLimit
                ? "bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300 font-medium"
                : "bg-background/60 text-muted-foreground"
          }`}>
            {wipLimit != null && wipLimit > 0 ? `${count}/${wipLimit}` : count}
          </span>
        </div>
        {onAdd && (
          <button
            type="button"
            onClick={() => onAdd(state)}
            className="inline-flex h-6 w-6 items-center justify-center rounded-md text-muted-foreground/70 hover:text-foreground hover:bg-accent transition-colors"
            aria-label={`Add card to ${state}`}
          >
            <Plus className="h-3.5 w-3.5" />
          </button>
        )}
      </div>

      <div
        ref={setNodeRef}
        className="flex-1 overflow-auto kiwi-scroll pr-1 space-y-2 min-h-[2rem] max-h-[calc(100vh-225px)]"
      >
        <SortableContext items={items} strategy={verticalListSortingStrategy}>
          {children}
        </SortableContext>
      </div>
    </div>
  );
}

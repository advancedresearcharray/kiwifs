import { Plus } from "lucide-react";
import { useDroppable } from "@dnd-kit/core";
import {
  SortableContext,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import type { ReactNode } from "react";
import { useKanbanStore } from "./kanbanStore";

type Props = {
  id: string;
  state: string;
  color: string;
  count: number;
  children: ReactNode;
  items: string[];
  wipLimit?: number;
};

function getColumnClassName(isOver: boolean, isOverLimit: boolean): string {
  const classes = [
    "flex flex-col min-w-[18rem] max-w-[18rem] h-fit rounded-lg border bg-muted py-2 pl-2 pr-1 transition-colors",
  ];
  if (isOver) {
    classes.push("border-primary/40 bg-primary/[0.04]");
    return classes.join(" ");
  }
  if (isOverLimit) {
    classes.push("border-destructive/40");
    return classes.join(" ");
  }
  classes.push("border-border/70");
  return classes.join(" ");
}

function getWipBadgeClassName(isOverLimit: boolean, isAtLimit: boolean): string {
  const classes = ["ml-0.5 rounded-full px-1.5 py-px text-[11px] tabular-nums"];
  if (isOverLimit) {
    classes.push("bg-destructive/10 text-destructive font-medium");
    return classes.join(" ");
  }
  if (isAtLimit) {
    classes.push("bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300 font-medium");
    return classes.join(" ");
  }
  classes.push("bg-background/60 text-muted-foreground");
  return classes.join(" ");
}

function formatWipCount(count: number, wipLimit: number | undefined): string | number {
  if (wipLimit != null && wipLimit > 0) {
    return `${count}/${wipLimit}`;
  }
  return count;
}

export function KanbanColumn({ id, state, color, count, children, items, wipLimit }: Props) {
  const { setNodeRef, isOver } = useDroppable({ id });
  const openAddCard = useKanbanStore((kanban) => kanban.openAddCard);
  const isOverLimit = wipLimit != null && wipLimit > 0 && count > wipLimit;
  const isAtLimit = wipLimit != null && wipLimit > 0 && count === wipLimit;

  return (
    <div
      className={getColumnClassName(isOver, isOverLimit)}
    >
      <div className="mb-2 flex items-center justify-between px-2">
        <div className="flex items-center gap-2 min-w-0">
          <span
            className="h-2 w-2 rounded-full shrink-0"
            style={{ backgroundColor: color }}
          />
          <span className="font-medium text-sm truncate">{state}</span>
          <span className={getWipBadgeClassName(isOverLimit, isAtLimit)}>
            {formatWipCount(count, wipLimit)}
          </span>
        </div>
        <button
          type="button"
          onClick={() => openAddCard(state)}
          className="inline-flex h-6 w-6 items-center justify-center rounded-md text-muted-foreground/70 hover:text-foreground hover:bg-accent transition-colors"
          aria-label={`Add card to ${state}`}
        >
          <Plus className="h-3.5 w-3.5" />
        </button>
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

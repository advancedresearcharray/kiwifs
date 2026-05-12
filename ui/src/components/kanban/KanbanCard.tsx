// Draggable Kanban card using @dnd-kit/sortable.

import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { formatDistanceToNow, parseISO } from "date-fns";
import { Badge } from "@kw/components/ui/badge";
import type { WorkflowPage } from "@kw/lib/api";

type Props = {
  page: WorkflowPage;
  onNavigate: (path: string) => void;
};

export function KanbanCard({ page, onNavigate }: Props) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: page.path });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      className="border border-border rounded-md bg-card p-2.5 cursor-grab active:cursor-grabbing hover:border-primary/40 transition-colors shadow-sm"
    >
      <button
        type="button"
        className="text-left w-full"
        onClick={(e) => {
          e.stopPropagation();
          onNavigate(page.path);
        }}
        onPointerDown={(e) => e.stopPropagation()}
      >
        <div className="font-medium text-sm truncate">{page.title}</div>
      </button>
      <div className="flex flex-wrap items-center gap-1.5 mt-1.5">
        {page.tags?.slice(0, 3).map((t) => (
          <Badge key={t} variant="secondary" className="text-[10px] px-1 py-0">
            {t}
          </Badge>
        ))}
      </div>
      <div className="flex items-center gap-2 text-[10px] text-muted-foreground mt-1.5">
        {page.author && <span>{page.author}</span>}
        {page.modified && (
          <span>
            {formatDistanceToNow(parseISO(page.modified), { addSuffix: true })}
          </span>
        )}
      </div>
    </div>
  );
}

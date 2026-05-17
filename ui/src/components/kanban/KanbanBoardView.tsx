import { DragOverlay } from "@dnd-kit/core";
import { KanbanCard } from "./KanbanCard";
import { KanbanColumn } from "./KanbanColumn";
import { useKanbanStore } from "./kanbanStore";
import type { WorkflowPage } from "@kw/lib/api";

type Props = {
  onNavigate: (path: string) => void;
};

function KanbanDragOverlayCard({ page }: { page: WorkflowPage | null }) {
  if (!page) {
    return null;
  }
  return (
    <div className="min-w-[18rem] max-w-[18rem] rounded-lg border border-border/40 bg-card px-3 py-2.5 shadow-xl shadow-black/10 dark:shadow-black/30 rotate-[2deg]">
      <span className="break-words text-[13px] leading-snug">{page.title}</span>
    </div>
  );
}

export function KanbanBoardView({ onNavigate }: Props) {
  const columns = useKanbanStore((state) => state.columns);
  const draggingPage = useKanbanStore((state) => state.draggingPage);

  return (
    <>
      <div className="flex px-6 pt-6 pb-4">
        <div className="min-w-[2rem]" />
        {columns.map((col) => (
          <div key={col.state} className="mr-5">
            <KanbanColumn
              id={col.state}
              state={col.state}
              color={col.color}
              count={col.pages.length}
              items={col.pages.map((page) => page.path)}
              wipLimit={col.wip_limit}
            >
              {col.pages.map((page) => (
                <KanbanCard key={page.path} page={page} onNavigate={onNavigate} />
              ))}
            </KanbanColumn>
          </div>
        ))}
        <div className="min-w-[0.75rem]" />
      </div>

      <DragOverlay>
        <KanbanDragOverlayCard page={draggingPage} />
      </DragOverlay>
    </>
  );
}

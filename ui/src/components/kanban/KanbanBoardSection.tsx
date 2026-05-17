import { AlertTriangle } from "lucide-react";
import { KanbanBoardView } from "./KanbanBoardView";
import { getBoardViewState, getUnmatchedPagesMessage, useKanbanStore } from "./kanbanStore";

type Props = {
  onNavigate: (path: string) => void;
};

/** Renders Kanban board warnings and board state from feature store state. */
export function KanbanBoardSection({ onNavigate }: Props) {
  const workflows = useKanbanStore((state) => state.workflows);
  const columns = useKanbanStore((state) => state.columns);
  const unmatchedPages = useKanbanStore((state) => state.unmatchedPages);
  const loading = useKanbanStore((state) => state.loading);
  const boardError = useKanbanStore((state) => state.boardError);
  const loadErrors = useKanbanStore((state) => state.loadErrors);
  const boardViewState = getBoardViewState({ loading, boardError, columns, workflows });

  return (
    <>
      {loadErrors.length > 0 && (
        <div className="flex items-start gap-2 px-6 py-2 border-b border-border bg-amber-50 dark:bg-amber-950/30 text-amber-800 dark:text-amber-200 text-xs">
          <AlertTriangle className="h-3.5 w-3.5 mt-0.5 shrink-0" />
          <span>Some workflow files could not be loaded: {loadErrors.join("; ")}</span>
        </div>
      )}

      {unmatchedPages.length > 0 && (
        <div className="flex items-start gap-2 px-6 py-2 border-b border-border bg-amber-50 dark:bg-amber-950/30 text-amber-800 dark:text-amber-200 text-xs">
          <AlertTriangle className="h-3.5 w-3.5 mt-0.5 shrink-0" />
          <span>{getUnmatchedPagesMessage(unmatchedPages)}</span>
        </div>
      )}

      <div className="flex-1 overflow-y-hidden overflow-x-auto kiwi-board-scroll">
        {boardViewState.kind === "loading" && (
          <div className="flex px-6 pt-6 pb-4">
            <div className="min-w-[2rem]" />
            {[420, 240, 340].map((height, index) => (
              <div key={index} className="mr-5 flex flex-col min-w-[18rem] max-w-[18rem]">
                <div className="mb-3 h-7 w-24 animate-pulse rounded-md bg-muted" />
                <div className="flex flex-col gap-2">
                  {Array.from({ length: Math.ceil(height / 80) }).map((_, skeletonIndex) => (
                    <div key={skeletonIndex} className="h-16 animate-pulse rounded-lg bg-muted/70" />
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
        {boardViewState.kind === "error" && (
          <div className="flex items-center justify-center h-64 text-destructive text-sm">
            <AlertTriangle className="h-4 w-4 mr-2 shrink-0" />
            {boardViewState.message}
          </div>
        )}
        {boardViewState.kind === "empty" && (
          <div className="flex items-center justify-center h-64 text-muted-foreground text-sm">
            {boardViewState.message}
          </div>
        )}
        {boardViewState.kind === "board" && <KanbanBoardView onNavigate={onNavigate} />}
      </div>
    </>
  );
}

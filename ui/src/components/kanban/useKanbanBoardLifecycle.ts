import { useEffect } from "react";
import { useKanbanStore } from "./kanbanStore";

/**
 * Coordinates the board loading lifecycle for the Kanban screen.
 *
 * Note: `loadWorkflows` and `loadBoard` are Zustand store actions whose
 * references are stable across renders (guaranteed by createStore). The
 * effect deps are safe and will not cause infinite loops.
 */
export function useKanbanBoardLifecycle() {
  const activeWorkflow = useKanbanStore((state) => state.activeWorkflow);
  const loadWorkflows = useKanbanStore((state) => state.loadWorkflows);
  const loadBoard = useKanbanStore((state) => state.loadBoard);

  useEffect(() => {
    void loadWorkflows();
  }, [loadWorkflows]);

  useEffect(() => {
    if (activeWorkflow) void loadBoard(activeWorkflow);
  }, [activeWorkflow, loadBoard]);
}

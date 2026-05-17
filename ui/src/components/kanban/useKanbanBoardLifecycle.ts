import { useEffect } from "react";
import { useKanbanStore } from "./kanbanStore";

/** Coordinates the board loading lifecycle for the Kanban screen. */
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

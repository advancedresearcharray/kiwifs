import { useMemo } from "react";
import { useKanbanDragHandlers } from "./KanbanDragProvider";
import { useKanbanStore } from "./kanbanStore";

/** Registers store-backed drag handlers with the app-level Kanban drag provider. */
export function useKanbanDragRegistration() {
  const handleDragStart = useKanbanStore((state) => state.handleDragStart);
  const handleDragEnd = useKanbanStore((state) => state.handleDragEnd);

  const dragHandlers = useMemo(() => ({
    onDragStart: handleDragStart,
    onDragEnd: handleDragEnd,
  }), [handleDragStart, handleDragEnd]);

  useKanbanDragHandlers(dragHandlers);
}

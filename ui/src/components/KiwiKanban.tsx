// KiwiKanban — Drag-and-drop Kanban board showing pages grouped by workflow state.

import { KanbanAddCardDialog } from "./kanban/KanbanAddCardDialog";
import { KanbanBoardSection } from "./kanban/KanbanBoardSection";
import { KanbanToolbar } from "./kanban/KanbanToolbar";
import { CreateWorkflowDialog, DeleteWorkflowDialog, EditWorkflowDialog } from "./kanban/KanbanWorkflowDialogs";
import { useKanbanBoardLifecycle } from "./kanban/useKanbanBoardLifecycle";
import { useKanbanDragRegistration } from "./kanban/useKanbanDragRegistration";
import { KanbanStoreProvider } from "./kanban/kanbanStore";

type Props = {
  onClose: () => void;
  onNavigate: (path: string) => void;
};

/**
 * Provides feature-scoped Kanban state for one screen instance.
 *
 * @param props - Component props.
 * @param props.onClose - Called when the user leaves the Kanban screen.
 * @param props.onNavigate - Called to navigate to a page path.
 * @returns Kanban board screen.
 */
export function KiwiKanban(props: Props) {
  return (
    <KanbanStoreProvider>
      <KiwiKanbanContent {...props} />
    </KanbanStoreProvider>
  );
}

/**
 * Renders the Kanban screen shell and keeps application callbacks outside the store boundary.
 *
 * @param props - Component props.
 * @param props.onClose - Called when the user leaves the Kanban screen.
 * @param props.onNavigate - Called to navigate to a page path.
 * @returns Kanban board content.
 */
function KiwiKanbanContent({ onClose, onNavigate }: Props) {
  useKanbanBoardLifecycle();
  useKanbanDragRegistration();

  return (
    <div className="h-full flex flex-col">
      <KanbanToolbar onClose={onClose} />
      <KanbanBoardSection onNavigate={onNavigate} />
      <KanbanAddCardDialog onNavigate={onNavigate} />
      <CreateWorkflowDialog />
      <EditWorkflowDialog />
      <DeleteWorkflowDialog />
    </div>
  );
}

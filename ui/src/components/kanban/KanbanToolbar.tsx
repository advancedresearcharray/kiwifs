import { ArrowLeft, Pencil, Plus, Trash2 } from "lucide-react";
import { Button } from "@kw/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@kw/components/ui/select";
import { useKanbanStore } from "./kanbanStore";

type Props = {
  onClose: () => void;
};

/** Renders Kanban workflow controls while keeping app close behavior outside the store. */
export function KanbanToolbar({ onClose }: Props) {
  const workflows = useKanbanStore((state) => state.workflows);
  const activeWorkflow = useKanbanStore((state) => state.activeWorkflow);
  const setActiveWorkflow = useKanbanStore((state) => state.setActiveWorkflow);
  const openCreateWorkflow = useKanbanStore((state) => state.openCreateWorkflow);
  const openEditWorkflow = useKanbanStore((state) => state.openEditWorkflow);
  const openDeleteWorkflow = useKanbanStore((state) => state.openDeleteWorkflow);

  return (
    <div className="flex items-center gap-2 px-3 sm:px-6 py-3 border-b border-border bg-card">
      <Button variant="outline" size="sm" onClick={onClose}>
        <ArrowLeft className="h-3.5 w-3.5" />
        <span className="hidden sm:inline">Back</span>
      </Button>
      <div className="font-semibold text-sm">Kanban</div>

      {workflows.length > 0 && (
        <Select value={activeWorkflow || ""} onValueChange={setActiveWorkflow}>
          <SelectTrigger className="h-8 w-44 text-sm">
            <SelectValue placeholder="Select workflow" />
          </SelectTrigger>
          <SelectContent>
            {workflows.map((workflow) => (
              <SelectItem key={workflow.name} value={workflow.name}>
                {workflow.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}

      <Button variant="outline" size="sm" onClick={openCreateWorkflow}>
        <Plus className="h-3.5 w-3.5" />
        <span className="hidden sm:inline">New board</span>
      </Button>

      {activeWorkflow && (
        <Button variant="outline" size="sm" onClick={openEditWorkflow}>
          <Pencil className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Edit columns</span>
        </Button>
      )}

      {activeWorkflow && (
        <Button variant="outline" size="sm" onClick={openDeleteWorkflow}>
          <Trash2 className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Delete board</span>
        </Button>
      )}
    </div>
  );
}

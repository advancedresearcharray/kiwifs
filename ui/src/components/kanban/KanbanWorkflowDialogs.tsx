import { AlertTriangle, Loader2, Pencil, Plus, Trash2 } from "lucide-react";
import { Button } from "@kw/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@kw/components/ui/dialog";
import { Input } from "@kw/components/ui/input";
import { Label } from "@kw/components/ui/label";
import { ColumnRowsEditor } from "./ColumnRowsEditor";
import { useKanbanStore } from "./kanbanStore";

function CreateWorkflowButtonIcon({ creating }: { creating: boolean }) {
  if (creating) {
    return <Loader2 className="h-3.5 w-3.5 animate-spin" />;
  }
  return <Plus className="h-3.5 w-3.5" />;
}

function SaveWorkflowButtonIcon({ saving }: { saving: boolean }) {
  if (saving) {
    return <Loader2 className="h-3.5 w-3.5 animate-spin" />;
  }
  return <Pencil className="h-3.5 w-3.5" />;
}

function DeleteWorkflowButtonIcon({ deleting }: { deleting: boolean }) {
  if (deleting) {
    return <Loader2 className="h-3.5 w-3.5 animate-spin" />;
  }
  return <Trash2 className="h-3.5 w-3.5" />;
}

function formatCardCountText(cardCount: number): string {
  if (cardCount === 1) {
    return "1 card";
  }
  return `${cardCount} cards`;
}

export function CreateWorkflowDialog() {
  const open = useKanbanStore((state) => state.createOpen);
  const name = useKanbanStore((state) => state.newWorkflowName);
  const rows = useKanbanStore((state) => state.createRows);
  const error = useKanbanStore((state) => state.createError);
  const creating = useKanbanStore((state) => state.creating);
  const setOpen = useKanbanStore((state) => state.setCreateOpen);
  const setName = useKanbanStore((state) => state.setNewWorkflowName);
  const createWorkflow = useKanbanStore((state) => state.createWorkflow);
  const addRow = useKanbanStore((state) => state.addCreateRow);
  const removeRow = useKanbanStore((state) => state.removeCreateRow);
  const updateRowName = useKanbanStore((state) => state.updateCreateRowName);
  const updateRowColor = useKanbanStore((state) => state.updateCreateRowColor);
  const updateRowWipLimit = useKanbanStore((state) => state.updateCreateRowWipLimit);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create Kanban board</DialogTitle>
          <DialogDescription>
            Boards are saved as workflow JSON files under .kiwi/workflows. Markdown pages become cards only when their frontmatter references this workflow and a state.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="workflow-name">Board name</Label>
            <Input
              id="workflow-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter" && !creating) void createWorkflow();
              }}
              placeholder="e.g. content pipeline"
            />
          </div>

          <div className="space-y-2">
            <Label>Columns</Label>
            <ColumnRowsEditor
              rows={rows}
              disabled={creating}
              onAdd={addRow}
              onRemove={removeRow}
              onNameChange={updateRowName}
              onColorChange={updateRowColor}
              onWipLimitChange={updateRowWipLimit}
            />
            <p className="text-xs text-muted-foreground">
              Adjacent columns get two-way transitions. Pages become cards when their frontmatter uses this board name and one of these column names.
            </p>
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)} disabled={creating}>
            Cancel
          </Button>
          <Button onClick={() => void createWorkflow()} disabled={creating}>
            <CreateWorkflowButtonIcon creating={creating} />
            Create board
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function EditWorkflowDialog() {
  const open = useKanbanStore((state) => state.editOpen);
  const activeWorkflow = useKanbanStore((state) => state.activeWorkflow);
  const rows = useKanbanStore((state) => state.editRows);
  const error = useKanbanStore((state) => state.editError);
  const saving = useKanbanStore((state) => state.savingEdit);
  const setOpen = useKanbanStore((state) => state.setEditOpen);
  const saveWorkflow = useKanbanStore((state) => state.saveEditWorkflow);
  const addRow = useKanbanStore((state) => state.addEditRow);
  const removeRow = useKanbanStore((state) => state.removeEditRow);
  const updateRowName = useKanbanStore((state) => state.updateEditRowName);
  const updateRowColor = useKanbanStore((state) => state.updateEditRowColor);
  const updateRowWipLimit = useKanbanStore((state) => state.updateEditRowWipLimit);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Edit columns</DialogTitle>
          <DialogDescription>
            Add, remove, or rename columns for "{activeWorkflow}". This saves the workflow JSON and rebuilds adjacent two-way transitions. Existing card frontmatter is not rewritten automatically.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-3 py-2">
          <ColumnRowsEditor
            rows={rows}
            disabled={saving}
            onAdd={addRow}
            onRemove={removeRow}
            onNameChange={updateRowName}
            onColorChange={updateRowColor}
            onWipLimitChange={updateRowWipLimit}
          />

          <p className="text-xs text-muted-foreground">
            Renamed or removed columns may hide existing cards until those pages' frontmatter state values are updated.
          </p>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={() => void saveWorkflow()} disabled={saving}>
            <SaveWorkflowButtonIcon saving={saving} />
            Save columns
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function DeleteWorkflowDialog() {
  const open = useKanbanStore((state) => state.deleteOpen);
  const activeWorkflow = useKanbanStore((state) => state.activeWorkflow);
  const columns = useKanbanStore((state) => state.columns);
  const error = useKanbanStore((state) => state.deleteError);
  const deleting = useKanbanStore((state) => state.deleting);
  const setOpen = useKanbanStore((state) => state.setDeleteOpen);
  const deleteWorkflow = useKanbanStore((state) => state.deleteWorkflow);
  const cardCount = columns.reduce((n, column) => n + column.pages.length, 0);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Delete Kanban board</DialogTitle>
          <DialogDescription>
            Delete workflow JSON for "{activeWorkflow}". Existing markdown pages are not modified; cards that still reference this workflow will no longer appear until their frontmatter is changed.
          </DialogDescription>
        </DialogHeader>

        {cardCount > 0 && (
          <div className="flex items-start gap-2 rounded-md bg-amber-50 dark:bg-amber-950/30 text-amber-800 dark:text-amber-200 text-xs p-3">
            <AlertTriangle className="h-3.5 w-3.5 mt-0.5 shrink-0" />
            <span>
              This board has {formatCardCountText(cardCount)}.
              Their frontmatter will still reference "{activeWorkflow}" but the workflow definition will be gone. You'll need to manually update each page's frontmatter to remove the stale reference.
            </span>
          </div>
        )}

        {error && <p className="text-sm text-destructive">{error}</p>}

        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)} disabled={deleting}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={() => void deleteWorkflow()} disabled={deleting || !activeWorkflow}>
            <DeleteWorkflowButtonIcon deleting={deleting} />
            Delete board
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// KiwiKanban — Drag-and-drop Kanban board showing pages grouped by workflow state.

import { useCallback, useEffect, useState } from "react";
import {
  DndContext,
  DragOverlay,
  closestCorners,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import { ArrowLeft, Loader2, Pencil, Plus, Trash2, X } from "lucide-react";
import { api, type WorkflowColumn, type WorkflowDef, type WorkflowPage } from "@kw/lib/api";
import {
  createDefaultWorkflow,
  normalizeWorkflowName,
  parseWorkflowStates,
  updateWorkflowStates,
} from "@kw/lib/workflow";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@kw/components/ui/select";
import { Textarea } from "@kw/components/ui/textarea";
import { KanbanColumn } from "./kanban/KanbanColumn";
import { KanbanCard } from "./kanban/KanbanCard";

type Props = {
  onClose: () => void;
  onNavigate: (path: string) => void;
};

type Workflow = WorkflowDef;
type EditStateRow = WorkflowDef["states"][number] & { id: string };

const DEFAULT_WORKFLOW_STATES = "todo, doing, done";

function makeEditRows(workflow: WorkflowDef): EditStateRow[] {
  return workflow.states.map((state, index) => ({ ...state, id: `${state.name}-${index}` }));
}

export function KiwiKanban({ onClose, onNavigate }: Props) {
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [activeWorkflow, setActiveWorkflow] = useState<string | null>(null);
  const [columns, setColumns] = useState<WorkflowColumn[]>([]);
  const [loading, setLoading] = useState(true);
  const [draggingPage, setDraggingPage] = useState<WorkflowPage | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [newWorkflowName, setNewWorkflowName] = useState("");
  const [newWorkflowStates, setNewWorkflowStates] = useState(DEFAULT_WORKFLOW_STATES);
  const [createError, setCreateError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [editRows, setEditRows] = useState<EditStateRow[]>([]);
  const [editError, setEditError] = useState<string | null>(null);
  const [savingEdit, setSavingEdit] = useState(false);

  const activeWorkflowDef = workflows.find((workflow) => workflow.name === activeWorkflow) ?? null;

  const loadWorkflows = useCallback(async (preferredWorkflow?: string) => {
    setLoading(true);
    try {
      const result = await api.listWorkflows();
      const wfs = result.workflows || [];
      setWorkflows(wfs);

      if (preferredWorkflow && wfs.some((w) => w.name === preferredWorkflow)) {
        setActiveWorkflow(preferredWorkflow);
      } else if (wfs.length > 0) {
        setActiveWorkflow((current) =>
          current && wfs.some((w) => w.name === current) ? current : wfs[0]!.name,
        );
      } else {
        setActiveWorkflow(null);
        setColumns([]);
      }
    } catch {
      setWorkflows([]);
      setActiveWorkflow(null);
      setColumns([]);
    } finally {
      setLoading(false);
    }
  }, []);

  // Load workflows
  useEffect(() => {
    void loadWorkflows();
  }, [loadWorkflows]);

  // Load board
  const loadBoard = useCallback(async (name: string) => {
    setLoading(true);
    try {
      const result = await api.getWorkflowBoard(name);
      setColumns(result.columns || []);
    } catch {
      setColumns([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (activeWorkflow) loadBoard(activeWorkflow);
  }, [activeWorkflow, loadBoard]);

  // Find which column a page belongs to
  const findColumnForPage = (path: string): string | null => {
    for (const col of columns) {
      if (col.pages.some((p) => p.path === path)) return col.state;
    }
    return null;
  };

  // Drag handlers
  const handleDragStart = (event: DragStartEvent) => {
    const pageId = event.active.id as string;
    for (const col of columns) {
      const page = col.pages.find((p) => p.path === pageId);
      if (page) {
        setDraggingPage(page);
        break;
      }
    }
  };

  const handleDragEnd = async (event: DragEndEvent) => {
    setDraggingPage(null);
    const { active, over } = event;
    if (!over || !activeWorkflow) return;

    const pagePath = active.id as string;
    const sourceState = findColumnForPage(pagePath);
    // The over ID could be a column ID or another page's ID
    let targetState: string | null = null;

    // Check if dropped on a column
    const overCol = columns.find((c) => c.state === over.id);
    if (overCol) {
      targetState = overCol.state;
    } else {
      // Dropped on another page — find its column
      targetState = findColumnForPage(over.id as string);
    }

    if (!targetState || targetState === sourceState) return;

    // Optimistic update
    setColumns((prev) =>
      prev.map((col) => {
        if (col.state === sourceState) {
          return { ...col, pages: col.pages.filter((p) => p.path !== pagePath) };
        }
        if (col.state === targetState) {
          const page = prev
            .find((c) => c.state === sourceState)
            ?.pages.find((p) => p.path === pagePath);
          if (page) {
            return { ...col, pages: [...col.pages, page] };
          }
        }
        return col;
      }),
    );

    // Server call
    try {
      await api.advanceWorkflow(pagePath, activeWorkflow, targetState);
    } catch {
      // Revert on error
      loadBoard(activeWorkflow);
    }
  };

  const handleCreateWorkflow = async () => {
    const name = normalizeWorkflowName(newWorkflowName);
    const states = parseWorkflowStates(newWorkflowStates);

    if (!name) {
      setCreateError("Board name is required.");
      return;
    }
    if (workflows.some((w) => w.name === name)) {
      setCreateError(`Workflow "${name}" already exists.`);
      return;
    }
    if (states.length === 0) {
      setCreateError("Add at least one state.");
      return;
    }

    setCreating(true);
    setCreateError(null);
    try {
      await api.saveWorkflow(createDefaultWorkflow(name, states));
      setCreateOpen(false);
      setNewWorkflowName("");
      setNewWorkflowStates(DEFAULT_WORKFLOW_STATES);
      await loadWorkflows(name);
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : "Failed to create workflow.");
    } finally {
      setCreating(false);
    }
  };

  const handleDeleteWorkflow = async () => {
    if (!activeWorkflow) return;

    setDeleting(true);
    setDeleteError(null);
    try {
      await api.deleteWorkflow(activeWorkflow);
      setDeleteOpen(false);
      await loadWorkflows();
    } catch (err) {
      setDeleteError(err instanceof Error ? err.message : "Failed to delete workflow.");
    } finally {
      setDeleting(false);
    }
  };

  const handleOpenEdit = () => {
    if (!activeWorkflowDef) return;
    setEditRows(makeEditRows(activeWorkflowDef));
    setEditError(null);
    setEditOpen(true);
  };

  const handleSaveEdit = async () => {
    if (!activeWorkflowDef || !activeWorkflow) return;

    setSavingEdit(true);
    setEditError(null);
    try {
      const updatedWorkflow = updateWorkflowStates(
        activeWorkflowDef,
        editRows.map((row) => ({ name: row.name, color: row.color })),
      );
      await api.saveWorkflow(updatedWorkflow);
      setEditOpen(false);
      await loadWorkflows(activeWorkflow);
      await loadBoard(activeWorkflow);
    } catch (err) {
      setEditError(err instanceof Error ? err.message : "Failed to save columns.");
    } finally {
      setSavingEdit(false);
    }
  };

  const handleAddEditRow = () => {
    setEditRows((rows) => [
      ...rows,
      { id: `new-${Date.now()}`, name: "", color: "#9B59B6" },
    ]);
  };

  const handleRemoveEditRow = (id: string) => {
    setEditRows((rows) => rows.filter((row) => row.id !== id));
  };

  const handleEditRowName = (id: string, name: string) => {
    setEditRows((rows) => rows.map((row) => (row.id === id ? { ...row, name } : row)));
  };

  return (
    <div className="h-full flex flex-col">
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-3 sm:px-6 py-3 border-b border-border bg-card">
        <Button variant="outline" size="sm" onClick={onClose}>
          <ArrowLeft className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Back</span>
        </Button>
        <div className="font-semibold text-sm">Kanban</div>

        {workflows.length > 0 && (
          <Select
            value={activeWorkflow || ""}
            onValueChange={setActiveWorkflow}
          >
            <SelectTrigger className="h-8 w-44 text-sm">
              <SelectValue placeholder="Select workflow" />
            </SelectTrigger>
            <SelectContent>
              {workflows.map((w) => (
                <SelectItem key={w.name} value={w.name}>
                  {w.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}

        <Button variant="outline" size="sm" onClick={() => setCreateOpen(true)}>
          <Plus className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">New board</span>
        </Button>

        {activeWorkflow && (
          <Button
            variant="outline"
            size="sm"
            onClick={handleOpenEdit}
          >
            <Pencil className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">Edit columns</span>
          </Button>
        )}

        {activeWorkflow && (
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              setDeleteError(null);
              setDeleteOpen(true);
            }}
          >
            <Trash2 className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">Delete board</span>
          </Button>
        )}
      </div>

      {/* Board */}
      <div className="flex-1 overflow-auto">
        {loading ? (
          <div className="flex items-center justify-center h-64 text-muted-foreground">
            <Loader2 className="h-5 w-5 animate-spin mr-2" /> Loading...
          </div>
        ) : columns.length === 0 ? (
          <div className="flex items-center justify-center h-64 text-muted-foreground text-sm">
            {workflows.length === 0
              ? "No workflows configured. Create a board to add a workflow JSON file."
              : "No pages in this workflow yet."}
          </div>
        ) : (
          <DndContext
            collisionDetection={closestCorners}
            onDragStart={handleDragStart}
            onDragEnd={handleDragEnd}
          >
            <div className="flex gap-4 p-4 min-w-max">
              {columns.map((col) => (
                <KanbanColumn
                  key={col.state}
                  id={col.state}
                  state={col.state}
                  color={col.color}
                  count={col.pages.length}
                  items={col.pages.map((p) => p.path)}
                >
                  {col.pages.map((page) => (
                    <KanbanCard
                      key={page.path}
                      page={page}
                      onNavigate={onNavigate}
                    />
                  ))}
                </KanbanColumn>
              ))}
            </div>

            <DragOverlay>
              {draggingPage ? (
                <div className="w-72 border border-primary rounded-md bg-card p-2.5 shadow-lg opacity-90">
                  <div className="font-medium text-sm truncate">
                    {draggingPage.title}
                  </div>
                </div>
              ) : null}
            </DragOverlay>
          </DndContext>
        )}
      </div>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
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
                value={newWorkflowName}
                onChange={(event) => setNewWorkflowName(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" && !creating) void handleCreateWorkflow();
                }}
                placeholder="e.g. content pipeline"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="workflow-states">States</Label>
              <Textarea
                id="workflow-states"
                value={newWorkflowStates}
                onChange={(event) => setNewWorkflowStates(event.target.value)}
                placeholder="todo, doing, done"
                rows={3}
              />
              <p className="text-xs text-muted-foreground">
                Separate states with commas or new lines. Adjacent states get two-way transitions.
              </p>
            </div>

            {createError && <p className="text-sm text-destructive">{createError}</p>}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)} disabled={creating}>
              Cancel
            </Button>
            <Button onClick={() => void handleCreateWorkflow()} disabled={creating}>
              {creating ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
              Create board
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Edit columns</DialogTitle>
            <DialogDescription>
              Add, remove, or rename columns for "{activeWorkflow}". This saves the workflow JSON and rebuilds adjacent two-way transitions. Existing card frontmatter is not rewritten automatically.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-3 py-2">
            {editRows.map((row, index) => (
              <div key={row.id} className="flex items-center gap-2">
                <div
                  className="h-4 w-4 rounded-full border border-border shrink-0"
                  style={{ backgroundColor: row.color }}
                  aria-hidden="true"
                />
                <Input
                  value={row.name}
                  onChange={(event) => handleEditRowName(row.id, event.target.value)}
                  placeholder={`Column ${index + 1}`}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  onClick={() => handleRemoveEditRow(row.id)}
                  disabled={savingEdit || editRows.length <= 1}
                  aria-label={`Remove column ${row.name || index + 1}`}
                >
                  <X className="h-3.5 w-3.5" />
                </Button>
              </div>
            ))}

            <Button type="button" variant="outline" size="sm" onClick={handleAddEditRow} disabled={savingEdit}>
              <Plus className="h-3.5 w-3.5" />
              Add column
            </Button>

            <p className="text-xs text-muted-foreground">
              Renamed or removed columns may hide existing cards until those pages' frontmatter state values are updated.
            </p>
            {editError && <p className="text-sm text-destructive">{editError}</p>}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditOpen(false)} disabled={savingEdit}>
              Cancel
            </Button>
            <Button onClick={() => void handleSaveEdit()} disabled={savingEdit}>
              {savingEdit ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Pencil className="h-3.5 w-3.5" />}
              Save columns
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Delete Kanban board</DialogTitle>
            <DialogDescription>
              Delete workflow JSON for "{activeWorkflow}". Existing markdown pages are not modified; cards that still reference this workflow will no longer appear until their frontmatter is changed.
            </DialogDescription>
          </DialogHeader>

          {deleteError && <p className="text-sm text-destructive">{deleteError}</p>}

          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteOpen(false)} disabled={deleting}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={() => void handleDeleteWorkflow()} disabled={deleting || !activeWorkflow}>
              {deleting ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Trash2 className="h-3.5 w-3.5" />}
              Delete board
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

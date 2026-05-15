// KiwiKanban — Drag-and-drop Kanban board showing pages grouped by workflow state.

import { useCallback, useEffect, useState } from "react";
import {
  DndContext,
  DragOverlay,
  closestCorners,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import { ArrowLeft, Loader2, Plus } from "lucide-react";
import { api, type WorkflowColumn, type WorkflowDef, type WorkflowPage } from "@kw/lib/api";
import { createDefaultWorkflow, normalizeWorkflowName, parseWorkflowStates } from "@kw/lib/workflow";
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

const DEFAULT_WORKFLOW_STATES = "todo, doing, done";

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
    </div>
  );
}

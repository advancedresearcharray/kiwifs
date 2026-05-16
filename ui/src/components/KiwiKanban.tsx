// KiwiKanban — Drag-and-drop Kanban board showing pages grouped by workflow state.

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  DragOverlay,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import { AlertTriangle, ArrowLeft, Loader2, Pencil, Plus, Trash2, X } from "lucide-react";
import { api, type SearchResult, type WorkflowColumn, type WorkflowDef, type WorkflowPage } from "@kw/lib/api";
import {
  createKanbanCardMarkdown,
  defaultKanbanCardPath,
  createDefaultWorkflow,
  normalizeWorkflowName,
  updateWorkflowStates,
} from "@kw/lib/workflow";
import { getKanbanDragData, isKanbanCardDragData, isTreePageDragData } from "@kw/lib/kanbanDnd";
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
import { Textarea } from "@kw/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@kw/components/ui/select";
import { KanbanColumn } from "./kanban/KanbanColumn";
import { KanbanCard } from "./kanban/KanbanCard";
import { useKanbanDragHandlers } from "./kanban/KanbanDragProvider";

type Props = {
  onClose: () => void;
  onNavigate: (path: string) => void;
};

type Workflow = WorkflowDef;
type EditStateRow = WorkflowDef["states"][number] & { id: string; wip_limit?: number };
type AddCardMode = "new" | "existing";

type AddCardDialogState = {
  open: boolean;
  state: string;
  mode: AddCardMode;
  title: string;
  path: string;
  body: string;
  query: string;
  results: SearchResult[];
  error: string | null;
  busy: boolean;
};

const DEFAULT_WORKFLOW_STATES = ["todo", "doing", "done"];
const emptyAddCardDialog: AddCardDialogState = {
  open: false,
  state: "",
  mode: "new",
  title: "",
  path: "",
  body: "",
  query: "",
  results: [],
  error: null,
  busy: false,
};

function makeEditRows(workflow: WorkflowDef): EditStateRow[] {
  return workflow.states.map((state, index) => ({
    ...state,
    id: `${state.name}-${index}`,
    wip_limit: state.wip_limit,
  }));
}

function makeDefaultRows(): EditStateRow[] {
  return makeEditRows(createDefaultWorkflow("", DEFAULT_WORKFLOW_STATES));
}

type ColumnRowsEditorProps = {
  rows: EditStateRow[];
  disabled: boolean;
  onAdd: () => void;
  onRemove: (id: string) => void;
  onNameChange: (id: string, name: string) => void;
  onColorChange: (id: string, color: string) => void;
  onWipLimitChange?: (id: string, limit: number | undefined) => void;
};

function ColumnRowsEditor({
  rows,
  disabled,
  onAdd,
  onRemove,
  onNameChange,
  onColorChange,
  onWipLimitChange,
}: ColumnRowsEditorProps) {
  return (
    <div className="space-y-3">
      {rows.map((row, index) => (
        <div key={row.id} className="flex items-center gap-2">
          <Input
            type="color"
            value={row.color}
            onChange={(event) => onColorChange(row.id, event.target.value)}
            disabled={disabled}
            className="h-9 w-12 shrink-0 cursor-pointer p-1"
            aria-label={`Column ${index + 1} color`}
          />
          <Input
            value={row.name}
            onChange={(event) => onNameChange(row.id, event.target.value)}
            placeholder={`Column ${index + 1}`}
            disabled={disabled}
          />
          {onWipLimitChange && (
            <Input
              type="number"
              min={0}
              value={row.wip_limit ?? ""}
              onChange={(event) => {
                const v = event.target.value ? parseInt(event.target.value, 10) : undefined;
                onWipLimitChange(row.id, v && v > 0 ? v : undefined);
              }}
              placeholder="WIP"
              disabled={disabled}
              className="h-9 w-16 shrink-0 text-xs"
              title="Work-in-progress limit (0 = unlimited)"
              aria-label={`WIP limit for ${row.name || `column ${index + 1}`}`}
            />
          )}
          <Button
            type="button"
            variant="ghost"
            size="icon"
            onClick={() => onRemove(row.id)}
            disabled={disabled || rows.length <= 1}
            aria-label={`Remove column ${row.name || index + 1}`}
          >
            <X className="h-3.5 w-3.5" />
          </Button>
        </div>
      ))}

      <Button type="button" variant="outline" size="sm" onClick={onAdd} disabled={disabled}>
        <Plus className="h-3.5 w-3.5" />
        Add column
      </Button>
    </div>
  );
}

export function KiwiKanban({ onClose, onNavigate }: Props) {
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [activeWorkflow, setActiveWorkflow] = useState<string | null>(null);
  const [columns, setColumns] = useState<WorkflowColumn[]>([]);
  const [unmatchedPages, setUnmatchedPages] = useState<WorkflowPage[]>([]);
  const [loading, setLoading] = useState(true);
  const [boardError, setBoardError] = useState<string | null>(null);
  const [loadErrors, setLoadErrors] = useState<string[]>([]);
  const [draggingPage, setDraggingPage] = useState<WorkflowPage | null>(null);
  const [createOpen, setCreateOpen] = useState(false);
  const [newWorkflowName, setNewWorkflowName] = useState("");
  const [createRows, setCreateRows] = useState<EditStateRow[]>(() => makeDefaultRows());
  const [createError, setCreateError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const [editRows, setEditRows] = useState<EditStateRow[]>([]);
  const [editError, setEditError] = useState<string | null>(null);
  const [savingEdit, setSavingEdit] = useState(false);
  const [addCard, setAddCard] = useState<AddCardDialogState>(emptyAddCardDialog);

  const activeWorkflowDef = workflows.find((workflow) => workflow.name === activeWorkflow) ?? null;

  const loadWorkflows = useCallback(async (preferredWorkflow?: string) => {
    setLoading(true);
    setBoardError(null);
    try {
      const result = await api.listWorkflows();
      const wfs = result.workflows || [];
      setWorkflows(wfs);
      setLoadErrors(result.errors ?? []);

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
    } catch (err) {
      setWorkflows([]);
      setActiveWorkflow(null);
      setColumns([]);
      setBoardError(err instanceof Error ? err.message : "Failed to load workflows.");
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
    setBoardError(null);
    try {
      const result = await api.getWorkflowBoard(name);
      setColumns(result.columns || []);
      setUnmatchedPages(result.unmatchedPages ?? []);
    } catch (err) {
      setColumns([]);
      setUnmatchedPages([]);
      setBoardError(err instanceof Error ? err.message : "Failed to load board.");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (activeWorkflow) loadBoard(activeWorkflow);
  }, [activeWorkflow, loadBoard]);

  // Find which column a page belongs to
  const findColumnForPage = useCallback((path: string): string | null => {
    for (const col of columns) {
      if (col.pages.some((p) => p.path === path)) return col.state;
    }
    return null;
  }, [columns]);

  const findTargetState = useCallback((overId: string): string | null => {
    const overCol = columns.find((c) => c.state === overId);
    if (overCol) return overCol.state;
    return findColumnForPage(overId);
  }, [columns, findColumnForPage]);

  // Drag handlers
  const handleDragStart = useCallback((event: DragStartEvent) => {
    const dragData = getKanbanDragData(event.active.data.current);
    if (isTreePageDragData(dragData)) {
      setDraggingPage({ path: dragData.path, title: dragData.title });
      return;
    }

    if (!isKanbanCardDragData(dragData)) return;
    for (const col of columns) {
      const page = col.pages.find((p) => p.path === dragData.path);
      if (page) {
        setDraggingPage(page);
        break;
      }
    }
  }, [columns]);

  // Compute the ordinal midpoint between two neighbours for within-column
  // reorder. Uses the ordinalStep constant (1000) as default spacing.
  const computeOrdinal = useCallback(
    (colPages: WorkflowPage[], targetIndex: number): number => {
      const prev = targetIndex > 0 ? (colPages[targetIndex - 1]?.ordinal ?? (targetIndex - 1) * 1000) : 0;
      const next =
        targetIndex < colPages.length
          ? (colPages[targetIndex]?.ordinal ?? (targetIndex) * 1000 + 1000)
          : prev + 1000;
      return Math.round((prev + next) / 2);
    },
    [],
  );

  const handleDragEnd = useCallback(async (event: DragEndEvent) => {
    setDraggingPage(null);
    const { active, over } = event;
    if (!over || !activeWorkflow) return;

    const dragData = getKanbanDragData(active.data.current);
    if (!dragData) return;

    const pagePath = dragData.path;
    const sourceState = isKanbanCardDragData(dragData) ? findColumnForPage(pagePath) : null;
    const targetState = findTargetState(String(over.id));

    if (!targetState) return;

    // Within-column reorder: same column, different position.
    if (targetState === sourceState && isKanbanCardDragData(dragData)) {
      const col = columns.find((c) => c.state === targetState);
      if (!col) return;
      const oldIndex = col.pages.findIndex((p) => p.path === pagePath);
      const overPath = String(over.id);
      let newIndex = col.pages.findIndex((p) => p.path === overPath);
      if (newIndex === -1) newIndex = col.pages.length - 1;
      if (oldIndex === newIndex) return;

      // Optimistic reorder
      const reordered = [...col.pages];
      const [moved] = reordered.splice(oldIndex, 1);
      if (!moved) return;
      reordered.splice(newIndex, 0, moved);
      setColumns((prev) =>
        prev.map((c) => (c.state === targetState ? { ...c, pages: reordered } : c)),
      );

      // Persist the new position via ordinal.
      const ordinal = computeOrdinal(
        reordered.filter((p) => p.path !== pagePath),
        newIndex,
      );
      try {
        await api.reorderCard(pagePath, ordinal);
      } catch {
        await loadBoard(activeWorkflow);
      }
      return;
    }

    if (targetState === sourceState) return;

    if (isTreePageDragData(dragData)) {
      try {
        await api.assignWorkflow(pagePath, activeWorkflow, targetState);
        await loadBoard(activeWorkflow);
      } catch {
        await loadBoard(activeWorkflow);
      }
      return;
    }

    // Optimistic update for cross-column move
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
      await loadBoard(activeWorkflow);
    }
  }, [activeWorkflow, columns, findColumnForPage, findTargetState, loadBoard, computeOrdinal]);

  const dragHandlers = useMemo(() => ({
    onDragStart: handleDragStart,
    onDragEnd: handleDragEnd,
  }), [handleDragStart, handleDragEnd]);

  useKanbanDragHandlers(dragHandlers);

  const handleCreateWorkflow = async () => {
    const name = normalizeWorkflowName(newWorkflowName);

    if (!name) {
      setCreateError("Board name is required.");
      return;
    }
    if (workflows.some((w) => w.name === name)) {
      setCreateError(`Workflow "${name}" already exists.`);
      return;
    }

    setCreating(true);
    setCreateError(null);
    try {
      const workflow = updateWorkflowStates(
        { name, states: [], transitions: [] },
        createRows.map((row) => ({ name: row.name, color: row.color, ...(row.wip_limit ? { wip_limit: row.wip_limit } : {}) })),
      );
      await api.saveWorkflow(workflow);
      setCreateOpen(false);
      setNewWorkflowName("");
      setCreateRows(makeDefaultRows());
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

  const handleOpenCreate = () => {
    setNewWorkflowName("");
    setCreateRows(makeDefaultRows());
    setCreateError(null);
    setCreateOpen(true);
  };

  const handleSaveEdit = async () => {
    if (!activeWorkflowDef || !activeWorkflow) return;

    setSavingEdit(true);
    setEditError(null);
    try {
      const updatedWorkflow = updateWorkflowStates(
        activeWorkflowDef,
        editRows.map((row) => ({ name: row.name, color: row.color, ...(row.wip_limit ? { wip_limit: row.wip_limit } : {}) })),
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

  const handleAddCreateRow = () => {
    setCreateRows((rows) => [
      ...rows,
      { id: `new-${Date.now()}`, name: "", color: "#9B59B6" },
    ]);
  };

  const handleRemoveCreateRow = (id: string) => {
    setCreateRows((rows) => rows.filter((row) => row.id !== id));
  };

  const handleCreateRowName = (id: string, name: string) => {
    setCreateRows((rows) => rows.map((row) => (row.id === id ? { ...row, name } : row)));
  };

  const handleCreateRowColor = (id: string, color: string) => {
    setCreateRows((rows) => rows.map((row) => (row.id === id ? { ...row, color } : row)));
  };

  const handleCreateRowWipLimit = (id: string, wip_limit: number | undefined) => {
    setCreateRows((rows) => rows.map((row) => (row.id === id ? { ...row, wip_limit } : row)));
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

  const handleEditRowColor = (id: string, color: string) => {
    setEditRows((rows) => rows.map((row) => (row.id === id ? { ...row, color } : row)));
  };

  const handleEditRowWipLimit = (id: string, wip_limit: number | undefined) => {
    setEditRows((rows) => rows.map((row) => (row.id === id ? { ...row, wip_limit } : row)));
  };

  const handleOpenAddCard = (state: string) => {
    if (!activeWorkflow) return;
    const title = "";
    setAddCard({
      ...emptyAddCardDialog,
      open: true,
      state,
      title,
      path: "",
    });
  };

  const handleNewCardTitleChange = (title: string) => {
    setAddCard((current) => ({
      ...current,
      title,
      path: current.path && current.path !== defaultKanbanCardPath(current.title, activeWorkflow || "kanban")
        ? current.path
        : defaultKanbanCardPath(title, activeWorkflow || "kanban"),
    }));
  };

  const handleCreateCard = async () => {
    if (!activeWorkflow) return;
    const title = addCard.title.trim();
    const path = addCard.path.trim();
    if (!title) {
      setAddCard((current) => ({ ...current, error: "Card title is required." }));
      return;
    }
    if (!path.endsWith(".md")) {
      setAddCard((current) => ({ ...current, error: "Card path must end with .md." }));
      return;
    }

    setAddCard((current) => ({ ...current, busy: true, error: null }));
    try {
      await api.writeFile(
        path,
        createKanbanCardMarkdown({
          title,
          workflow: activeWorkflow,
          state: addCard.state,
          body: addCard.body,
        }),
      );
      setAddCard(emptyAddCardDialog);
      await loadBoard(activeWorkflow);
      onNavigate(path);
    } catch (err) {
      setAddCard((current) => ({
        ...current,
        busy: false,
        error: err instanceof Error ? err.message : "Failed to create card.",
      }));
    }
  };

  const handleSearchExistingPages = async () => {
    const query = addCard.query.trim();
    if (!query) {
      setAddCard((current) => ({ ...current, error: "Search query is required." }));
      return;
    }
    setAddCard((current) => ({ ...current, busy: true, error: null }));
    try {
      const response = await api.search(query);
      setAddCard((current) => ({
        ...current,
        busy: false,
        results: (response.results || []).filter((result) => result.path.endsWith(".md")),
      }));
    } catch (err) {
      setAddCard((current) => ({
        ...current,
        busy: false,
        error: err instanceof Error ? err.message : "Failed to search pages.",
      }));
    }
  };

  const handleAssignExistingPage = async (path: string) => {
    if (!activeWorkflow) return;
    setAddCard((current) => ({ ...current, busy: true, error: null }));
    try {
      await api.assignWorkflow(path, activeWorkflow, addCard.state);
      setAddCard(emptyAddCardDialog);
      await loadBoard(activeWorkflow);
    } catch (err) {
      setAddCard((current) => ({
        ...current,
        busy: false,
        error: err instanceof Error ? err.message : "Failed to add page to board.",
      }));
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

        <Button variant="outline" size="sm" onClick={handleOpenCreate}>
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

      {/* Broken workflow file warnings */}
      {loadErrors.length > 0 && (
        <div className="flex items-start gap-2 px-6 py-2 border-b border-border bg-amber-50 dark:bg-amber-950/30 text-amber-800 dark:text-amber-200 text-xs">
          <AlertTriangle className="h-3.5 w-3.5 mt-0.5 shrink-0" />
          <span>Some workflow files could not be loaded: {loadErrors.join("; ")}</span>
        </div>
      )}

      {/* Cards with unrecognized states */}
      {unmatchedPages.length > 0 && (
        <div className="flex items-start gap-2 px-6 py-2 border-b border-border bg-amber-50 dark:bg-amber-950/30 text-amber-800 dark:text-amber-200 text-xs">
          <AlertTriangle className="h-3.5 w-3.5 mt-0.5 shrink-0" />
          <span>
            {unmatchedPages.length} card{unmatchedPages.length > 1 ? "s have" : " has"} a state
            that doesn't match any column ({unmatchedPages.map((p) => p.title || p.path).join(", ")}).
            Edit their frontmatter or add the missing column.
          </span>
        </div>
      )}

      {/* Board */}
      <div className="flex-1 overflow-y-hidden overflow-x-auto kiwi-board-scroll">
        {loading ? (
          <div className="flex px-6 pt-6 pb-4">
            <div className="min-w-[2rem]" />
            {[420, 240, 340].map((h, i) => (
              <div key={i} className="mr-5 flex flex-col min-w-[18rem] max-w-[18rem]">
                <div className="mb-3 h-7 w-24 animate-pulse rounded-md bg-muted" />
                <div className="flex flex-col gap-2">
                  {Array.from({ length: Math.ceil(h / 80) }).map((_, j) => (
                    <div key={j} className="h-16 animate-pulse rounded-lg bg-muted/70" />
                  ))}
                </div>
              </div>
            ))}
          </div>
        ) : boardError ? (
          /* Error state */
          <div className="flex items-center justify-center h-64 text-destructive text-sm">
            <AlertTriangle className="h-4 w-4 mr-2 shrink-0" />
            {boardError}
          </div>
        ) : columns.length === 0 ? (
          <div className="flex items-center justify-center h-64 text-muted-foreground text-sm">
            {workflows.length === 0
              ? "No workflows configured. Create a board to add a workflow JSON file."
              : "No pages in this workflow yet."}
          </div>
        ) : (
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
                    items={col.pages.map((p) => p.path)}
                    wipLimit={col.wip_limit}
                    onAdd={handleOpenAddCard}
                  >
                    {col.pages.map((page) => (
                      <KanbanCard
                        key={page.path}
                        page={page}
                        onNavigate={onNavigate}
                      />
                    ))}
                  </KanbanColumn>
                </div>
              ))}
              <div className="min-w-[0.75rem]" />
            </div>

            <DragOverlay>
              {draggingPage ? (
                <div className="min-w-[18rem] max-w-[18rem] rounded-lg border border-border/40 bg-card px-3 py-2.5 shadow-xl shadow-black/10 dark:shadow-black/30 rotate-[2deg]">
                  <span className="break-words text-[13px] leading-snug">
                    {draggingPage.title}
                  </span>
                </div>
              ) : null}
            </DragOverlay>
          </>
        )}
      </div>

      <Dialog open={addCard.open} onOpenChange={(open) => setAddCard((current) => ({ ...current, open }))}>
        <DialogContent className="w-[calc(100vw-2rem)] overflow-hidden sm:max-w-lg">
          <DialogHeader className="min-w-0">
            <DialogTitle>Add card to {addCard.state}</DialogTitle>
            <DialogDescription>
              Create a new markdown page or attach an existing page by setting its workflow/state frontmatter.
            </DialogDescription>
          </DialogHeader>

          <div className="min-w-0 space-y-4 py-2">
            <div className="flex min-w-0 gap-2">
              <Button
                type="button"
                size="sm"
                variant={addCard.mode === "new" ? "default" : "outline"}
                onClick={() => setAddCard((current) => ({ ...current, mode: "new", error: null }))}
                disabled={addCard.busy}
              >
                New card
              </Button>
              <Button
                type="button"
                size="sm"
                variant={addCard.mode === "existing" ? "default" : "outline"}
                onClick={() => setAddCard((current) => ({ ...current, mode: "existing", error: null }))}
                disabled={addCard.busy}
              >
                Add existing page
              </Button>
            </div>

            {addCard.mode === "new" ? (
              <div className="space-y-3">
                <div className="space-y-2">
                  <Label htmlFor="kanban-card-title">Title</Label>
                  <Input
                    id="kanban-card-title"
                    value={addCard.title}
                    onChange={(event) => handleNewCardTitleChange(event.target.value)}
                    placeholder="e.g. Draft launch note"
                    disabled={addCard.busy}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="kanban-card-path">Path</Label>
                  <Input
                    id="kanban-card-path"
                    value={addCard.path}
                    onChange={(event) => setAddCard((current) => ({ ...current, path: event.target.value }))}
                    placeholder="tasks/draft-launch-note.md"
                    disabled={addCard.busy}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="kanban-card-body">Body</Label>
                  <Textarea
                    id="kanban-card-body"
                    value={addCard.body}
                    onChange={(event) => setAddCard((current) => ({ ...current, body: event.target.value }))}
                    placeholder="Optional notes..."
                    disabled={addCard.busy}
                    rows={4}
                  />
                </div>
              </div>
            ) : (
              <div className="min-w-0 space-y-3">
                <div className="flex min-w-0 gap-2">
                  <Input
                    className="min-w-0 flex-1"
                    value={addCard.query}
                    onChange={(event) => setAddCard((current) => ({ ...current, query: event.target.value }))}
                    onKeyDown={(event) => {
                      if (event.key === "Enter" && !addCard.busy) void handleSearchExistingPages();
                    }}
                    placeholder="Search markdown pages"
                    disabled={addCard.busy}
                  />
                  <Button type="button" variant="outline" className="shrink-0" onClick={() => void handleSearchExistingPages()} disabled={addCard.busy}>
                    Search
                  </Button>
                </div>
                <div className="max-h-56 w-full min-w-0 overflow-auto rounded-md border border-border divide-y divide-border/50">
                  {addCard.results.length === 0 ? (
                    <div className="p-3 text-sm text-muted-foreground">No search results yet.</div>
                  ) : (
                    addCard.results.map((result) => (
                      <div key={result.path} className="flex min-w-0 items-start gap-2 p-2">
                        <div className="min-w-0 flex-1 overflow-hidden">
                          <div className="truncate text-sm font-medium" title={result.path}>{result.path}</div>
                          {result.snippet && <div className="line-clamp-2 break-words text-xs text-muted-foreground">{result.snippet}</div>}
                        </div>
                        <Button className="shrink-0" size="sm" onClick={() => void handleAssignExistingPage(result.path)} disabled={addCard.busy}>
                          Add
                        </Button>
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}

            {addCard.error && <p className="text-sm text-destructive">{addCard.error}</p>}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setAddCard(emptyAddCardDialog)} disabled={addCard.busy}>
              Cancel
            </Button>
            {addCard.mode === "new" && (
              <Button onClick={() => void handleCreateCard()} disabled={addCard.busy}>
                {addCard.busy ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Plus className="h-3.5 w-3.5" />}
                Create card
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

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
              <Label>Columns</Label>
              <ColumnRowsEditor
                rows={createRows}
                disabled={creating}
                onAdd={handleAddCreateRow}
                onRemove={handleRemoveCreateRow}
                onNameChange={handleCreateRowName}
                onColorChange={handleCreateRowColor}
                onWipLimitChange={handleCreateRowWipLimit}
              />
              <p className="text-xs text-muted-foreground">
                Adjacent columns get two-way transitions. Pages become cards when their frontmatter uses this board name and one of these column names.
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
            <ColumnRowsEditor
              rows={editRows}
              disabled={savingEdit}
              onAdd={handleAddEditRow}
              onRemove={handleRemoveEditRow}
              onNameChange={handleEditRowName}
              onColorChange={handleEditRowColor}
              onWipLimitChange={handleEditRowWipLimit}
            />

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

          {/* Orphan warning when board has cards */}
          {columns.some((c) => c.pages.length > 0) && (
            <div className="flex items-start gap-2 rounded-md bg-amber-50 dark:bg-amber-950/30 text-amber-800 dark:text-amber-200 text-xs p-3">
              <AlertTriangle className="h-3.5 w-3.5 mt-0.5 shrink-0" />
              <span>
                This board has {columns.reduce((n, c) => n + c.pages.length, 0)} card{columns.reduce((n, c) => n + c.pages.length, 0) > 1 ? "s" : ""}.
                Their frontmatter will still reference "{activeWorkflow}" but the workflow definition will be gone. You'll need to manually update each page's frontmatter to remove the stale reference.
              </span>
            </div>
          )}

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

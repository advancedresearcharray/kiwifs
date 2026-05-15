export type KanbanCardDragData = {
  type: "kanban-card";
  path: string;
};

export type TreePageDragData = {
  type: "tree-page";
  path: string;
  title: string;
};

export type KanbanDragData = KanbanCardDragData | TreePageDragData;

export function createKanbanCardDragData(path: string): KanbanCardDragData {
  return { type: "kanban-card", path };
}

export function createTreePageDragData(path: string, title: string): TreePageDragData {
  return { type: "tree-page", path, title };
}

export function isKanbanCardDragData(value: unknown): value is KanbanCardDragData {
  return isRecord(value)
    && value.type === "kanban-card"
    && typeof value.path === "string";
}

export function isTreePageDragData(value: unknown): value is TreePageDragData {
  return isRecord(value)
    && value.type === "tree-page"
    && typeof value.path === "string"
    && typeof value.title === "string";
}

export function getKanbanDragData(value: unknown): KanbanDragData | null {
  if (isKanbanCardDragData(value) || isTreePageDragData(value)) return value;
  return null;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

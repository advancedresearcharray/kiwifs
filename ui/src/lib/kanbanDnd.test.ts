import { describe, expect, it } from "vitest";
import {
  createKanbanCardDragData,
  createTreePageDragData,
  getKanbanDragData,
  isKanbanCardDragData,
  isTreePageDragData,
} from "./kanbanDnd";

describe("kanbanDnd", () => {
  it("creates and narrows kanban card drag data", () => {
    const data = createKanbanCardDragData("tasks/a.md");

    expect(isKanbanCardDragData(data)).toBe(true);
    expect(isTreePageDragData(data)).toBe(false);
    expect(getKanbanDragData(data)).toEqual({ type: "kanban-card", path: "tasks/a.md" });
  });

  it("creates and narrows tree page drag data", () => {
    const data = createTreePageDragData("notes/idea.md", "Idea");

    expect(isTreePageDragData(data)).toBe(true);
    expect(isKanbanCardDragData(data)).toBe(false);
    expect(getKanbanDragData(data)).toEqual({ type: "tree-page", path: "notes/idea.md", title: "Idea" });
  });

  it("rejects malformed drag data", () => {
    expect(getKanbanDragData(null)).toBeNull();
    expect(getKanbanDragData({ type: "tree-page", path: "notes/idea.md" })).toBeNull();
    expect(getKanbanDragData({ type: "kanban-card", path: 42 })).toBeNull();
  });
});

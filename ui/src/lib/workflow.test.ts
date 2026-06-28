import { describe, expect, it } from "vitest";
import {
  addWorkflowState,
  createDefaultWorkflow,
  createKanbanCardMarkdown,
  defaultKanbanCardPath,
  deleteWorkflowState,
  normalizeWorkflowName,
  parseWorkflowStates,
  renameWorkflowState,
  updateWorkflowStates,
} from "./workflow";

describe("workflow helpers", () => {
  it("normalizes board names without changing user-visible spaces", () => {
    expect(normalizeWorkflowName("  새로운 칸반  ")).toBe("새로운 칸반");
    expect(normalizeWorkflowName("content/pipeline\\draft")).toBe("content-pipeline-draft");
  });

  it("parses comma or newline separated states and removes duplicates", () => {
    expect(parseWorkflowStates("todo, doing\ndone, doing, ")).toEqual(["todo", "doing", "done"]);
  });

  it("creates all-to-all transitions for a default board workflow", () => {
    expect(createDefaultWorkflow("새로운 칸반", ["todo", "doing", "done"])).toEqual({
      name: "새로운 칸반",
      states: [
        { name: "todo", color: "#9B59B6" },
        { name: "doing", color: "#3498DB" },
        { name: "done", color: "#2ECC71" },
      ],
      transitions: [
        { from: "todo", to: "doing" },
        { from: "todo", to: "done" },
        { from: "doing", to: "todo" },
        { from: "doing", to: "done" },
        { from: "done", to: "todo" },
        { from: "done", to: "doing" },
      ],
    });
  });

  it("creates a board workflow from editable column rows", () => {
    expect(
      updateWorkflowStates(
        { name: "launch", states: [], transitions: [] },
        [
          { name: " idea ", color: "#111111" },
          { name: "shipping", color: "#222222" },
        ],
      ),
    ).toEqual({
      name: "launch",
      states: [
        { name: "idea", color: "#111111" },
        { name: "shipping", color: "#222222" },
      ],
      transitions: [
        { from: "idea", to: "shipping" },
        { from: "shipping", to: "idea" },
      ],
    });
  });

  it("adds a state and rebuilds all-to-all transitions", () => {
    const workflow = createDefaultWorkflow("content", ["todo", "done"]);

    expect(addWorkflowState(workflow, "review")).toEqual({
      name: "content",
      states: [
        { name: "todo", color: "#9B59B6" },
        { name: "done", color: "#3498DB" },
        { name: "review", color: "#2ECC71" },
      ],
      transitions: [
        { from: "todo", to: "done" },
        { from: "todo", to: "review" },
        { from: "done", to: "todo" },
        { from: "done", to: "review" },
        { from: "review", to: "todo" },
        { from: "review", to: "done" },
      ],
    });
  });

  it("renames a state while preserving its color and position", () => {
    const workflow = createDefaultWorkflow("content", ["todo", "doing", "done"]);

    expect(renameWorkflowState(workflow, "doing", "review")).toEqual({
      name: "content",
      states: [
        { name: "todo", color: "#9B59B6" },
        { name: "review", color: "#3498DB" },
        { name: "done", color: "#2ECC71" },
      ],
      transitions: [
        { from: "todo", to: "review" },
        { from: "todo", to: "done" },
        { from: "review", to: "todo" },
        { from: "review", to: "done" },
        { from: "done", to: "todo" },
        { from: "done", to: "review" },
      ],
    });
  });

  it("deletes a state and reconnects transitions", () => {
    const workflow = createDefaultWorkflow("content", ["todo", "review", "done"]);

    expect(deleteWorkflowState(workflow, "review")).toEqual({
      name: "content",
      states: [
        { name: "todo", color: "#9B59B6" },
        { name: "done", color: "#2ECC71" },
      ],
      transitions: [
        { from: "todo", to: "done" },
        { from: "done", to: "todo" },
      ],
    });
  });

  it("builds markdown for a new Kanban card", () => {
    expect(
      createKanbanCardMarkdown({
        title: 'Fix "billing" flow',
        workflow: "앞으로할일",
        state: "todo",
        body: "Check the retry path.",
      }),
    ).toBe(
      '---\ntitle: "Fix \\"billing\\" flow"\nworkflow: "앞으로할일"\nstate: "todo"\n---\n# Fix "billing" flow\n\nCheck the retry path.\n',
    );
  });

  it("builds markdown with tags, priority, and due date", () => {
    expect(
      createKanbanCardMarkdown({
        title: "Upgrade deps",
        workflow: "sprint",
        state: "todo",
        tags: ["chore", "deps"],
        priority: "high",
        due: "2025-06-01",
      }),
    ).toBe(
      '---\ntitle: "Upgrade deps"\nworkflow: "sprint"\nstate: "todo"\ntags: ["chore", "deps"]\npriority: "high"\ndue: "2025-06-01"\n---\n# Upgrade deps\n',
    );
  });

  it("omits optional fields when not provided", () => {
    const md = createKanbanCardMarkdown({
      title: "Simple card",
      workflow: "board",
      state: "doing",
    });
    expect(md).not.toContain("tags:");
    expect(md).not.toContain("priority:");
    expect(md).not.toContain("due:");
  });

  it("creates a safe default path for a new Kanban card", () => {
    expect(defaultKanbanCardPath("Bug: 로그인/결제 flow!", "tasks")).toBe(
      "tasks/bug-로그인-결제-flow.md",
    );
    expect(defaultKanbanCardPath("   ", "tasks")).toMatch(/^tasks\/card-\d+\.md$/);
  });
});

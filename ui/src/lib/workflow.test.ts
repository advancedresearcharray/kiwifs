import { describe, expect, it } from "vitest";
import {
  addWorkflowState,
  createDefaultWorkflow,
  deleteWorkflowState,
  normalizeWorkflowName,
  parseWorkflowStates,
  renameWorkflowState,
} from "./workflow";

describe("workflow helpers", () => {
  it("normalizes board names without changing user-visible spaces", () => {
    expect(normalizeWorkflowName("  새로운 칸반  ")).toBe("새로운 칸반");
    expect(normalizeWorkflowName("content/pipeline\\draft")).toBe("content-pipeline-draft");
  });

  it("parses comma or newline separated states and removes duplicates", () => {
    expect(parseWorkflowStates("todo, doing\ndone, doing, ")).toEqual(["todo", "doing", "done"]);
  });

  it("creates adjacent two-way transitions for a default board workflow", () => {
    expect(createDefaultWorkflow("새로운 칸반", ["todo", "doing", "done"])).toEqual({
      name: "새로운 칸반",
      states: [
        { name: "todo", color: "#9B59B6" },
        { name: "doing", color: "#3498DB" },
        { name: "done", color: "#2ECC71" },
      ],
      transitions: [
        { from: "todo", to: "doing" },
        { from: "doing", to: "todo" },
        { from: "doing", to: "done" },
        { from: "done", to: "doing" },
      ],
    });
  });

  it("adds a state and rebuilds adjacent transitions", () => {
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
        { from: "done", to: "todo" },
        { from: "done", to: "review" },
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
        { from: "review", to: "todo" },
        { from: "review", to: "done" },
        { from: "done", to: "review" },
      ],
    });
  });

  it("deletes a state and reconnects adjacent transitions", () => {
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
});

import { describe, expect, it } from "vitest";
import { createDefaultWorkflow, normalizeWorkflowName, parseWorkflowStates } from "./workflow";

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
});

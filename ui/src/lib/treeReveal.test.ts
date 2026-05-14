import { describe, expect, it } from "vitest";
import { nextExpandedForReveal, shouldFocusRevealTarget } from "./treeReveal";

describe("tree reveal helpers", () => {
  it("expands root and every parent folder for a reveal target without mutating the previous set", () => {
    const previous = new Set(["existing"]);

    const next = nextExpandedForReveal(previous, "notes/projects/today.md");

    expect([...next].sort()).toEqual(["", "existing", "notes", "notes/projects"].sort());
    expect([...previous]).toEqual(["existing"]);
    expect(next).not.toBe(previous);
  });

  it("does nothing when no reveal path is available", () => {
    const previous = new Set(["existing"]);

    expect(nextExpandedForReveal(previous, null)).toBe(previous);
    expect(nextExpandedForReveal(previous, "")).toBe(previous);
  });

  it("matches a node only when the reveal request path equals the node path", () => {
    const request = { path: "notes/today.md", nonce: 2 };

    expect(shouldFocusRevealTarget(request, "notes/today.md")).toBe(true);
    expect(shouldFocusRevealTarget(request, "notes")).toBe(false);
    expect(shouldFocusRevealTarget(null, "notes/today.md")).toBe(false);
  });
});

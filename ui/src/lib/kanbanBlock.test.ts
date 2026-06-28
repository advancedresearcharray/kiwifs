import { describe, expect, it } from "vitest";
import { buildKanbanBlockExportText, parseKanbanBlockConfig } from "./kanbanBlock";

describe("parseKanbanBlockConfig", () => {
  it("parses JSON config", () => {
    const config = parseKanbanBlockConfig(JSON.stringify({
      title: "Sprint",
      columns: [{ name: "Now", cards: [{ id: "a", title: "Ship" }] }],
    }));

    expect(config.title).toBe("Sprint");
    expect(config.columns[0]?.cards[0]).toMatchObject({ id: "a", title: "Ship" });
  });

  it("parses YAML-like columns, card fields, tags, and export config", () => {
    const config = parseKanbanBlockConfig(`
title: "Sprint Planning"
columns:
  - name: Now
    color: "#22c55e"
    cards:
      - id: auth
        title: Fix auth token refresh
        description: Refresh before expiry
        priority: critical
        assignee: cinos
        tags: [backend, critical]
  - name: Next
    color: "#3b82f6"
    cards:
      - id: search
        title: Add semantic search
        tags: [backend, feature]
export:
  format: json
  copyLabel: Copy JSON
`);

    expect(config.title).toBe("Sprint Planning");
    expect(config.export).toEqual({ format: "json", copyLabel: "Copy JSON" });
    expect(config.columns).toHaveLength(2);
    expect(config.columns[0]).toMatchObject({ name: "Now", color: "#22c55e" });
    expect(config.columns[0]?.cards[0]).toMatchObject({
      id: "auth",
      title: "Fix auth token refresh",
      description: "Refresh before expiry",
      priority: "critical",
      assignee: "cinos",
      tags: ["backend", "critical"],
    });
  });

  it("returns no columns for an empty config", () => {
    expect(parseKanbanBlockConfig("# empty").columns).toEqual([]);
  });
});

describe("buildKanbanBlockExportText", () => {
  const columns = [
    { name: "Now", cards: [{ id: "auth", title: "Fix auth", tags: ["backend"] }] },
    { name: "Next", cards: [{ id: "search", title: "Add search" }] },
  ];

  it("builds markdown export text", () => {
    expect(buildKanbanBlockExportText(columns)).toBe([
      "## Now",
      "- **Fix auth** [backend]",
      "",
      "## Next",
      "- **Add search**",
      "",
    ].join("\n"));
  });

  it("builds json export text", () => {
    expect(JSON.parse(buildKanbanBlockExportText(columns, "json"))).toEqual([
      { column: "Now", cards: [{ id: "auth", title: "Fix auth", tags: ["backend"] }] },
      { column: "Next", cards: [{ id: "search", title: "Add search" }] },
    ]);
  });
});

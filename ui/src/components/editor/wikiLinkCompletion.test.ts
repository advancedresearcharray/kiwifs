import { describe, expect, it } from "vitest";
import { EditorState } from "@codemirror/state";
import { wikiLinkCompletionSource, type WikiPage } from "./wikiLinkCompletion";

const pages: WikiPage[] = [
  { path: "pages/authentication.md", title: "Authentication" },
  { path: "pages/authorization.md", title: "Authorization" },
  { path: "pages/getting-started.md", title: "Getting Started" },
  { path: "runbook/deploy.md", title: "Deploy Runbook" },
];

function completionsAt(doc: string, pos: number) {
  const state = EditorState.create({ doc });
  const source = wikiLinkCompletionSource(pages);
  const context = {
    state,
    pos,
    explicit: false,
  } as any;
  return source(context);
}

describe("wikiLinkCompletionSource", () => {
  it("returns null when no [[ trigger is present", () => {
    expect(completionsAt("hello world", 5)).toBeNull();
  });

  it("returns completions after [[", () => {
    const doc = "See [[";
    const result = completionsAt(doc, doc.length);
    expect(result).not.toBeNull();
    expect(result!.options.length).toBe(4);
  });

  it("filters by typed query after [[", () => {
    const doc = "See [[auth";
    const result = completionsAt(doc, doc.length);
    expect(result).not.toBeNull();
    const labels = result!.options.map((o) => o.displayLabel);
    expect(labels).toContain("Authentication");
    expect(labels).toContain("Authorization");
    expect(labels).not.toContain("Getting Started");
  });

  it("matches against path as well as title", () => {
    const doc = "See [[runbook";
    const result = completionsAt(doc, doc.length);
    expect(result).not.toBeNull();
    expect(result!.options.length).toBe(1);
    expect(result!.options[0].displayLabel).toBe("Deploy Runbook");
  });

  it("strips .md extension from inserted wiki link", () => {
    const doc = "See [[deploy";
    const result = completionsAt(doc, doc.length);
    expect(result).not.toBeNull();
    expect(result!.options[0].label).toBe("[[runbook/deploy]]");
  });

  it("returns null for regular brackets", () => {
    expect(completionsAt("array[0]", 8)).toBeNull();
    expect(completionsAt("a [single bracket", 17)).toBeNull();
  });
});

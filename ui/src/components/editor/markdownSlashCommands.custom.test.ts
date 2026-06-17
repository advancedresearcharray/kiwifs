import { describe, expect, it, vi } from "vitest";
import { EditorState } from "@codemirror/state";
import type { CompletionContext, CompletionResult, CompletionSource } from "@codemirror/autocomplete";
import type { EditorView } from "@codemirror/view";
import { customSlashCompletionSource } from "./markdownSlashCommands";

async function invokeCompletionSource(
  source: CompletionSource,
  context: CompletionContext,
): Promise<CompletionResult | null> {
  const result = source(context);
  return result instanceof Promise ? await result : result;
}

function mockView(doc: string) {
  let currentDoc = doc;
  const state = {
    doc: {
      toString: () => currentDoc,
      length: currentDoc.length,
      sliceString: (from: number, to: number) => currentDoc.slice(from, to),
      lineAt: (pos: number) => {
        const before = currentDoc.slice(0, pos);
        const lineStart = before.lastIndexOf("\n") + 1;
        return { from: lineStart, text: currentDoc.slice(lineStart, currentDoc.indexOf("\n", lineStart) === -1 ? undefined : currentDoc.indexOf("\n", lineStart)) };
      },
    },
  };
  const view = {
    get state() {
      return state;
    },
    dispatch: (update: { changes: { from: number; to: number; insert: string }; selection: { anchor: number } }) => {
      currentDoc = update.changes.insert;
      state.doc = {
        toString: () => currentDoc,
        length: currentDoc.length,
        sliceString: (from: number, to: number) => currentDoc.slice(from, to),
        lineAt: state.doc.lineAt,
      };
    },
  } as unknown as EditorView;
  return { view, getDoc: () => currentDoc };
}

describe("customSlashCompletionSource", () => {
  it("offers configured commands matching the slash query", async () => {
    const commands = [
      { id: "adr", label: "ADR", icon: "FileCheck", description: "Insert ADR", template: "templates/adr.md" },
      { id: "runbook", label: "Runbook", icon: "Zap", description: "Runbook step", template: "templates/runbook.md" },
    ];
    const loadTemplate = vi.fn().mockResolvedValue("# ADR\n");
    const onError = vi.fn();

    const source = customSlashCompletionSource(commands, loadTemplate, onError);
    const state = EditorState.create({ doc: "# Note\n\n/ad" });
    const result = await invokeCompletionSource(source, {
      state,
      pos: state.doc.length,
      explicit: false,
      match: undefined,
    } as unknown as CompletionContext);

    expect(result?.options).toHaveLength(1);
    expect(result?.options?.[0]?.label).toBe("/adr");
  });

  it("loads template content on apply and reports errors", async () => {
    const commands = [
      { id: "adr", label: "ADR", icon: "", description: "", template: "templates/adr.md" },
    ];
    const loadTemplate = vi.fn().mockRejectedValue(new Error("missing file"));
    const onError = vi.fn();

    const source = customSlashCompletionSource(commands, loadTemplate, onError);
    const state = EditorState.create({ doc: "/adr" });
    const result = await invokeCompletionSource(source, {
      state,
      pos: state.doc.length,
      explicit: false,
      match: undefined,
    } as unknown as CompletionContext);

    const apply = result?.options?.[0]?.apply;
    expect(typeof apply).toBe("function");
    if (typeof apply === "function") {
      const { view } = mockView("/adr");
      apply(view, result!.options![0], 0, 4);
      await new Promise((resolve) => setTimeout(resolve, 0));
      expect(onError).toHaveBeenCalledWith(expect.stringContaining("templates/adr.md"));
    }
  });

  it("inserts loaded template markdown into the document", async () => {
    const commands = [
      { id: "adr", label: "ADR", icon: "", description: "", template: "templates/adr.md" },
    ];
    const loadTemplate = vi.fn().mockResolvedValue("# ADR template\n");
    const onError = vi.fn();

    const source = customSlashCompletionSource(commands, loadTemplate, onError);
    const state = EditorState.create({ doc: "Intro\n/adr" });
    const result = await invokeCompletionSource(source, {
      state,
      pos: state.doc.length,
      explicit: false,
      match: undefined,
    } as unknown as CompletionContext);

    const apply = result?.options?.[0]?.apply;
    if (typeof apply === "function") {
      const { view, getDoc } = mockView("Intro\n/adr");
      apply(view, result!.options![0], 6, 10);
      await new Promise((resolve) => setTimeout(resolve, 0));
      expect(getDoc()).toBe("Intro\n# ADR template\n");
    }
  });
});

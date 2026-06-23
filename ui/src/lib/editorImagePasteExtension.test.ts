import { describe, expect, it, vi } from "vitest";
import type { EditorView } from "@codemirror/view";
import { codeMirrorImageUpload } from "./editorImagePasteExtension";
import { UPLOADING_PLACEHOLDER } from "./editorImagePaste";

function createMockView(initial = "") {
  let doc = initial;
  const selection = { main: { head: initial.length } };
  const view = {
    state: {
      selection,
      doc: {
        sliceString: (from: number, to: number) => doc.slice(from, to),
      },
    },
    dispatch(update: {
      changes?: { from?: number; to?: number; insert?: string };
      selection?: { anchor: number };
    }) {
      if (update.changes) {
        const { from = 0, to = from, insert = "" } = update.changes;
        doc = doc.slice(0, from) + insert + doc.slice(to);
      }
      if (update.selection) {
        selection.main.head = update.selection.anchor;
      }
    },
    getDoc: () => doc,
  } as EditorView & { getDoc: () => string };
  return view;
}

describe("codeMirrorImageUpload", () => {
  it("inserts placeholder then replaces with markdown on successful upload", async () => {
    const uploadImage = vi.fn(async () => "/raw/pages/paste-20260623.png");
    const onError = vi.fn();
    const view = createMockView("hello ");
    const file = new File([""], "clip.png", { type: "image/png" });

    await codeMirrorImageUpload(view, file, { uploadImage, onError });

    expect(view.getDoc()).toMatch(
      /^hello !\[paste-\d{8}-\d{6}\]\(\/raw\/pages\/paste-20260623\.png\)$/,
    );
    expect(view.getDoc()).not.toContain(UPLOADING_PLACEHOLDER);
    expect(onError).not.toHaveBeenCalled();
    expect(uploadImage).toHaveBeenCalledOnce();
  });

  it("removes placeholder and reports error on upload failure", async () => {
    const uploadImage = vi.fn(async () => {
      throw new Error("File too large");
    });
    const onError = vi.fn();
    const view = createMockView("start");
    const file = new File([""], "big.png", { type: "image/png" });

    await codeMirrorImageUpload(view, file, { uploadImage, onError });

    expect(view.getDoc()).toBe("start");
    expect(onError).toHaveBeenCalledWith("File too large");
  });

  it("renames pasted files before upload", async () => {
    const uploadImage = vi.fn(async (f: File) => {
      expect(f.name).toMatch(/^paste-\d{8}-\d{6}\.png$/);
      return "/raw/pages/paste.png";
    });
    const view = createMockView("");
    const file = new File([""], "screenshot.png", { type: "image/png" });

    await codeMirrorImageUpload(view, file, { uploadImage });

    expect(uploadImage).toHaveBeenCalledOnce();
  });
});

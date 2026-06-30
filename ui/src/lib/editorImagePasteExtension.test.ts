import { describe, expect, it, vi } from "vitest";
import {
  beginImageInsert,
  editorImagePasteDomHandlers,
  insertUploadedImage,
  replacePlaceholderInView,
  removePlaceholderFromView,
} from "./editorImagePasteExtension";

function mockView(doc: string, head = doc.length) {
  let current = doc;
  let cursor = head;
  const dispatch = vi.fn((update: {
    changes?: { from: number; to?: number; insert?: string };
    selection?: { anchor: number };
  }) => {
    const { changes, selection } = update;
    if (changes) {
      const before = current.slice(0, changes.from);
      const after = current.slice(changes.to ?? changes.from);
      current = before + (changes.insert ?? "") + after;
      if (selection) cursor = selection.anchor;
      else if (changes.insert) cursor = changes.from + changes.insert.length;
    } else if (selection) {
      cursor = selection.anchor;
    }
  });
  return {
    get doc() {
      return current;
    },
    state: {
      get doc() {
        return { toString: () => current };
      },
      selection: { main: { get head() { return cursor; } } },
    },
    dispatch,
    posAtCoords: vi.fn(() => 0),
  };
}

describe("editorImagePasteExtension", () => {
  it("replaces uploading placeholder with markdown image on success", async () => {
    const view = mockView("Hello ![Uploading...](kiwi-upload://tok)");
    const placeholder = "![Uploading...](kiwi-upload://tok)";
    const file = new File(["x"], "clip.png", { type: "image/png" });
    const uploadImage = vi.fn().mockResolvedValue("/raw/notes/paste-20260623-143022.png");

    await insertUploadedImage(view as any, file, placeholder, {
      pagePath: "notes/page.md",
      uploadImage,
      onError: vi.fn(),
    });

    expect(uploadImage).toHaveBeenCalled();
    expect(view.doc).toMatch(/^Hello !\[paste-\d{8}-\d{6}\.png\]\(paste-\d{8}-\d{6}\.png\)$/);
  });

  it("removes placeholder and reports error when upload fails", async () => {
    const view = mockView("![Uploading...](kiwi-upload://tok)");
    const onError = vi.fn();
    const file = new File(["x"], "clip.png", { type: "image/png" });

    await insertUploadedImage(view as any, file, "![Uploading...](kiwi-upload://tok)", {
      pagePath: "notes/page.md",
      uploadImage: vi.fn().mockRejectedValue(new Error("File too large")),
      onError,
    });

    expect(onError).toHaveBeenCalledWith("File too large");
    expect(view.doc).toBe("");
  });

  it("replacePlaceholderInView swaps text in document", () => {
    const view = mockView("aaPLACEHOLDERbb");
    replacePlaceholderInView(view as any, "PLACEHOLDER", "OK");
    expect(view.doc).toBe("aaOKbb");
  });

  it("removePlaceholderFromView deletes marker text", () => {
    const view = mockView("aaPLACEHOLDERbb");
    removePlaceholderFromView(view as any, "PLACEHOLDER");
    expect(view.doc).toBe("aabb");
  });

  it("beginImageInsert inserts uploading placeholder and starts upload", async () => {
    const view = mockView("Hello ");
    const file = new File(["x"], "clip.png", { type: "image/png" });
    const uploadImage = vi.fn().mockResolvedValue("/raw/notes/paste-20260623-143022.png");
    const onUploadStart = vi.fn();
    const onUploadEnd = vi.fn();

    beginImageInsert(view as any, file, {
      pagePath: "notes/page.md",
      uploadImage,
      onError: vi.fn(),
      onUploadStart,
      onUploadEnd,
    });

    expect(onUploadStart).toHaveBeenCalled();
    expect(view.doc).toContain("Uploading...");
    expect(view.doc).toContain("kiwi-upload://");
    await vi.waitFor(() => expect(onUploadEnd).toHaveBeenCalled());
    expect(uploadImage).toHaveBeenCalledWith(file);
  });

  it("paste handler prevents default and inserts placeholder", async () => {
    const view = mockView("");
    const png = new File(["a"], "a.png", { type: "image/png" });
    const items = [{ kind: "file", type: "image/png", getAsFile: () => png }];
    const preventDefault = vi.fn();
    const event = {
      clipboardData: { items, files: [png] },
      preventDefault,
    } as unknown as ClipboardEvent;
    const uploadImage = vi.fn().mockResolvedValue("/raw/notes/paste-20260623-143022.png");
    const handlers = editorImagePasteDomHandlers({
      pagePath: "notes/page.md",
      uploadImage,
      onError: vi.fn(),
    });

    expect(handlers.paste(event, view as any)).toBe(true);
    expect(preventDefault).toHaveBeenCalled();
    expect(view.doc).toContain("Uploading...");
    await vi.waitFor(() => expect(uploadImage).toHaveBeenCalled());
  });

  it("drop handler prevents default at drop coordinates", async () => {
    const view = mockView("drop here");
    view.posAtCoords = vi.fn(() => 4);
    const png = new File(["a"], "a.png", { type: "image/png" });
    const preventDefault = vi.fn();
    const event = {
      clientX: 10,
      clientY: 20,
      dataTransfer: { types: ["Files"], items: [], files: [png] },
      preventDefault,
    } as unknown as DragEvent;
    const uploadImage = vi.fn().mockResolvedValue("/raw/notes/paste-20260623-143022.png");
    const handlers = editorImagePasteDomHandlers({
      pagePath: "notes/page.md",
      uploadImage,
      onError: vi.fn(),
    });

    expect(handlers.drop(event, view as any)).toBe(true);
    expect(preventDefault).toHaveBeenCalled();
    expect(view.doc).toContain("Uploading...");
    await vi.waitFor(() => expect(uploadImage).toHaveBeenCalled());
  });
});

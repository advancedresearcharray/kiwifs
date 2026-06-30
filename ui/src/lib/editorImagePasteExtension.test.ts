import { describe, expect, it, vi } from "vitest";
import {
  insertUploadedImage,
  replacePlaceholderInView,
  removePlaceholderFromView,
} from "./editorImagePasteExtension";

function mockView(doc: string) {
  let current = doc;
  const dispatch = vi.fn((update: { changes?: { from: number; to?: number; insert?: string } }) => {
    const { changes } = update;
    if (!changes) return;
    const before = current.slice(0, changes.from);
    const after = current.slice(changes.to ?? changes.from);
    current = before + (changes.insert ?? "") + after;
  });
  return {
    get doc() {
      return current;
    },
    state: {
      get doc() {
        return { toString: () => current };
      },
      selection: { main: { head: current.length } },
    },
    dispatch,
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
});

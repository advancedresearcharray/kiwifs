import { describe, expect, it } from "vitest";
import {
  assetUrlToMarkdownRef,
  extractImagesFromDataTransfer,
  hasImageInDataTransfer,
  isPasteableImageType,
  markdownImageRef,
  pasteImageFileName,
  renameFileForPaste,
  uploadingPlaceholder,
} from "./editorImagePaste";

describe("editorImagePaste", () => {
  it("detects supported image MIME types", () => {
    expect(isPasteableImageType("image/png")).toBe(true);
    expect(isPasteableImageType("image/jpeg")).toBe(true);
    expect(isPasteableImageType("text/plain")).toBe(false);
  });

  it("generates paste-YYYYMMDD-HHMMSS filenames", () => {
    const now = new Date("2026-06-23T14:30:22");
    expect(pasteImageFileName("image/png", now)).toBe("paste-20260623-143022.png");
    expect(pasteImageFileName("image/jpeg", now)).toBe("paste-20260623-143022.jpg");
  });

  it("renames clipboard files for upload", () => {
    const file = new File(["x"], "screenshot.png", { type: "image/png" });
    const now = new Date("2026-06-23T14:30:22");
    const renamed = renameFileForPaste(file, now);
    expect(renamed.name).toBe("paste-20260623-143022.png");
    expect(renamed.type).toBe("image/png");
  });

  it("extracts image files from DataTransfer items", () => {
    const png = new File(["a"], "a.png", { type: "image/png" });
    const items = [
      { kind: "file", type: "image/png", getAsFile: () => png },
      { kind: "file", type: "text/plain", getAsFile: () => new File(["b"], "b.txt", { type: "text/plain" }) },
    ];
    const dt = { items, files: [] } as unknown as DataTransfer;
    expect(extractImagesFromDataTransfer(dt)).toEqual([png]);
    expect(hasImageInDataTransfer(dt)).toBe(true);
  });

  it("maps uploaded /raw/ URLs to page-relative markdown refs", () => {
    expect(assetUrlToMarkdownRef("/raw/notes/paste-1.png", "notes/page.md")).toBe(
      "paste-1.png",
    );
    expect(assetUrlToMarkdownRef("/raw/paste-1.png", "page.md")).toBe("paste-1.png");
  });

  it("builds markdown image syntax", () => {
    expect(markdownImageRef("shot", "paste-1.png")).toBe("![shot](paste-1.png)");
  });

  it("creates unique uploading placeholders", () => {
    const a = uploadingPlaceholder("abc");
    const b = uploadingPlaceholder("def");
    expect(a).toContain("Uploading...");
    expect(a).not.toBe(b);
  });
});

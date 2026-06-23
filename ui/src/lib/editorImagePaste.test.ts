import { describe, expect, it } from "vitest";
import {
  altFromPasteFilename,
  extractImageFromClipboard,
  extractImageFilesFromDataTransfer,
  formatPasteTimestamp,
  imageMarkdown,
  isEditorImageFile,
  isEditorImageMime,
  isOsFileDrag,
  renameFileForPaste,
  UPLOADING_PLACEHOLDER,
} from "./editorImagePaste";

describe("editorImagePaste", () => {
  it("recognizes supported image mime types", () => {
    expect(isEditorImageMime("image/png")).toBe(true);
    expect(isEditorImageMime("image/jpeg")).toBe(true);
    expect(isEditorImageMime("image/gif")).toBe(true);
    expect(isEditorImageMime("image/webp")).toBe(true);
    expect(isEditorImageMime("image/svg+xml")).toBe(false);
    expect(isEditorImageMime("text/plain")).toBe(false);
  });

  it("formats paste timestamps for filenames", () => {
    const ts = formatPasteTimestamp(new Date("2026-06-23T14:30:22Z"));
    expect(ts).toMatch(/^20260623-\d{6}$/);
  });

  it("renames clipboard files to paste-<timestamp>.<ext>", () => {
    const original = new File([new Uint8Array([1, 2, 3])], "screenshot.png", {
      type: "image/png",
    });
    const renamed = renameFileForPaste(original, new Date("2026-06-23T14:30:22Z"));
    expect(renamed.name).toBe("paste-20260623-143022.png");
    expect(renamed.type).toBe("image/png");
  });

  it("builds markdown image syntax", () => {
    expect(imageMarkdown("/raw/pages/paste.png", "paste-20260623")).toBe(
      "![paste-20260623](/raw/pages/paste.png)",
    );
    expect(UPLOADING_PLACEHOLDER).toBe("![Uploading...]()");
  });

  it("derives alt text from paste filename", () => {
    expect(altFromPasteFilename("paste-20260623-143022.png")).toBe(
      "paste-20260623-143022",
    );
  });

  it("detects image files by extension when mime is empty", () => {
    const file = new File([""], "photo.jpeg", { type: "" });
    expect(isEditorImageFile(file)).toBe(true);
  });

  it("extracts image from clipboard data transfer items", () => {
    const file = new File([""], "clip.png", { type: "image/png" });
    const data = {
      items: [
        {
          kind: "file",
          type: "image/png",
          getAsFile: () => file,
        },
      ],
      files: [] as unknown as FileList,
    } as unknown as DataTransfer;
    expect(extractImageFromClipboard(data)?.name).toBe("clip.png");
  });

  it("extracts image files from drop data transfer", () => {
    const a = new File([""], "a.png", { type: "image/png" });
    const b = new File([""], "b.txt", { type: "text/plain" });
    const data = {
      files: [a, b],
    } as unknown as DataTransfer;
    const images = extractImageFilesFromDataTransfer(data);
    expect(images).toHaveLength(1);
    expect(images[0].name).toBe("a.png");
  });

  it("extracts image from files when clipboard items are empty", () => {
    const file = new File([""], "screenshot.png", { type: "image/png" });
    const data = {
      items: [] as unknown as DataTransferItemList,
      files: [file],
    } as unknown as DataTransfer;
    expect(extractImageFromClipboard(data)?.name).toBe("screenshot.png");
  });

  it("recognizes image mime types with parameters", () => {
    expect(isEditorImageMime("image/png; charset=binary")).toBe(true);
    expect(isEditorImageMime("IMAGE/JPEG")).toBe(true);
  });

  it("detects OS file drag from dataTransfer types", () => {
    expect(isOsFileDrag({ dataTransfer: { types: ["Files"] } as unknown as DataTransfer })).toBe(true);
    expect(isOsFileDrag({ dataTransfer: { types: ["text/plain"] } as unknown as DataTransfer })).toBe(false);
    expect(isOsFileDrag({})).toBe(false);
  });
});

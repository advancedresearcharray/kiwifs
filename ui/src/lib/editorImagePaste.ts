/** Clipboard image paste helpers for KiwiEditor (visual + source modes). */

export const EDITOR_IMAGE_MIME_TYPES = [
  "image/png",
  "image/jpeg",
  "image/gif",
  "image/webp",
] as const;

export const UPLOADING_PLACEHOLDER = "![Uploading...]()";

export function isEditorImageMime(type: string): boolean {
  const normalized = type.toLowerCase().split(";")[0].trim();
  return (EDITOR_IMAGE_MIME_TYPES as readonly string[]).includes(normalized);
}

export function isEditorImageFile(file: File): boolean {
  if (file.type && isEditorImageMime(file.type)) return true;
  return /\.(png|jpe?g|gif|webp)$/i.test(file.name);
}

export function extensionForImageMime(mime: string): string {
  const normalized = mime.toLowerCase().split(";")[0].trim();
  switch (normalized) {
    case "image/png":
      return "png";
    case "image/jpeg":
      return "jpg";
    case "image/gif":
      return "gif";
    case "image/webp":
      return "webp";
    default:
      return "png";
  }
}

export function extensionFromFilename(name: string): string | null {
  const match = name.match(/\.(png|jpe?g|gif|webp)$/i);
  if (!match) return null;
  const ext = match[1].toLowerCase();
  return ext === "jpeg" ? "jpg" : ext;
}

export function formatPasteTimestamp(date: Date): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${date.getFullYear()}${pad(date.getMonth() + 1)}${pad(date.getDate())}-${pad(date.getHours())}${pad(date.getMinutes())}${pad(date.getSeconds())}`;
}

export function renameFileForPaste(file: File, now = new Date()): File {
  const ext =
    extensionForImageMime(file.type) ||
    extensionFromFilename(file.name) ||
    "png";
  const mime =
    file.type ||
    (ext === "jpg" ? "image/jpeg" : `image/${ext === "jpg" ? "jpeg" : ext}`);
  const name = `paste-${formatPasteTimestamp(now)}.${ext}`;
  return new File([file], name, { type: mime });
}

export function altFromPasteFilename(filename: string): string {
  return filename.replace(/\.[^.]+$/, "") || filename;
}

export function imageMarkdown(url: string, alt?: string): string {
  const label = alt ?? url.split("/").pop() ?? "image";
  return `![${label}](${url})`;
}

export function extractImageFromClipboard(data: DataTransfer | null): File | null {
  if (!data) return null;

  const items = data.items;
  if (items) {
    for (let i = 0; i < items.length; i++) {
      const item = items[i];
      if (item.kind === "file" && isEditorImageMime(item.type)) {
        const file = item.getAsFile();
        if (file) return file;
      }
    }
  }

  const files = data.files;
  if (files) {
    for (let i = 0; i < files.length; i++) {
      if (isEditorImageFile(files[i])) return files[i];
    }
  }

  return null;
}

export function extractImageFilesFromDataTransfer(data: DataTransfer | null): File[] {
  if (!data) return [];
  const out: File[] = [];
  const files = data.files;
  if (!files) return out;
  for (let i = 0; i < files.length; i++) {
    if (isEditorImageFile(files[i])) out.push(files[i]);
  }
  return out;
}

export function isOsFileDrag(e: { dataTransfer?: DataTransfer | null }): boolean {
  const types = e.dataTransfer?.types;
  return types ? Array.from(types).includes("Files") : false;
}

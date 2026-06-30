import { dirOf } from "./paths";

/** Clipboard / drag image MIME types supported by the asset upload API. */
export const PASTE_IMAGE_MIME_TYPES = [
  "image/png",
  "image/jpeg",
  "image/gif",
  "image/webp",
] as const;

export const UPLOADING_ALT = "Uploading...";
const UPLOADING_URL_PREFIX = "kiwi-upload://";

export function isPasteableImageType(mime: string): boolean {
  return (PASTE_IMAGE_MIME_TYPES as readonly string[]).includes(mime);
}

export function extensionForImageMime(mime: string): string {
  switch (mime) {
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

/** Standardize clipboard paste filenames: paste-YYYYMMDD-HHMMSS.ext */
export function pasteImageFileName(mime: string, now = new Date()): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  const stamp = [
    now.getFullYear(),
    pad(now.getMonth() + 1),
    pad(now.getDate()),
    "-",
    pad(now.getHours()),
    pad(now.getMinutes()),
    pad(now.getSeconds()),
  ].join("");
  return `paste-${stamp}.${extensionForImageMime(mime)}`;
}

export function renameFileForPaste(file: File, now = new Date()): File {
  const mime = file.type || "image/png";
  const name = pasteImageFileName(mime, now);
  return new File([file], name, { type: mime, lastModified: file.lastModified });
}

export function extractImagesFromDataTransfer(
  dataTransfer: DataTransfer | null | undefined,
): File[] {
  if (!dataTransfer) return [];
  const files: File[] = [];
  const items = dataTransfer.items;
  if (items && items.length > 0) {
    for (let i = 0; i < items.length; i++) {
      const item = items[i];
      if (item.kind !== "file") continue;
      const file = item.getAsFile();
      if (!file) continue;
      const mime = item.type || file.type;
      if (!isPasteableImageType(mime)) continue;
      files.push(file);
    }
    if (files.length > 0) return files;
  }
  return Array.from(dataTransfer.files).filter((f) => isPasteableImageType(f.type));
}

export function hasImageInDataTransfer(
  dataTransfer: DataTransfer | null | undefined,
): boolean {
  return extractImagesFromDataTransfer(dataTransfer).length > 0;
}

export function isOsImageDrag(event: { dataTransfer?: DataTransfer | null }): boolean {
  const types = event.dataTransfer?.types;
  if (!types) return false;
  return Array.from(types).includes("Files");
}

export function uploadingPlaceholder(token: string): string {
  return `![${UPLOADING_ALT}](${UPLOADING_URL_PREFIX}${token})`;
}

export function isUploadingPlaceholder(text: string): boolean {
  return text.includes(UPLOADING_URL_PREFIX);
}

/** Map /raw/workspace/path.png to a markdown ref relative to the edited page. */
export function assetUrlToMarkdownRef(rawUrl: string, pagePath: string): string {
  const assetPath = rawUrl.replace(/^\/raw\//, "");
  const pageDir = dirOf(pagePath);
  if (!pageDir) return assetPath;
  if (assetPath === pageDir) return basename(assetPath);
  const prefix = `${pageDir}/`;
  if (assetPath.startsWith(prefix)) {
    return assetPath.slice(prefix.length);
  }
  return assetPath;
}

export function markdownImageRef(alt: string, ref: string): string {
  const safeAlt = alt.replace(/[\[\]]/g, "");
  return `![${safeAlt}](${ref})`;
}

function basename(p: string): string {
  const idx = p.lastIndexOf("/");
  return idx < 0 ? p : p.slice(idx + 1);
}

import { EditorView } from "@codemirror/view";
import type { Extension } from "@codemirror/state";
import {
  assetUrlToMarkdownRef,
  extractImagesFromDataTransfer,
  hasImageInDataTransfer,
  isOsImageDrag,
  markdownImageRef,
  renameFileForPaste,
  uploadingPlaceholder,
} from "./editorImagePaste";

export type EditorImagePasteOptions = {
  pagePath: string;
  uploadImage: (file: File) => Promise<string>;
  onError: (message: string) => void;
};

export function replacePlaceholderInView(
  view: EditorView,
  placeholder: string,
  replacement: string,
): void {
  const doc = view.state.doc.toString();
  const idx = doc.indexOf(placeholder);
  if (idx < 0) return;
  view.dispatch({
    changes: { from: idx, to: idx + placeholder.length, insert: replacement },
  });
}

export function removePlaceholderFromView(view: EditorView, placeholder: string): void {
  const doc = view.state.doc.toString();
  const idx = doc.indexOf(placeholder);
  if (idx < 0) return;
  view.dispatch({
    changes: { from: idx, to: idx + placeholder.length, insert: "" },
  });
}

export async function insertUploadedImage(
  view: EditorView,
  file: File,
  placeholder: string,
  opts: EditorImagePasteOptions,
): Promise<void> {
  try {
    const renamed = renameFileForPaste(file);
    const rawUrl = await opts.uploadImage(renamed);
    const ref = assetUrlToMarkdownRef(rawUrl, opts.pagePath);
    const replacement = markdownImageRef(renamed.name, ref);
    replacePlaceholderInView(view, placeholder, replacement);
  } catch (e) {
    removePlaceholderFromView(view, placeholder);
    const msg = e instanceof Error ? e.message : String(e);
    opts.onError(msg || "Image upload failed");
  }
}

export function beginImageInsert(
  view: EditorView,
  file: File,
  opts: EditorImagePasteOptions,
): string {
  const token = crypto.randomUUID();
  const placeholder = uploadingPlaceholder(token);
  const pos = view.state.selection.main.head;
  view.dispatch({
    changes: { from: pos, insert: placeholder },
    selection: { anchor: pos + placeholder.length },
  });
  void insertUploadedImage(view, file, placeholder, opts);
  return placeholder;
}

export function editorImagePasteExtension(opts: EditorImagePasteOptions): Extension {
  return EditorView.domEventHandlers({
    paste(event, view) {
      if (!hasImageInDataTransfer(event.clipboardData)) return false;
      const files = extractImagesFromDataTransfer(event.clipboardData);
      if (files.length === 0) return false;
      event.preventDefault();
      beginImageInsert(view, files[0], opts);
      return true;
    },
    drop(event, view) {
      if (!isOsImageDrag(event)) return false;
      const files = extractImagesFromDataTransfer(event.dataTransfer);
      if (files.length === 0) return false;
      event.preventDefault();
      const pos = view.posAtCoords({ x: event.clientX, y: event.clientY });
      if (pos) {
        view.dispatch({ selection: { anchor: pos } });
      }
      beginImageInsert(view, files[0], opts);
      return true;
    },
    dragover(event) {
      if (!isOsImageDrag(event) || !hasImageInDataTransfer(event.dataTransfer)) return false;
      event.preventDefault();
      return true;
    },
  });
}

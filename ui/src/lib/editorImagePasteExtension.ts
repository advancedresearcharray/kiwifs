import { EditorView } from "@codemirror/view";
import type { Extension } from "@codemirror/state";
import {
  altFromPasteFilename,
  extractImageFromClipboard,
  extractImageFilesFromDataTransfer,
  imageMarkdown,
  renameFileForPaste,
  UPLOADING_PLACEHOLDER,
} from "@kw/lib/editorImagePaste";

export type EditorImageUploadOptions = {
  uploadImage: (file: File) => Promise<string>;
  onError?: (message: string) => void;
};

export async function codeMirrorImageUpload(
  view: EditorView,
  file: File,
  options: EditorImageUploadOptions,
): Promise<void> {
  const named = renameFileForPaste(file);
  const pos = view.state.selection.main.head;
  view.dispatch({
    changes: { from: pos, insert: UPLOADING_PLACEHOLDER },
    selection: { anchor: pos + UPLOADING_PLACEHOLDER.length },
  });
  const start = pos;

  try {
    const url = await options.uploadImage(named);
    const md = imageMarkdown(url, altFromPasteFilename(named.name));
    const end = start + UPLOADING_PLACEHOLDER.length;
    if (view.state.doc.sliceString(start, end) === UPLOADING_PLACEHOLDER) {
      view.dispatch({
        changes: { from: start, to: end, insert: md },
        selection: { anchor: start + md.length },
      });
    }
  } catch (e) {
    const end = start + UPLOADING_PLACEHOLDER.length;
    if (view.state.doc.sliceString(start, end) === UPLOADING_PLACEHOLDER) {
      view.dispatch({ changes: { from: start, to: end } });
    }
    const msg = e instanceof Error ? e.message : String(e);
    options.onError?.(msg);
  }
}

export function editorImagePasteExtension(options: EditorImageUploadOptions): Extension {
  return EditorView.domEventHandlers({
    paste(event, view) {
      const file = extractImageFromClipboard(event.clipboardData);
      if (!file) return false;
      event.preventDefault();
      void codeMirrorImageUpload(view, file, options);
      return true;
    },
    drop(event, view) {
      const files = extractImageFilesFromDataTransfer(event.dataTransfer);
      if (files.length === 0) return false;
      event.preventDefault();
      for (const file of files) void codeMirrorImageUpload(view, file, options);
      return true;
    },
    dragover(event) {
      if (extractImageFilesFromDataTransfer(event.dataTransfer).length > 0) {
        event.preventDefault();
      }
    },
  });
}

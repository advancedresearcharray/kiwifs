import { Plugin, PluginKey } from "prosemirror-state";
import type { EditorView } from "prosemirror-view";
import {
  altFromPasteFilename,
  extractImageFromClipboard,
  extractImageFilesFromDataTransfer,
  imageMarkdown,
  renameFileForPaste,
  UPLOADING_PLACEHOLDER,
} from "@kw/lib/editorImagePaste";
import type { EditorImageUploadOptions } from "@kw/lib/editorImagePasteExtension";

async function insertUploadedImage(
  view: EditorView,
  file: File,
  options: EditorImageUploadOptions,
): Promise<void> {
  const named = renameFileForPaste(file);
  const { state } = view;
  const pos = state.selection.from;
  view.dispatch(state.tr.insertText(UPLOADING_PLACEHOLDER, pos));
  const start = pos;

  try {
    const url = await options.uploadImage(named);
    const md = imageMarkdown(url, altFromPasteFilename(named.name));
    const current = view.state;
    const end = start + UPLOADING_PLACEHOLDER.length;
    if (current.doc.textBetween(start, end) === UPLOADING_PLACEHOLDER) {
      view.dispatch(current.tr.replaceWith(start, end, current.schema.text(md)));
    }
  } catch (e) {
    const current = view.state;
    const end = start + UPLOADING_PLACEHOLDER.length;
    if (current.doc.textBetween(start, end) === UPLOADING_PLACEHOLDER) {
      view.dispatch(current.tr.delete(start, end));
    }
    const msg = e instanceof Error ? e.message : String(e);
    options.onError?.(msg);
  }
}

export const imagePastePluginKey = new PluginKey("kiwi-image-paste");

export function imagePasteProsemirrorPlugin(options: EditorImageUploadOptions): Plugin {
  return new Plugin({
    key: imagePastePluginKey,
    props: {
      handlePaste(view, event) {
        const file = extractImageFromClipboard(event.clipboardData);
        if (!file) return false;
        event.preventDefault();
        void insertUploadedImage(view, file, options);
        return true;
      },
      handleDrop(view, event) {
        const files = extractImageFilesFromDataTransfer(event.dataTransfer);
        if (files.length === 0) return false;
        event.preventDefault();
        for (const file of files) void insertUploadedImage(view, file, options);
        return true;
      },
    },
  });
}

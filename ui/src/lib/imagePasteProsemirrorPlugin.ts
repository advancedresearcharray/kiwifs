import { Plugin, PluginKey } from "prosemirror-state";
import type { EditorView } from "prosemirror-view";
import {
  extractImagesFromDataTransfer,
  hasImageInDataTransfer,
  isOsImageDrag,
  renameFileForPaste,
} from "./editorImagePaste";

export type ImagePastePluginOptions = {
  uploadImage: (file: File) => Promise<string>;
  onError: (message: string) => void;
  onUploaded?: () => void;
};

const imagePasteKey = new PluginKey("kiwi-image-paste");

async function uploadAndInsertImage(
  view: EditorView,
  file: File,
  pos: number,
  opts: ImagePastePluginOptions,
): Promise<void> {
  const renamed = renameFileForPaste(file);
  try {
    const url = await opts.uploadImage(renamed);
    const imageNode = view.state.schema.nodes.image?.create({ src: url, alt: renamed.name });
    if (!imageNode) {
      opts.onError("Editor does not support image nodes");
      return;
    }
    const tr = view.state.tr.insert(pos, imageNode);
    view.dispatch(tr);
    opts.onUploaded?.();
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e);
    opts.onError(msg || "Image upload failed");
  }
}

export function imagePasteProsemirrorPlugin(opts: ImagePastePluginOptions): Plugin {
  return new Plugin({
    key: imagePasteKey,
    props: {
      handleDOMEvents: {
        paste(view, event) {
          if (!hasImageInDataTransfer(event.clipboardData)) return false;
          const files = extractImagesFromDataTransfer(event.clipboardData);
          if (files.length === 0) return false;
          event.preventDefault();
          const pos = view.state.selection.from;
          void uploadAndInsertImage(view, files[0], pos, opts);
          return true;
        },
        drop(view, event) {
          if (!isOsImageDrag(event)) return false;
          const files = extractImagesFromDataTransfer(event.dataTransfer);
          if (files.length === 0) return false;
          event.preventDefault();
          const coords = view.posAtCoords({ left: event.clientX, top: event.clientY });
          const pos = coords?.pos ?? view.state.selection.from;
          void uploadAndInsertImage(view, files[0], pos, opts);
          return true;
        },
        dragover(_view, event) {
          if (!isOsImageDrag(event) || !hasImageInDataTransfer(event.dataTransfer)) return false;
          event.preventDefault();
          return true;
        },
      },
    },
  });
}

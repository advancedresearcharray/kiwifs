import { Plugin, PluginKey } from "prosemirror-state";
import type { EditorView } from "prosemirror-view";
import {
  assetUrlToMarkdownRef,
  extractImagesFromDataTransfer,
  hasImageInDataTransfer,
  isOsImageDrag,
} from "./editorImagePaste";

export type ImagePastePluginOptions = {
  /** Should rename via renameFileForPaste before upload (e.g. uploadAssetForEditor). */
  uploadImage: (file: File) => Promise<string>;
  pagePath: string;
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
  try {
    const rawUrl = await opts.uploadImage(file);
    const ref = assetUrlToMarkdownRef(rawUrl, opts.pagePath);
    const alt = ref.includes("/") ? ref.slice(ref.lastIndexOf("/") + 1) : ref;
    const imageNode = view.state.schema.nodes.image?.create({ src: rawUrl, alt });
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

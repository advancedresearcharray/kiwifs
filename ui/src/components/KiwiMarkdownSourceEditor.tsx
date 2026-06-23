import { useCallback, useMemo, useRef, useState } from "react";
import CodeMirror from "@uiw/react-codemirror";
import { autocompletion, type CompletionSource } from "@codemirror/autocomplete";
import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
import { highlightSelectionMatches, searchKeymap } from "@codemirror/search";
import { type Extension } from "@codemirror/state";
import { EditorView, keymap } from "@codemirror/view";
import { cn } from "@kw/lib/cn";
import { editorImagePasteExtension } from "@kw/lib/editorImagePasteExtension";
import { isOsFileDrag } from "@kw/lib/editorImagePaste";
import { markdownEditorExtensions } from "./editor/markdownLanguage";
import { markdownEditorTheme } from "./editor/markdownEditorTheme";
import { slashCompletionSource } from "./editor/markdownSlashCommands";
import {
  wikiLinkCompletionSource,
  type WikiPage,
} from "./editor/wikiLinkCompletion";
import { EditorImageDropOverlay } from "./EditorImageDropOverlay";

export type KiwiMarkdownSourceEditorProps = {
  value: string;
  onChange: (next: string) => void;
  readOnly?: boolean;
  dark?: boolean;
  minHeight?: string;
  className?: string;
  onSaveShortcut?: () => void;
  pages?: WikiPage[];
  uploadImage?: (file: File) => Promise<string>;
  onImageUploadError?: (message: string) => void;
};

export function KiwiMarkdownSourceEditor({
  value,
  onChange,
  readOnly = false,
  dark = false,
  minHeight = "60vh",
  className,
  onSaveShortcut,
  pages = [],
  uploadImage,
  onImageUploadError,
}: KiwiMarkdownSourceEditorProps) {
  const [fileDragActive, setFileDragActive] = useState(false);
  const fileDragDepthRef = useRef(0);

  const resetFileDrag = useCallback(() => {
    fileDragDepthRef.current = 0;
    setFileDragActive(false);
  }, []);

  const extensions = useMemo(() => {
    const saveKeymap = keymap.of([
      {
        key: "Mod-s",
        preventDefault: true,
        run: () => {
          onSaveShortcut?.();
          return true;
        },
      },
    ]);

    const completionSources: CompletionSource[] = [slashCompletionSource];
    if (pages.length > 0) {
      completionSources.push(wikiLinkCompletionSource(pages));
    }

    const exts: Extension[] = [
      history(),
      ...markdownEditorExtensions(),
      highlightSelectionMatches(),
      autocompletion({
        override: completionSources,
        activateOnTyping: true,
        closeOnBlur: true,
      }),
      EditorView.contentAttributes.of({
        "aria-label": "Markdown source editor",
        "aria-multiline": "true",
      }),
      keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap]),
      saveKeymap,
    ];

    if (uploadImage) {
      exts.push(
        editorImagePasteExtension({
          uploadImage,
          onError: onImageUploadError,
        }),
      );
    }

    return exts;
  }, [onSaveShortcut, pages, uploadImage, onImageUploadError]);

  const theme = useMemo(() => markdownEditorTheme({ dark }), [dark]);

  const handleDragEnter = useCallback((e: React.DragEvent) => {
    if (!uploadImage || !isOsFileDrag(e)) return;
    e.preventDefault();
    fileDragDepthRef.current += 1;
    setFileDragActive(true);
  }, [uploadImage]);

  const handleDragLeave = useCallback(
    (e: React.DragEvent) => {
      if (!uploadImage || !isOsFileDrag(e)) return;
      if (e.currentTarget.contains(e.relatedTarget as Node)) return;
      fileDragDepthRef.current = Math.max(0, fileDragDepthRef.current - 1);
      if (fileDragDepthRef.current === 0) resetFileDrag();
    },
    [uploadImage, resetFileDrag],
  );

  const handleDragOver = useCallback(
    (e: React.DragEvent) => {
      if (!uploadImage || !isOsFileDrag(e)) return;
      e.preventDefault();
      e.dataTransfer.dropEffect = "copy";
    },
    [uploadImage],
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      if (!uploadImage || !isOsFileDrag(e)) return;
      e.preventDefault();
      resetFileDrag();
    },
    [uploadImage, resetFileDrag],
  );

  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-md border bg-background shadow-sm",
        className,
      )}
      data-testid="markdown-source-editor"
      onDragEnter={handleDragEnter}
      onDragLeave={handleDragLeave}
      onDragOver={handleDragOver}
      onDrop={handleDrop}
    >
      <EditorImageDropOverlay active={fileDragActive} />
      <CodeMirror
        value={value}
        height="auto"
        minHeight={minHeight}
        basicSetup={{
          lineNumbers: true,
          foldGutter: true,
          highlightActiveLine: true,
          highlightSelectionMatches: false,
          autocompletion: false,
          bracketMatching: true,
          closeBrackets: true,
          defaultKeymap: false,
          history: false,
          searchKeymap: false,
        }}
        editable={!readOnly}
        readOnly={readOnly}
        theme={theme}
        extensions={extensions}
        onChange={onChange}
      />
    </div>
  );
}

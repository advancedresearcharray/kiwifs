import { useCallback, useMemo, useRef, useState } from "react";
import CodeMirror from "@uiw/react-codemirror";
import { autocompletion, type CompletionSource } from "@codemirror/autocomplete";
import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
import { highlightSelectionMatches, searchKeymap } from "@codemirror/search";
import { type Extension } from "@codemirror/state";
import { EditorView, keymap } from "@codemirror/view";
import { cn } from "@kw/lib/cn";
import type { EditorSlashCommandConfig } from "@kw/lib/editorSlashCommands";
import { editorImagePasteExtension } from "@kw/lib/editorImagePasteExtension";
import { isOsFileDrag } from "@kw/lib/editorImagePaste";
import { EditorImageDropOverlay } from "../EditorImageDropOverlay";
import { markdownEditorExtensions } from "./markdownLanguage";
import { markdownEditorTheme } from "./markdownEditorTheme";
import { customSlashCompletionSource, slashCompletionSource } from "./markdownSlashCommands";
import {
  wikiLinkCompletionSource,
  type WikiPage,
} from "./wikiLinkCompletion";

export type MarkdownSourceEditorProps = {
  value: string;
  onChange: (next: string) => void;
  readOnly?: boolean;
  dark?: boolean;
  minHeight?: string;
  className?: string;
  onSaveShortcut?: () => void;
  pages?: WikiPage[];
  customSlashCommands?: EditorSlashCommandConfig[];
  loadSlashTemplate?: (templatePath: string) => Promise<string>;
  onSlashTemplateError?: (message: string) => void;
  uploadImage?: (file: File) => Promise<string>;
  onImageUploadError?: (message: string) => void;
};

export function MarkdownSourceEditor({
  value,
  onChange,
  readOnly = false,
  dark = false,
  minHeight = "60vh",
  className,
  onSaveShortcut,
  pages = [],
  customSlashCommands = [],
  loadSlashTemplate,
  onSlashTemplateError,
  uploadImage,
  onImageUploadError,
}: MarkdownSourceEditorProps) {
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
    if (customSlashCommands.length > 0 && loadSlashTemplate && onSlashTemplateError) {
      completionSources.push(
        customSlashCompletionSource(customSlashCommands, loadSlashTemplate, onSlashTemplateError),
      );
    }
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
  }, [
    onSaveShortcut,
    pages,
    customSlashCommands,
    loadSlashTemplate,
    onSlashTemplateError,
    uploadImage,
    onImageUploadError,
  ]);

  const theme = useMemo(() => markdownEditorTheme({ dark }), [dark]);

  const handleDragEnter = useCallback(
    (e: React.DragEvent) => {
      if (!uploadImage || !isOsFileDrag(e)) return;
      e.preventDefault();
      fileDragDepthRef.current += 1;
      setFileDragActive(true);
    },
    [uploadImage],
  );

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

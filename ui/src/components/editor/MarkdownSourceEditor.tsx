import CodeMirror from "@uiw/react-codemirror";
import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
import { highlightSelectionMatches, searchKeymap } from "@codemirror/search";
import { type Extension } from "@codemirror/state";
import { keymap } from "@codemirror/view";
import { cn } from "@/lib/cn";
import { markdownEditorExtensions } from "./markdownLanguage";
import { markdownEditorTheme } from "./markdownEditorTheme";
import { markdownSlashCommandExtension } from "./markdownSlashCommands";

export type MarkdownSourceEditorProps = {
  value: string;
  onChange: (next: string) => void;
  readOnly?: boolean;
  dark?: boolean;
  minHeight?: string;
  className?: string;
  onSaveShortcut?: () => void;
};

export function MarkdownSourceEditor({
  value,
  onChange,
  readOnly = false,
  dark = false,
  minHeight = "60vh",
  className,
  onSaveShortcut,
}: MarkdownSourceEditorProps) {
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
  const extensions: Extension[] = [
    history(),
    ...markdownEditorExtensions(),
    highlightSelectionMatches(),
    markdownSlashCommandExtension(),
    keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap]),
    saveKeymap,
  ];

  return (
    <div className={cn("overflow-hidden rounded-md border bg-background shadow-sm", className)} data-testid="markdown-source-editor">
      <CodeMirror
        value={value}
        height="auto"
        minHeight={minHeight}
        basicSetup={{
          lineNumbers: true,
          foldGutter: true,
          highlightActiveLine: true,
          highlightSelectionMatches: false,
          autocompletion: true,
          bracketMatching: true,
          closeBrackets: true,
          defaultKeymap: false,
          history: false,
          searchKeymap: false,
        }}
        editable={!readOnly}
        readOnly={readOnly}
        theme={markdownEditorTheme({ dark })}
        extensions={extensions}
        onChange={onChange}
      />
    </div>
  );
}

import { markdown, markdownLanguage } from "@codemirror/lang-markdown";
import { languages } from "@codemirror/language-data";
import { Table } from "@lezer/markdown";

export function markdownEditorExtensions() {
  return [
    markdown({
      base: markdownLanguage,
      codeLanguages: languages,
      extensions: [Table],
    }),
  ];
}

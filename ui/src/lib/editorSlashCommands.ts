import * as LucideIcons from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { BlockNoteEditor } from "@blocknote/core";
import { createElement } from "react";
import { api } from "./api";

export type EditorSlashCommandConfig = {
  id: string;
  label: string;
  icon: string;
  description: string;
  template: string;
};

export function resolveLucideIcon(name: string): LucideIcon {
  const icons = LucideIcons as unknown as Record<string, LucideIcon | undefined>;
  const trimmed = name.trim();
  if (!trimmed) return LucideIcons.FileText;
  return icons[trimmed] ?? LucideIcons.FileText;
}

export async function loadSlashCommandTemplate(templatePath: string): Promise<string> {
  const { content } = await api.readFile(templatePath);
  return content;
}

export function templateLoadErrorMessage(templatePath: string, err: unknown): string {
  const detail = err instanceof Error ? err.message : String(err);
  return `Could not load template "${templatePath}": ${detail}`;
}

export async function insertTemplateAtCursor(editor: BlockNoteEditor, markdown: string): Promise<void> {
  const cur = editor.getTextCursorPosition().block;
  try {
    const blocks = await editor.tryParseMarkdownToBlocks(markdown);
    if (blocks?.length) {
      editor.insertBlocks(blocks, cur, "after");
      return;
    }
  } catch {
    // fall through to plain paragraph insert
  }
  editor.insertBlocks([{ type: "paragraph", content: markdown }], cur, "after");
}

export function blockNoteSlashItems(
  editor: BlockNoteEditor,
  commands: EditorSlashCommandConfig[],
  onError: (message: string) => void,
) {
  return commands.map((cmd) => ({
    title: cmd.label || cmd.id,
    subtext: cmd.description || `Insert from ${cmd.template}`,
    aliases: [cmd.id, cmd.label].filter(Boolean),
    group: "Templates",
    icon: createElement(resolveLucideIcon(cmd.icon), { size: 18 }),
    onItemClick: () => {
      void loadSlashCommandTemplate(cmd.template)
        .then((content) => insertTemplateAtCursor(editor, content))
        .catch((err) => onError(templateLoadErrorMessage(cmd.template, err)));
    },
  }));
}

export function matchesSlashQuery(value: string, query: string): boolean {
  const normalized = query.toLowerCase();
  if (!normalized) return true;
  return value.toLowerCase().startsWith(normalized);
}

export function filterSlashCommands(
  commands: EditorSlashCommandConfig[],
  query: string,
): EditorSlashCommandConfig[] {
  return commands.filter(
    (cmd) => matchesSlashQuery(cmd.id, query) || matchesSlashQuery(cmd.label, query),
  );
}

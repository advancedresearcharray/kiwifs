import { autocompletion, type Completion, type CompletionContext } from "@codemirror/autocomplete";
import { type Extension } from "@codemirror/state";
import { type EditorView } from "@codemirror/view";

export type MarkdownSlashCommandName = "table" | "todo" | "code" | "quote" | "frontmatter";

export type MarkdownSlashCommand = {
  name: MarkdownSlashCommandName;
  icon: string;
  label: string;
  detail: string;
  insert: string;
  cursorOffset: number;
};

export type SlashTriggerRange = {
  from: number;
  to: number;
};

const frontmatterFieldInsert = "key: \n";

export const markdownSlashCommands: MarkdownSlashCommand[] = [
  {
    name: "table",
    icon: "📊",
    label: "Table",
    detail: "Insert a GFM table",
    insert: "| Column | Value |\n| --- | --- |\n|  |  |\n",
    cursorOffset: "| Column | Value |\n| --- | --- |\n| ".length,
  },
  {
    name: "todo",
    icon: "✅",
    label: "Todo",
    detail: "Insert a task list item",
    insert: "- [ ] ",
    cursorOffset: "- [ ] ".length,
  },
  {
    name: "code",
    icon: "💻",
    label: "Code block",
    detail: "Insert a fenced code block",
    insert: "```\n\n```",
    cursorOffset: "```\n".length,
  },
  {
    name: "quote",
    icon: "💬",
    label: "Quote",
    detail: "Insert a block quote",
    insert: "> ",
    cursorOffset: "> ".length,
  },
  {
    name: "frontmatter",
    icon: "🧾",
    label: "Frontmatter",
    detail: "Insert or edit YAML frontmatter",
    insert: "---\ntitle: \ntags: []\n---\n\n",
    cursorOffset: "---\ntitle: ".length,
  },
];

export function hasYamlFrontmatter(doc: string): boolean {
  return /^---\s*\n[\s\S]*?\n---(?:\s*\n|$)/.test(doc);
}

function frontmatterCloseDelimiterStart(doc: string): number | null {
  if (!doc.startsWith("---")) return null;
  const match = doc.match(/^---\s*\n[\s\S]*?\n---(?:\s*\n|$)/);
  if (!match) return null;
  return match[0].lastIndexOf("---");
}

export function getAvailableMarkdownSlashCommands(doc: string, query: string): MarkdownSlashCommand[] {
  const normalizedQuery = query.toLowerCase();
  return markdownSlashCommands.filter((command) => {
    if (command.name === "frontmatter" && hasYamlFrontmatter(doc)) return false;
    return command.name.startsWith(normalizedQuery);
  });
}

function applyFrontmatterCommand(
  doc: string,
  trigger: SlashTriggerRange,
  command: MarkdownSlashCommand,
): { doc: string; selection: number } {
  const delimiterStart = frontmatterCloseDelimiterStart(doc);
  if (delimiterStart === null) {
    const nextDoc = `${doc.slice(0, trigger.from)}${command.insert}${doc.slice(trigger.to)}`;
    return { doc: nextDoc, selection: trigger.from + command.cursorOffset };
  }

  const docWithoutTrigger = `${doc.slice(0, trigger.from)}${doc.slice(trigger.to)}`;
  const nextDoc = `${docWithoutTrigger.slice(0, delimiterStart)}${frontmatterFieldInsert}${docWithoutTrigger.slice(delimiterStart)}`;
  return {
    doc: nextDoc,
    selection: delimiterStart + "key: ".length,
  };
}

export function replaceSlashCommandTrigger(
  doc: string,
  trigger: SlashTriggerRange,
  command: MarkdownSlashCommand,
): { doc: string; selection: number } {
  if (command.name === "frontmatter") {
    return applyFrontmatterCommand(doc, trigger, command);
  }

  const nextDoc = `${doc.slice(0, trigger.from)}${command.insert}${doc.slice(trigger.to)}`;
  return {
    doc: nextDoc,
    selection: trigger.from + command.cursorOffset,
  };
}

function slashTriggerBeforeCursor(context: CompletionContext): SlashTriggerRange | null {
  const line = context.state.doc.lineAt(context.pos);
  const beforeCursor = line.text.slice(0, context.pos - line.from);
  const match = beforeCursor.match(/(?:^|\s)(\/[\w-]*)$/);
  if (!match || match.index === undefined) return null;

  const slashText = match[1];
  const from = line.from + match.index + match[0].length - slashText.length;
  return { from, to: context.pos };
}

function completionFor(command: MarkdownSlashCommand, trigger: SlashTriggerRange): Completion {
  return {
    label: `/${command.name}`,
    displayLabel: `${command.icon} /${command.name}`,
    type: "keyword",
    detail: command.detail,
    apply: (view: EditorView) => {
      const currentDoc = view.state.doc.toString();
      const { doc, selection } = replaceSlashCommandTrigger(currentDoc, trigger, command);
      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: doc },
        selection: { anchor: selection },
        userEvent: "input.complete",
      });
    },
  };
}

function slashCompletionSource(context: CompletionContext) {
  const trigger = slashTriggerBeforeCursor(context);
  if (!trigger) return null;

  const doc = context.state.doc.toString();
  const query = context.state.doc.sliceString(trigger.from + 1, trigger.to).toLowerCase();
  const options = getAvailableMarkdownSlashCommands(doc, query).map((command) => completionFor(command, trigger));

  return {
    from: trigger.from,
    to: trigger.to,
    options,
    validFor: /^\/[\w-]*$/,
  };
}

export { slashCompletionSource };

export function markdownSlashCommandExtension(): Extension {
  return autocompletion({
    override: [slashCompletionSource],
    activateOnTyping: true,
    closeOnBlur: true,
  });
}

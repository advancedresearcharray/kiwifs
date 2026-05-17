export type KanbanBlockCard = {
  id: string;
  title: string;
  tags?: string[];
  description?: string;
  priority?: string;
  assignee?: string;
};

export type KanbanBlockColumn = {
  name: string;
  color?: string;
  cards: KanbanBlockCard[];
};

export type KanbanBlockExportConfig = {
  format?: "markdown" | "json";
  copyLabel?: string;
};

export type KanbanBlockConfig = {
  title?: string;
  columns: KanbanBlockColumn[];
  export?: KanbanBlockExportConfig;
};

type ParserSection = "root" | "columns" | "column" | "cards" | "card" | "export";

export function parseKanbanBlockConfig(source: string): KanbanBlockConfig {
  const trimmed = source.trim();

  if (trimmed.startsWith("{")) {
    return JSON.parse(trimmed) as KanbanBlockConfig;
  }

  let title: string | undefined;
  const exportConfig: KanbanBlockExportConfig = {};
  const columns: KanbanBlockColumn[] = [];
  const lines = trimmed.split("\n");
  let section: ParserSection = "root";
  let currentColumn: KanbanBlockColumn = { name: "", cards: [] };
  let currentCard: Partial<KanbanBlockCard> = {};

  for (const line of lines) {
    const l = line.trim();
    if (!l || l.startsWith("#")) continue;

    const indent = line.length - line.trimStart().length;

    if (l.startsWith("title:") && indent === 0) {
      title = stripYamlQuotes(l.slice(6).trim());
      section = "root";
    } else if (l === "export:" && indent === 0) {
      section = "export";
    } else if (section === "export") {
      const kvMatch = l.match(/^([A-Za-z]+):\s*(.*)$/);
      if (kvMatch) {
        const [, key, value] = kvMatch;
        assignExportValue(exportConfig, key!, stripYamlQuotes(value!.trim()));
      }
      if (indent === 0 && l !== "export:") section = "root";
    } else if (l === "columns:" && indent === 0) {
      section = "columns";
    } else if (section === "columns" && l.startsWith("- name:")) {
      if (currentColumn.name) {
        if (currentCard.id) {
          currentColumn.cards.push(finalizeKanbanBlockCard(currentCard));
          currentCard = {};
        }
        columns.push(currentColumn);
      }
      currentColumn = { name: stripYamlQuotes(l.slice("- name:".length).trim()), cards: [] };
      section = "column";
    } else if (section === "column" && l.startsWith("color:")) {
      currentColumn.color = stripYamlQuotes(l.slice(6).trim());
    } else if ((section === "column" || section === "cards") && l === "cards:") {
      section = "cards";
    } else if (section === "cards" && l.startsWith("- id:")) {
      if (currentCard.id) {
        currentColumn.cards.push(finalizeKanbanBlockCard(currentCard));
      }
      currentCard = { id: stripYamlQuotes(l.slice("- id:".length).trim()) };
      section = "card";
    } else if (section === "card" || section === "cards") {
      const kvMatch = l.match(/^([A-Za-z]+):\s*(.*)$/);
      if (kvMatch) {
        const [, key, rawValue] = kvMatch;
        const value = stripYamlQuotes(rawValue!.trim());
        assignCardValue(currentCard, key!, value);
      }

      if (l.startsWith("- name:")) {
        if (currentCard.id) {
          currentColumn.cards.push(finalizeKanbanBlockCard(currentCard));
          currentCard = {};
        }
        columns.push(currentColumn);
        currentColumn = { name: stripYamlQuotes(l.slice("- name:".length).trim()), cards: [] };
        section = "column";
      }
    }
  }

  if (currentCard.id) {
    currentColumn.cards.push(finalizeKanbanBlockCard(currentCard));
  }
  if (currentColumn.name) {
    columns.push(currentColumn);
  }

  return { title, columns, export: exportConfig };
}

export function buildKanbanBlockExportText(
  columns: KanbanBlockColumn[],
  format: KanbanBlockExportConfig["format"] = "markdown",
): string {
  if (format === "json") {
    return JSON.stringify(
      columns.map((col) => ({
        column: col.name,
        cards: col.cards.map((card) => ({ id: card.id, title: card.title, tags: card.tags })),
      })),
      null,
      2,
    );
  }

  const parts: string[] = [];
  for (const col of columns) {
    parts.push(`## ${col.name}`);
    for (const card of col.cards) {
      const tagStr = card.tags?.length ? ` [${card.tags.join(", ")}]` : "";
      parts.push(`- **${card.title}**${tagStr}`);
    }
    parts.push("");
  }
  return parts.join("\n");
}

function finalizeKanbanBlockCard(raw: Partial<KanbanBlockCard>): KanbanBlockCard {
  return {
    id: raw.id || `card-${Math.random().toString(36).slice(2, 8)}`,
    title: raw.title || "Untitled",
    tags: raw.tags,
    description: raw.description,
    priority: raw.priority,
    assignee: raw.assignee,
  };
}

function assignExportValue(config: KanbanBlockExportConfig, key: string, value: string) {
  if (key === "format" && (value === "markdown" || value === "json")) {
    config.format = value;
  } else if (key === "copyLabel") {
    config.copyLabel = value;
  }
}

function assignCardValue(card: Partial<KanbanBlockCard>, key: string, value: string) {
  if (key === "title") card.title = value;
  else if (key === "description") card.description = value;
  else if (key === "priority") card.priority = value;
  else if (key === "assignee") card.assignee = value;
  else if (key === "tags" && value.startsWith("[") && value.endsWith("]")) {
    card.tags = value.slice(1, -1).split(",").map((s) => stripYamlQuotes(s.trim()));
  }
}

function stripYamlQuotes(value: string): string {
  return value.replace(/^["']|["']$/g, "");
}

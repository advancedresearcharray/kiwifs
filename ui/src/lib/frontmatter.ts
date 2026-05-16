export type FrontmatterSplit = {
  frontmatter: string | null;
  body: string;
};

const FRONTMATTER_DELIMITER = "---";
const FRONTMATTER_OPEN_RE = /^\uFEFF?---[ \t]*\r?\n/;
const FRONTMATTER_CLOSE_RE = /^---[ \t]*$/;

export function splitFrontmatter(markdown: string): FrontmatterSplit {
  const open = markdown.match(FRONTMATTER_OPEN_RE);
  if (!open) return { frontmatter: null, body: markdown };

  const bodyStart = open[0].length;
  const lines = markdown.slice(bodyStart).split(/(\r?\n)/);
  let cursor = bodyStart;

  for (let i = 0; i < lines.length; i += 2) {
    const line = lines[i] ?? "";
    const newline = lines[i + 1] ?? "";
    if (FRONTMATTER_CLOSE_RE.test(line)) {
      const closeEnd = cursor + line.length + newline.length;
      const frontmatter = markdown.slice(0, cursor + line.length);
      const body = markdown.slice(closeEnd).replace(/^\r?\n/, "");
      return { frontmatter, body };
    }
    cursor += line.length + newline.length;
  }

  return { frontmatter: null, body: markdown };
}

export function frontmatterToText(frontmatter: string | null): string {
  if (!frontmatter) return "";
  const withoutBom = frontmatter.replace(/^\uFEFF/, "");
  const lines = withoutBom.split(/\r?\n/);
  if (lines[0]?.trim() !== FRONTMATTER_DELIMITER) return "";
  if (lines[lines.length - 1]?.trim() === FRONTMATTER_DELIMITER) {
    return lines.slice(1, -1).join("\n");
  }
  return lines.slice(1).join("\n");
}

export function textToFrontmatter(text: string): string | null {
  const trimmed = text.trim();
  if (!trimmed) return null;
  return `${FRONTMATTER_DELIMITER}\n${trimmed}\n${FRONTMATTER_DELIMITER}`;
}

export function joinFrontmatter(frontmatterText: string, body: string): string {
  const frontmatter = textToFrontmatter(frontmatterText);
  if (!frontmatter) return body;
  return `${frontmatter}\n\n${body.replace(/^\s*\n/, "")}`;
}

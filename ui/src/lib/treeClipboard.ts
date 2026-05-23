export type TreeClipboardEntry = {
  path: string;
  content?: string;
};

export type TreeClipboard = {
  mode: "copy" | "cut";
  entries: TreeClipboardEntry[];
};

let clipboard: TreeClipboard | null = null;

export function getTreeClipboard(): TreeClipboard | null {
  return clipboard;
}

export function setTreeClipboard(data: TreeClipboard | null): void {
  clipboard = data;
}

export function clearTreeClipboard(): void {
  clipboard = null;
}

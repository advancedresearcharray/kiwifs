import { stripTrailingSlash } from "@kw/lib/paths";

export type TreeRevealRequest = {
  path: string;
  nonce: number;
};

export function parentPathsFor(path: string): string[] {
  const parts = stripTrailingSlash(path).split("/").filter(Boolean);
  const parents: string[] = [];
  for (let i = 1; i < parts.length; i += 1) {
    parents.push(parts.slice(0, i).join("/"));
  }
  return parents;
}

export function nextExpandedForReveal(previous: Set<string>, revealPath: string | null | undefined): Set<string> {
  if (!revealPath) return previous;

  const next = new Set(previous);
  next.add("");
  for (const parent of parentPathsFor(revealPath)) {
    next.add(parent);
  }
  return next;
}

export function shouldFocusRevealTarget(
  revealRequest: TreeRevealRequest | null | undefined,
  nodePath: string,
): boolean {
  return revealRequest?.path === nodePath;
}

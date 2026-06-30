export type ParsedWikiLinkHref =
  | { kind: "resolved"; pagePath: string; anchor?: string }
  | { kind: "missing"; pagePath: string }
  | { kind: "other" };

/** True when hover previews are appropriate (fine pointer + hover capability). */
export function canOpenHoverPreview(
  matchMedia: (query: string) => MediaQueryList = typeof window !== "undefined"
    ? window.matchMedia.bind(window)
    : () => ({ matches: true }) as MediaQueryList,
): boolean {
  return matchMedia("(hover: hover) and (pointer: fine)").matches;
}

export function parseWikiLinkHref(href: string): ParsedWikiLinkHref {
  if (href.startsWith("#kiwi:")) {
    const raw = href.slice("#kiwi:".length);
    const hashIdx = raw.indexOf("#");
    const pagePath = hashIdx >= 0 ? raw.slice(0, hashIdx) : raw;
    const anchor = hashIdx >= 0 ? raw.slice(hashIdx) : undefined;
    return { kind: "resolved", pagePath, anchor: anchor || undefined };
  }

  if (href.startsWith("#kiwi-missing:")) {
    return { kind: "missing", pagePath: href.slice("#kiwi-missing:".length) };
  }

  return { kind: "other" };
}

import { api, type PeekResponse } from "@kw/lib/api";

export type WikiLinkPeekData = {
  path: string;
  title: string;
  snippet: string;
  tags: string[];
};

export type WikiLinkPeekResult =
  | { status: "ok"; data: WikiLinkPeekData }
  | { status: "not_found" }
  | { status: "error"; message: string };

const SNIPPET_MAX = 200;

const cache = new Map<string, WikiLinkPeekResult>();
const inflight = new Map<string, Promise<WikiLinkPeekResult>>();

export function peekTags(frontmatter: unknown): string[] {
  if (!frontmatter || typeof frontmatter !== "object") return [];
  const tags = (frontmatter as Record<string, unknown>).tags;
  if (Array.isArray(tags)) return tags.filter((t): t is string => typeof t === "string");
  if (typeof tags === "string" && tags.trim()) return [tags.trim()];
  return [];
}

export function peekTitle(response: PeekResponse): string {
  const fm = response.frontmatter;
  if (fm && typeof fm === "object") {
    const title = (fm as Record<string, unknown>).title;
    if (typeof title === "string" && title.trim()) return title.trim();
  }
  return response.title;
}

export function truncateSnippet(snippet: string, max = SNIPPET_MAX): string {
  const trimmed = snippet.trim();
  if (trimmed.length <= max) return trimmed;
  return `${trimmed.slice(0, max).trimEnd()}…`;
}

function toPeekData(response: PeekResponse): WikiLinkPeekData {
  return {
    path: response.path,
    title: peekTitle(response),
    snippet: truncateSnippet(response.snippet),
    tags: peekTags(response.frontmatter),
  };
}

export function clearWikiLinkPeekCache(): void {
  cache.clear();
  inflight.clear();
}

export function getCachedWikiLinkPeek(path: string): WikiLinkPeekResult | undefined {
  return cache.get(path);
}

export async function fetchWikiLinkPeek(path: string): Promise<WikiLinkPeekResult> {
  const cached = cache.get(path);
  if (cached) return cached;

  const pending = inflight.get(path);
  if (pending) return pending;

  const promise = (async (): Promise<WikiLinkPeekResult> => {
    try {
      const response = await api.peek(path);
      const result: WikiLinkPeekResult = { status: "ok", data: toPeekData(response) };
      cache.set(path, result);
      return result;
    } catch (error) {
      const status = (error as { status?: number }).status;
      if (status === 404) {
        const result: WikiLinkPeekResult = { status: "not_found" };
        cache.set(path, result);
        return result;
      }
      const message = error instanceof Error ? error.message : "Failed to load preview";
      return { status: "error", message };
    } finally {
      inflight.delete(path);
    }
  })();

  inflight.set(path, promise);
  return promise;
}

import { ApiError, type PeekResponse } from "@kw/lib/api";

export type PeekData = {
  path: string;
  title: string;
  snippet: string;
  tags: string[];
};

export type PeekResult = PeekData | { notFound: true };

export function extractTagsFromFrontmatter(frontmatter: unknown): string[] {
  if (!frontmatter || typeof frontmatter !== "object") return [];
  const tags = (frontmatter as Record<string, unknown>).tags;
  if (Array.isArray(tags)) {
    return tags.filter((tag): tag is string => typeof tag === "string" && tag.length > 0);
  }
  if (typeof tags === "string" && tags.length > 0) return [tags];
  return [];
}

export function truncateSnippet(snippet: string, max = 200): string {
  const trimmed = snippet.trim();
  if (trimmed.length <= max) return trimmed;
  return `${trimmed.slice(0, max).trimEnd()}…`;
}

const cache = new Map<string, PeekResult>();
const pending = new Map<string, Promise<PeekResult>>();

export function clearPeekCache(): void {
  cache.clear();
  pending.clear();
}

export async function fetchPeekData(
  path: string,
  peekFn: (path: string) => Promise<PeekResponse>,
): Promise<PeekResult> {
  const cached = cache.get(path);
  if (cached) return cached;

  const inflight = pending.get(path);
  if (inflight) return inflight;

  const request = (async () => {
    try {
      const data = await peekFn(path);
      const result: PeekData = {
        path: data.path,
        title: data.title,
        snippet: truncateSnippet(data.snippet),
        tags: extractTagsFromFrontmatter(data.frontmatter),
      };
      cache.set(path, result);
      return result;
    } catch (error) {
      if (error instanceof ApiError && error.status === 404) {
        const notFound = { notFound: true as const };
        cache.set(path, notFound);
        return notFound;
      }
      throw error;
    } finally {
      pending.delete(path);
    }
  })();

  pending.set(path, request);
  return request;
}

export function supportsHoverPreview(): boolean {
  if (typeof window === "undefined") return false;
  return window.matchMedia("(hover: hover) and (pointer: fine)").matches;
}

import { api, ApiError, type PeekResponse } from "@kw/lib/api";

export type WikiLinkPreviewData = {
  path: string;
  title: string;
  snippet: string;
  tags: string[];
};

export function extractTagsFromFrontmatter(frontmatter: unknown): string[] {
  if (!frontmatter || typeof frontmatter !== "object") return [];
  const tags = (frontmatter as Record<string, unknown>).tags;
  if (Array.isArray(tags)) {
    return tags.filter((tag): tag is string => typeof tag === "string");
  }
  if (typeof tags === "string" && tags.trim()) return [tags.trim()];
  return [];
}

export function titleFromPeek(peek: PeekResponse): string {
  const fm = peek.frontmatter;
  if (fm && typeof fm === "object") {
    const title = (fm as Record<string, unknown>).title;
    if (typeof title === "string" && title.trim()) return title.trim();
  }
  return peek.title;
}

export function peekToPreview(peek: PeekResponse): WikiLinkPreviewData {
  return {
    path: peek.path,
    title: titleFromPeek(peek),
    snippet: peek.snippet,
    tags: extractTagsFromFrontmatter(peek.frontmatter),
  };
}

type CacheEntry = WikiLinkPreviewData | "missing";

export class WikiLinkPreviewCache {
  private cache = new Map<string, CacheEntry>();
  private inflight = new Map<string, Promise<CacheEntry>>();

  get(path: string): CacheEntry | undefined {
    return this.cache.get(path);
  }

  clear(): void {
    this.cache.clear();
    this.inflight.clear();
  }

  async load(path: string): Promise<CacheEntry> {
    const cached = this.cache.get(path);
    if (cached) return cached;

    let pending = this.inflight.get(path);
    if (!pending) {
      pending = api
        .peek(path)
        .then((peek) => {
          const data = peekToPreview(peek);
          this.cache.set(path, data);
          return data;
        })
        .catch((error: unknown) => {
          if (error instanceof ApiError && error.status === 404) {
            this.cache.set(path, "missing");
            return "missing" as const;
          }
          this.inflight.delete(path);
          throw error;
        })
        .finally(() => {
          this.inflight.delete(path);
        });
      this.inflight.set(path, pending);
    }

    return pending;
  }
}

export const wikiLinkPreviewCache = new WikiLinkPreviewCache();

import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError } from "./api";
import {
  WikiLinkPreviewCache,
  extractTagsFromFrontmatter,
  peekToPreview,
  titleFromPeek,
} from "./wikiLinkPreviewCache";

describe("wikiLinkPreviewCache", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("extracts string and array tags from frontmatter", () => {
    expect(extractTagsFromFrontmatter({ tags: ["docs", "links"] })).toEqual([
      "docs",
      "links",
    ]);
    expect(extractTagsFromFrontmatter({ tags: "solo" })).toEqual(["solo"]);
    expect(extractTagsFromFrontmatter({})).toEqual([]);
  });

  it("prefers frontmatter title over peek heading title", () => {
    expect(
      titleFromPeek({
        path: "pages/auth.md",
        title: "Authentication",
        frontmatter: { title: "Auth Guide" },
        snippet: "",
        links_out: [],
        links_in: [],
        word_count: 0,
        headings: [],
      }),
    ).toBe("Auth Guide");
  });

  it("maps peek responses into preview cards", () => {
    expect(
      peekToPreview({
        path: "pages/wikilinks.md",
        title: "Wiki Links",
        frontmatter: { tags: ["documentation"] },
        snippet: "Cross-link pages with [[wikilinks]].",
        links_out: [],
        links_in: [],
        word_count: 120,
        headings: ["Wiki Links"],
      }),
    ).toEqual({
      path: "pages/wikilinks.md",
      title: "Wiki Links",
      snippet: "Cross-link pages with [[wikilinks]].",
      tags: ["documentation"],
    });
  });

  it("deduplicates concurrent peek requests for the same path", async () => {
    const peek = vi.fn(async () => ({
      path: "pages/auth.md",
      title: "Authentication",
      frontmatter: { tags: ["security"] },
      snippet: "Sign in flow.",
      links_out: [],
      links_in: [],
      word_count: 10,
      headings: [],
    }));
    vi.spyOn(await import("./api"), "api", "get").mockReturnValue({ peek } as never);

    const cache = new WikiLinkPreviewCache();
    const [first, second] = await Promise.all([
      cache.load("pages/auth.md"),
      cache.load("pages/auth.md"),
    ]);

    expect(peek).toHaveBeenCalledTimes(1);
    expect(first).toEqual(second);
    expect(cache.get("pages/auth.md")).toEqual({
      path: "pages/auth.md",
      title: "Authentication",
      snippet: "Sign in flow.",
      tags: ["security"],
    });
  });

  it("caches missing pages after a 404 peek", async () => {
    const peek = vi.fn(async () => {
      throw new ApiError(404, "Not Found", "missing", "/peek");
    });
    vi.spyOn(await import("./api"), "api", "get").mockReturnValue({ peek } as never);

    const cache = new WikiLinkPreviewCache();
    await expect(cache.load("ghost.md")).resolves.toBe("missing");
    await expect(cache.load("ghost.md")).resolves.toBe("missing");
    expect(peek).toHaveBeenCalledTimes(1);
  });

  it("rethrows non-404 peek errors without caching", async () => {
    const peek = vi.fn(async () => {
      throw new ApiError(500, "Internal Server Error", "boom", "/peek");
    });
    vi.spyOn(await import("./api"), "api", "get").mockReturnValue({ peek } as never);

    const cache = new WikiLinkPreviewCache();
    await expect(cache.load("broken.md")).rejects.toMatchObject({ status: 500 });
    expect(cache.get("broken.md")).toBeUndefined();
    expect(peek).toHaveBeenCalledTimes(1);
  });
});

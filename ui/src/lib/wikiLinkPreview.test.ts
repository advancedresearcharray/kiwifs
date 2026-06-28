import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError } from "./api";
import {
  clearPeekCache,
  extractTagsFromFrontmatter,
  fetchPeekData,
  truncateSnippet,
} from "./wikiLinkPreview";

describe("extractTagsFromFrontmatter", () => {
  it("returns string tags as a single-item array", () => {
    expect(extractTagsFromFrontmatter({ tags: "docs" })).toEqual(["docs"]);
  });

  it("returns array tags in order", () => {
    expect(extractTagsFromFrontmatter({ tags: ["alpha", "beta"] })).toEqual([
      "alpha",
      "beta",
    ]);
  });

  it("ignores non-string tag entries", () => {
    expect(extractTagsFromFrontmatter({ tags: ["ok", 42, null] })).toEqual(["ok"]);
  });

  it("returns empty array when tags are absent", () => {
    expect(extractTagsFromFrontmatter({ title: "Hello" })).toEqual([]);
    expect(extractTagsFromFrontmatter(null)).toEqual([]);
  });
});

describe("truncateSnippet", () => {
  it("leaves short snippets unchanged", () => {
    expect(truncateSnippet("Short preview text.")).toBe("Short preview text.");
  });

  it("truncates long snippets to 200 characters with ellipsis", () => {
    const long = "word ".repeat(60).trim();
    const result = truncateSnippet(long);
    expect(result.length).toBeLessThanOrEqual(201);
    expect(result.endsWith("…")).toBe(true);
  });
});

describe("fetchPeekData", () => {
  afterEach(() => {
    clearPeekCache();
    vi.restoreAllMocks();
  });

  it("maps peek responses into preview data", async () => {
    const peekFn = vi.fn(async () => ({
      path: "pages/guide.md",
      title: "Guide",
      snippet: "Getting started with KiwiFS.",
      frontmatter: { tags: ["docs", "guide"] },
    }));

    await expect(fetchPeekData("pages/guide.md", peekFn)).resolves.toEqual({
      path: "pages/guide.md",
      title: "Guide",
      snippet: "Getting started with KiwiFS.",
      tags: ["docs", "guide"],
    });
    expect(peekFn).toHaveBeenCalledTimes(1);
  });

  it("deduplicates concurrent requests for the same path", async () => {
    let resolvePeek: ((value: {
      path: string;
      title: string;
      snippet: string;
      frontmatter: unknown;
    }) => void) | undefined;
    const peekFn = vi.fn(
      () =>
        new Promise<{ path: string; title: string; snippet: string; frontmatter: unknown }>(
          (resolve) => {
            resolvePeek = resolve;
          },
        ),
    );

    const first = fetchPeekData("pages/shared.md", peekFn);
    const second = fetchPeekData("pages/shared.md", peekFn);
    expect(peekFn).toHaveBeenCalledTimes(1);

    resolvePeek?.({
      path: "pages/shared.md",
      title: "Shared",
      snippet: "Body",
      frontmatter: { tags: ["x"] },
    });

    await expect(first).resolves.toMatchObject({ title: "Shared", tags: ["x"] });
    await expect(second).resolves.toMatchObject({ title: "Shared", tags: ["x"] });
  });

  it("caches 404 responses as not found", async () => {
    const peekFn = vi
      .fn()
      .mockRejectedValueOnce(new ApiError(404, "Not Found", "missing", "/peek"));

    await expect(fetchPeekData("pages/missing.md", peekFn)).resolves.toEqual({
      notFound: true,
    });
    await expect(fetchPeekData("pages/missing.md", peekFn)).resolves.toEqual({
      notFound: true,
    });
    expect(peekFn).toHaveBeenCalledTimes(1);
  });
});

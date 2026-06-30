import { afterEach, describe, expect, it, vi } from "vitest";
import { api, setBaseOverride } from "./api";
import {
  clearWikiLinkPeekCache,
  fetchWikiLinkPeek,
  getCachedWikiLinkPeek,
  peekTags,
  peekTitle,
  truncateSnippet,
} from "./wikiLinkPeek";

describe("wikiLinkPeek", () => {
  afterEach(() => {
    clearWikiLinkPeekCache();
    setBaseOverride(null);
    vi.restoreAllMocks();
  });

  describe("peekTags", () => {
    it("extracts string array tags from frontmatter", () => {
      expect(peekTags({ tags: ["docs", "api"] })).toEqual(["docs", "api"]);
    });

    it("wraps a single string tag", () => {
      expect(peekTags({ tags: "guide" })).toEqual(["guide"]);
    });

    it("returns empty for missing frontmatter", () => {
      expect(peekTags(null)).toEqual([]);
    });
  });

  describe("peekTitle", () => {
    it("prefers frontmatter title over API title", () => {
      expect(
        peekTitle({
          path: "pages/foo.md",
          title: "Heading Title",
          frontmatter: { title: "Frontmatter Title" },
          snippet: "",
          links_out: [],
          links_in: [],
          word_count: 0,
          headings: [],
        }),
      ).toBe("Frontmatter Title");
    });
  });

  describe("truncateSnippet", () => {
    it("truncates long snippets with ellipsis", () => {
      const long = "a".repeat(250);
      expect(truncateSnippet(long, 200)).toBe(`${"a".repeat(200)}…`);
    });

    it("leaves short snippets unchanged", () => {
      expect(truncateSnippet("hello world")).toBe("hello world");
    });
  });

  describe("fetchWikiLinkPeek", () => {
    it("caches successful peek responses", async () => {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () =>
          new Response(
            JSON.stringify({
              path: "pages/guide.md",
              title: "Guide",
              frontmatter: { tags: ["docs"] },
              snippet: "Getting started with KiwiFS.",
              links_out: [],
              links_in: [],
              word_count: 42,
              headings: ["Guide"],
            }),
            { status: 200, headers: { "Content-Type": "application/json" } },
          ),
        ),
      );
      setBaseOverride("/api/kiwi");

      const first = await fetchWikiLinkPeek("pages/guide.md");
      const second = await fetchWikiLinkPeek("pages/guide.md");

      expect(first.status).toBe("ok");
      if (first.status === "ok") {
        expect(first.data.title).toBe("Guide");
        expect(first.data.tags).toEqual(["docs"]);
        expect(first.data.snippet).toBe("Getting started with KiwiFS.");
      }
      expect(second).toBe(first);
      expect(getCachedWikiLinkPeek("pages/guide.md")).toBe(first);
      expect(fetch).toHaveBeenCalledTimes(1);
    });

    it("deduplicates concurrent fetches for the same path", async () => {
      let resolveFetch: (value: Response) => void = () => {};
      const fetchMock = vi.fn(
        () =>
          new Promise<Response>((resolve) => {
            resolveFetch = resolve;
          }),
      );
      vi.stubGlobal("fetch", fetchMock);
      setBaseOverride("/api/kiwi");

      const pendingA = fetchWikiLinkPeek("pages/shared.md");
      const pendingB = fetchWikiLinkPeek("pages/shared.md");

      resolveFetch!(
        new Response(
          JSON.stringify({
            path: "pages/shared.md",
            title: "Shared",
            frontmatter: {},
            snippet: "Body",
            links_out: [],
            links_in: [],
            word_count: 1,
            headings: [],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      );

      const [a, b] = await Promise.all([pendingA, pendingB]);
      expect(a).toBe(b);
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });

    it("maps 404 responses to not_found", async () => {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () => new Response("not found", { status: 404, statusText: "Not Found" })),
      );
      setBaseOverride("/api/kiwi");

      const result = await fetchWikiLinkPeek("pages/missing.md");
      expect(result).toEqual({ status: "not_found" });
      expect(getCachedWikiLinkPeek("pages/missing.md")).toEqual({ status: "not_found" });
    });
  });
});

describe("api.peek", () => {
  afterEach(() => {
    setBaseOverride(null);
    vi.restoreAllMocks();
  });

  it("requests /api/kiwi/peek with path query", async () => {
    const fetchMock = vi.fn(async () =>
      new Response(
        JSON.stringify({
          path: "pages/foo.md",
          title: "Foo",
          frontmatter: {},
          snippet: "Snippet",
          links_out: [],
          links_in: [],
          word_count: 3,
          headings: [],
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    );
    vi.stubGlobal("fetch", fetchMock);
    setBaseOverride("/api/kiwi");

    const result = await api.peek("pages/foo.md");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/kiwi/peek?path=pages%2Ffoo.md",
      expect.any(Object),
    );
    expect(result.path).toBe("pages/foo.md");
    expect(result.snippet).toBe("Snippet");
  });
});

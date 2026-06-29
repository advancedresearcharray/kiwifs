import { afterEach, describe, expect, it, vi } from "vitest";
import { api, ApiError, setBaseOverride } from "./api";

describe("api.peek", () => {
  afterEach(() => {
    setBaseOverride(null);
    vi.restoreAllMocks();
  });

  it("fetches lightweight page summary from /peek", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string) => {
        expect(url).toContain("/api/kiwi/peek?path=pages%2Fauth.md");
        return new Response(
          JSON.stringify({
            path: "pages/auth.md",
            title: "Authentication",
            frontmatter: { title: "Auth Guide", tags: ["security"] },
            snippet: "How users sign in.",
            links_out: [],
            links_in: [],
            word_count: 42,
            headings: ["Authentication"],
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }),
    );

    setBaseOverride("/api/kiwi");

    await expect(api.peek("pages/auth.md")).resolves.toMatchObject({
      path: "pages/auth.md",
      title: "Authentication",
      snippet: "How users sign in.",
    });
  });

  it("surfaces 404 responses for missing pages", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response("not found", { status: 404, statusText: "Not Found" })),
    );

    setBaseOverride("/api/kiwi");

    await expect(api.peek("missing.md")).rejects.toBeInstanceOf(ApiError);
  });
});

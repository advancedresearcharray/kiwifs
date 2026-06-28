import { afterEach, describe, expect, it, vi } from "vitest";
import { api, setBaseOverride } from "./api";

describe("api error handling", () => {
  afterEach(() => {
    setBaseOverride(null);
    vi.restoreAllMocks();
  });

  it("uses canonical merge=frontmatter PATCH with flat JSON body", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({ path: "doc.md", etag: "abc123" })
    );
    vi.stubGlobal("fetch", fetchMock);

    setBaseOverride("/api/kiwi");

    await api.patchFrontmatter("doc.md", { order: 2 }, '"etag-1"');

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/kiwi/file?path=doc.md&merge=frontmatter",
      expect.objectContaining({
        method: "PATCH",
        headers: expect.objectContaining({
          "Content-Type": "application/json",
          "If-Match": '"etag-1"',
        }),
        body: JSON.stringify({ order: 2 }),
      })
    );
  });

  it("preserves status and response body for failed frontmatter patches", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        new Response(
          'validation failed: frontmatter-yaml-invalid mapping key "last_reviewed" already defined',
          { status: 422, statusText: "Unprocessable Entity" }
        )
      )
    );

    setBaseOverride("/api/kiwi");

    await expect(api.patchFrontmatter("bad.md", { order: 1 })).rejects.toMatchObject({
      name: "ApiError",
      status: 422,
      statusText: "Unprocessable Entity",
      body: 'validation failed: frontmatter-yaml-invalid mapping key "last_reviewed" already defined',
    });
  });

  it("fetches page peek metadata from /peek", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        path: "pages/guide.md",
        title: "Guide",
        snippet: "Intro paragraph",
        frontmatter: { tags: ["docs"] },
      })
    );
    vi.stubGlobal("fetch", fetchMock);
    setBaseOverride("/api/kiwi");

    await expect(api.peek("pages/guide.md")).resolves.toMatchObject({
      path: "pages/guide.md",
      title: "Guide",
      snippet: "Intro paragraph",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/kiwi/peek?path=pages%2Fguide.md",
      expect.objectContaining({
        headers: expect.objectContaining({ "X-Actor": expect.any(String) }),
      })
    );
  });
});

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

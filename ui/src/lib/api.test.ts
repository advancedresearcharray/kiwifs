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

    expect(fetchMock).toHaveBeenCalledOnce();
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe("/api/kiwi/file?path=doc.md&merge=frontmatter");
    expect(init.method).toBe("PATCH");
    expect(init.headers).toMatchObject({
      "Content-Type": "application/json",
      "If-Match": '"etag-1"',
    });
    expect(init.body).toBe(JSON.stringify({ order: 2 }));
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
});

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

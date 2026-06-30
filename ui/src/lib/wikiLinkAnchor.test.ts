import { describe, expect, it } from "vitest";
import { canOpenHoverPreview, parseWikiLinkHref } from "./wikiLinkAnchor";

describe("canOpenHoverPreview", () => {
  it("allows preview on fine pointer with hover", () => {
    const matchMedia = (query: string) =>
      ({ matches: query === "(hover: hover) and (pointer: fine)" }) as MediaQueryList;
    expect(canOpenHoverPreview(matchMedia)).toBe(true);
  });

  it("disables preview on coarse pointer devices", () => {
    const matchMedia = () => ({ matches: false }) as MediaQueryList;
    expect(canOpenHoverPreview(matchMedia)).toBe(false);
  });
});

describe("parseWikiLinkHref", () => {
  it("returns other for same-page hash links", () => {
    expect(parseWikiLinkHref("#intro")).toEqual({ kind: "other" });
  });

  it("returns other for external links", () => {
    expect(parseWikiLinkHref("https://example.com")).toEqual({ kind: "other" });
  });

  it("parses resolved wiki links", () => {
    expect(parseWikiLinkHref("#kiwi:pages/guide.md")).toEqual({
      kind: "resolved",
      pagePath: "pages/guide.md",
      anchor: undefined,
    });
  });

  it("parses anchor suffix on resolved wiki links", () => {
    expect(parseWikiLinkHref("#kiwi:pages/guide.md#section")).toEqual({
      kind: "resolved",
      pagePath: "pages/guide.md",
      anchor: "#section",
    });
  });

  it("parses missing wiki links", () => {
    expect(parseWikiLinkHref("#kiwi-missing:pages/new-page")).toEqual({
      kind: "missing",
      pagePath: "pages/new-page",
    });
  });
});

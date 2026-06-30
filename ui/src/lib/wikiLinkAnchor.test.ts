import { describe, expect, it } from "vitest";
import { parseWikiLinkHref } from "./wikiLinkAnchor";

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

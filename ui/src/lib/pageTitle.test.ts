import { describe, expect, it } from "vitest";
import { formatDocumentTitle } from "./pageTitle";

describe("formatDocumentTitle", () => {
  it("returns app name when no page is active", () => {
    expect(formatDocumentTitle(null, "KiwiFS")).toBe("KiwiFS");
    expect(formatDocumentTitle(null, "Acme KB")).toBe("Acme KB");
  });

  it("combines page title and app name when navigating", () => {
    expect(formatDocumentTitle("getting-started.md", "KiwiFS")).toBe(
      "Getting Started · KiwiFS",
    );
    expect(formatDocumentTitle("docs/api-reference.md", "Acme KB")).toBe(
      "Api Reference · Acme KB",
    );
  });

  it("falls back to app name when titleize yields empty", () => {
    expect(formatDocumentTitle(".md", "KiwiFS")).toBe("KiwiFS");
  });
});

import { describe, expect, it } from "vitest";
import { sanitizeCustomCSS } from "./kiwiTheme";

describe("sanitizeCustomCSS", () => {
  it("strips script tags case-insensitively", () => {
    expect(sanitizeCustomCSS(".x{color:red}<script>alert(1)</script>")).toBe(".x{color:red}");
    expect(sanitizeCustomCSS("a{}<SCRIPT>x</SCRIPT>b{}")).toBe("a{}b{}");
  });

  it("preserves valid CSS", () => {
    const css = ".kiwi-admonition-note { border-color: hotpink; }";
    expect(sanitizeCustomCSS(css)).toBe(css);
  });
});

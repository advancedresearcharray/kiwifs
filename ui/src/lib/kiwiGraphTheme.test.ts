import { describe, expect, it } from "vitest";
import { colorForGraphCommunity, type KiwiGraphTheme } from "./kiwiGraphTheme";

const theme: KiwiGraphTheme = {
  defaultNode: "#7c8a6e",
  edge: "#c8c8c8",
  nodeDim: "#333333",
  edgeGhost: "#2a2a2a",
  edgeStrong: "#cccccc",
  palette: ["#111111", "#222222"],
};

describe("colorForGraphCommunity", () => {
  it("uses configured palette colors first", () => {
    expect(colorForGraphCommunity(0, theme)).toBe("#111111");
    expect(colorForGraphCommunity(1, theme)).toBe("#222222");
  });

  it("uses chromatic fallback colors beyond the configured palette", () => {
    expect(colorForGraphCommunity(2, theme)).toMatch(/^hsl\(\d+ 72% 58%\)$/);
    expect(colorForGraphCommunity(3, theme)).toMatch(/^hsl\(\d+ 72% 58%\)$/);
    expect(colorForGraphCommunity(2, theme)).not.toBe(colorForGraphCommunity(3, theme));
    expect(colorForGraphCommunity(2, theme)).not.toContain("0 0%");
  });
});

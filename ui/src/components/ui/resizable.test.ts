import { describe, expect, it } from "vitest";
import { resolveInitialPanelSizes } from "./resizable";

describe("resolveInitialPanelSizes", () => {
  it("honors defaultLayout when length matches panel count", () => {
    expect(resolveInitialPanelSizes(2, [60, 40])).toEqual([60, 40]);
  });

  it("falls back to even split when layout length mismatches", () => {
    expect(resolveInitialPanelSizes(2, [60])).toEqual([50, 50]);
  });

  it("falls back when layout contains non-positive values", () => {
    expect(resolveInitialPanelSizes(2, [0, 100])).toEqual([50, 50]);
    expect(resolveInitialPanelSizes(2, [NaN, 40])).toEqual([50, 50]);
  });

  it("defaults to two equal panes when layout is omitted", () => {
    expect(resolveInitialPanelSizes(2)).toEqual([50, 50]);
  });
});

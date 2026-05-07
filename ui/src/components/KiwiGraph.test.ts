import { describe, expect, it } from "vitest";
import { getGraphPerformanceProfile } from "../lib/graphPerformance";

describe("graph performance profile", () => {
  it("keeps labels and full layout iterations for small graphs", () => {
    expect(getGraphPerformanceProfile(250)).toEqual({
      forceAtlas2Iterations: 200,
      renderLabelsByDefault: true,
      largeGraph: false,
    });
  });

  it("disables default labels and reduces layout work for large graphs", () => {
    expect(getGraphPerformanceProfile(1500)).toEqual({
      forceAtlas2Iterations: 80,
      renderLabelsByDefault: false,
      largeGraph: true,
    });
  });

  it("uses the lightest layout pass for very large graphs", () => {
    expect(getGraphPerformanceProfile(3500)).toEqual({
      forceAtlas2Iterations: 40,
      renderLabelsByDefault: false,
      largeGraph: true,
    });
  });
});

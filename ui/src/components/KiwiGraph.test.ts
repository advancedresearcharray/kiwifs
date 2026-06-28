import { describe, expect, it } from "vitest";
import { getGraphPerformanceProfile } from "../lib/graphPerformance";

describe("graph performance profile", () => {
  it("keeps labels and full layout iterations for small graphs", () => {
    expect(getGraphPerformanceProfile(250)).toMatchObject({
      forceAtlas2Iterations: 200,
      renderLabelsByDefault: true,
      largeGraph: false,
      d3: {
        linkDistance: 60,
        collideIterations: 2,
        alphaDecay: 0.02,
      },
    });
  });

  it("disables default labels and reduces layout work for large graphs", () => {
    expect(getGraphPerformanceProfile(1500)).toMatchObject({
      forceAtlas2Iterations: 80,
      renderLabelsByDefault: false,
      largeGraph: true,
      d3: {
        linkDistance: 50,
        collideIterations: 1,
        alphaDecay: 0.045,
      },
    });
  });

  it("uses the lightest layout pass for very large graphs", () => {
    expect(getGraphPerformanceProfile(3500)).toMatchObject({
      forceAtlas2Iterations: 40,
      renderLabelsByDefault: false,
      largeGraph: true,
      d3: {
        linkDistance: 44,
        collideIterations: 1,
        alphaDecay: 0.075,
      },
    });
  });
});

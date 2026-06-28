export type GraphPerformanceProfile = {
  /** Historical Sigma/ForceAtlas2 consumers still use this field. */
  forceAtlas2Iterations: number;
  renderLabelsByDefault: boolean;
  largeGraph: boolean;
  d3: {
    linkDistance: number;
    linkStrength: number;
    chargeStrength: number;
    chargeDistanceMax: number;
    collideIterations: number;
    alphaDecay: number;
    velocityDecay: number;
  };
};

export function getGraphPerformanceProfile(order: number): GraphPerformanceProfile {
  if (order >= 3000) {
    return {
      forceAtlas2Iterations: 40,
      renderLabelsByDefault: false,
      largeGraph: true,
      d3: {
        linkDistance: 44,
        linkStrength: 0.22,
        chargeStrength: -55,
        chargeDistanceMax: 180,
        collideIterations: 1,
        alphaDecay: 0.075,
        velocityDecay: 0.55,
      },
    };
  }

  if (order >= 1000) {
    return {
      forceAtlas2Iterations: 80,
      renderLabelsByDefault: false,
      largeGraph: true,
      d3: {
        linkDistance: 50,
        linkStrength: 0.3,
        chargeStrength: -80,
        chargeDistanceMax: 240,
        collideIterations: 1,
        alphaDecay: 0.045,
        velocityDecay: 0.48,
      },
    };
  }

  return {
    forceAtlas2Iterations: 200,
    renderLabelsByDefault: true,
    largeGraph: false,
    d3: {
      linkDistance: 60,
      linkStrength: 0.4,
      chargeStrength: -120,
      chargeDistanceMax: 300,
      collideIterations: 2,
      alphaDecay: 0.02,
      velocityDecay: 0.4,
    },
  };
}

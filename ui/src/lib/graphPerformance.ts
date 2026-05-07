export type GraphPerformanceProfile = {
  forceAtlas2Iterations: number;
  renderLabelsByDefault: boolean;
  largeGraph: boolean;
};

export function getGraphPerformanceProfile(order: number): GraphPerformanceProfile {
  if (order >= 3000) {
    return {
      forceAtlas2Iterations: 40,
      renderLabelsByDefault: false,
      largeGraph: true,
    };
  }

  if (order >= 1000) {
    return {
      forceAtlas2Iterations: 80,
      renderLabelsByDefault: false,
      largeGraph: true,
    };
  }

  return {
    forceAtlas2Iterations: 200,
    renderLabelsByDefault: true,
    largeGraph: false,
  };
}

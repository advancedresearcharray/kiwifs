// Graph analytics helpers — thin wrappers around graphology-metrics and
// graphology-shortest-path that keep KiwiGraph.tsx lean and make the analysis
// logic independently testable.

import Graph from "graphology";
import { centrality } from "graphology-metrics";
import { bidirectional as _bidirectional } from "graphology-shortest-path";

// ── PageRank ────────────────────────────────────────────────────────────────

/** Compute PageRank scores for every node. Returns a Map<nodeKey, score>. */
export function computePageRank(graph: Graph): Map<string, number> {
  if (graph.order === 0) return new Map();

  // graphology-metrics pagerank returns { [node]: score }
  const scores = centrality.pagerank(graph, {
    alpha: 0.85,
    maxIterations: 100,
    tolerance: 1e-6,
    getEdgeWeight: null, // unweighted
  });

  return new Map(Object.entries(scores));
}

// ── Community Detection ─────────────────────────────────────────────────────

// Re-export the existing louvain integration. The caller (KiwiGraph) already
// uses `louvain.assign()` directly — this wrapper just provides a functional
// interface that returns a community map instead of mutating the graph.
import louvain from "graphology-communities-louvain";

/** Detect communities via Louvain. Returns Map<nodeKey, communityId>. */
export function detectCommunities(graph: Graph): Map<string, number> {
  if (graph.order === 0 || graph.size === 0) return new Map();
  const mapping = louvain(graph);
  return new Map(Object.entries(mapping).map(([k, v]) => [k, v as number]));
}

// ── Shortest Path ───────────────────────────────────────────────────────────

/**
 * Find the shortest (unweighted) path between two nodes.
 * Returns the ordered list of node keys, or an empty array if no path exists.
 */
export function findShortestPath(
  graph: Graph,
  from: string,
  to: string,
): string[] {
  if (!graph.hasNode(from) || !graph.hasNode(to)) return [];
  if (from === to) return [from];

  const result = _bidirectional(graph, from, to);
  return result ?? [];
}

// ── Community Palette ───────────────────────────────────────────────────────

const PALETTE = [
  "#5b9e4f", "#4a89c8", "#d97b3e", "#c254a5", "#3db5a6",
  "#c9534e", "#8b6cc1", "#c4a832", "#4eadd4", "#7a8f3e",
];

/**
 * Return a deterministic hex color for a community index.
 * First 10 communities get a hand-picked palette; beyond that we walk the
 * golden angle to generate perceptually distinct hues.
 */
export function communityPalette(communityId: number): string {
  if (communityId >= 0 && communityId < PALETTE.length) {
    return PALETTE[communityId]!;
  }

  // Golden-angle hue walk for overflow communities
  const hue = (communityId * 137.508) % 360;
  const saturation = 0.72;
  const lightness = 0.58;
  const chroma = (1 - Math.abs(2 * lightness - 1)) * saturation;
  const x = chroma * (1 - Math.abs(((hue / 60) % 2) - 1));
  const m = lightness - chroma / 2;
  let r = 0, g = 0, b = 0;

  if (hue < 60) [r, g, b] = [chroma, x, 0];
  else if (hue < 120) [r, g, b] = [x, chroma, 0];
  else if (hue < 180) [r, g, b] = [0, chroma, x];
  else if (hue < 240) [r, g, b] = [0, x, chroma];
  else if (hue < 300) [r, g, b] = [x, 0, chroma];
  else [r, g, b] = [chroma, 0, x];

  const toHex = (v: number) =>
    Math.round((v + m) * 255)
      .toString(16)
      .padStart(2, "0");
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
}

// ── Node Size from PageRank ─────────────────────────────────────────────────

/**
 * Map a PageRank score to a pixel size, linearly interpolated between
 * `minSize` and `maxSize`.
 */
export function pagerankToSize(
  score: number,
  maxScore: number,
  minSize = 4,
  maxSize = 24,
): number {
  if (maxScore <= 0) return minSize;
  const t = Math.min(score / maxScore, 1);
  return minSize + t * (maxSize - minSize);
}

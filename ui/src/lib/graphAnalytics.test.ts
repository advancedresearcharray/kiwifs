import { describe, it, expect } from "vitest";
import Graph from "graphology";
import {
  computePageRank,
  detectCommunities,
  findShortestPath,
  communityPalette,
  pagerankToSize,
} from "./graphAnalytics";

function makeTriangle(): Graph {
  const g = new Graph({ type: "undirected", multi: false });
  g.addNode("a", { label: "A" });
  g.addNode("b", { label: "B" });
  g.addNode("c", { label: "C" });
  g.addEdge("a", "b");
  g.addEdge("b", "c");
  g.addEdge("a", "c");
  return g;
}

function makeChain(): Graph {
  const g = new Graph({ type: "undirected", multi: false });
  g.addNode("1");
  g.addNode("2");
  g.addNode("3");
  g.addNode("4");
  g.addEdge("1", "2");
  g.addEdge("2", "3");
  g.addEdge("3", "4");
  return g;
}

function makeTwoClusters(): Graph {
  const g = new Graph({ type: "undirected", multi: false });
  // Cluster 1
  g.addNode("a1");
  g.addNode("a2");
  g.addNode("a3");
  g.addEdge("a1", "a2");
  g.addEdge("a2", "a3");
  g.addEdge("a1", "a3");
  // Cluster 2
  g.addNode("b1");
  g.addNode("b2");
  g.addNode("b3");
  g.addEdge("b1", "b2");
  g.addEdge("b2", "b3");
  g.addEdge("b1", "b3");
  // Weak bridge
  g.addEdge("a3", "b1");
  return g;
}

// ── PageRank ────────────────────────────────────────────────────────────────

describe("computePageRank", () => {
  it("returns an empty map for an empty graph", () => {
    const g = new Graph();
    expect(computePageRank(g).size).toBe(0);
  });

  it("assigns scores to every node", () => {
    const scores = computePageRank(makeTriangle());
    expect(scores.size).toBe(3);
    for (const [, score] of scores) {
      expect(score).toBeGreaterThan(0);
      expect(score).toBeLessThanOrEqual(1);
    }
  });

  it("gives equal scores to nodes with equal structure", () => {
    const scores = computePageRank(makeTriangle());
    const values = Array.from(scores.values());
    // All nodes in a triangle have identical topology
    expect(values[0]).toBeCloseTo(values[1]!, 4);
    expect(values[1]).toBeCloseTo(values[2]!, 4);
  });

  it("gives higher scores to better-connected nodes", () => {
    const g = new Graph({ type: "undirected", multi: false });
    g.addNode("hub");
    g.addNode("leaf1");
    g.addNode("leaf2");
    g.addNode("leaf3");
    g.addEdge("hub", "leaf1");
    g.addEdge("hub", "leaf2");
    g.addEdge("hub", "leaf3");

    const scores = computePageRank(g);
    const hubScore = scores.get("hub")!;
    const leafScore = scores.get("leaf1")!;
    expect(hubScore).toBeGreaterThan(leafScore);
  });
});

// ── Community Detection ─────────────────────────────────────────────────────

describe("detectCommunities", () => {
  it("returns empty for an empty graph", () => {
    expect(detectCommunities(new Graph()).size).toBe(0);
  });

  it("returns empty for a graph with no edges", () => {
    const g = new Graph();
    g.addNode("a");
    g.addNode("b");
    expect(detectCommunities(g).size).toBe(0);
  });

  it("detects at least two communities in two clusters", () => {
    const communities = detectCommunities(makeTwoClusters());
    const ids = new Set(communities.values());
    expect(ids.size).toBeGreaterThanOrEqual(2);
  });

  it("places connected clique members in the same community", () => {
    const communities = detectCommunities(makeTwoClusters());
    expect(communities.get("a1")).toBe(communities.get("a2"));
    expect(communities.get("a2")).toBe(communities.get("a3"));
    expect(communities.get("b1")).toBe(communities.get("b2"));
    expect(communities.get("b2")).toBe(communities.get("b3"));
  });
});

// ── Shortest Path ───────────────────────────────────────────────────────────

describe("findShortestPath", () => {
  it("returns empty for missing source", () => {
    const g = makeChain();
    expect(findShortestPath(g, "missing", "1")).toEqual([]);
  });

  it("returns empty for missing target", () => {
    const g = makeChain();
    expect(findShortestPath(g, "1", "missing")).toEqual([]);
  });

  it("returns single node for self-path", () => {
    const g = makeChain();
    expect(findShortestPath(g, "2", "2")).toEqual(["2"]);
  });

  it("finds direct neighbor path", () => {
    const g = makeChain();
    const path = findShortestPath(g, "1", "2");
    expect(path).toEqual(["1", "2"]);
  });

  it("finds multi-hop path", () => {
    const g = makeChain();
    const path = findShortestPath(g, "1", "4");
    expect(path).toEqual(["1", "2", "3", "4"]);
  });

  it("returns empty for disconnected nodes", () => {
    const g = new Graph({ type: "undirected", multi: false });
    g.addNode("x");
    g.addNode("y");
    // no edge
    expect(findShortestPath(g, "x", "y")).toEqual([]);
  });
});

// ── Palette ─────────────────────────────────────────────────────────────────

describe("communityPalette", () => {
  it("returns predefined colors for ids 0-9", () => {
    for (let i = 0; i < 10; i++) {
      const color = communityPalette(i);
      expect(color).toMatch(/^#[0-9a-f]{6}$/i);
    }
  });

  it("returns distinct colors for ids 0-9", () => {
    const colors = new Set(Array.from({ length: 10 }, (_, i) => communityPalette(i)));
    expect(colors.size).toBe(10);
  });

  it("generates valid hex for overflow ids", () => {
    for (const id of [10, 15, 50, 100]) {
      const color = communityPalette(id);
      expect(color).toMatch(/^#[0-9a-f]{6}$/i);
    }
  });

  it("generates distinct colors for overflow ids", () => {
    const colors = new Set(Array.from({ length: 20 }, (_, i) => communityPalette(i + 10)));
    expect(colors.size).toBe(20);
  });
});

// ── Size Mapping ────────────────────────────────────────────────────────────

describe("pagerankToSize", () => {
  it("returns minSize when score is 0", () => {
    expect(pagerankToSize(0, 1, 4, 24)).toBe(4);
  });

  it("returns maxSize when score equals maxScore", () => {
    expect(pagerankToSize(1, 1, 4, 24)).toBe(24);
  });

  it("returns minSize when maxScore is 0", () => {
    expect(pagerankToSize(0.5, 0, 4, 24)).toBe(4);
  });

  it("linearly interpolates", () => {
    expect(pagerankToSize(0.5, 1, 0, 10)).toBeCloseTo(5, 4);
  });

  it("clamps at maxSize for oversized scores", () => {
    expect(pagerankToSize(2, 1, 4, 24)).toBe(24);
  });
});

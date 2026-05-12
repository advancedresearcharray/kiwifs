// Knowledge graph view: nodes are markdown pages, edges are [[wiki-link]]
// references between them. Powered by Sigma.js (WebGL) + Graphology with a
// ForceAtlas2 layout; Louvain community detection drives the per-cluster
// palette.
//
// Enhanced with:
//   - PageRank-based node sizing (toggle)
//   - Community detection coloring (toggle)
//   - Shortest-path finding between two selected nodes
//   - Layout switcher (ForceAtlas2 / Circular / Random)
//   - Rich hover tooltip (title, PageRank, community, degree)

import { useEffect, useMemo, useRef, useState } from "react";
import Graph from "graphology";
import circular from "graphology-layout/circular";
import random from "graphology-layout/random";
import forceAtlas2 from "graphology-layout-forceatlas2";
import louvain from "graphology-communities-louvain";
import {
  SigmaContainer,
  useLoadGraph,
  useRegisterEvents,
  useSetSettings,
  useSigma,
} from "@react-sigma/core";
import {
  ArrowLeft,
  Loader2,
  Route,
  Search as SearchIcon,
  Tag,
} from "lucide-react";
import "@react-sigma/core/lib/style.css";
import { api, type GraphResponse, type TreeEntry } from "@kw/lib/api";
import { buildResolver } from "@kw/lib/wikiLinks";
import { titleize } from "@kw/lib/paths";
import { cn } from "@kw/lib/cn";
import {
  colorForGraphCommunity,
  readKiwiGraphTheme,
  type KiwiGraphTheme,
} from "@kw/lib/kiwiGraphTheme";
import {
  computePageRank,
  findShortestPath,
  pagerankToSize,
} from "@kw/lib/graphAnalytics";
import { Button } from "@kw/components/ui/button";
import { Card } from "@kw/components/ui/card";
import { Input } from "@kw/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@kw/components/ui/select";

type Props = {
  tree: TreeEntry | null;
  activePath?: string | null;
  onNavigate: (path: string) => void;
  onClose: () => void;
};

type LayoutKind = "forceatlas2" | "circular" | "random";

function topDir(path: string): string {
  const i = path.indexOf("/");
  return i < 0 ? "(root)" : path.slice(0, i);
}

type Built = {
  graph: Graph;
  dirs: string[];
  tags: string[];
  theme: KiwiGraphTheme;
  pagerank: Map<string, number>;
  communities: Map<string, number>;
};

// Build the Graphology instance from the server response plus the file tree.
function buildGraph(
  resp: GraphResponse,
  tree: TreeEntry | null,
  theme: KiwiGraphTheme,
  layout: LayoutKind,
  sizeByPageRank: boolean,
  colorByCommunity: boolean,
): Built {
  const g = new Graph({ type: "undirected", multi: false });
  const resolver = buildResolver(tree);

  const tagSet = new Set<string>();
  for (const n of resp.nodes) {
    g.addNode(n.path, {
      label: titleize(n.path),
      path: n.path,
      dir: topDir(n.path),
      tags: n.tags || [],
      size: 4,
      color: theme.defaultNode,
    });
    if (n.tags) n.tags.forEach((t) => tagSet.add(t));
  }

  const dirSet = new Set<string>();
  for (const n of resp.nodes) dirSet.add(topDir(n.path));

  for (const e of resp.edges) {
    if (!g.hasNode(e.source)) continue;
    const resolved = resolver(e.target);
    if (!resolved || !g.hasNode(resolved)) continue;
    if (resolved === e.source) continue; // skip self-loops
    const edgeKey = g.hasEdge(e.source, resolved)
      ? g.edge(e.source, resolved)
      : g.addEdge(e.source, resolved, { size: 0.8, color: theme.edge });
    void edgeKey;
  }

  // PageRank computation
  const prScores = computePageRank(g);
  const maxPR = Math.max(...Array.from(prScores.values()), 0);

  // Node sizing — PageRank or degree-based
  g.forEachNode((node) => {
    if (sizeByPageRank && maxPR > 0) {
      const score = prScores.get(node) ?? 0;
      g.setNodeAttribute(node, "size", pagerankToSize(score, maxPR, 4, 24));
    } else {
      const deg = g.degree(node);
      g.setNodeAttribute(node, "size", Math.max(6, Math.min(22, 6 + Math.sqrt(deg) * 2.5)));
    }
  });

  // Community detection + coloring
  const communityMap = new Map<string, number>();
  if (g.size > 0) {
    louvain.assign(g, { nodeCommunityAttribute: "community" });
    g.forEachNode((node) => {
      const c = g.getNodeAttribute(node, "community") as number | undefined;
      const idx = typeof c === "number" ? c : 0;
      communityMap.set(node, idx);
      if (colorByCommunity) {
        g.setNodeAttribute(node, "color", colorForGraphCommunity(idx, theme));
      }
    });
  }

  // Layout
  switch (layout) {
    case "circular":
      circular.assign(g, { scale: 100 });
      break;
    case "random":
      random.assign(g, { scale: 200 });
      break;
    case "forceatlas2":
    default:
      circular.assign(g, { scale: 100 });
      forceAtlas2.assign(g, {
        iterations: 200,
        settings: {
          gravity: 1,
          scalingRatio: 10,
          slowDown: 2,
          barnesHutOptimize: g.order > 200,
          strongGravityMode: false,
        },
      });
      break;
  }

  return {
    graph: g,
    dirs: Array.from(dirSet).sort(),
    tags: Array.from(tagSet).sort(),
    theme,
    pagerank: prScores,
    communities: communityMap,
  };
}

export function KiwiGraph({ tree, activePath, onNavigate, onClose }: Props) {
  const [resp, setResp] = useState<GraphResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [dirFilter, setDirFilter] = useState<string>("");
  const [tagFilter, setTagFilter] = useState<string>("");
  const [query, setQuery] = useState<string>("");
  const [hovered, setHovered] = useState<string | null>(null);
  const [htmlClassEpoch, setHtmlClassEpoch] = useState(0);

  // New analytics controls
  const [sizeByPageRank, setSizeByPageRank] = useState(true);
  const [colorByCommunity, setColorByCommunity] = useState(true);
  const [layout, setLayout] = useState<LayoutKind>("forceatlas2");

  // Path finding state
  const [pathFindActive, setPathFindActive] = useState(false);
  const [pathSource, setPathSource] = useState<string | null>(null);
  const [pathTarget, setPathTarget] = useState<string | null>(null);
  const [foundPath, setFoundPath] = useState<string[] | null>(null);

  useEffect(() => {
    let cancelled = false;
    setError(null);
    api
      .graph()
      .then((r) => {
        if (!cancelled) setResp(r);
      })
      .catch((e) => {
        if (!cancelled) setError(String(e));
      });
    return () => {
      cancelled = true;
    };
  }, [tree]);

  useEffect(() => {
    const obs = new MutationObserver(() =>
      setHtmlClassEpoch((n: number) => n + 1),
    );
    obs.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    });
    return () => obs.disconnect();
  }, []);

  const built = useMemo<Built | null>(() => {
    if (!resp) return null;
    return buildGraph(resp, tree, readKiwiGraphTheme(), layout, sizeByPageRank, colorByCommunity);
  }, [resp, tree, htmlClassEpoch, layout, sizeByPageRank, colorByCommunity]);

  // Compute path when source and target are both set
  useEffect(() => {
    if (!built || !pathSource || !pathTarget) {
      setFoundPath(null);
      return;
    }
    const path = findShortestPath(built.graph, pathSource, pathTarget);
    setFoundPath(path.length > 0 ? path : null);
  }, [built, pathSource, pathTarget]);

  // Clear path finding state when deactivated
  useEffect(() => {
    if (!pathFindActive) {
      setPathSource(null);
      setPathTarget(null);
      setFoundPath(null);
    }
  }, [pathFindActive]);

  function handlePathNodeClick(node: string) {
    if (!pathFindActive) return false;
    if (!pathSource) {
      setPathSource(node);
      return true;
    }
    if (!pathTarget) {
      setPathTarget(node);
      return true;
    }
    // Both already set — reset and start over
    setPathSource(node);
    setPathTarget(null);
    setFoundPath(null);
    return true;
  }

  return (
    <div className="h-full w-full flex flex-col relative">
      {/* ── Toolbar ── */}
      <div className="flex flex-wrap items-center gap-2 sm:gap-3 px-3 sm:px-6 py-3 border-b border-border bg-card">
        <Button variant="outline" size="sm" onClick={onClose}>
          <ArrowLeft className="h-3.5 w-3.5" /> <span className="hidden sm:inline">Back</span>
        </Button>
        <div className="font-semibold text-sm">Knowledge graph</div>
        <div className="text-xs text-muted-foreground hidden sm:block">
          {built
            ? `${built.graph.order} pages · ${built.graph.size} links`
            : null}
        </div>
        <div className="ml-auto flex items-center gap-2 flex-wrap">
          {/* Search / highlight */}
          <div className="relative">
            <SearchIcon className="h-3.5 w-3.5 absolute left-2 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none" />
            <Input
              type="text"
              placeholder="Highlight..."
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="h-8 pl-7 w-32 sm:w-48 text-sm"
            />
          </div>

          {/* Directory filter */}
          {built && built.dirs.length > 1 && (
            <Select
              value={dirFilter || "__all__"}
              onValueChange={(v) => setDirFilter(v === "__all__" ? "" : v)}
            >
              <SelectTrigger className="h-8 w-32 sm:w-44 text-sm">
                <SelectValue placeholder="All folders" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__all__">All folders</SelectItem>
                {built.dirs.map((d) => (
                  <SelectItem key={d} value={d}>
                    {d}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}

          {/* Tag filter */}
          {built && built.tags.length > 0 && (
            <Select
              value={tagFilter || "__all__"}
              onValueChange={(v) => setTagFilter(v === "__all__" ? "" : v)}
            >
              <SelectTrigger className="h-8 w-32 sm:w-44 text-sm">
                <Tag className="h-3 w-3 mr-1" />
                <SelectValue placeholder="All tags" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__all__">All tags</SelectItem>
                {built.tags.map((t) => (
                  <SelectItem key={t} value={t}>
                    {t}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}

          {/* Layout selector */}
          <Select
            value={layout}
            onValueChange={(v) => setLayout(v as LayoutKind)}
          >
            <SelectTrigger className="h-8 w-32 text-sm">
              <SelectValue placeholder="Layout" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="forceatlas2">ForceAtlas2</SelectItem>
              <SelectItem value="circular">Circular</SelectItem>
              <SelectItem value="random">Random</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* ── Analytics toggles ── */}
      <div className="flex flex-wrap items-center gap-2 px-3 sm:px-6 py-1.5 border-b border-border/50 bg-card/50 text-xs">
        <label className="flex items-center gap-1.5 cursor-pointer select-none">
          <input
            type="checkbox"
            checked={sizeByPageRank}
            onChange={(e) => setSizeByPageRank(e.target.checked)}
            className="accent-primary h-3 w-3"
          />
          Size by PageRank
        </label>
        <label className="flex items-center gap-1.5 cursor-pointer select-none">
          <input
            type="checkbox"
            checked={colorByCommunity}
            onChange={(e) => setColorByCommunity(e.target.checked)}
            className="accent-primary h-3 w-3"
          />
          Color by community
        </label>
        <Button
          variant={pathFindActive ? "secondary" : "outline"}
          size="sm"
          className="h-6 px-2 text-xs gap-1"
          onClick={() => setPathFindActive((v) => !v)}
        >
          <Route className="h-3 w-3" />
          Find path
        </Button>
        <Button
          variant="outline"
          size="sm"
          className="h-6 px-2 text-xs opacity-50 cursor-not-allowed"
          disabled
        >
          Show semantic edges
        </Button>
        {pathFindActive && (
          <span className="text-muted-foreground">
            {!pathSource
              ? "Click source node..."
              : !pathTarget
                ? `Source: ${titleize(pathSource)} — click target...`
                : foundPath
                  ? `Path: ${foundPath.length - 1} hop${foundPath.length - 1 !== 1 ? "s" : ""}`
                  : "No path found"}
          </span>
        )}
      </div>

      {/* ── Graph canvas ── */}
      <div className="flex-1 relative">
        {error && (
          <div className="absolute inset-0 grid place-items-center text-sm text-destructive font-mono">
            {error}
          </div>
        )}
        {!error && !built && (
          <div className="absolute inset-0 grid place-items-center text-sm text-muted-foreground">
            <div className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" /> Building graph...
            </div>
          </div>
        )}
        {built && built.graph.order === 0 && (
          <div className="absolute inset-0 grid place-items-center text-sm text-muted-foreground">
            No pages yet.
          </div>
        )}
        {built && built.graph.order > 0 && (
          <SigmaContainer
            key={`${resp ? resp.nodes.length : 0}-${htmlClassEpoch}-${layout}`}
            graph={built.graph as any}
            className="!bg-background"
            style={{ height: "100%", width: "100%" }}
            settings={{
              renderLabels: true,
              labelColor: { attribute: "color" },
              labelSize: 12,
              labelWeight: "500",
              labelDensity: 0.7,
              labelGridCellSize: 80,
              defaultEdgeColor: built.theme.edge,
              zIndex: true,
            }}
          >
            <GraphInteractions
              onNavigate={onNavigate}
              hovered={hovered}
              setHovered={setHovered}
              query={query.trim().toLowerCase()}
              dirFilter={dirFilter}
              tagFilter={tagFilter}
              activePath={activePath || undefined}
              colors={built.theme}
              foundPath={foundPath}
              pathFindActive={pathFindActive}
              onPathNodeClick={handlePathNodeClick}
            />
          </SigmaContainer>
        )}

        {/* ── Hover tooltip (enhanced) ── */}
        {hovered && built && built.graph.hasNode(hovered) && (
          <Card className="absolute bottom-3 left-3 px-3 py-2 text-xs pointer-events-none max-w-xs">
            <div className="font-medium">
              {built.graph.getNodeAttribute(hovered, "label") as string}
            </div>
            <div className="text-muted-foreground font-mono text-[10px]">{hovered}</div>
            <div className="flex flex-wrap gap-x-3 gap-y-0.5 text-muted-foreground mt-1">
              <span>
                {built.graph.degree(hovered)} connection
                {built.graph.degree(hovered) === 1 ? "" : "s"}
              </span>
              {built.pagerank.has(hovered) && (
                <span>PR: {(built.pagerank.get(hovered)! * 100).toFixed(2)}%</span>
              )}
              {built.communities.has(hovered) && (
                <span>Community: {built.communities.get(hovered)}</span>
              )}
            </div>
            {(() => {
              const tags = (built.graph.getNodeAttribute(hovered, "tags") as string[]) || [];
              if (tags.length === 0) return null;
              return (
                <div className="flex flex-wrap gap-1 mt-1">
                  {tags.map((t) => (
                    <span key={t} className="bg-muted px-1 rounded text-[10px]">{t}</span>
                  ))}
                </div>
              );
            })()}
          </Card>
        )}

        {/* ── Path length overlay ── */}
        {foundPath && foundPath.length > 1 && (
          <Card className="absolute bottom-3 left-1/2 -translate-x-1/2 px-3 py-1.5 text-xs pointer-events-none bg-primary text-primary-foreground">
            Shortest path: {foundPath.length - 1} hop{foundPath.length - 1 !== 1 ? "s" : ""} ({foundPath.map(n => titleize(n)).join(" -> ")})
          </Card>
        )}

        {built && built.graph.order > 0 && (
          <GraphLegend graph={built.graph} theme={built.theme} />
        )}
        <div
          className={cn(
            "absolute bottom-3 right-3 text-[10px] text-muted-foreground font-mono",
            "pointer-events-none",
          )}
        >
          drag to pan · scroll to zoom
        </div>
      </div>
    </div>
  );
}

// GraphInteractions runs inside <SigmaContainer> so it can use the sigma
// hooks. It owns: click->navigate, hover highlighting, query-based dimming,
// the directory filter, and path highlighting.
function GraphInteractions({
  onNavigate,
  hovered,
  setHovered,
  query,
  dirFilter,
  tagFilter,
  activePath,
  colors,
  foundPath,
  pathFindActive,
  onPathNodeClick,
}: {
  onNavigate: (path: string) => void;
  hovered: string | null;
  setHovered: (s: string | null) => void;
  query: string;
  dirFilter: string;
  tagFilter: string;
  activePath?: string;
  colors: KiwiGraphTheme;
  foundPath: string[] | null;
  pathFindActive: boolean;
  onPathNodeClick: (node: string) => boolean;
}) {
  const sigma = useSigma();
  const registerEvents = useRegisterEvents();
  const setSettings = useSetSettings();
  const loadGraph = useLoadGraph();
  const loadedRef = useRef(false);

  useEffect(() => {
    if (loadedRef.current) return;
    loadedRef.current = true;
    const g = sigma.getGraph();
    if (g) loadGraph(g as any);
  }, [loadGraph, sigma]);

  useEffect(() => {
    registerEvents({
      clickNode: (e) => {
        const node = e.node;
        if (pathFindActive) {
          onPathNodeClick(node);
          return;
        }
        onNavigate(node);
      },
      enterNode: (e) => setHovered(e.node),
      leaveNode: () => setHovered(null),
    });
  }, [registerEvents, onNavigate, setHovered, pathFindActive, onPathNodeClick]);

  useEffect(() => {
    const graph = sigma.getGraph();
    const neighbors = new Set<string>();
    if (hovered && graph.hasNode(hovered)) {
      neighbors.add(hovered);
      graph.forEachNeighbor(hovered, (n: string) => neighbors.add(n));
    }

    const pathSet = foundPath ? new Set(foundPath) : null;
    // Build set of edges on the found path
    const pathEdges = new Set<string>();
    if (foundPath && foundPath.length > 1) {
      for (let i = 0; i < foundPath.length - 1; i++) {
        const a = foundPath[i]!;
        const b = foundPath[i + 1]!;
        if (graph.hasEdge(a, b)) pathEdges.add(graph.edge(a, b)!);
        if (graph.hasEdge(b, a)) pathEdges.add(graph.edge(b, a)!);
      }
    }

    setSettings({
      nodeReducer: (node, data) => {
        const out: any = { ...data };
        const path = (data as any).path as string;
        const dir = (data as any).dir as string;
        const label = ((data as any).label as string) || "";
        const tags = ((data as any).tags as string[]) || [];
        const tagOut = tagFilter && !tags.includes(tagFilter);
        const filteredOut = (dirFilter && dir !== dirFilter) || tagOut;
        const queryMatch = query
          ? path.toLowerCase().includes(query) || label.toLowerCase().includes(query)
          : true;
        if (filteredOut) {
          out.hidden = true;
          return out;
        }
        if (activePath && node === activePath) {
          out.size = Math.max((out.size as number) || 6, 10);
          out.zIndex = 3;
          out.forceLabel = true;
          out.borderColor = "#ffffff";
          out.borderSize = 2;
        }

        // Path highlighting takes precedence over hover/search
        if (pathSet) {
          if (pathSet.has(node)) {
            out.zIndex = 3;
            out.forceLabel = true;
            out.size = Math.max((out.size as number) || 6, 12);
          } else {
            out.color = colors.nodeDim;
            out.label = "";
            out.zIndex = 0;
          }
          return out;
        }

        if (hovered) {
          if (!neighbors.has(node)) {
            if (node !== activePath) {
              out.color = colors.nodeDim;
              out.label = "";
              out.zIndex = 0;
            }
          } else {
            out.zIndex = 2;
            out.forceLabel = true;
          }
        } else if (query) {
          if (!queryMatch) {
            out.color = colors.nodeDim;
            out.label = "";
          } else {
            out.forceLabel = true;
            out.zIndex = 2;
          }
        }
        return out;
      },
      edgeReducer: (edge, data) => {
        const out: any = { ...data };
        const g = sigma.getGraph();
        const [s, t] = g.extremities(edge);
        if (dirFilter) {
          const sDir = g.getNodeAttribute(s, "dir") as string;
          const tDir = g.getNodeAttribute(t, "dir") as string;
          if (sDir !== dirFilter && tDir !== dirFilter) {
            out.hidden = true;
            return out;
          }
        }
        if (tagFilter) {
          const sTags = (g.getNodeAttribute(s, "tags") as string[]) || [];
          const tTags = (g.getNodeAttribute(t, "tags") as string[]) || [];
          if (!sTags.includes(tagFilter) && !tTags.includes(tagFilter)) {
            out.hidden = true;
            return out;
          }
        }

        // Path highlighting
        if (pathSet) {
          if (pathEdges.has(edge)) {
            out.color = "#f59e0b"; // amber highlight for path edges
            out.size = 3;
            out.zIndex = 2;
          } else {
            out.color = colors.edgeGhost;
            out.size = 0.2;
          }
          return out;
        }

        if (hovered) {
          if (s !== hovered && t !== hovered) {
            out.color = colors.edgeGhost;
            out.size = 0.3;
          } else {
            out.color = colors.edgeStrong;
            out.size = 1.5;
            out.zIndex = 1;
          }
        }
        return out;
      },
    });

    sigma.refresh();
  }, [sigma, setSettings, hovered, query, dirFilter, tagFilter, activePath, colors, foundPath, pathFindActive]);

  return null;
}

function GraphLegend({ graph, theme }: { graph: Graph; theme: KiwiGraphTheme }) {
  const communities = new Map<number, { color: string; count: number; dirs: Map<string, number> }>();
  graph.forEachNode((_, attrs) => {
    const c = (attrs as any).community as number | undefined;
    if (c == null) return;
    const dir = (attrs as any).dir as string || "(root)";
    const existing = communities.get(c);
    if (existing) {
      existing.count++;
      existing.dirs.set(dir, (existing.dirs.get(dir) || 0) + 1);
    } else {
      const dirs = new Map<string, number>();
      dirs.set(dir, 1);
      communities.set(c, {
        color: colorForGraphCommunity(c, theme),
        count: 1,
        dirs,
      });
    }
  });
  if (communities.size <= 1) return null;

  const sorted = Array.from(communities.entries()).sort((a, b) => b[1].count - a[1].count);

  return (
    <Card className="absolute top-3 right-3 px-3 py-2 text-xs">
      <div className="text-muted-foreground mb-1.5 font-medium">Communities</div>
      <div className="space-y-1">
        {sorted.map(([idx, { color, count, dirs }]) => {
          const topDirEntry = Array.from(dirs.entries()).sort((a, b) => b[1] - a[1])[0]?.[0] || "unknown";
          return (
            <div key={idx} className="flex items-center gap-2">
              <span
                className="h-2.5 w-2.5 rounded-full shrink-0"
                style={{ background: color }}
              />
              <span className="text-muted-foreground">
                {topDirEntry} ({count})
              </span>
            </div>
          );
        })}
      </div>
    </Card>
  );
}

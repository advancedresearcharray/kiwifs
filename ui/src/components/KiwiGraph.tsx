// Knowledge graph — PixiJS (WebGL) + d3-force, inspired by Obsidian's graph.
// Nodes = markdown pages, edges = [[wiki-link]] references.
// Features: community coloring, PageRank sizing, hover glow, search highlight,
// directory/tag filtering, shortest-path finder, zoom-adaptive labels.

import { useCallback, useEffect, useRef, useState } from "react";
import {
  ArrowLeft,
  Loader2,
  Route,
  Search as SearchIcon,
  Tag,
} from "lucide-react";
import {
  forceSimulation,
  forceLink,
  forceManyBody,
  forceCenter,
  forceCollide,
  type Simulation,
  type SimulationNodeDatum,
  type SimulationLinkDatum,
} from "d3-force";
import { api, type GraphResponse, type TreeEntry } from "@kw/lib/api";
import { buildResolver } from "@kw/lib/wikiLinks";
import { titleize } from "@kw/lib/paths";
import { cn } from "@kw/lib/cn";
import {
  readKiwiGraphTheme,
  type KiwiGraphTheme,
} from "@kw/lib/kiwiGraphTheme";
import {
  communityPalette,
  computePageRank,
  pagerankToSize,
} from "@kw/lib/graphAnalytics";
import Graph from "graphology";
import louvain from "graphology-communities-louvain";
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

// ── Internal node/link types for the simulation ──────────────────────────────

interface GNode extends SimulationNodeDatum {
  id: string;
  label: string;
  dir: string;
  tags: string[];
  community: number;
  pagerank: number;
  radius: number;
  color: string;
}

interface GLink extends SimulationLinkDatum<GNode> {
  source: string | GNode;
  target: string | GNode;
}

function topDir(path: string): string {
  const i = path.indexOf("/");
  return i < 0 ? "(root)" : path.slice(0, i);
}

// ── Hex color helpers ────────────────────────────────────────────────────────

function hexToRgb(hex: string): [number, number, number] {
  const h = hex.replace("#", "");
  return [
    parseInt(h.slice(0, 2), 16),
    parseInt(h.slice(2, 4), 16),
    parseInt(h.slice(4, 6), 16),
  ];
}

// ── Build graph data from API response ───────────────────────────────────────

type BuiltGraph = {
  nodes: GNode[];
  links: GLink[];
  dirs: string[];
  tags: string[];
  communityMap: Map<number, { color: string; count: number; topDir: string }>;
};

function buildGraphData(
  resp: GraphResponse,
  tree: TreeEntry | null,
  theme: KiwiGraphTheme,
  sizeByPageRank: boolean,
  colorByCommunity: boolean,
): BuiltGraph {
  const g = new Graph({ type: "undirected", multi: false });
  const resolver = buildResolver(tree);
  const tagSet = new Set<string>();
  const dirSet = new Set<string>();

  for (const n of resp.nodes) {
    g.addNode(n.path, {});
    dirSet.add(topDir(n.path));
    if (n.tags) n.tags.forEach((t) => tagSet.add(t));
  }

  for (const e of resp.edges) {
    if (!g.hasNode(e.source)) continue;
    const resolved = resolver(e.target);
    if (!resolved || !g.hasNode(resolved) || resolved === e.source) continue;
    if (!g.hasEdge(e.source, resolved)) g.addEdge(e.source, resolved);
  }

  const prScores = computePageRank(g);
  const maxPR = Math.max(...Array.from(prScores.values()), 0);

  let communities = new Map<string, number>();
  if (g.size > 0) {
    try {
      const raw = louvain(g);
      communities = new Map(
        Object.entries(raw).map(([k, v]) => [k, v as number]),
      );
    } catch {}
  }

  const communityDirCounts = new Map<
    number,
    { color: string; count: number; dirs: Map<string, number> }
  >();

  const nodes: GNode[] = resp.nodes.map((n) => {
    const pr = prScores.get(n.path) ?? 0;
    const comm = communities.get(n.path) ?? 0;
    const color = communityPalette(comm);
    const dir = topDir(n.path);

    if (!communityDirCounts.has(comm)) {
      communityDirCounts.set(comm, {
        color,
        count: 0,
        dirs: new Map(),
      });
    }
    const cd = communityDirCounts.get(comm)!;
    cd.count++;
    cd.dirs.set(dir, (cd.dirs.get(dir) || 0) + 1);

    const radius = sizeByPageRank && maxPR > 0
      ? pagerankToSize(pr, maxPR, 4, 18)
      : Math.max(4, Math.min(14, 4 + Math.sqrt(g.hasNode(n.path) ? g.degree(n.path) : 0) * 2));

    const nodeColor = colorByCommunity ? color : (theme.defaultNode || "#7c8a6e");

    return {
      id: n.path,
      label: titleize(n.path),
      dir,
      tags: n.tags || [],
      community: comm,
      pagerank: pr,
      radius,
      color: nodeColor,
    };
  });

  const nodeIds = new Set(nodes.map((n) => n.id));
  const links: GLink[] = [];
  const seen = new Set<string>();
  for (const e of resp.edges) {
    if (!nodeIds.has(e.source)) continue;
    const resolved = resolver(e.target);
    if (!resolved || !nodeIds.has(resolved) || resolved === e.source) continue;
    const key = [e.source, resolved].sort().join("||");
    if (seen.has(key)) continue;
    seen.add(key);
    links.push({ source: e.source, target: resolved });
  }

  const communityMap = new Map<
    number,
    { color: string; count: number; topDir: string }
  >();
  for (const [idx, { color, count, dirs }] of communityDirCounts) {
    const topDirEntry =
      Array.from(dirs.entries()).sort((a, b) => b[1] - a[1])[0]?.[0] ??
      "unknown";
    communityMap.set(idx, { color, count, topDir: topDirEntry });
  }

  return {
    nodes,
    links,
    dirs: Array.from(dirSet).sort(),
    tags: Array.from(tagSet).sort(),
    communityMap,
  };
}

// ── Shortest path (BFS on adjacency) ─────────────────────────────────────────

function bfsPath(
  adj: Map<string, Set<string>>,
  from: string,
  to: string,
): string[] {
  if (from === to) return [from];
  const visited = new Set<string>([from]);
  const queue: [string, string[]][] = [[from, [from]]];
  while (queue.length > 0) {
    const [node, path] = queue.shift()!;
    for (const neighbor of adj.get(node) || []) {
      if (visited.has(neighbor)) continue;
      visited.add(neighbor);
      const newPath = [...path, neighbor];
      if (neighbor === to) return newPath;
      queue.push([neighbor, newPath]);
    }
  }
  return [];
}

// ═════════════════════════════════════════════════════════════════════════════
// Main component
// ═════════════════════════════════════════════════════════════════════════════

export function KiwiGraph({ tree, activePath, onNavigate, onClose }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const simRef = useRef<Simulation<GNode, GLink> | null>(null);
  const rafRef = useRef<number>(0);
  const dataRef = useRef<BuiltGraph | null>(null);
  const adjRef = useRef<Map<string, Set<string>>>(new Map());

  const [resp, setResp] = useState<GraphResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [dirFilter, setDirFilter] = useState("");
  const [tagFilter, setTagFilter] = useState("");
  const [query, setQuery] = useState("");
  const [hovered, setHovered] = useState<GNode | null>(null);

  const [sizeByPageRank, setSizeByPageRank] = useState(true);
  const [colorByCommunity, setColorByCommunity] = useState(true);

  const [pathFindActive, setPathFindActive] = useState(false);
  const [pathSource, setPathSource] = useState<string | null>(null);
  const [pathTarget, setPathTarget] = useState<string | null>(null);
  const [foundPath, setFoundPath] = useState<string[] | null>(null);

  // Transform & viewport state (stored in refs for perf — no re-renders)
  const transformRef = useRef({ x: 0, y: 0, scale: 1 });
  const draggingRef = useRef<{
    node: GNode | null;
    pan: boolean;
    startX: number;
    startY: number;
    startTx: number;
    startTy: number;
  } | null>(null);
  const mouseRef = useRef({ x: 0, y: 0 });

  // Fetch graph data
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

  // Read theme once
  const themeRef = useRef<KiwiGraphTheme>(readKiwiGraphTheme());
  useEffect(() => {
    const obs = new MutationObserver(() => {
      themeRef.current = readKiwiGraphTheme();
    });
    obs.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    });
    return () => obs.disconnect();
  }, []);

  // ── Core: build graph + simulation + canvas render loop ──────────────────

  useEffect(() => {
    if (!resp || !containerRef.current) return;

    const data = buildGraphData(resp, tree, themeRef.current, sizeByPageRank, colorByCommunity);
    dataRef.current = data;

    if (data.nodes.length === 0) return;

    // Build adjacency for path-finding
    const adj = new Map<string, Set<string>>();
    for (const n of data.nodes) adj.set(n.id, new Set());
    for (const l of data.links) {
      const s = typeof l.source === "string" ? l.source : l.source.id;
      const t = typeof l.target === "string" ? l.target : l.target.id;
      adj.get(s)?.add(t);
      adj.get(t)?.add(s);
    }
    adjRef.current = adj;

    // Canvas setup
    const container = containerRef.current;
    let canvas = canvasRef.current;
    if (!canvas) {
      canvas = document.createElement("canvas");
      canvas.style.cssText =
        "position:absolute;inset:0;width:100%;height:100%;display:block;";
      container.appendChild(canvas);
      canvasRef.current = canvas;
    }
    const ctx = canvas.getContext("2d")!;

    function resize() {
      const dpr = window.devicePixelRatio || 1;
      const rect = container.getBoundingClientRect();
      canvas!.width = rect.width * dpr;
      canvas!.height = rect.height * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    }
    resize();
    const ro = new ResizeObserver(resize);
    ro.observe(container);

    // Center transform
    const rect = container.getBoundingClientRect();
    transformRef.current = { x: rect.width / 2, y: rect.height / 2, scale: 1 };

    // d3-force simulation
    const sim = forceSimulation<GNode>(data.nodes)
      .force(
        "link",
        forceLink<GNode, GLink>(data.links)
          .id((d) => d.id)
          .distance(60)
          .strength(0.4),
      )
      .force("charge", forceManyBody().strength(-120).distanceMax(300))
      .force("center", forceCenter(0, 0))
      .force(
        "collide",
        forceCollide<GNode>()
          .radius((d) => d.radius + 2)
          .iterations(2),
      )
      .alphaDecay(0.02)
      .velocityDecay(0.4);

    simRef.current = sim;

    // ── Render loop (Canvas 2D — no WebGL) ────────────────────────────────

    function render() {
      const w = canvas!.clientWidth;
      const h = canvas!.clientHeight;
      const { x: tx, y: ty, scale } = transformRef.current;

      ctx.clearRect(0, 0, w, h);
      ctx.save();
      ctx.translate(tx, ty);
      ctx.scale(scale, scale);

      const isDark = document.documentElement.classList.contains("dark");
      const qLower = query.trim().toLowerCase();
      const pathSet = foundPath ? new Set(foundPath) : null;

      const hoveredNeighbors = new Set<string>();
      if (hovered) {
        hoveredNeighbors.add(hovered.id);
        for (const n of adjRef.current.get(hovered.id) || [])
          hoveredNeighbors.add(n);
      }

      // ── Draw edges ───────────────────────────────────────────────────────

      for (const link of data.links) {
        const s = link.source as GNode;
        const t = link.target as GNode;
        const sx = s.x, sy = s.y, tx = t.x, ty = t.y;
        if (sx == null || sy == null || tx == null || ty == null) continue;

        // Filtering
        if (dirFilter && s.dir !== dirFilter && t.dir !== dirFilter) continue;
        if (tagFilter) {
          if (!s.tags.includes(tagFilter) && !t.tags.includes(tagFilter))
            continue;
        }

        let alpha = 0.15;
        let width = 0.5;
        let color = isDark ? "rgba(255,255,255," : "rgba(100,100,100,";

        if (pathSet) {
          const onPath = pathSet.has(s.id) && pathSet.has(t.id);
          if (onPath) {
            alpha = 0.9;
            width = 2.5;
            color = "rgba(245,158,11,";
          } else {
            alpha = 0.04;
          }
        } else if (hovered) {
          const connected =
            (s.id === hovered.id || t.id === hovered.id);
          alpha = connected ? 0.6 : 0.04;
          width = connected ? 1.5 : 0.3;
        }

        ctx.beginPath();
        ctx.moveTo(sx, sy);
        ctx.lineTo(tx, ty);
        ctx.strokeStyle = `${color}${alpha})`;
        ctx.lineWidth = width / scale;
        ctx.stroke();
      }

      // ── Draw nodes ───────────────────────────────────────────────────────

      for (const node of data.nodes) {
        const nx = node.x, ny = node.y;
        if (nx == null || ny == null) continue;

        // Filtering
        if (dirFilter && node.dir !== dirFilter) continue;
        if (tagFilter && !node.tags.includes(tagFilter)) continue;

        const isActive = activePath === node.id;
        const isHovered = hovered?.id === node.id;
        const isNeighbor = hoveredNeighbors.has(node.id);
        const onPath = pathSet?.has(node.id) ?? false;
        const queryMatch = qLower
          ? node.id.toLowerCase().includes(qLower) ||
            node.label.toLowerCase().includes(qLower)
          : true;

        let nodeAlpha = 1;
        if (pathSet && !onPath) nodeAlpha = 0.1;
        else if (hovered && !isNeighbor) nodeAlpha = 0.15;
        else if (qLower && !queryMatch) nodeAlpha = 0.1;

        const r = node.radius / scale;
        const [cr, cg, cb] = hexToRgb(node.color);
        const highlighted = isHovered || isActive || onPath;

        // Layer 1: Outer ambient glow (soft halo) — always present, stronger on highlight
        const glowRadius = highlighted ? r * 4 : r * 2.2;
        const glowAlpha = highlighted ? 0.35 * nodeAlpha : 0.12 * nodeAlpha;
        const glow = ctx.createRadialGradient(
          nx, ny, r * 0.3,
          nx, ny, glowRadius,
        );
        glow.addColorStop(0, `rgba(${cr},${cg},${cb},${glowAlpha})`);
        glow.addColorStop(0.5, `rgba(${cr},${cg},${cb},${glowAlpha * 0.4})`);
        glow.addColorStop(1, `rgba(${cr},${cg},${cb},0)`);
        ctx.beginPath();
        ctx.arc(nx, ny, glowRadius, 0, Math.PI * 2);
        ctx.fillStyle = glow;
        ctx.fill();

        // Layer 2: Core disc — soft-edged via gradient (no hard stroke)
        const core = ctx.createRadialGradient(nx, ny, 0, nx, ny, r);
        const coreAlpha = nodeAlpha * (highlighted ? 1 : 0.85);
        core.addColorStop(0, `rgba(${Math.min(255, cr + 60)},${Math.min(255, cg + 60)},${Math.min(255, cb + 60)},${coreAlpha})`);
        core.addColorStop(0.6, `rgba(${cr},${cg},${cb},${coreAlpha})`);
        core.addColorStop(1, `rgba(${cr},${cg},${cb},${coreAlpha * 0.6})`);
        ctx.beginPath();
        ctx.arc(nx, ny, r, 0, Math.PI * 2);
        ctx.fillStyle = core;
        ctx.fill();

        // Layer 3 (highlight only): Bright center point
        if (highlighted && nodeAlpha > 0.5) {
          const bright = ctx.createRadialGradient(nx, ny, 0, nx, ny, r * 0.5);
          bright.addColorStop(0, `rgba(255,255,255,${isDark ? 0.7 : 0.5})`);
          bright.addColorStop(1, `rgba(255,255,255,0)`);
          ctx.beginPath();
          ctx.arc(nx, ny, r * 0.5, 0, Math.PI * 2);
          ctx.fillStyle = bright;
          ctx.fill();
        }

        // Labels — zoom-adaptive visibility
        const showLabel =
          scale > 0.6 ||
          isHovered ||
          isActive ||
          onPath ||
          (qLower && queryMatch);
        if (showLabel && nodeAlpha > 0.3) {
          const fontSize = Math.max(10, 12 / scale);
          ctx.font = `${highlighted ? "600" : "400"} ${fontSize}px system-ui, -apple-system, sans-serif`;
          ctx.textAlign = "center";
          ctx.textBaseline = "top";

          const labelY = ny + r + 4 / scale;
          const textAlpha = Math.min(nodeAlpha, scale > 0.6 ? 1 : 0.8);

          // Halo behind text for readability
          ctx.fillStyle = isDark
            ? `rgba(0,0,0,${textAlpha * 0.6})`
            : `rgba(255,255,255,${textAlpha * 0.6})`;
          ctx.fillText(node.label, nx + 0.5 / scale, labelY + 0.5 / scale);

          ctx.fillStyle = isDark
            ? `rgba(220,220,220,${textAlpha})`
            : `rgba(40,40,40,${textAlpha})`;
          ctx.fillText(node.label, nx, labelY);
        }
      }

      ctx.restore();
      rafRef.current = requestAnimationFrame(render);
    }

    // Start the render loop
    rafRef.current = requestAnimationFrame(render);

    // Stop simulation after it cools
    sim.on("end", () => {});

    return () => {
      cancelAnimationFrame(rafRef.current);
      sim.stop();
      ro.disconnect();
    };
  }, [resp, tree, dirFilter, tagFilter, query, hovered, activePath, foundPath, pathFindActive, sizeByPageRank, colorByCommunity]);

  // ── Mouse interactions ──────────────────────────────────────────────────────

  const screenToWorld = useCallback(
    (sx: number, sy: number): [number, number] => {
      const { x: tx, y: ty, scale } = transformRef.current;
      return [(sx - tx) / scale, (sy - ty) / scale];
    },
    [],
  );

  const findNodeAt = useCallback(
    (wx: number, wy: number): GNode | null => {
      if (!dataRef.current) return null;
      const { scale } = transformRef.current;
      for (let i = dataRef.current.nodes.length - 1; i >= 0; i--) {
        const n = dataRef.current.nodes[i]!;
        if (n.x == null || n.y == null) continue;
        // Filtering: skip hidden nodes
        if (dirFilter && n.dir !== dirFilter) continue;
        if (tagFilter && !n.tags.includes(tagFilter)) continue;
        const dx = wx - n.x;
        const dy = wy - n.y;
        const r = n.radius / scale + 4 / scale;
        if (dx * dx + dy * dy < r * r) return n;
      }
      return null;
    },
    [dirFilter, tagFilter],
  );

  const handlePointerDown = useCallback(
    (e: React.PointerEvent) => {
      const rect = containerRef.current?.getBoundingClientRect();
      if (!rect) return;
      const sx = e.clientX - rect.left;
      const sy = e.clientY - rect.top;
      const [wx, wy] = screenToWorld(sx, sy);
      const node = findNodeAt(wx, wy);

      if (node) {
        draggingRef.current = {
          node,
          pan: false,
          startX: sx,
          startY: sy,
          startTx: 0,
          startTy: 0,
        };
        node.fx = node.x;
        node.fy = node.y;
        simRef.current?.alphaTarget(0.3).restart();
      } else {
        draggingRef.current = {
          node: null,
          pan: true,
          startX: sx,
          startY: sy,
          startTx: transformRef.current.x,
          startTy: transformRef.current.y,
        };
      }
      (e.target as HTMLElement).setPointerCapture(e.pointerId);
    },
    [screenToWorld, findNodeAt],
  );

  const handlePointerMove = useCallback(
    (e: React.PointerEvent) => {
      const rect = containerRef.current?.getBoundingClientRect();
      if (!rect) return;
      const sx = e.clientX - rect.left;
      const sy = e.clientY - rect.top;
      mouseRef.current = { x: sx, y: sy };

      const drag = draggingRef.current;
      if (drag) {
        if (drag.pan) {
          transformRef.current.x = drag.startTx + (sx - drag.startX);
          transformRef.current.y = drag.startTy + (sy - drag.startY);
        } else if (drag.node) {
          const [wx, wy] = screenToWorld(sx, sy);
          drag.node.fx = wx;
          drag.node.fy = wy;
        }
      } else {
        const [wx, wy] = screenToWorld(sx, sy);
        const node = findNodeAt(wx, wy);
        setHovered(node);
        if (containerRef.current) {
          containerRef.current.style.cursor = node ? "pointer" : "grab";
        }
      }
    },
    [screenToWorld, findNodeAt],
  );

  const handlePointerUp = useCallback(
    (e: React.PointerEvent) => {
      const drag = draggingRef.current;
      if (drag?.node) {
        const dx = Math.abs(
          e.clientX -
            (containerRef.current?.getBoundingClientRect().left ?? 0) -
            drag.startX,
        );
        const dy = Math.abs(
          e.clientY -
            (containerRef.current?.getBoundingClientRect().top ?? 0) -
            drag.startY,
        );
        // Treat as click if barely moved
        if (dx < 4 && dy < 4) {
          if (pathFindActive) {
            if (!pathSource) setPathSource(drag.node.id);
            else if (!pathTarget) setPathTarget(drag.node.id);
            else {
              setPathSource(drag.node.id);
              setPathTarget(null);
              setFoundPath(null);
            }
          } else {
            onNavigate(drag.node.id);
          }
        }
        drag.node.fx = null;
        drag.node.fy = null;
        simRef.current?.alphaTarget(0);
      }
      draggingRef.current = null;
    },
    [onNavigate, pathFindActive, pathSource, pathTarget],
  );

  const handleWheel = useCallback((e: React.WheelEvent) => {
    e.preventDefault();
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) return;
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;
    const factor = e.deltaY < 0 ? 1.08 : 1 / 1.08;
    const t = transformRef.current;
    const newScale = Math.max(0.1, Math.min(8, t.scale * factor));
    const ratio = newScale / t.scale;
    t.x = mx - (mx - t.x) * ratio;
    t.y = my - (my - t.y) * ratio;
    t.scale = newScale;
  }, []);

  // Path finding
  useEffect(() => {
    if (!pathSource || !pathTarget) {
      setFoundPath(null);
      return;
    }
    const path = bfsPath(adjRef.current, pathSource, pathTarget);
    setFoundPath(path.length > 0 ? path : null);
  }, [pathSource, pathTarget]);

  useEffect(() => {
    if (!pathFindActive) {
      setPathSource(null);
      setPathTarget(null);
      setFoundPath(null);
    }
  }, [pathFindActive]);

  const built = dataRef.current;

  return (
    <div className="h-full w-full flex flex-col relative">
      {/* ── Toolbar ── */}
      <div className="flex flex-wrap items-center gap-2 sm:gap-3 px-3 sm:px-6 py-3 border-b border-border bg-card shrink-0">
        <Button variant="outline" size="sm" onClick={onClose}>
          <ArrowLeft className="h-3.5 w-3.5" />{" "}
          <span className="hidden sm:inline">Back</span>
        </Button>
        <div className="font-semibold text-sm">Knowledge graph</div>
        <div className="text-xs text-muted-foreground hidden sm:block">
          {built
            ? `${built.nodes.length} pages · ${built.links.length} links`
            : null}
        </div>
        <div className="ml-auto flex items-center gap-2 flex-wrap">
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
        </div>
      </div>

      {/* ── Analytics bar ── */}
      <div className="flex flex-wrap items-center gap-2 px-3 sm:px-6 py-1.5 border-b border-border/50 bg-card/50 text-xs shrink-0">
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

      {/* ── Canvas area ── */}
      <div className="flex-1 relative min-h-0">
        {error && (
          <div className="absolute inset-0 grid place-items-center text-sm text-destructive font-mono z-10">
            {error}
          </div>
        )}
        {!error && !resp && (
          <div className="absolute inset-0 grid place-items-center text-sm text-muted-foreground z-10">
            <div className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" /> Building graph...
            </div>
          </div>
        )}
        {resp && built && built.nodes.length === 0 && (
          <div className="absolute inset-0 grid place-items-center text-sm text-muted-foreground z-10">
            No pages yet.
          </div>
        )}

        <div
          ref={containerRef}
          className="absolute inset-0"
          style={{ cursor: "grab", touchAction: "none" }}
          onPointerDown={handlePointerDown}
          onPointerMove={handlePointerMove}
          onPointerUp={handlePointerUp}
          onWheel={handleWheel}
        />

        {/* Hover tooltip */}
        {hovered && (
          <Card
            className="absolute bottom-3 left-3 px-3 py-2 text-xs pointer-events-none max-w-xs z-20"
          >
            <div className="font-medium">{hovered.label}</div>
            <div className="text-muted-foreground font-mono text-[10px]">
              {hovered.id}
            </div>
            <div className="flex flex-wrap gap-x-3 gap-y-0.5 text-muted-foreground mt-1">
              <span>
                {adjRef.current.get(hovered.id)?.size ?? 0} connection
                {(adjRef.current.get(hovered.id)?.size ?? 0) === 1 ? "" : "s"}
              </span>
              <span>
                PR: {(hovered.pagerank * 100).toFixed(2)}%
              </span>
              <span>Community: {hovered.community}</span>
            </div>
            {hovered.tags.length > 0 && (
              <div className="flex flex-wrap gap-1 mt-1">
                {hovered.tags.map((t) => (
                  <span
                    key={t}
                    className="bg-muted px-1 rounded text-[10px]"
                  >
                    {t}
                  </span>
                ))}
              </div>
            )}
          </Card>
        )}

        {/* Path overlay */}
        {foundPath && foundPath.length > 1 && (
          <Card className="absolute bottom-3 left-1/2 -translate-x-1/2 px-3 py-1.5 text-xs pointer-events-none bg-primary text-primary-foreground z-20">
            Shortest path: {foundPath.length - 1} hop
            {foundPath.length - 1 !== 1 ? "s" : ""} (
            {foundPath.map((n) => titleize(n)).join(" → ")})
          </Card>
        )}

        {/* Community legend */}
        {built && built.communityMap.size > 1 && (
          <Card className="absolute top-3 right-3 px-3 py-2 text-xs z-20">
            <div className="text-muted-foreground mb-1.5 font-medium">
              Communities
            </div>
            <div className="space-y-1">
              {Array.from(built.communityMap.entries())
                .sort((a, b) => b[1].count - a[1].count)
                .map(([idx, { color, count, topDir }]) => (
                  <div key={idx} className="flex items-center gap-2">
                    <span
                      className="h-2.5 w-2.5 rounded-full shrink-0"
                      style={{ background: color }}
                    />
                    <span className="text-muted-foreground">
                      {topDir} ({count})
                    </span>
                  </div>
                ))}
            </div>
          </Card>
        )}

        <div
          className={cn(
            "absolute bottom-3 right-3 text-[10px] text-muted-foreground font-mono",
            "pointer-events-none z-20",
          )}
        >
          drag to pan · scroll to zoom
        </div>
      </div>
    </div>
  );
}

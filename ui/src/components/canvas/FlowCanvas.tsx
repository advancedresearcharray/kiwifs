// FlowCanvas — React Flow renderer for JSON Canvas documents.
// Supports custom node types (text, file, link, group), SSE live reload,
// and client-side dagre auto-layout.

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  ReactFlow,
  MiniMap,
  Controls,
  Background,
  BackgroundVariant,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { LayoutGrid, Loader2, Save } from "lucide-react";
import { Button } from "@kw/components/ui/button";
import { api, sseUrl } from "@kw/lib/api";
import { applyDagreLayout } from "@kw/lib/canvasLayout";
import CanvasTextNode from "./CanvasTextNode";
import CanvasFileNode from "./CanvasFileNode";
import CanvasLinkNode from "./CanvasLinkNode";
import CanvasGroupNode from "./CanvasGroupNode";

const nodeTypes = {
  text: CanvasTextNode,
  file: CanvasFileNode,
  link: CanvasLinkNode,
  group: CanvasGroupNode,
};

type CanvasNode = {
  id: string;
  type: string;
  x: number;
  y: number;
  width: number;
  height: number;
  text?: string;
  file?: string;
  url?: string;
  color?: string;
};
type CanvasEdge = {
  id: string;
  fromNode: string;
  toNode: string;
  fromSide?: string;
  toSide?: string;
  label?: string;
  color?: string;
};
type CanvasDoc = { nodes: CanvasNode[]; edges: CanvasEdge[] };

function toFlowNodes(raw: CanvasNode[]): Node[] {
  return raw.map((n) => ({
    id: n.id,
    type: n.type === "file" || n.type === "link" || n.type === "group" ? n.type : "text",
    position: { x: n.x ?? 0, y: n.y ?? 0 },
    data: { text: n.text, file: n.file, url: n.url, color: n.color },
    width: n.width,
    height: n.height,
  }));
}

function toFlowEdges(raw: CanvasEdge[]): Edge[] {
  return raw.map((e) => ({
    id: e.id,
    source: e.fromNode,
    target: e.toNode,
    label: e.label,
    animated: false,
    style: e.color ? { stroke: e.color } : undefined,
  }));
}

function toCanvasDoc(nodes: Node[], edges: Edge[]): CanvasDoc {
  return {
    nodes: nodes.map((n) => ({
      id: n.id,
      type: n.type ?? "text",
      x: n.position.x,
      y: n.position.y,
      width: n.width ?? 240,
      height: n.height ?? 80,
      ...n.data,
    })),
    edges: edges.map((e) => ({
      id: e.id,
      fromNode: e.source,
      toNode: e.target,
      ...(e.label ? { label: String(e.label) } : {}),
    })),
  };
}

type Props = {
  path: string;
  onNavigate: (path: string) => void;
};

export function FlowCanvas({ path, onNavigate: _onNavigate }: Props) {
  void _onNavigate;
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const suppressSSERef = useRef(false);

  const loadCanvas = useCallback(async () => {
    try {
      const data = await api.getCanvas(path);
      const doc = data as unknown as CanvasDoc;
      const flowNodes = toFlowNodes(doc.nodes ?? []);
      const flowEdges = toFlowEdges(doc.edges ?? []);
      setNodes(flowNodes);
      setEdges(flowEdges);
    } catch {
      setNodes([]);
      setEdges([]);
    } finally {
      setLoading(false);
    }
  }, [path, setNodes, setEdges]);

  useEffect(() => {
    loadCanvas();
  }, [loadCanvas]);

  // SSE: live reload when another actor writes to this canvas
  useEffect(() => {
    const es = new EventSource(sseUrl());
    const onWrite = (ev: MessageEvent) => {
      if (suppressSSERef.current) return;
      try {
        const data = JSON.parse(ev.data);
        if (data.path === path) {
          loadCanvas();
        }
      } catch { /* ignore */ }
    };
    es.addEventListener("write", onWrite);
    return () => {
      es.removeEventListener("write", onWrite);
      es.close();
    };
  }, [path, loadCanvas]);

  const save = useCallback(async () => {
    setSaving(true);
    suppressSSERef.current = true;
    try {
      const doc = toCanvasDoc(nodes, edges);
      await api.saveCanvas(path, doc as unknown as Record<string, unknown>);
    } catch (e) {
      console.error("Canvas save failed:", e);
    } finally {
      setSaving(false);
      setTimeout(() => { suppressSSERef.current = false; }, 1000);
    }
  }, [path, nodes, edges]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "s") {
        e.preventDefault();
        save();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [save]);

  const scheduleAutosave = useCallback(() => {
    if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    saveTimerRef.current = setTimeout(save, 3000);
  }, [save]);

  useEffect(() => {
    return () => {
      if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    };
  }, []);

  const handleNodesChange = useCallback(
    (changes: Parameters<typeof onNodesChange>[0]) => {
      onNodesChange(changes);
      if (changes.some((c) => c.type === "position" && "dragging" in c && !c.dragging)) {
        scheduleAutosave();
      }
    },
    [onNodesChange, scheduleAutosave],
  );

  const handleEdgesChange = useCallback(
    (changes: Parameters<typeof onEdgesChange>[0]) => {
      onEdgesChange(changes);
      scheduleAutosave();
    },
    [onEdgesChange, scheduleAutosave],
  );

  const handleAutoLayout = useCallback(() => {
    const { nodes: laid, edges: laidEdges } = applyDagreLayout(nodes, edges, "TB");
    setNodes(laid);
    setEdges(laidEdges);
    scheduleAutosave();
  }, [nodes, edges, setNodes, setEdges, scheduleAutosave]);

  const proOptions = useMemo(() => ({ hideAttribution: true }), []);

  if (loading) {
    return (
      <div className="h-full grid place-items-center text-muted-foreground">
        <div className="flex items-center gap-2">
          <Loader2 className="h-4 w-4 animate-spin" /> Loading canvas...
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center gap-2 px-3 py-1.5 border-b border-border bg-card shrink-0">
        <div className="ml-auto flex items-center gap-1.5">
          <Button variant="outline" size="sm" className="gap-1 h-7 text-xs" onClick={handleAutoLayout}>
            <LayoutGrid className="h-3 w-3" /> Auto-layout
          </Button>
          <Button variant="outline" size="sm" className="gap-1 h-7 text-xs" onClick={save} disabled={saving}>
            {saving ? <Loader2 className="h-3 w-3 animate-spin" /> : <Save className="h-3 w-3" />}
            Save
          </Button>
        </div>
      </div>
      <div className="flex-1">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={handleNodesChange}
          onEdgesChange={handleEdgesChange}
          nodeTypes={nodeTypes}
          proOptions={proOptions}
          fitView
          fitViewOptions={{ padding: 0.2 }}
          minZoom={0.1}
          maxZoom={4}
          defaultEdgeOptions={{ animated: false, type: "smoothstep" }}
        >
          <Background variant={BackgroundVariant.Dots} gap={20} size={1} />
          <Controls showInteractive={false} />
          <MiniMap
            nodeStrokeWidth={2}
            pannable
            zoomable
            className="!bg-card !border-border"
          />
        </ReactFlow>
      </div>
    </div>
  );
}

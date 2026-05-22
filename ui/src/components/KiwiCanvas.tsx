// KiwiCanvas — Dual-mode canvas: React Flow (structured) or Excalidraw (freeform).

import { Suspense, lazy, useCallback, useEffect, useRef, useState } from "react";
import { ArrowLeft, Loader2, Paintbrush, Save, Workflow } from "lucide-react";
import { api } from "@kw/lib/api";
import { Button } from "@kw/components/ui/button";
import {
  excalidrawSceneToJsonCanvas,
  jsonCanvasToExcalidrawScene,
  type ExcalidrawSceneLike,
  type JSONCanvas,
} from "@kw/lib/jsonCanvasExcalidraw";
import { FlowCanvas } from "./canvas/FlowCanvas";

type Props = {
  path: string | null;
  embedded?: boolean;
  onClose?: () => void;
  onNavigate: (path: string) => void;
};

type Mode = "flow" | "excalidraw";
const MODE_KEY = "kiwifs-canvas-mode";

const ExcalidrawCanvas = lazy(() =>
  import("@excalidraw/excalidraw").then((module) => ({ default: module.Excalidraw })),
);

export function KiwiCanvas({ path, embedded = false, onClose, onNavigate }: Props) {
  const [mode, setMode] = useState<Mode>(() => {
    try {
      const saved = localStorage.getItem(MODE_KEY);
      return saved === "excalidraw" ? "excalidraw" : "flow";
    } catch {
      return "flow";
    }
  });

  const toggleMode = useCallback(() => {
    const next = mode === "flow" ? "excalidraw" : "flow";
    setMode(next);
    try { localStorage.setItem(MODE_KEY, next); } catch { /* */ }
  }, [mode]);

  if (mode === "flow" && path) {
    return (
      <div className="h-full flex flex-col">
        {!embedded && (
          <div className="flex items-center gap-2 px-3 py-2 border-b border-border bg-card shrink-0">
            {onClose && (
              <Button variant="outline" size="sm" onClick={onClose}>
                <ArrowLeft className="h-3.5 w-3.5" />
                <span className="hidden sm:inline">Back to pages</span>
              </Button>
            )}
            <div className="font-semibold text-sm">Canvas: {path}</div>
            <div className="ml-auto">
              <Button variant="ghost" size="sm" className="gap-1 h-7 text-xs" onClick={toggleMode}>
                <Paintbrush className="h-3 w-3" /> Freeform
              </Button>
            </div>
          </div>
        )}
        {embedded && (
          <div className="flex items-center gap-1 px-2 py-1 bg-card shrink-0">
            <Button variant="ghost" size="sm" className="gap-1 h-7 text-xs ml-auto" onClick={toggleMode}>
              <Paintbrush className="h-3 w-3" /> Freeform
            </Button>
          </div>
        )}
        <div className="flex-1 min-h-0">
          <FlowCanvas key={path} path={path} onNavigate={onNavigate} />
        </div>
      </div>
    );
  }

  return (
    <ExcalidrawMode
      path={path}
      embedded={embedded}
      onClose={onClose}
      onToggle={toggleMode}
    />
  );
}

// Excalidraw freeform mode (legacy)
function ExcalidrawMode({
  path,
  embedded,
  onClose,
  onToggle,
}: {
  path: string | null;
  embedded: boolean;
  onClose?: () => void;
  onToggle: () => void;
}) {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [initialScene, setInitialScene] = useState<ExcalidrawSceneLike | null>(null);
  const sceneRef = useRef<ExcalidrawSceneLike | null>(null);
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (!path) {
      const blankScene = jsonCanvasToExcalidrawScene({ nodes: [], edges: [] });
      sceneRef.current = blankScene;
      setInitialScene(blankScene);
      setLoading(false);
      return;
    }
    setLoading(true);
    api
      .getCanvas(path)
      .then((data) => {
        const canvas: JSONCanvas = {
          nodes: (data.nodes as JSONCanvas["nodes"]) || [],
          edges: (data.edges as JSONCanvas["edges"]) || [],
        };
        const scene = jsonCanvasToExcalidrawScene(canvas);
        sceneRef.current = scene;
        setInitialScene(scene);
      })
      .catch(() => {
        const scene = jsonCanvasToExcalidrawScene({ nodes: [], edges: [] });
        sceneRef.current = scene;
        setInitialScene(scene);
      })
      .finally(() => setLoading(false));
  }, [path]);

  const save = useCallback(async () => {
    if (!sceneRef.current || !path) return;
    setSaving(true);
    try {
      const canvas = excalidrawSceneToJsonCanvas(sceneRef.current.elements);
      await api.saveCanvas(path, canvas as unknown as Record<string, unknown>);
    } catch (e) {
      console.error("Canvas save failed:", e);
    } finally {
      setSaving(false);
    }
  }, [path]);

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

  const handleChange = useCallback(
    (elements: readonly unknown[], appState: unknown, files: unknown) => {
      if (!sceneRef.current) return;
      sceneRef.current = {
        ...sceneRef.current,
        elements: elements as ExcalidrawSceneLike["elements"],
        appState: { ...(appState as Record<string, unknown>), collaborators: new Map() },
        files: (files as Record<string, unknown>) ?? {},
      };
      scheduleAutosave();
    },
    [scheduleAutosave],
  );

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center gap-2 px-3 py-2 border-b border-border bg-card shrink-0">
        {!embedded && onClose && (
          <Button variant="outline" size="sm" onClick={onClose}>
            <ArrowLeft className="h-3.5 w-3.5" />
            <span className="hidden sm:inline">Back to pages</span>
          </Button>
        )}
        {!embedded && (
          <div className="font-semibold text-sm">Canvas{path ? `: ${path}` : ""}</div>
        )}
        <div className={embedded ? "ml-auto flex items-center gap-2" : "ml-auto flex items-center gap-2"}>
          <Button variant="ghost" size="sm" className="gap-1 h-7 text-xs" onClick={onToggle}>
            <Workflow className="h-3 w-3" /> Structured
          </Button>
          <Button variant="outline" size="sm" className="gap-1" onClick={save} disabled={saving || !path}>
            {saving ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Save className="h-3.5 w-3.5" />}
            Save
          </Button>
        </div>
      </div>
      <div className="flex-1 relative">
        {loading || !initialScene ? (
          <div className="absolute inset-0 grid place-items-center text-muted-foreground">
            <div className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" /> Loading canvas...
            </div>
          </div>
        ) : (
          <div className="absolute inset-0 overflow-hidden bg-background">
            <Suspense fallback={<div className="p-4 text-sm text-muted-foreground">Loading Excalidraw...</div>}>
              <ExcalidrawCanvas
                key={path ?? "new-canvas"}
                initialData={initialScene as never}
                onChange={handleChange as never}
              />
            </Suspense>
          </div>
        )}
      </div>
    </div>
  );
}

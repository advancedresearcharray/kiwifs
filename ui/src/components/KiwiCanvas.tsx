// KiwiCanvas — Infinite whiteboard powered by Excalidraw for spatial note
// organization. Loads/saves in JSON Canvas format.

import { Suspense, lazy, useCallback, useEffect, useRef, useState } from "react";
import "@excalidraw/excalidraw/index.css";
import { ArrowLeft, Loader2, Save } from "lucide-react";
import { api } from "@kw/lib/api";
import { Button } from "@kw/components/ui/button";
import {
  excalidrawSceneToJsonCanvas,
  jsonCanvasToExcalidrawScene,
  type ExcalidrawSceneLike,
  type JSONCanvas,
} from "@kw/lib/jsonCanvasExcalidraw";

type Props = {
  path: string | null;
  onClose: () => void;
  onNavigate: (path: string) => void;
};

const ExcalidrawCanvas = lazy(() =>
  import("@excalidraw/excalidraw").then((module) => ({ default: module.Excalidraw })),
);

export function KiwiCanvas({ path, onClose, onNavigate: _onNavigate }: Props) {
  void _onNavigate; // Reserved for future page-shape double-click navigation
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [initialScene, setInitialScene] = useState<ExcalidrawSceneLike | null>(null);
  const sceneRef = useRef<ExcalidrawSceneLike | null>(null);
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Load canvas
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
        // 404 or any error: start with a blank canvas (will be created on first save)
        const scene = jsonCanvasToExcalidrawScene({ nodes: [], edges: [] });
        sceneRef.current = scene;
        setInitialScene(scene);
      })
      .finally(() => setLoading(false));
  }, [path]);

  // Save handler
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

  // Ctrl/Cmd+S save
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

  // Auto-save debounce on editor changes
  const scheduleAutosave = useCallback(() => {
    if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    saveTimerRef.current = setTimeout(() => {
      save();
    }, 3000);
  }, [save]);

  // Clean up timer
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
        appState: {
          ...(appState as Record<string, unknown>),
          collaborators: new Map(),
        },
        files: (files as Record<string, unknown>) ?? {},
      };
      scheduleAutosave();
    },
    [scheduleAutosave],
  );

  return (
    <div className="h-full flex flex-col">
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-3 sm:px-6 py-3 border-b border-border bg-card shrink-0">
        <Button variant="outline" size="sm" onClick={onClose}>
          <ArrowLeft className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Back to pages</span>
        </Button>
        <div className="font-semibold text-sm">
          Canvas{path ? `: ${path}` : ""}
        </div>
        <div className="ml-auto flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            className="gap-1"
            onClick={save}
            disabled={saving || !path}
          >
            {saving ? (
              <Loader2 className="h-3.5 w-3.5 animate-spin" />
            ) : (
              <Save className="h-3.5 w-3.5" />
            )}
            Save
          </Button>
        </div>
      </div>

      {/* Canvas area */}
      <div className="flex-1 relative">
        {loading || !initialScene ? (
          <div className="absolute inset-0 grid place-items-center text-muted-foreground">
            <div className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" /> Loading canvas...
            </div>
          </div>
        ) : (
          <div className="absolute inset-0 overflow-hidden bg-background">
            <Suspense fallback={<div className="p-4 text-sm text-muted-foreground">Loading Excalidraw canvas…</div>}>
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

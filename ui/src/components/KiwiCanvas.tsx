// KiwiCanvas — Infinite whiteboard powered by tldraw for spatial note
// organization. Loads/saves in JSON Canvas format.

import { useCallback, useEffect, useRef, useState } from "react";
import { Tldraw, type Editor } from "tldraw";
import "tldraw/tldraw.css";
import { ArrowLeft, Loader2, Save } from "lucide-react";
import { api } from "@kw/lib/api";
import { Button } from "@kw/components/ui/button";
import {
  jsonCanvasToShapes,
  shapesToJsonCanvas,
  type JSONCanvas,
  type TldrawShapeRecord,
} from "./canvas/canvasAdapter";

type Props = {
  path: string | null;
  onClose: () => void;
  onNavigate: (path: string) => void;
};

export function KiwiCanvas({ path, onClose, onNavigate: _onNavigate }: Props) {
  void _onNavigate; // Reserved for future page-shape double-click navigation
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [initialCanvas, setInitialCanvas] = useState<JSONCanvas | null>(null);
  const editorRef = useRef<Editor | null>(null);
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Load canvas
  useEffect(() => {
    if (!path) {
      setLoading(false);
      setInitialCanvas({ nodes: [], edges: [] });
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
        setInitialCanvas(canvas);
      })
      .catch(() => {
        // 404 or any error: start with a blank canvas (will be created on first save)
        setInitialCanvas({ nodes: [], edges: [] });
      })
      .finally(() => setLoading(false));
  }, [path]);

  // Save handler
  const save = useCallback(async () => {
    if (!editorRef.current || !path) return;
    setSaving(true);
    try {
      const editor = editorRef.current;
      // Get all shapes from the editor
      const allShapes = editor.getCurrentPageShapes();
      const records: TldrawShapeRecord[] = allShapes.map((s) => ({
        id: s.id,
        type: s.type,
        x: s.x,
        y: s.y,
        props: s.props as Record<string, unknown>,
        meta: (s.meta as Record<string, unknown>) || {},
      }));
      const canvas = shapesToJsonCanvas(records);
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

  const handleMount = useCallback(
    (editor: Editor) => {
      editorRef.current = editor;

      // Load initial shapes from JSON Canvas
      if (initialCanvas && initialCanvas.nodes.length > 0) {
        const shapes = jsonCanvasToShapes(initialCanvas);
        for (const shape of shapes) {
          try {
            editor.createShape({
              id: shape.id as any,
              type: shape.type as any,
              x: shape.x,
              y: shape.y,
              props: shape.props,
              meta: shape.meta as any,
            });
          } catch {
            // Shape type might not be supported, skip
          }
        }

        // Zoom to fit after adding shapes
        setTimeout(() => {
          editor.zoomToFit({ animation: { duration: 200 } });
        }, 100);
      }

      // Listen for changes for auto-save
      editor.store.listen(() => {
        scheduleAutosave();
      });

      // Double-click note shapes to navigate to kiwi pages
      editor.on("event", (info) => {
        if (info.type === "pointer" && info.name === "pointer_down") {
          // Check for double-click navigation
          const selectedShapes = editor.getSelectedShapes();
          if (selectedShapes.length === 1) {
            const shape = selectedShapes[0]!;
            const meta = (shape.meta as Record<string, unknown>) || {};
            if (meta.kiwiPath && typeof meta.kiwiPath === "string") {
              // Will be navigated on double-click via a timeout check
              // tldraw handles double-click natively for text editing
            }
          }
        }
      });
    },
    [initialCanvas, scheduleAutosave],
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
        {loading ? (
          <div className="absolute inset-0 grid place-items-center text-muted-foreground">
            <div className="flex items-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" /> Loading canvas...
            </div>
          </div>
        ) : (
          <Tldraw
            onMount={handleMount}
            autoFocus
          />
        )}
      </div>
    </div>
  );
}

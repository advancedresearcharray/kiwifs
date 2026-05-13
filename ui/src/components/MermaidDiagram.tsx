import { useEffect, useRef, useState } from "react";

let mermaidInitTheme: "dark" | "default" | null = null;

const MIN_ZOOM = 0.5;
const MAX_ZOOM = 5;
const ZOOM_STEP = 0.25;

async function getMermaid() {
  const { default: mermaid } = await import("mermaid");
  return mermaid;
}

async function ensureInit(theme: "dark" | "default") {
  const mermaid = await getMermaid();
  if (mermaidInitTheme !== theme) {
    mermaid.initialize({
      startOnLoad: false,
      securityLevel: "strict",
      theme,
    });
    mermaidInitTheme = theme;
  }
  return mermaid;
}

function clampZoom(nextZoom: number) {
  return Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, nextZoom));
}

export function MermaidDiagram({ chart }: { chart: string }) {
  const [svg, setSvg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [zoom, setZoom] = useState(1);
  const containerRef = useRef<HTMLDivElement>(null);
  const renderIdRef = useRef(`kiwi-mermaid-${Math.random().toString(36).slice(2)}`);

  const [isDark, setIsDark] = useState<boolean>(
    () =>
      typeof document !== "undefined" &&
      document.documentElement.classList.contains("dark"),
  );

  useEffect(() => {
    const obs = new MutationObserver(() =>
      setIsDark(document.documentElement.classList.contains("dark")),
    );
    obs.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    });
    return () => obs.disconnect();
  }, []);

  useEffect(() => {
    let cancelled = false;

    async function renderDiagram() {
      setSvg(null);
      setError(null);
      setZoom(1);

      try {
        const mermaid = await ensureInit(isDark ? "dark" : "default");
        const rendered = await mermaid.render(renderIdRef.current, chart);
        if (cancelled) return;

        setSvg(rendered.svg);
        queueMicrotask(() => {
          if (containerRef.current) rendered.bindFunctions?.(containerRef.current);
        });
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : String(e));
      }
    }

    renderDiagram();
    return () => {
      cancelled = true;
    };
  }, [chart, isDark]);

  if (error) {
    return (
      <figure className="kiwi-mermaid rounded-md border border-destructive/30 bg-destructive/5 p-4">
        <figcaption className="mb-2 text-sm font-medium text-destructive">
          Mermaid render error
        </figcaption>
        <pre className="overflow-x-auto text-xs">
          <code>{error}</code>
        </pre>
      </figure>
    );
  }

  return (
    <figure className="kiwi-mermaid relative rounded-md border border-border bg-card p-4">
      {svg ? (
        <>
          <div
            className="absolute right-3 top-3 z-10 flex items-center gap-1 rounded-md border border-border bg-background px-1.5 py-1 text-xs shadow-md"
            aria-label="Mermaid diagram zoom controls"
          >
            <button
              type="button"
              className="h-7 w-7 rounded-sm border border-border bg-card font-semibold leading-none hover:bg-accent focus:outline-none focus:ring-2 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              onClick={() => setZoom((current) => clampZoom(current - ZOOM_STEP))}
              disabled={zoom <= MIN_ZOOM}
              aria-label="Zoom out Mermaid diagram"
              title="Zoom out"
            >
              −
            </button>
            <span className="min-w-12 px-1.5 py-1 text-center tabular-nums">
              {Math.round(zoom * 100)}%
            </span>
            {zoom !== 1 && (
              <button
                type="button"
                className="rounded-sm border border-border bg-card px-2 py-1 font-medium hover:bg-accent focus:outline-none focus:ring-2 focus:ring-ring"
                onClick={() => setZoom(1)}
                aria-label="Reset Mermaid diagram zoom"
                title="Reset zoom"
              >
                Reset
              </button>
            )}
            <button
              type="button"
              className="h-7 w-7 rounded-sm border border-border bg-card font-semibold leading-none hover:bg-accent focus:outline-none focus:ring-2 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              onClick={() => setZoom((current) => clampZoom(current + ZOOM_STEP))}
              disabled={zoom >= MAX_ZOOM}
              aria-label="Zoom in Mermaid diagram"
              title="Zoom in"
            >
              +
            </button>
          </div>
          <div className="overflow-auto pt-10">
            <div
              ref={containerRef}
              className="mx-auto [&_svg]:h-auto [&_svg]:w-full"
              style={{ width: `${zoom * 100}%` }}
              dangerouslySetInnerHTML={{ __html: svg }}
            />
          </div>
        </>
      ) : (
        <div className="text-sm text-muted-foreground">
          Rendering Mermaid diagram&hellip;
        </div>
      )}
    </figure>
  );
}

import { useEffect, useRef, useState } from "react";

let mermaidInitTheme: "dark" | "default" | null = null;

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

export function MermaidDiagram({ chart }: { chart: string }) {
  const [svg, setSvg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
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
    <figure className="kiwi-mermaid overflow-x-auto rounded-md border border-border bg-card p-4">
      {svg ? (
        <div
          ref={containerRef}
          className="min-w-fit [&_svg]:mx-auto [&_svg]:max-w-full"
          dangerouslySetInnerHTML={{ __html: svg }}
        />
      ) : (
        <div className="text-sm text-muted-foreground">
          Rendering Mermaid diagram&hellip;
        </div>
      )}
    </figure>
  );
}

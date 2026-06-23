/**
 * Resizable split panes — shadcn-compatible API.
 * Uses pointer-driven sizing (same pattern as the app sidebar) so we do not
 * require react-resizable-panels at build time in restricted environments.
 */
import * as React from "react";
import { GripVertical } from "lucide-react";
import { cn } from "@kw/lib/cn";

type PanelGroupContextValue = {
  direction: "horizontal" | "vertical";
  sizes: number[];
  setSizes: React.Dispatch<React.SetStateAction<number[]>>;
  panelCount: number;
  onLayout?: (sizes: number[]) => void;
};

const PanelGroupContext = React.createContext<PanelGroupContextValue | null>(null);

type ResizablePanelGroupProps = React.HTMLAttributes<HTMLDivElement> & {
  direction?: "horizontal" | "vertical";
  /** Initial panel sizes (percentages). Avoids resetting persisted layout on mount. */
  defaultSizes?: number[];
  onLayout?: (sizes: number[]) => void;
};

function ResizablePanelGroup({
  className,
  direction = "horizontal",
  defaultSizes,
  onLayout,
  children,
  ...props
}: ResizablePanelGroupProps) {
  const panels = React.Children.toArray(children).filter(Boolean);
  const panelCount = panels.length;
  const [sizes, setSizes] = React.useState<number[]>(() => {
    if (defaultSizes?.length === panelCount) return [...defaultSizes];
    return Array.from({ length: panelCount }, () => 100 / Math.max(panelCount, 1));
  });
  const onLayoutRef = React.useRef(onLayout);
  onLayoutRef.current = onLayout;

  React.useEffect(() => {
    setSizes((prev) => {
      if (prev.length === panelCount) return prev;
      if (defaultSizes?.length === panelCount) return [...defaultSizes];
      return Array.from({ length: panelCount }, () => 100 / Math.max(panelCount, 1));
    });
  }, [panelCount, defaultSizes]);

  const notifyLayout = React.useCallback((next: number[]) => {
    onLayoutRef.current?.(next);
  }, []);

  return (
    <PanelGroupContext.Provider value={{ direction, sizes, setSizes, panelCount, onLayout: notifyLayout }}>
      <div
        className={cn(
          "flex h-full w-full",
          direction === "vertical" ? "flex-col" : "flex-row",
          className,
        )}
        {...props}
      >
        {panels}
      </div>
    </PanelGroupContext.Provider>
  );
}

type ResizablePanelProps = React.HTMLAttributes<HTMLDivElement> & {
  defaultSize?: number;
  minSize?: number;
  index?: number;
};

const ResizablePanel = React.forwardRef<HTMLDivElement, ResizablePanelProps>(
  ({ className, defaultSize, minSize = 10, index = 0, style, ...props }, ref) => {
    const ctx = React.useContext(PanelGroupContext);
    const size = ctx?.sizes[index] ?? defaultSize ?? 50;

    React.useEffect(() => {
      if (defaultSize == null || !ctx) return;
      ctx.setSizes((prev) => {
        if (prev[index] != null && prev[index] !== 100 / ctx.panelCount) return prev;
        const next = [...prev];
        next[index] = defaultSize;
        const other = 100 - defaultSize;
        for (let i = 0; i < next.length; i++) {
          if (i !== index) next[i] = other / Math.max(ctx.panelCount - 1, 1);
        }
        return next;
      });
    }, [defaultSize, ctx, index]);

    return (
      <div
        ref={ref}
        className={cn("min-h-0 min-w-0 overflow-hidden", className)}
        style={{
          flexBasis: `${size}%`,
          flexGrow: 0,
          flexShrink: 0,
          minWidth: ctx?.direction === "horizontal" ? `${minSize}%` : undefined,
          minHeight: ctx?.direction === "vertical" ? `${minSize}%` : undefined,
          ...style,
        }}
        {...props}
      />
    );
  },
);
ResizablePanel.displayName = "ResizablePanel";

type ResizableHandleProps = React.HTMLAttributes<HTMLDivElement> & {
  withHandle?: boolean;
  /** Index of the panel to the left (or above) of this handle. */
  index?: number;
};

function ResizableHandle({
  className,
  withHandle,
  index = 0,
  ...props
}: ResizableHandleProps) {
  const ctx = React.useContext(PanelGroupContext);
  const dragging = React.useRef(false);

  const onPointerDown = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!ctx) return;
    e.preventDefault();
    dragging.current = true;
    const start = ctx.direction === "horizontal" ? e.clientX : e.clientY;
    const startSizes = [...ctx.sizes];
    let latestSizes = startSizes;
    const el = e.currentTarget.parentElement;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    const total = ctx.direction === "horizontal" ? rect.width : rect.height;

    const onMove = (ev: PointerEvent) => {
      if (!dragging.current) return;
      const delta = (ctx.direction === "horizontal" ? ev.clientX : ev.clientY) - start;
      const deltaPct = (delta / total) * 100;
      const left = Math.max(10, Math.min(90, startSizes[index] + deltaPct));
      const right = Math.max(10, Math.min(90, startSizes[index + 1] - deltaPct));
      const next = [...startSizes];
      next[index] = left;
      next[index + 1] = right;
      latestSizes = next;
      ctx.setSizes(next);
    };

    const onUp = () => {
      dragging.current = false;
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerup", onUp);
      ctx.onLayout?.(latestSizes);
    };

    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", onUp);
  };

  return (
    <div
      role="separator"
      aria-orientation={ctx?.direction === "vertical" ? "horizontal" : "vertical"}
      className={cn(
        "relative flex shrink-0 items-center justify-center bg-border",
        ctx?.direction === "vertical"
          ? "h-1.5 w-full cursor-row-resize"
          : "w-1.5 h-full cursor-col-resize",
        className,
      )}
      onPointerDown={onPointerDown}
      {...props}
    >
      {withHandle ? (
        <div className="z-10 flex h-4 w-3 items-center justify-center rounded-sm border bg-border">
          <GripVertical className="h-2.5 w-2.5" />
        </div>
      ) : null}
    </div>
  );
}

export { ResizablePanelGroup, ResizablePanel, ResizableHandle };

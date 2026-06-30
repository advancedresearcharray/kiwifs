import * as React from "react";
import { cn } from "@kw/lib/cn";

type PanelGroupContextValue = {
  direction: "horizontal" | "vertical";
  sizes: number[];
  setSizes: (sizes: number[]) => void;
};

const PanelGroupContext = React.createContext<PanelGroupContextValue | null>(null);

type ResizablePanelGroupProps = React.HTMLAttributes<HTMLDivElement> & {
  direction?: "horizontal" | "vertical";
  onLayout?: (sizes: number[]) => void;
};

export function ResizablePanelGroup({
  direction = "horizontal",
  onLayout,
  className,
  children,
  ...props
}: ResizablePanelGroupProps) {
  const childArray = React.Children.toArray(children);
  const panelCount = childArray.filter(
    (child) => React.isValidElement(child) && (child.type as { displayName?: string }).displayName === "ResizablePanel",
  ).length;
  const [sizes, setSizesState] = React.useState<number[]>(() =>
    Array.from({ length: Math.max(panelCount, 2) }, () => 100 / Math.max(panelCount, 2)),
  );

  const setSizes = React.useCallback(
    (next: number[]) => {
      setSizesState(next);
      onLayout?.(next);
    },
    [onLayout],
  );

  return (
    <PanelGroupContext.Provider value={{ direction, sizes, setSizes }}>
      <div
        className={cn(
          "flex h-full w-full",
          direction === "vertical" ? "flex-col" : "flex-row",
          className,
        )}
        {...props}
      >
        {children}
      </div>
    </PanelGroupContext.Provider>
  );
}

type ResizablePanelProps = React.HTMLAttributes<HTMLDivElement> & {
  defaultSize?: number;
  minSize?: number;
  index?: number;
};

export function ResizablePanel({
  className,
  defaultSize,
  minSize = 15,
  index = 0,
  style,
  ...props
}: ResizablePanelProps) {
  const ctx = React.useContext(PanelGroupContext);
  const size = ctx?.sizes[index] ?? defaultSize ?? 50;

  return (
    <div
      className={cn("min-h-0 min-w-0 overflow-hidden", className)}
      style={{
        flexBasis: 0,
        flexGrow: size,
        flexShrink: 1,
        minWidth: ctx?.direction === "horizontal" ? `${minSize}%` : undefined,
        minHeight: ctx?.direction === "vertical" ? `${minSize}%` : undefined,
        ...style,
      }}
      {...props}
    />
  );
}
ResizablePanel.displayName = "ResizablePanel";

type ResizableHandleProps = React.HTMLAttributes<HTMLDivElement> & {
  index?: number;
};

export function ResizableHandle({ className, index = 0, ...props }: ResizableHandleProps) {
  const ctx = React.useContext(PanelGroupContext);
  const dragging = React.useRef(false);

  const onPointerDown = (e: React.PointerEvent<HTMLDivElement>) => {
    if (!ctx) return;
    e.preventDefault();
    dragging.current = true;
    const start = ctx.direction === "horizontal" ? e.clientX : e.clientY;
    const startSizes = [...ctx.sizes];
    const group = e.currentTarget.parentElement;
    if (!group) return;

    const onMove = (ev: PointerEvent) => {
      if (!dragging.current) return;
      const rect = group.getBoundingClientRect();
      const total = ctx.direction === "horizontal" ? rect.width : rect.height;
      if (total <= 0) return;
      const delta = (ctx.direction === "horizontal" ? ev.clientX : ev.clientY) - start;
      const deltaPct = (delta / total) * 100;
      const left = Math.max(15, Math.min(85, startSizes[index] + deltaPct));
      const right = 100 - left;
      ctx.setSizes([left, right]);
    };

    const onUp = () => {
      dragging.current = false;
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerup", onUp);
    };

    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", onUp);
  };

  return (
    <div
      role="separator"
      aria-orientation={ctx?.direction === "vertical" ? "horizontal" : "vertical"}
      onPointerDown={onPointerDown}
      className={cn(
        "relative z-10 shrink-0 bg-border transition-colors hover:bg-primary/30",
        ctx?.direction === "vertical" ? "h-1.5 w-full cursor-row-resize" : "w-1.5 h-full cursor-col-resize",
        className,
      )}
      {...props}
    />
  );
}

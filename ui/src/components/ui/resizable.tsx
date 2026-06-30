/**
 * shadcn-compatible resizable panels (react-resizable-panels API surface).
 * Uses pointer-drag resize like the sidebar handle when the package is unavailable.
 */
import * as React from "react";
import { GripVertical } from "lucide-react";
import { cn } from "@kw/lib/cn";

type Direction = "horizontal" | "vertical";

type PanelGroupContextValue = {
  direction: Direction;
  sizes: number[];
  setSizes: (sizes: number[]) => void;
  registerPanel: (index: number, defaultSize: number) => void;
  panelCount: number;
};

const PanelGroupContext = React.createContext<PanelGroupContextValue | null>(null);

type PanelGroupProps = React.HTMLAttributes<HTMLDivElement> & {
  direction?: Direction;
  onLayout?: (sizes: number[]) => void;
};

function ResizablePanelGroup({
  className,
  direction = "horizontal",
  onLayout,
  children,
  ...props
}: PanelGroupProps) {
  const childArray = React.Children.toArray(children);
  const panelDefaults = React.useRef<number[]>([]);
  const [sizes, setSizesState] = React.useState<number[]>([]);

  const setSizes = React.useCallback(
    (next: number[]) => {
      setSizesState(next);
      onLayout?.(next);
    },
    [onLayout],
  );

  const registerPanel = React.useCallback((index: number, defaultSize: number) => {
    panelDefaults.current[index] = defaultSize;
    setSizesState((prev) => {
      if (prev.length >= panelDefaults.current.filter(Boolean).length) return prev;
      const total = panelDefaults.current.reduce((sum, v) => sum + (v ?? 0), 0);
      if (total <= 0) return prev;
      return panelDefaults.current.map((v) => ((v ?? 0) / total) * 100);
    });
  }, []);

  const value = React.useMemo(
    () => ({
      direction,
      sizes,
      setSizes,
      registerPanel,
      panelCount: childArray.filter((child) =>
        React.isValidElement(child) && (child.type as { displayName?: string }).displayName === "ResizablePanel",
      ).length,
    }),
    [direction, sizes, setSizes, registerPanel, childArray],
  );

  return (
    <PanelGroupContext.Provider value={value}>
      <div
        className={cn(
          "flex h-full w-full",
          direction === "vertical" && "flex-col",
          className,
        )}
        data-panel-group-direction={direction}
        {...props}
      >
        {children}
      </div>
    </PanelGroupContext.Provider>
  );
}

type PanelProps = React.HTMLAttributes<HTMLDivElement> & {
  defaultSize?: number;
  minSize?: number;
  index?: number;
};

function ResizablePanel({
  className,
  defaultSize = 50,
  minSize = 15,
  index = 0,
  style,
  ...props
}: PanelProps) {
  const ctx = React.useContext(PanelGroupContext);
  if (!ctx) throw new Error("ResizablePanel must be used within ResizablePanelGroup");

  React.useEffect(() => {
    ctx.registerPanel(index, defaultSize);
  }, [ctx, index, defaultSize]);

  const size = ctx.sizes[index] ?? defaultSize;
  const flexBasis = `${size}%`;

  return (
    <div
      className={cn("min-h-0 min-w-0 overflow-hidden", className)}
      style={{
        flexBasis,
        flexGrow: 0,
        flexShrink: 0,
        minWidth: ctx.direction === "horizontal" ? `${minSize}%` : undefined,
        minHeight: ctx.direction === "vertical" ? `${minSize}%` : undefined,
        ...style,
      }}
      data-panel-index={index}
      {...props}
    />
  );
}
ResizablePanel.displayName = "ResizablePanel";

type HandleProps = React.HTMLAttributes<HTMLDivElement> & {
  withHandle?: boolean;
  panelIndex?: number;
};

function ResizableHandle({
  className,
  withHandle,
  panelIndex = 0,
  ...props
}: HandleProps) {
  const ctx = React.useContext(PanelGroupContext);
  if (!ctx) throw new Error("ResizableHandle must be used within ResizablePanelGroup");

  const onMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    const group = (e.currentTarget as HTMLElement).parentElement;
    if (!group) return;
    const rect = group.getBoundingClientRect();
    const isHorizontal = ctx.direction === "horizontal";
    const start = isHorizontal ? e.clientX : e.clientY;
    const total = isHorizontal ? rect.width : rect.height;
    const startSizes = [...ctx.sizes];

    const onMove = (ev: MouseEvent) => {
      const delta = (isHorizontal ? ev.clientX : ev.clientY) - start;
      const deltaPercent = (delta / total) * 100;
      const left = Math.max(15, Math.min(85, (startSizes[panelIndex] ?? 50) + deltaPercent));
      const right = 100 - left;
      ctx.setSizes([left, right]);
    };

    const onUp = () => {
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
    };

    document.addEventListener("mousemove", onMove);
    document.addEventListener("mouseup", onUp);
  };

  return (
    <div
      role="separator"
      aria-orientation={ctx.direction === "horizontal" ? "vertical" : "horizontal"}
      className={cn(
        "relative flex shrink-0 items-center justify-center bg-border",
        ctx.direction === "horizontal" ? "w-px cursor-col-resize" : "h-px cursor-row-resize",
        "after:absolute after:bg-border/80",
        ctx.direction === "horizontal"
          ? "after:inset-y-0 after:left-1/2 after:w-1 after:-translate-x-1/2"
          : "after:inset-x-0 after:top-1/2 after:h-1 after:-translate-y-1/2",
        className,
      )}
      onMouseDown={onMouseDown}
      {...props}
    >
      {withHandle && (
        <div className="z-10 flex h-4 w-3 items-center justify-center rounded-sm border bg-border">
          <GripVertical className="h-2.5 w-2.5" />
        </div>
      )}
    </div>
  );
}

export { ResizablePanelGroup, ResizablePanel, ResizableHandle };

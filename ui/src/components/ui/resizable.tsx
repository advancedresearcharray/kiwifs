import * as React from "react";
import { GripVertical } from "lucide-react";
import { cn } from "@kw/lib/cn";

type PanelGroupContextValue = {
  direction: "horizontal" | "vertical";
  registerPanel: (id: string, defaultSize: number) => void;
  sizes: Record<string, number>;
  setSizes: React.Dispatch<React.SetStateAction<Record<string, number>>>;
  onLayout?: (sizes: number[]) => void;
};

const PanelGroupContext = React.createContext<PanelGroupContextValue | null>(null);

type ResizablePanelGroupProps = React.HTMLAttributes<HTMLDivElement> & {
  direction?: "horizontal" | "vertical";
  onLayout?: (sizes: number[]) => void;
};

export function ResizablePanelGroup({
  className,
  direction = "horizontal",
  onLayout,
  children,
  ...props
}: ResizablePanelGroupProps) {
  const [sizes, setSizes] = React.useState<Record<string, number>>({});
  const panelsRef = React.useRef<Map<string, number>>(new Map());

  const registerPanel = React.useCallback((id: string, defaultSize: number) => {
    if (!panelsRef.current.has(id)) {
      panelsRef.current.set(id, defaultSize);
      setSizes((prev) => (prev[id] != null ? prev : { ...prev, [id]: defaultSize }));
    }
  }, []);

  const value = React.useMemo(
    () => ({ direction, registerPanel, sizes, setSizes, onLayout }),
    [direction, registerPanel, sizes, onLayout],
  );

  return (
    <PanelGroupContext.Provider value={value}>
      <div
        className={cn(
          "flex h-full w-full",
          direction === "horizontal" ? "flex-row" : "flex-col",
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
  id: string;
  defaultSize?: number;
  minSize?: number;
};

export function ResizablePanel({
  id,
  defaultSize = 50,
  minSize = 20,
  className,
  children,
  style,
  ...props
}: ResizablePanelProps) {
  const ctx = React.useContext(PanelGroupContext);
  if (!ctx) throw new Error("ResizablePanel must be used within ResizablePanelGroup");

  React.useEffect(() => {
    ctx.registerPanel(id, defaultSize);
  }, [ctx, id, defaultSize]);

  const size = ctx.sizes[id] ?? defaultSize;

  return (
    <div
      className={cn("min-h-0 min-w-0 overflow-hidden", className)}
      style={{
        flexBasis: 0,
        flexGrow: size,
        flexShrink: 1,
        minWidth: ctx.direction === "horizontal" ? `${minSize}%` : undefined,
        minHeight: ctx.direction === "vertical" ? `${minSize}%` : undefined,
        ...style,
      }}
      data-panel-id={id}
      {...props}
    >
      {children}
    </div>
  );
}

type ResizableHandleProps = React.HTMLAttributes<HTMLDivElement> & {
  withHandle?: boolean;
  panelIds: [string, string];
};

export function ResizableHandle({
  className,
  withHandle,
  panelIds,
  ...props
}: ResizableHandleProps) {
  const ctx = React.useContext(PanelGroupContext);
  const dragging = React.useRef(false);

  if (!ctx) throw new Error("ResizableHandle must be used within ResizablePanelGroup");

  const onMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    dragging.current = true;
    const [leftId, rightId] = panelIds;
    const startX = e.clientX;
    const startY = e.clientY;
    const startLeft = ctx.sizes[leftId] ?? 50;
    const startRight = ctx.sizes[rightId] ?? 50;
    const container = (e.currentTarget as HTMLElement).parentElement;
    if (!container) return;

    const onMove = (ev: MouseEvent) => {
      if (!dragging.current) return;
      const rect = container.getBoundingClientRect();
      let nextLeft = startLeft;
      if (ctx.direction === "horizontal") {
        const deltaPct = ((ev.clientX - startX) / rect.width) * 100;
        nextLeft = Math.max(20, Math.min(80, startLeft + deltaPct));
      } else {
        const deltaPct = ((ev.clientY - startY) / rect.height) * 100;
        nextLeft = Math.max(20, Math.min(80, startLeft + deltaPct));
      }
      const nextRight = startLeft + startRight - nextLeft;
      ctx.setSizes((prev) => ({ ...prev, [leftId]: nextLeft, [rightId]: nextRight }));
    };

    const onUp = () => {
      dragging.current = false;
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mouseup", onUp);
      ctx.setSizes((prev) => {
        const left = prev[leftId] ?? startLeft;
        const right = prev[rightId] ?? startRight;
        ctx.onLayout?.([left, right]);
        return prev;
      });
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
        className,
      )}
      onMouseDown={onMouseDown}
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

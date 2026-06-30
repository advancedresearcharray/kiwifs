import { useCallback, useRef, type ReactNode } from "react";
import { X } from "lucide-react";
import { Button } from "@kw/components/ui/button";
import { clampPaneSize } from "@kw/lib/splitView";

type Props = {
  leftSize: number;
  onLeftSizeChange: (size: number) => void;
  left: ReactNode;
  right: ReactNode;
  onCloseRight: () => void;
};

export function SplitPageLayout({ leftSize, onLeftSizeChange, left, right, onCloseRight }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const dragging = useRef(false);

  const onResizeStart = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      dragging.current = true;
      const container = containerRef.current;
      if (!container) return;
      const rect = container.getBoundingClientRect();
      const startX = e.clientX;
      const startSize = leftSize;

      const onMove = (ev: MouseEvent) => {
        const delta = ev.clientX - startX;
        const next = clampPaneSize(startSize + (delta / rect.width) * 100);
        onLeftSizeChange(next);
      };

      const onUp = () => {
        dragging.current = false;
        document.removeEventListener("mousemove", onMove);
        document.removeEventListener("mouseup", onUp);
      };

      document.addEventListener("mousemove", onMove);
      document.addEventListener("mouseup", onUp);
    },
    [leftSize, onLeftSizeChange],
  );

  return (
    <div ref={containerRef} className="flex h-full min-h-0 w-full">
      <div
        className="min-w-0 overflow-auto kiwi-scroll border-r border-border"
        style={{ width: `${leftSize}%` }}
      >
        {left}
      </div>
      <div
        role="separator"
        aria-orientation="vertical"
        aria-label="Resize split panes"
        className="kiwi-resize-handle w-1 cursor-col-resize hover:bg-primary/30 active:bg-primary/50 transition-colors shrink-0"
        onMouseDown={onResizeStart}
      />
      <div className="relative min-w-0 flex-1 overflow-auto kiwi-scroll">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="absolute top-2 right-2 z-20 h-7 w-7 bg-background/80 backdrop-blur-sm border border-border shadow-sm"
          aria-label="Close split view"
          onClick={onCloseRight}
        >
          <X className="h-3.5 w-3.5" />
        </Button>
        {right}
      </div>
    </div>
  );
}

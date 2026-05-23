import { memo, useCallback } from "react";
import { Handle, Position, NodeResizer, type NodeProps } from "@xyflow/react";
import { ExternalLink } from "lucide-react";

type Data = { url?: string; color?: string };

function CanvasLinkNode({ data, selected }: NodeProps) {
  const { url, color } = data as Data;
  let display = url ?? "link";
  try {
    display = new URL(url ?? "").hostname;
  } catch { /* keep raw */ }

  const handleDoubleClick = useCallback(
    (e: React.MouseEvent) => {
      e.stopPropagation();
      if (url) window.open(url, "_blank", "noopener,noreferrer");
    },
    [url],
  );

  return (
    <>
      <NodeResizer
        minWidth={140}
        minHeight={50}
        isVisible={!!selected}
        lineClassName="!border-primary/40"
        handleClassName="!w-2 !h-2 !bg-primary !border-primary"
      />
      <div
        className="rounded-lg border border-border bg-green-50 dark:bg-green-950/30 shadow-sm px-3 py-2 min-w-[140px] h-full cursor-pointer"
        style={color ? { borderLeftColor: color, borderLeftWidth: 3 } : undefined}
        onDoubleClick={handleDoubleClick}
      >
        <Handle type="target" position={Position.Top} id="top-target" className="!bg-muted-foreground !w-2 !h-2" />
        <Handle type="target" position={Position.Left} id="left-target" className="!bg-muted-foreground !w-2 !h-2" />
        <div className="flex items-center gap-1.5">
          <ExternalLink className="h-3.5 w-3.5 text-green-600 shrink-0" />
          <span className="text-sm font-medium truncate">{display}</span>
        </div>
        {url && <div className="text-[10px] text-muted-foreground mt-0.5 truncate">{url}</div>}
        <Handle type="source" position={Position.Bottom} id="bottom-source" className="!bg-muted-foreground !w-2 !h-2" />
        <Handle type="source" position={Position.Right} id="right-source" className="!bg-muted-foreground !w-2 !h-2" />
      </div>
    </>
  );
}

export default memo(CanvasLinkNode);

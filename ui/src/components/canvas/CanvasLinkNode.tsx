import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import { ExternalLink } from "lucide-react";

type Data = { url?: string; color?: string };

function CanvasLinkNode({ data }: NodeProps) {
  const { url, color } = data as Data;
  let display = url ?? "link";
  try {
    display = new URL(url ?? "").hostname;
  } catch { /* keep raw */ }
  return (
    <div
      className="rounded-lg border border-border bg-green-50 dark:bg-green-950/30 shadow-sm px-3 py-2 min-w-[140px] max-w-[320px]"
      style={color ? { borderLeftColor: color, borderLeftWidth: 3 } : undefined}
    >
      <Handle type="target" position={Position.Top} className="!bg-muted-foreground !w-2 !h-2" />
      <div className="flex items-center gap-1.5">
        <ExternalLink className="h-3.5 w-3.5 text-green-600 shrink-0" />
        <span className="text-sm font-medium truncate">{display}</span>
      </div>
      {url && <div className="text-[10px] text-muted-foreground mt-0.5 truncate">{url}</div>}
      <Handle type="source" position={Position.Bottom} className="!bg-muted-foreground !w-2 !h-2" />
    </div>
  );
}

export default memo(CanvasLinkNode);

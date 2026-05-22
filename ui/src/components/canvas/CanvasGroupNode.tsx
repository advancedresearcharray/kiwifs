import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import { Group } from "lucide-react";

type Data = { text?: string; color?: string };

function CanvasGroupNode({ data }: NodeProps) {
  const { text, color } = data as Data;
  return (
    <div
      className="rounded-lg border-2 border-dashed border-border bg-muted/30 px-4 py-3 min-w-[200px] min-h-[120px]"
      style={color ? { borderColor: color } : undefined}
    >
      <Handle type="target" position={Position.Top} className="!bg-muted-foreground !w-2 !h-2" />
      <div className="flex items-center gap-1.5 mb-2">
        <Group className="h-3.5 w-3.5 text-muted-foreground" />
        <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">{text || "Group"}</span>
      </div>
      <Handle type="source" position={Position.Bottom} className="!bg-muted-foreground !w-2 !h-2" />
    </div>
  );
}

export default memo(CanvasGroupNode);

import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";

type Data = { text?: string; color?: string };

function CanvasTextNode({ data }: NodeProps) {
  const { text, color } = data as Data;
  return (
    <div
      className="rounded-lg border border-border bg-card shadow-sm px-3 py-2 min-w-[120px] max-w-[320px]"
      style={color ? { borderLeftColor: color, borderLeftWidth: 3 } : undefined}
    >
      <Handle type="target" position={Position.Top} className="!bg-muted-foreground !w-2 !h-2" />
      <div className="text-xs text-muted-foreground mb-1 font-medium uppercase tracking-wider">text</div>
      <div className="text-sm whitespace-pre-wrap break-words">{text || "..."}</div>
      <Handle type="source" position={Position.Bottom} className="!bg-muted-foreground !w-2 !h-2" />
    </div>
  );
}

export default memo(CanvasTextNode);

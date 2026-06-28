import { memo } from "react";
import { Handle, Position, NodeResizer, type NodeProps } from "@xyflow/react";

type Data = { text?: string; color?: string };

function CanvasGroupNode({ data, selected }: NodeProps) {
  const { text, color } = data as Data;
  return (
    <>
      <NodeResizer
        minWidth={200}
        minHeight={120}
        isVisible={!!selected}
        lineClassName="!border-primary/40"
        handleClassName="!w-2 !h-2 !bg-primary !border-primary"
      />
      <div
        className="rounded-lg border-2 border-dashed border-border bg-muted/30 px-4 py-3 min-w-[200px] min-h-[120px] h-full"
        style={color ? { borderColor: color } : undefined}
      >
        <Handle type="target" position={Position.Top} id="top-target" className="!bg-muted-foreground !w-2 !h-2" />
        <Handle type="target" position={Position.Left} id="left-target" className="!bg-muted-foreground !w-2 !h-2" />
        <div className="text-xs font-semibold text-muted-foreground">
          {text || "Group"}
        </div>
        <Handle type="source" position={Position.Bottom} id="bottom-source" className="!bg-muted-foreground !w-2 !h-2" />
        <Handle type="source" position={Position.Right} id="right-source" className="!bg-muted-foreground !w-2 !h-2" />
      </div>
    </>
  );
}

export default memo(CanvasGroupNode);

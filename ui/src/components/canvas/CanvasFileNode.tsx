import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import { FileText } from "lucide-react";

type Data = { file?: string; color?: string };

function CanvasFileNode({ data }: NodeProps) {
  const { file, color } = data as Data;
  const display = file?.replace(/\.md$/, "").split("/").pop() ?? "untitled";
  return (
    <div
      className="rounded-lg border border-border bg-blue-50 dark:bg-blue-950/30 shadow-sm px-3 py-2 min-w-[140px] max-w-[320px]"
      style={color ? { borderLeftColor: color, borderLeftWidth: 3 } : undefined}
    >
      <Handle type="target" position={Position.Top} className="!bg-muted-foreground !w-2 !h-2" />
      <div className="flex items-center gap-1.5">
        <FileText className="h-3.5 w-3.5 text-blue-500 shrink-0" />
        <span className="text-sm font-medium truncate">{display}</span>
      </div>
      {file && <div className="text-[10px] text-muted-foreground mt-0.5 truncate">{file}</div>}
      <Handle type="source" position={Position.Bottom} className="!bg-muted-foreground !w-2 !h-2" />
    </div>
  );
}

export default memo(CanvasFileNode);

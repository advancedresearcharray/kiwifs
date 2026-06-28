import { memo, useCallback } from "react";
import { Handle, Position, NodeResizer, type NodeProps } from "@xyflow/react";
import { FileText } from "lucide-react";

type Data = {
  file?: string;
  color?: string;
  onNavigate?: (path: string) => void;
};

function CanvasFileNode({ data, selected }: NodeProps) {
  const { file, color, onNavigate } = data as Data;
  const display = file?.replace(/\.md$/, "").split("/").pop() ?? "untitled";

  const handleClick = useCallback(
    (e: React.MouseEvent) => {
      if (e.detail === 2 && file && onNavigate) {
        // Double-click navigates to the file
        e.stopPropagation();
        onNavigate(file);
      }
    },
    [file, onNavigate],
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
        className="rounded-lg border border-border bg-blue-50 dark:bg-blue-950/30 shadow-sm px-3 py-2 min-w-[140px] h-full cursor-pointer"
        style={color ? { borderLeftColor: color, borderLeftWidth: 3 } : undefined}
        onClick={handleClick}
      >
        <Handle type="target" position={Position.Top} id="top-target" className="!bg-muted-foreground !w-2 !h-2" />
        <Handle type="target" position={Position.Left} id="left-target" className="!bg-muted-foreground !w-2 !h-2" />
        <div className="flex items-center gap-1.5">
          <FileText className="h-3.5 w-3.5 text-blue-500 shrink-0" />
          <span className="text-sm font-medium truncate">{display}</span>
        </div>
        {file && <div className="text-[10px] text-muted-foreground mt-0.5 truncate">{file}</div>}
        <Handle type="source" position={Position.Bottom} id="bottom-source" className="!bg-muted-foreground !w-2 !h-2" />
        <Handle type="source" position={Position.Right} id="right-source" className="!bg-muted-foreground !w-2 !h-2" />
      </div>
    </>
  );
}

export default memo(CanvasFileNode);

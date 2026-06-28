import { memo, useCallback, useEffect, useRef, useState } from "react";
import { Handle, Position, NodeResizer, type NodeProps } from "@xyflow/react";

type Data = {
  text?: string;
  color?: string;
  onTextChange?: (id: string, text: string) => void;
};

function CanvasTextNode({ id, data, selected }: NodeProps) {
  const { text, color, onTextChange } = data as Data;
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(text ?? "");
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (editing && textareaRef.current) {
      textareaRef.current.focus();
      textareaRef.current.select();
    }
  }, [editing]);

  // Sync external text changes when not editing
  useEffect(() => {
    if (!editing) setDraft(text ?? "");
  }, [text, editing]);

  const commitEdit = useCallback(() => {
    setEditing(false);
    if (draft !== (text ?? "") && onTextChange) {
      onTextChange(id, draft);
    }
  }, [id, draft, text, onTextChange]);

  const handleDoubleClick = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    setEditing(true);
  }, []);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Escape") {
        setDraft(text ?? "");
        setEditing(false);
      } else if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        commitEdit();
      }
      // Stop propagation so React Flow doesn't capture the key
      e.stopPropagation();
    },
    [text, commitEdit],
  );

  const colorStyle = color
    ? { borderLeftColor: color, borderLeftWidth: 3 }
    : undefined;

  return (
    <>
      <NodeResizer
        minWidth={120}
        minHeight={60}
        isVisible={!!selected}
        lineClassName="!border-primary/40"
        handleClassName="!w-2 !h-2 !bg-primary !border-primary"
      />
      <div
        className="rounded-lg border border-border bg-card shadow-sm px-3 py-2 min-w-[120px] h-full"
        style={colorStyle}
        onDoubleClick={handleDoubleClick}
      >
        <Handle type="target" position={Position.Top} id="top-target" className="!bg-muted-foreground !w-2 !h-2" />
        <Handle type="target" position={Position.Left} id="left-target" className="!bg-muted-foreground !w-2 !h-2" />
        {editing ? (
          <textarea
            ref={textareaRef}
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onBlur={commitEdit}
            onKeyDown={handleKeyDown}
            className="w-full h-full resize-none bg-transparent text-sm border-none outline-none p-0 m-0"
            style={{ minHeight: 40 }}
          />
        ) : (
          <div className="text-sm whitespace-pre-wrap break-words">
            {text || (
              <span className="text-muted-foreground italic">Double-click to edit</span>
            )}
          </div>
        )}
        <Handle type="source" position={Position.Bottom} id="bottom-source" className="!bg-muted-foreground !w-2 !h-2" />
        <Handle type="source" position={Position.Right} id="right-source" className="!bg-muted-foreground !w-2 !h-2" />
      </div>
    </>
  );
}

export default memo(CanvasTextNode);

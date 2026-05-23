import { useRef, type ReactNode } from "react";
import type { NodeApi } from "react-arborist";
import { cn } from "@kw/lib/cn";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@kw/components/ui/tooltip";
import { useTreeRevealTargetFocus } from "@kw/hooks/useTreeReveal";
import type { TreeRevealRequest } from "@kw/lib/treeReveal";
import type { FlatNode } from "@kw/lib/treeTransform";

export const TREE_INDENT = 14;
const ROW_PADDING_X = 8;

export function IndentGuides({ level, indent }: { level: number; indent: number }) {
  if (level <= 0) return null;
  return (
    <>
      {Array.from({ length: level }, (_, i) => (
        <span
          key={i}
          aria-hidden
          className="pointer-events-none absolute top-0 bottom-0 w-px bg-border/40"
          style={{ left: (i + 1) * indent - Math.floor(indent / 2) }}
        />
      ))}
    </>
  );
}

function treeRowBgClass(opts: {
  isActive: boolean;
  isSelected: boolean;
  osDropHighlight: string;
}): string {
  return cn(
    "pointer-events-none absolute inset-0 transition-colors",
    !opts.isActive && !opts.osDropHighlight && "group-hover:bg-accent/50",
    opts.isActive && "bg-accent",
    opts.isSelected && !opts.isActive && "bg-accent/60",
    opts.osDropHighlight,
  );
}

function treeRowContentClass(opts: {
  isActive: boolean;
  isExcluded: boolean;
  isDragging?: boolean;
  osDropHighlight: string;
}): string {
  return cn(
    "relative z-[1] flex h-full w-full min-w-0 items-center gap-1.5 pr-2 text-left",
    "text-foreground/90 group-hover:text-accent-foreground",
    opts.isActive && "text-accent-foreground font-medium",
    opts.osDropHighlight && "text-accent-foreground",
    opts.isExcluded && "opacity-40",
    opts.isDragging && "opacity-50",
  );
}

type TreeRowShellProps = {
  node: NodeApi<FlatNode>;
  revealRequest?: TreeRevealRequest | null;
  isActive: boolean;
  osDropHighlight: string;
  className?: string;
  onDragOver?: (e: React.DragEvent) => void;
  onDrop?: (e: React.DragEvent) => void;
  onClick?: (e: React.MouseEvent) => void;
  children: ReactNode;
};

export function TreeRowShell({
  node,
  revealRequest,
  isActive,
  osDropHighlight,
  className,
  onDragOver,
  onDrop,
  onClick,
  children,
}: TreeRowShellProps) {
  const rowRef = useRef<HTMLDivElement>(null);
  useTreeRevealTargetFocus(revealRequest, node.id, rowRef);
  const contentPaddingLeft = ROW_PADDING_X + node.level * TREE_INDENT;

  return (
    <Tooltip delayDuration={400}>
      <TooltipTrigger asChild>
        <div
          ref={rowRef}
          data-level={node.level}
          data-row-path={node.id}
          className={cn(
            "kiwi-tree-row group relative h-full w-full min-w-0 cursor-pointer",
            node.isFocused && "ring-1 ring-inset ring-ring z-[1]",
            className,
          )}
          onClick={onClick}
          onDragOver={onDragOver}
          onDrop={onDrop}
        >
          <IndentGuides level={node.level} indent={TREE_INDENT} />
          <div
            aria-hidden
            className={treeRowBgClass({
              isActive,
              isSelected: node.isSelected,
              osDropHighlight,
            })}
          />
          <div
            className={treeRowContentClass({
              isActive,
              isExcluded: !!node.data.excluded,
              isDragging: node.isDragging,
              osDropHighlight,
            })}
            style={{ paddingLeft: contentPaddingLeft }}
          >
            {children}
          </div>
        </div>
      </TooltipTrigger>
      <TooltipContent side="right" className="font-mono text-xs max-w-sm">
        {node.id}
      </TooltipContent>
    </Tooltip>
  );
}

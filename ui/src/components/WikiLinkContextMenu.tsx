import type { ReactNode } from "react";
import { Columns2 } from "lucide-react";
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from "@kw/components/ui/context-menu";
import { useSplitView } from "@kw/contexts/SplitViewContext";

type Props = {
  pagePath: string;
  children: ReactNode;
  className?: string;
  onNavigate: (path: string) => void;
};

export function WikiLinkContextMenu({ pagePath, children, className, onNavigate }: Props) {
  const split = useSplitView();

  return (
    <ContextMenu>
      <ContextMenuTrigger asChild>
        <span className={className}>{children}</span>
      </ContextMenuTrigger>
      <ContextMenuContent>
        <ContextMenuItem onClick={() => onNavigate(pagePath)}>
          Open
        </ContextMenuItem>
        <ContextMenuItem onClick={() => split.openInSplit(pagePath, "right")}>
          <Columns2 className="h-3.5 w-3.5" />
          Open in Split View
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  );
}

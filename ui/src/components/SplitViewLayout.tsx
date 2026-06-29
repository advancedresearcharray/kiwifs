import { useCallback, useRef, useState } from "react";
import { X } from "lucide-react";
import { KiwiPage } from "@kw/components/KiwiPage";
import { KiwiEditor } from "@kw/components/KiwiEditor";
import { Button } from "@kw/components/ui/button";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@kw/components/ui/resizable";
import { useSplitView } from "@kw/contexts/SplitViewContext";
import type { TreeEntry } from "@kw/lib/api";
import type { SplitPane } from "@kw/lib/splitView";

type PaneEditing = { left: boolean; right: boolean };

type Props = {
  tree: TreeEntry | null;
  refreshKey: number;
  onPrimaryNavigate: (path: string) => void;
  onRevealInTree: (path: string) => void;
  onToggleStar: (path: string) => void;
  isStarred: (path: string) => boolean;
  onTogglePin: (path: string) => void;
  isPinned: (path: string) => boolean;
  onTreeRefresh: () => void;
  onPublishedChanged: () => void;
  onOpenHistory: (path: string) => void;
  onTagClick: (tag: string) => void;
};

export function SplitViewLayout({
  tree,
  refreshKey,
  onPrimaryNavigate,
  onRevealInTree,
  onToggleStar,
  isStarred,
  onTogglePin,
  isPinned,
  onTreeRefresh,
  onPublishedChanged,
  onOpenHistory,
  onTagClick,
}: Props) {
  const split = useSplitView();
  const [editing, setEditing] = useState<PaneEditing>({ left: false, right: false });
  const leftEditorRef = useRef<{ save: () => Promise<void>; toggleMode?: () => void } | null>(null);
  const rightEditorRef = useRef<{ save: () => Promise<void>; toggleMode?: () => void } | null>(null);

  const handleNavigate = useCallback(
    (pane: SplitPane, path: string) => {
      split.navigatePane(pane, path);
      if (pane === "left") onPrimaryNavigate(path);
    },
    [onPrimaryNavigate, split],
  );

  const renderPane = (pane: SplitPane) => {
    const path = pane === "left" ? split.leftPath : split.rightPath;
    const versionHash = pane === "left" ? split.leftVersionHash : split.rightVersionHash;
    const isEditing = pane === "left" ? editing.left : editing.right;
    const setPaneEditing = (value: boolean) =>
      setEditing((prev) => ({ ...prev, [pane]: value }));
    const editorRef = pane === "left" ? leftEditorRef : rightEditorRef;
    const navigate = (p: string) => handleNavigate(pane, p);
    const showClose = pane === "right";

    return (
      <div className="relative flex h-full min-h-0 flex-col">
        {showClose ? (
          <div className="absolute right-2 top-2 z-20">
            <Button
              type="button"
              variant="outline"
              size="icon"
              className="h-7 w-7 bg-background/90 backdrop-blur-sm"
              aria-label="Close split pane"
              onClick={() => split.closeSplit()}
            >
              <X className="h-3.5 w-3.5" />
            </Button>
          </div>
        ) : null}
        <div className="flex-1 min-h-0 overflow-auto kiwi-scroll">
          {isEditing ? (
            <KiwiEditor
              path={path}
              tree={tree}
              saveRef={editorRef}
              onClose={() => setPaneEditing(false)}
              onNavigate={navigate}
              onSaved={() => {
                setPaneEditing(false);
                onTreeRefresh();
              }}
            />
          ) : (
            <KiwiPage
              path={path}
              tree={tree}
              versionHash={versionHash}
              onNavigate={navigate}
              onEdit={() => {
                if (versionHash) return;
                setPaneEditing(true);
              }}
              onHistory={() => onOpenHistory(path)}
              onRevealInTree={() => onRevealInTree(path)}
              onToggleStar={() => onToggleStar(path)}
              isStarred={isStarred(path)}
              onTogglePin={() => onTogglePin(path)}
              isPinned={isPinned(path)}
              onDeleted={() => {
                if (pane === "left") onPrimaryNavigate("");
                else split.closeSplit();
                onTreeRefresh();
              }}
              onDuplicated={(newPath) => {
                onTreeRefresh();
                navigate(newPath);
              }}
              onMoved={(newPath) => {
                onTreeRefresh();
                navigate(newPath);
              }}
              onTagClick={onTagClick}
              refreshKey={refreshKey}
              onPublishedChanged={onPublishedChanged}
            />
          )}
        </div>
      </div>
    );
  };

  return (
    <ResizablePanelGroup
      direction="horizontal"
      className="h-full"
      onLayout={(sizes) => {
        if (sizes[0] != null) split.setPaneSizes(sizes[0]);
      }}
    >
      <ResizablePanel id="left" defaultSize={split.leftSize} minSize={20}>
        {renderPane("left")}
      </ResizablePanel>
      <ResizableHandle withHandle panelIds={["left", "right"]} />
      <ResizablePanel id="right" defaultSize={split.rightSize} minSize={20}>
        {renderPane("right")}
      </ResizablePanel>
    </ResizablePanelGroup>
  );
}

export function SplitViewMobileNotice({ onDismiss }: { onDismiss: () => void }) {
  return (
    <div className="flex h-full items-center justify-center p-8">
      <div className="max-w-sm rounded-lg border border-border bg-card p-6 text-center shadow-sm">
        <h2 className="text-lg font-semibold">Split view unavailable</h2>
        <p className="mt-2 text-sm text-muted-foreground">
          Side-by-side page view is not available on mobile viewports. Use a wider screen to compare pages.
        </p>
        <Button type="button" className="mt-4" variant="outline" onClick={onDismiss}>
          Dismiss
        </Button>
      </div>
    </div>
  );
}

import { useCallback } from "react";
import { X } from "lucide-react";
import type { TreeEntry } from "@kw/lib/api";
import type { EditorMode } from "@kw/lib/editorMode";
import type { SplitViewState } from "@kw/lib/splitView";
import { clampPaneSize, saveSplitViewState } from "@kw/lib/splitView";
import { KiwiPage } from "./KiwiPage";
import { KiwiEditor } from "./KiwiEditor";
import { Button } from "./ui/button";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "./resizable";

export type SplitPagePaneProps = {
  tree: TreeEntry | null;
  refreshKey: number;
  onToggleStar: (path: string) => void;
  isStarred: (path: string) => boolean;
  onTogglePin: (path: string) => void;
  isPinned: (path: string) => boolean;
  onPublishedChanged?: () => void;
  onRevealInTree: (path: string) => void;
  onHistory: (path: string) => void;
  onRefresh: () => void;
};

type Props = SplitPagePaneProps & {
  split: SplitViewState;
  editingPath: string | null;
  onSplitChange: (next: SplitViewState | null) => void;
  onNavigatePane: (pane: "left" | "right", path: string) => void;
  onEditPane: (path: string) => void;
  onCloseEdit: () => void;
  editorRef: React.RefObject<{ save: () => Promise<void>; toggleMode?: () => void } | null>;
  editorModePref?: "editor" | "source";
  onEditorModeChange?: (mode: EditorMode) => void;
};

export function KiwiSplitPageView({
  split,
  tree,
  refreshKey,
  editingPath,
  onSplitChange,
  onNavigatePane,
  onEditPane,
  onCloseEdit,
  onToggleStar,
  isStarred,
  onTogglePin,
  isPinned,
  onPublishedChanged,
  onRevealInTree,
  onHistory,
  onRefresh,
  editorRef,
  editorModePref,
  onEditorModeChange,
}: Props) {
  const leftSize = split.leftSize;
  const rightSize = 100 - leftSize;

  const handleLayout = useCallback(
    (sizes: number[]) => {
      const nextLeft = clampPaneSize(sizes[0] ?? leftSize);
      if (nextLeft === split.leftSize) return;
      const next: SplitViewState = { ...split, leftSize: nextLeft };
      saveSplitViewState(next);
      onSplitChange(next);
    },
    [split, leftSize, onSplitChange],
  );

  const closeSecondary = () => {
    saveSplitViewState(null);
    onSplitChange(null);
  };

  const renderPane = (
    pane: "left" | "right",
    path: string,
    versionHash: string | null,
    showClose: boolean,
  ) => {
    if (editingPath === path && !versionHash) {
      return (
        <KiwiEditor
          path={path}
          tree={tree}
          saveRef={editorRef}
          editorModePref={editorModePref}
          onEditorModeChange={onEditorModeChange}
          onClose={onCloseEdit}
          onNavigate={(p) => onNavigatePane(pane, p)}
          onSaved={() => {
            onCloseEdit();
            onRefresh();
          }}
        />
      );
    }

    return (
      <div className="flex flex-col h-full relative">
        {showClose && (
          <div className="absolute top-2 right-2 z-20">
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7 bg-background/80 backdrop-blur-sm border border-border shadow-sm"
              aria-label="Close split pane"
              onClick={closeSecondary}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        )}
        <KiwiPage
          path={path}
          tree={tree}
          versionHash={versionHash ?? undefined}
          onNavigate={(p) => onNavigatePane(pane, p)}
          onEdit={() => onEditPane(path)}
          onHistory={() => onHistory(path)}
          onRevealInTree={() => onRevealInTree(path)}
          onToggleStar={() => onToggleStar(path)}
          isStarred={isStarred(path)}
          onTogglePin={() => onTogglePin(path)}
          isPinned={isPinned(path)}
          onDeleted={() => {
            if (pane === "left") onSplitChange(null);
            else closeSecondary();
            onRefresh();
          }}
          onDuplicated={(p) => {
            onRefresh();
            onNavigatePane(pane, p);
          }}
          onMoved={(p) => {
            onRefresh();
            onNavigatePane(pane, p);
          }}
          onTagClick={() => {}}
          refreshKey={refreshKey}
          onPublishedChanged={onPublishedChanged}
          onOpenInSplitView={(p) => {
            const next: SplitViewState = {
              ...split,
              right: { path: p, versionHash: null },
            };
            saveSplitViewState(next);
            onSplitChange(next);
          }}
        />
      </div>
    );
  };

  return (
    <ResizablePanelGroup direction="horizontal" onLayout={handleLayout} className="h-full">
      <ResizablePanel index={0} defaultSize={leftSize} minSize={20} className="h-full">
        {renderPane("left", split.left.path, split.left.versionHash, false)}
      </ResizablePanel>
      <ResizableHandle withHandle index={0} className="kiwi-resize-handle hover:bg-primary/30 active:bg-primary/50" />
      <ResizablePanel index={1} defaultSize={rightSize} minSize={20} className="h-full border-l border-border">
        {renderPane("right", split.right.path, split.right.versionHash, true)}
      </ResizablePanel>
    </ResizablePanelGroup>
  );
}

export function SplitViewMobileBlocked() {
  return (
    <div className="flex h-full items-center justify-center px-6 text-center">
      <div className="max-w-sm space-y-2">
        <p className="text-base font-medium text-foreground">Split view unavailable</p>
        <p className="text-sm text-muted-foreground">
          Side-by-side page view is not available on mobile viewports. Widen the window or use a desktop browser.
        </p>
      </div>
    </div>
  );
}

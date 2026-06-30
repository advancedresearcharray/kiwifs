import { useCallback, useRef, useState } from "react";
import { KiwiPage } from "./KiwiPage";
import { KiwiEditor } from "./KiwiEditor";
import { Button } from "@kw/components/ui/button";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@kw/components/ui/resizable";
import { useSplitView } from "@kw/contexts/SplitViewProvider";
import type { TreeEntry } from "@kw/lib/api";
import type { SplitPaneSide } from "@kw/lib/splitView";
import type { EditorMode } from "@kw/lib/editorMode";

export type KiwiSplitViewProps = {
  tree: TreeEntry | null;
  refreshKey: number;
  leftPath: string;
  rightPath: string;
  rightVersionHash: string | null;
  sizes: [number, number];
  onLeftNavigate: (path: string) => void;
  onRightNavigate: (path: string) => void;
  onSyncActivePath: (path: string) => void;
  onHistory: (path: string) => void;
  onRevealInTree: (path: string) => void;
  onToggleStar: (path: string) => void;
  isStarred: (path: string) => boolean;
  onTogglePin: (path: string) => void;
  isPinned: (path: string) => boolean;
  onDeleted: (path: string) => void;
  onDuplicated: (path: string) => void;
  onMoved: (path: string) => void;
  onTagClick: (tag: string) => void;
  onPublishedChanged: () => void;
  onTreeRefresh: () => void;
  editorModePref?: "editor" | "source";
  onEditorModeChange?: (mode: EditorMode) => void;
};

export function KiwiSplitView({
  tree,
  refreshKey,
  leftPath,
  rightPath,
  rightVersionHash,
  sizes,
  onLeftNavigate,
  onRightNavigate,
  onSyncActivePath,
  onHistory,
  onRevealInTree,
  onToggleStar,
  isStarred,
  onTogglePin,
  isPinned,
  onDeleted,
  onDuplicated,
  onMoved,
  onTagClick,
  onPublishedChanged,
  onTreeRefresh,
  editorModePref,
  onEditorModeChange,
}: KiwiSplitViewProps) {
  const { closeSplit, setSizes, openInSplit } = useSplitView();
  const [editingPane, setEditingPane] = useState<SplitPaneSide | null>(null);
  const leftSaveRef = useRef<{ save: () => Promise<void>; toggleMode?: () => void } | null>(null);
  const rightSaveRef = useRef<{ save: () => Promise<void>; toggleMode?: () => void } | null>(null);

  const handleOpenInSplit = useCallback(
    (targetPath: string, sourcePane: SplitPaneSide) => {
      if (sourcePane === "left") {
        openInSplit(leftPath, targetPath);
      } else {
        openInSplit(rightPath, targetPath);
      }
    },
    [leftPath, rightPath, openInSplit],
  );

  const renderPane = (side: SplitPaneSide) => {
    const path = side === "left" ? leftPath : rightPath;
    const versionHash = side === "right" ? rightVersionHash : null;
    const onNavigate = side === "left" ? onLeftNavigate : onRightNavigate;
    const saveRef = side === "left" ? leftSaveRef : rightSaveRef;
    const isSecondary = side === "right";

    if (editingPane === side && !versionHash) {
      return (
        <KiwiEditor
          path={path}
          tree={tree}
          saveRef={saveRef}
          editorModePref={editorModePref}
          onEditorModeChange={onEditorModeChange}
          onClose={() => setEditingPane(null)}
          onNavigate={onNavigate}
          onSaved={() => {
            setEditingPane(null);
            onTreeRefresh();
          }}
        />
      );
    }

    return (
      <KiwiPage
        path={path}
        tree={tree}
        versionHash={versionHash}
        onNavigate={(p) => {
          onNavigate(p);
          onSyncActivePath(side === "left" ? p : leftPath);
        }}
        onEdit={() => setEditingPane(side)}
        onHistory={() => onHistory(path)}
        onRevealInTree={() => onRevealInTree(path)}
        onToggleStar={() => onToggleStar(path)}
        isStarred={isStarred(path)}
        onTogglePin={() => onTogglePin(path)}
        isPinned={isPinned(path)}
        onDeleted={() => onDeleted(path)}
        onDuplicated={(p) => onDuplicated(p)}
        onMoved={(p) => onMoved(p)}
        onTagClick={onTagClick}
        refreshKey={refreshKey}
        onPublishedChanged={onPublishedChanged}
        onOpenInSplit={(target) => handleOpenInSplit(target, side)}
        showClosePane={isSecondary}
        onClosePane={closeSplit}
        readOnly={Boolean(versionHash)}
      />
    );
  };

  return (
    <div className="flex h-full flex-col">
      <ResizablePanelGroup
        direction="horizontal"
        className="flex-1 min-h-0"
        defaultLayout={[sizes[0], sizes[1]]}
        onLayout={(next) => {
          if (next.length === 2) setSizes([next[0], next[1]]);
        }}
      >
        <ResizablePanel index={0} defaultSize={sizes[0]} className="border-r border-border">
          {renderPane("left")}
        </ResizablePanel>
        <ResizableHandle index={0} />
        <ResizablePanel index={1} defaultSize={sizes[1]} className="relative">
          {renderPane("right")}
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  );
}

export function SplitViewMobileNotice({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  if (!open) return null;
  return (
    <div className="absolute inset-0 z-30 flex items-center justify-center bg-background/90 p-6">
      <div className="max-w-sm rounded-lg border border-border bg-card p-6 text-center shadow-lg">
        <p className="text-sm text-muted-foreground">
          Split view is not available on mobile. Widen your window or use a desktop browser.
        </p>
        <Button className="mt-4" size="sm" onClick={onClose}>
          OK
        </Button>
      </div>
    </div>
  );
}

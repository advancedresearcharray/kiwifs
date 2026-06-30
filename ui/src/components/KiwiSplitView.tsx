import { X } from "lucide-react";
import type { TreeEntry } from "@kw/lib/api";
import { KiwiPage } from "./KiwiPage";
import { Button } from "@kw/components/ui/button";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@kw/components/ui/resizable";
import type { SplitPaneVersion, SplitViewState } from "@kw/lib/splitView";
import { splitViewHasSecondary } from "@kw/lib/splitView";

type PaneCallbacks = {
  onEdit: () => void;
  onHistory?: () => void;
  onRevealInTree?: () => void;
  onToggleStar?: () => void;
  isStarred?: boolean;
  onTogglePin?: () => void;
  isPinned?: boolean;
  onDeleted?: () => void;
  onDuplicated?: (newPath: string) => void;
  onMoved?: (newPath: string) => void;
  onTagClick?: (tag: string) => void;
  onPublishedChanged?: () => void;
};

type Props = {
  tree: TreeEntry | null;
  leftPath: string;
  splitView: SplitViewState;
  refreshKey: number;
  onLeftNavigate: (path: string) => void;
  onRightNavigate: (path: string) => void;
  onOpenInSplitView: (path: string) => void;
  onSizesChange: (sizes: [number, number]) => void;
  onCloseSecondary: () => void;
  leftPane: PaneCallbacks;
  rightPane: PaneCallbacks;
};

function SplitPaneHeader({
  label,
  onClose,
}: {
  label: string;
  onClose?: () => void;
}) {
  return (
    <div className="flex items-center justify-between gap-2 border-b border-border px-3 py-1.5 bg-muted/30 shrink-0">
      <span className="text-xs font-medium text-muted-foreground truncate">{label}</span>
      {onClose && (
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6 shrink-0"
          aria-label="Close secondary pane"
          onClick={onClose}
        >
          <X className="h-3.5 w-3.5" />
        </Button>
      )}
    </div>
  );
}

function SecondaryPlaceholder() {
  return (
    <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground text-center">
      Right-click a page or wiki-link and choose &ldquo;Open in Split View&rdquo;, or navigate from this pane.
    </div>
  );
}

export function KiwiSplitView({
  tree,
  leftPath,
  splitView,
  refreshKey,
  onLeftNavigate,
  onRightNavigate,
  onOpenInSplitView,
  onSizesChange,
  onCloseSecondary,
  leftPane,
  rightPane,
}: Props) {
  const hasSecondary = splitViewHasSecondary(splitView);
  const rightPath = splitView.rightPath;
  const rightVersion = splitView.rightVersion;

  const renderPage = (
    path: string,
    onNavigate: (p: string) => void,
    pane: PaneCallbacks,
    version?: SplitPaneVersion | null,
  ) => (
    <KiwiPage
      path={path}
      tree={tree}
      onNavigate={onNavigate}
      onOpenInSplitView={onOpenInSplitView}
      versionHash={version?.hash}
      readOnly={Boolean(version)}
      onEdit={pane.onEdit}
      onHistory={pane.onHistory}
      onRevealInTree={pane.onRevealInTree}
      onToggleStar={pane.onToggleStar}
      isStarred={pane.isStarred}
      onTogglePin={pane.onTogglePin}
      isPinned={pane.isPinned}
      onDeleted={pane.onDeleted}
      onDuplicated={pane.onDuplicated}
      onMoved={pane.onMoved}
      onTagClick={pane.onTagClick}
      refreshKey={refreshKey}
      onPublishedChanged={pane.onPublishedChanged}
    />
  );

  return (
    <ResizablePanelGroup
      direction="horizontal"
      className="h-full"
      onLayout={(sizes) => {
        if (sizes.length === 2) {
          onSizesChange([sizes[0], sizes[1]]);
        }
      }}
    >
      <ResizablePanel
        index={0}
        defaultSize={splitView.sizes[0]}
        className="flex flex-col min-h-0"
      >
        <SplitPaneHeader label="Primary" />
        <div className="flex-1 min-h-0 overflow-auto kiwi-scroll">
          {renderPage(leftPath, onLeftNavigate, leftPane)}
        </div>
      </ResizablePanel>

      <ResizableHandle withHandle />

      <ResizablePanel
        index={1}
        defaultSize={splitView.sizes[1]}
        className="flex flex-col min-h-0"
      >
        <SplitPaneHeader
          label={
            rightVersion
              ? `Version ${rightVersion.hash.slice(0, 7)}`
              : rightPath ?? "Secondary"
          }
          onClose={onCloseSecondary}
        />
        <div className="flex-1 min-h-0 overflow-auto kiwi-scroll">
          {!hasSecondary ? (
            <SecondaryPlaceholder />
          ) : rightVersion ? (
            renderPage(rightVersion.path, onRightNavigate, rightPane, rightVersion)
          ) : rightPath ? (
            renderPage(rightPath, onRightNavigate, rightPane)
          ) : (
            <SecondaryPlaceholder />
          )}
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  );
}

export function SplitViewMobileNotice({ onDismiss }: { onDismiss: () => void }) {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-4 p-8 text-center">
      <p className="text-sm text-muted-foreground max-w-sm">
        Split view is not available on mobile viewports. Widen the window or use a desktop browser.
      </p>
      <Button variant="outline" size="sm" onClick={onDismiss}>
        Dismiss
      </Button>
    </div>
  );
}

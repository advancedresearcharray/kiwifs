import type { ReactNode } from "react";
import { useSplitView } from "@kw/contexts/SplitViewContext";
import { SplitViewLayout, SplitViewMobileNotice } from "@kw/components/SplitViewLayout";

type Props = {
  activePath: string;
  tree: Parameters<typeof SplitViewLayout>[0]["tree"];
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
  singlePane: ReactNode;
};

export function SplitAwarePageContent({
  activePath,
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
  singlePane,
}: Props) {
  const split = useSplitView();

  if (split.mobileBlocked) {
    return <SplitViewMobileNotice onDismiss={split.dismissMobileNotice} />;
  }

  if (split.enabled && activePath) {
    return (
      <div className="h-full overflow-hidden">
        <SplitViewLayout
        tree={tree}
        refreshKey={refreshKey}
        onPrimaryNavigate={onPrimaryNavigate}
        onRevealInTree={onRevealInTree}
        onToggleStar={onToggleStar}
        isStarred={isStarred}
        onTogglePin={onTogglePin}
        isPinned={isPinned}
        onTreeRefresh={onTreeRefresh}
        onPublishedChanged={onPublishedChanged}
        onOpenHistory={onOpenHistory}
        onTagClick={onTagClick}
      />
      </div>
    );
  }

  return <>{singlePane}</>;
}

export function dispatchToggleSplitView(path?: string | null): void {
  window.dispatchEvent(new CustomEvent("kiwi:toggle-split-view", { detail: { path } }));
}

export function dispatchCompareWithCurrent(path: string, versionHash: string): void {
  window.dispatchEvent(
    new CustomEvent("kiwi:compare-with-current", { detail: { path, versionHash } }),
  );
}

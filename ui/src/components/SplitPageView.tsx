import { useState, type ReactNode } from "react";
import { KiwiEditor } from "@kw/components/KiwiEditor";
import { KiwiPage } from "@kw/components/KiwiPage";
import { SplitPageLayout } from "@kw/components/SplitPageLayout";
import { useSplitView } from "@kw/contexts/SplitViewContext";
import type { TreeEntry } from "@kw/lib/api";
import type { EditorMode } from "@kw/lib/editorMode";
import type { SplitPaneSpec } from "@kw/lib/splitView";

type PaneSide = "left" | "right";

type SharedPageProps = {
  tree: TreeEntry | null;
  refreshKey: number;
  onRevealInTree: () => void;
  onToggleStar: (path: string) => void;
  isStarred: (path: string) => boolean;
  onTogglePin: (path: string) => void;
  isPinned: (path: string) => boolean;
  onDeleted: () => void;
  onDuplicated: (path: string) => void;
  onMoved: (path: string) => void;
  onTagClick: (tag: string) => void;
  onPublishedChanged: () => void;
  onHistory: (path: string) => void;
  onOpenInSplit: (path: string) => void;
};

type Props = SharedPageProps & {
  editorRef: React.MutableRefObject<{ save: () => Promise<void>; toggleMode?: () => void } | null>;
  editorModePref?: "editor" | "source";
  onEditorModeChange?: (mode: EditorMode) => void;
  onSaved: () => void;
};

function renderPane(
  spec: SplitPaneSpec,
  side: PaneSide,
  editing: boolean,
  props: SharedPageProps & {
    onNavigate: (path: string) => void;
    onEdit: () => void;
    editorRef?: React.MutableRefObject<{ save: () => Promise<void>; toggleMode?: () => void } | null>;
    editorModePref?: "editor" | "source";
    onEditorModeChange?: (mode: EditorMode) => void;
    onSaved?: () => void;
    onCloseEditor?: () => void;
  },
): ReactNode {
  const readOnly = Boolean(spec.versionHash);

  if (editing && !readOnly && props.editorRef && props.onSaved && props.onCloseEditor) {
    return (
      <KiwiEditor
        path={spec.path}
        tree={props.tree}
        saveRef={props.editorRef}
        editorModePref={props.editorModePref}
        onEditorModeChange={props.onEditorModeChange}
        onClose={props.onCloseEditor}
        onNavigate={props.onNavigate}
        onSaved={props.onSaved}
      />
    );
  }

  return (
    <KiwiPage
      path={spec.path}
      versionHash={spec.versionHash ?? undefined}
      tree={props.tree}
      onNavigate={props.onNavigate}
      onEdit={readOnly ? undefined : props.onEdit}
      onHistory={() => props.onHistory(spec.path)}
      onRevealInTree={props.onRevealInTree}
      onToggleStar={() => props.onToggleStar(spec.path)}
      isStarred={props.isStarred(spec.path)}
      onTogglePin={() => props.onTogglePin(spec.path)}
      isPinned={props.isPinned(spec.path)}
      onDeleted={props.onDeleted}
      onDuplicated={props.onDuplicated}
      onMoved={props.onMoved}
      onTagClick={props.onTagClick}
      refreshKey={props.refreshKey}
      onPublishedChanged={props.onPublishedChanged}
      onOpenInSplit={props.onOpenInSplit}
      paneLabel={side === "left" ? "Primary" : spec.versionHash ? "Historical version" : "Secondary"}
    />
  );
}

export function SplitPageView({
  tree,
  refreshKey,
  editorRef,
  editorModePref,
  onEditorModeChange,
  onSaved,
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
  onHistory,
  onOpenInSplit,
}: Props) {
  const { state, navigateLeft, navigateRight, setLeftSize, closeSplit } = useSplitView();
  const [editingLeft, setEditingLeft] = useState(false);
  const [editingRight, setEditingRight] = useState(false);

  if (!state.left || !state.right) return null;

  const shared: SharedPageProps = {
    tree,
    refreshKey,
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
    onHistory,
    onOpenInSplit,
  };

  return (
    <SplitPageLayout
      leftSize={state.leftSize}
      onLeftSizeChange={setLeftSize}
      onCloseRight={closeSplit}
      left={renderPane(state.left, "left", editingLeft, {
        ...shared,
        onNavigate: (path) => {
          setEditingLeft(false);
          navigateLeft(path);
        },
        onEdit: () => setEditingLeft(true),
        editorRef,
        editorModePref,
        onEditorModeChange,
        onSaved: () => {
          setEditingLeft(false);
          onSaved();
        },
        onCloseEditor: () => setEditingLeft(false),
      })}
      right={renderPane(state.right, "right", editingRight, {
        ...shared,
        onNavigate: (path) => {
          setEditingRight(false);
          navigateRight(path);
        },
        onEdit: () => setEditingRight(true),
        onCloseEditor: () => setEditingRight(false),
      })}
    />
  );
}

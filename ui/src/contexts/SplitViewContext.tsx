import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import { getCurrentSpace } from "@kw/lib/api";
import {
  clampPaneSize,
  clearSplitViewState,
  defaultSplitViewState,
  isSplitViewMobileViewport,
  loadSplitViewState,
  saveSplitViewState,
  type SplitPane,
  type SplitViewPersisted,
} from "@kw/lib/splitView";

export type SplitViewContextValue = {
  enabled: boolean;
  leftPath: string;
  rightPath: string;
  leftSize: number;
  rightSize: number;
  leftVersionHash: string | null;
  rightVersionHash: string | null;
  mobileBlocked: boolean;
  dismissMobileNotice: () => void;
  openInSplit: (path: string, pane?: SplitPane) => void;
  compareWithCurrent: (path: string, versionHash: string) => void;
  closeSplit: () => void;
  toggleSplit: (secondaryPath?: string) => void;
  navigatePane: (pane: SplitPane, path: string) => void;
  setPaneSizes: (leftSize: number) => void;
  clearPaneVersion: (pane: SplitPane) => void;
};

const SplitViewContext = createContext<SplitViewContextValue | null>(null);

type Props = {
  primaryPath: string | null;
  children: ReactNode;
};

export function SplitViewProvider({ primaryPath, children }: Props) {
  const space = getCurrentSpace() || "default";
  const [mobileBlocked, setMobileBlocked] = useState(false);
  const [state, setState] = useState<SplitViewPersisted>(() =>
    loadSplitViewState(space, primaryPath ?? ""),
  );

  useEffect(() => {
    if (!primaryPath) return;
    setState((prev) => {
      const next = { ...prev, leftPath: prev.leftPath || primaryPath };
      if (!prev.enabled) next.leftPath = primaryPath;
      return next;
    });
  }, [primaryPath]);

  useEffect(() => {
    saveSplitViewState(space, state);
  }, [space, state]);

  const persist = useCallback((updater: (prev: SplitViewPersisted) => SplitViewPersisted) => {
    setState((prev) => updater(prev));
  }, []);

  const openInSplit = useCallback(
    (path: string, pane: SplitPane = "right") => {
      if (isSplitViewMobileViewport()) {
        setMobileBlocked(true);
        return;
      }
      setMobileBlocked(false);
      const left = primaryPath ?? state.leftPath ?? path;
      persist((prev) => ({
        ...prev,
        enabled: true,
        leftPath: pane === "left" ? path : left,
        rightPath: pane === "right" ? path : prev.rightPath || path,
        leftVersionHash: pane === "left" ? null : prev.leftVersionHash,
        rightVersionHash: pane === "right" ? null : prev.rightVersionHash,
      }));
    },
    [persist, primaryPath, state.leftPath],
  );

  const compareWithCurrent = useCallback(
    (path: string, versionHash: string) => {
      if (isSplitViewMobileViewport()) {
        setMobileBlocked(true);
        return;
      }
      setMobileBlocked(false);
      persist((prev) => ({
        ...prev,
        enabled: true,
        leftPath: path,
        rightPath: path,
        leftVersionHash: versionHash,
        rightVersionHash: null,
      }));
    },
    [persist],
  );

  const closeSplit = useCallback(() => {
    setMobileBlocked(false);
    persist((prev) => ({
      ...prev,
      enabled: false,
      leftVersionHash: null,
      rightVersionHash: null,
    }));
    clearSplitViewState(space);
  }, [persist, space]);

  const toggleSplit = useCallback(
    (secondaryPath?: string) => {
      if (state.enabled) {
        closeSplit();
        return;
      }
      if (isSplitViewMobileViewport()) {
        setMobileBlocked(true);
        return;
      }
      const left = primaryPath ?? state.leftPath;
      if (!left) return;
      const right = secondaryPath ?? state.rightPath ?? left;
      setMobileBlocked(false);
      persist((prev) => ({
        ...prev,
        enabled: true,
        leftPath: left,
        rightPath: right,
        leftVersionHash: null,
        rightVersionHash: null,
      }));
    },
    [closeSplit, persist, primaryPath, state.enabled, state.leftPath, state.rightPath],
  );

  const navigatePane = useCallback(
    (pane: SplitPane, path: string) => {
      persist((prev) => ({
        ...prev,
        [pane === "left" ? "leftPath" : "rightPath"]: path,
        [pane === "left" ? "leftVersionHash" : "rightVersionHash"]: null,
      }));
    },
    [persist],
  );

  const setPaneSizes = useCallback(
    (leftSize: number) => {
      const clamped = clampPaneSize(leftSize);
      persist((prev) => ({
        ...prev,
        leftSize: clamped,
        rightSize: 100 - clamped,
      }));
    },
    [persist],
  );

  const clearPaneVersion = useCallback(
    (pane: SplitPane) => {
      persist((prev) => ({
        ...prev,
        [pane === "left" ? "leftVersionHash" : "rightVersionHash"]: null,
      }));
    },
    [persist],
  );

  useEffect(() => {
    const onToggle = (event: Event) => {
      const detail = (event as CustomEvent<{ path?: string | null }>).detail;
      toggleSplit(detail?.path ?? undefined);
    };
    const onCompare = (event: Event) => {
      const detail = (event as CustomEvent<{ path?: string; versionHash?: string }>).detail;
      if (detail?.path && detail.versionHash) {
        compareWithCurrent(detail.path, detail.versionHash);
      }
    };
    const onOpenSplit = (event: Event) => {
      const detail = (event as CustomEvent<{ path?: string }>).detail;
      if (detail?.path) openInSplit(detail.path, "right");
    };
    window.addEventListener("kiwi:toggle-split-view", onToggle);
    window.addEventListener("kiwi:compare-with-current", onCompare);
    window.addEventListener("kiwi:open-split-view", onOpenSplit);
    return () => {
      window.removeEventListener("kiwi:toggle-split-view", onToggle);
      window.removeEventListener("kiwi:compare-with-current", onCompare);
      window.removeEventListener("kiwi:open-split-view", onOpenSplit);
    };
  }, [toggleSplit, compareWithCurrent, openInSplit]);

  const dismissMobileNotice = useCallback(() => setMobileBlocked(false), []);

  const value = useMemo<SplitViewContextValue>(
    () => ({
      enabled: state.enabled,
      leftPath: state.leftPath,
      rightPath: state.rightPath,
      leftSize: state.leftSize,
      rightSize: state.rightSize,
      leftVersionHash: state.leftVersionHash,
      rightVersionHash: state.rightVersionHash,
      mobileBlocked,
      dismissMobileNotice,
      openInSplit,
      compareWithCurrent,
      closeSplit,
      toggleSplit,
      navigatePane,
      setPaneSizes,
      clearPaneVersion,
    }),
    [
      state,
      mobileBlocked,
      dismissMobileNotice,
      openInSplit,
      compareWithCurrent,
      closeSplit,
      toggleSplit,
      navigatePane,
      setPaneSizes,
      clearPaneVersion,
    ],
  );

  return <SplitViewContext.Provider value={value}>{children}</SplitViewContext.Provider>;
}

export function useSplitView(): SplitViewContextValue {
  const ctx = useContext(SplitViewContext);
  if (!ctx) {
    return {
      enabled: false,
      leftPath: "",
      rightPath: "",
      leftSize: 50,
      rightSize: 50,
      leftVersionHash: null,
      rightVersionHash: null,
      mobileBlocked: false,
      dismissMobileNotice: () => {},
      openInSplit: () => {},
      compareWithCurrent: () => {},
      closeSplit: () => {},
      toggleSplit: () => {},
      navigatePane: () => {},
      setPaneSizes: () => {},
      clearPaneVersion: () => {},
    };
  }
  return ctx;
}

export function resetSplitViewForTests(): void {
  if (typeof window === "undefined") return;
  for (let i = sessionStorage.length - 1; i >= 0; i--) {
    const key = sessionStorage.key(i);
    if (key?.startsWith("kiwifs-split-view:")) sessionStorage.removeItem(key);
  }
}

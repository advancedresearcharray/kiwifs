import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";
import {
  closeSplitView,
  isSplitViewMobileViewport,
  loadSplitViewState,
  openSplitView,
  saveSplitViewState,
  toggleSplitView,
  type SplitViewState,
} from "@kw/lib/splitView";

type SplitViewContextValue = {
  state: SplitViewState;
  isSplit: boolean;
  mobileBlocked: boolean;
  dismissMobileBlocked: () => void;
  openInSplit: (path: string, primaryPath: string | null) => void;
  compareVersion: (path: string, versionHash: string) => void;
  toggleSplit: (primaryPath: string | null) => void;
  closeSplit: () => void;
  navigateLeft: (path: string) => void;
  navigateRight: (path: string) => void;
  setLeftSize: (size: number) => void;
  syncPrimaryPath: (path: string | null) => void;
};

const SplitViewContext = createContext<SplitViewContextValue | null>(null);

export function SplitViewProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<SplitViewState>(() => loadSplitViewState());
  const [mobileBlocked, setMobileBlocked] = useState(false);

  useEffect(() => {
    saveSplitViewState(state);
  }, [state]);

  const openInSplit = useCallback((path: string, primaryPath: string | null) => {
    if (isSplitViewMobileViewport()) {
      setMobileBlocked(true);
      return;
    }
    if (!primaryPath) {
      setState((prev) => openSplitView(prev, { path }, { path }));
      return;
    }
    setState((prev) => openSplitView(prev, { path: primaryPath }, { path }));
  }, []);

  const compareVersion = useCallback((path: string, versionHash: string) => {
    if (isSplitViewMobileViewport()) {
      setMobileBlocked(true);
      return;
    }
    setState((prev) => openSplitView(prev, { path }, { path, versionHash }));
  }, []);

  const toggleSplit = useCallback((primaryPath: string | null) => {
    if (!primaryPath) return;
    if (isSplitViewMobileViewport()) {
      setMobileBlocked(true);
      return;
    }
    setState((prev) => toggleSplitView(prev, primaryPath));
  }, []);

  const closeSplit = useCallback(() => {
    setState((prev) => closeSplitView(prev));
  }, []);

  const navigateLeft = useCallback((path: string) => {
    setState((prev) => {
      if (!prev.enabled || !prev.left) return prev;
      return { ...prev, left: { path, versionHash: null } };
    });
  }, []);

  const navigateRight = useCallback((path: string) => {
    setState((prev) => {
      if (!prev.enabled || !prev.right) return prev;
      return { ...prev, right: { path, versionHash: null } };
    });
  }, []);

  const setLeftSize = useCallback((size: number) => {
    setState((prev) => ({ ...prev, leftSize: size }));
  }, []);

  const syncPrimaryPath = useCallback((path: string | null) => {
    if (!path) return;
    setState((prev) => {
      if (!prev.enabled || !prev.left) return prev;
      if (prev.left.path === path && !prev.left.versionHash) return prev;
      return { ...prev, left: { path, versionHash: null } };
    });
  }, []);

  const value = useMemo(
    (): SplitViewContextValue => ({
      state,
      isSplit: state.enabled && Boolean(state.left && state.right),
      mobileBlocked,
      dismissMobileBlocked: () => setMobileBlocked(false),
      openInSplit,
      compareVersion,
      toggleSplit,
      closeSplit,
      navigateLeft,
      navigateRight,
      setLeftSize,
      syncPrimaryPath,
    }),
    [
      state,
      mobileBlocked,
      openInSplit,
      compareVersion,
      toggleSplit,
      closeSplit,
      navigateLeft,
      navigateRight,
      setLeftSize,
      syncPrimaryPath,
    ],
  );

  return <SplitViewContext.Provider value={value}>{children}</SplitViewContext.Provider>;
}

export function useSplitView(): SplitViewContextValue {
  const ctx = useContext(SplitViewContext);
  if (!ctx) {
    throw new Error("useSplitView must be used within SplitViewProvider");
  }
  return ctx;
}

export function useSplitViewOptional(): SplitViewContextValue | null {
  return useContext(SplitViewContext);
}

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
  closeSplitState,
  compareVersionSplitState,
  isMobileViewport,
  loadSplitViewFromSession,
  navigateSplitPane,
  openInSplitState,
  saveSplitViewToSession,
  toggleSplitState,
  type SplitPaneSide,
  type SplitViewState,
} from "@kw/lib/splitView";

type SplitViewContextValue = {
  split: SplitViewState;
  isMobile: boolean;
  mobileBlocked: boolean;
  clearMobileBlocked: () => void;
  toggleSplit: (activePath: string | null) => void;
  openInSplit: (activePath: string | null, targetPath: string) => void;
  compareWithCurrent: (path: string, versionHash: string) => void;
  closeSplit: () => void;
  navigatePane: (pane: SplitPaneSide, path: string) => void;
  setSizes: (sizes: [number, number]) => void;
};

const SplitViewContext = createContext<SplitViewContextValue | null>(null);

export function SplitViewProvider({
  children,
  isMobile,
}: {
  children: ReactNode;
  isMobile: boolean;
}) {
  const [split, setSplit] = useState<SplitViewState>(() => {
    const restored = loadSplitViewFromSession();
    if (restored?.enabled) return restored;
    return closeSplitState();
  });
  const [mobileBlocked, setMobileBlocked] = useState(false);

  useEffect(() => {
    saveSplitViewToSession(split);
  }, [split]);

  useEffect(() => {
    if (isMobile && split.enabled) {
      setSplit(closeSplitState());
    }
  }, [isMobile, split.enabled]);

  const guardMobile = useCallback(() => {
    if (!isMobileViewport() && !isMobile) return false;
    setMobileBlocked(true);
    return true;
  }, [isMobile]);

  const toggleSplit = useCallback(
    (activePath: string | null) => {
      if (guardMobile()) return;
      setSplit((current) => toggleSplitState(current, activePath));
    },
    [guardMobile],
  );

  const openInSplit = useCallback(
    (activePath: string | null, targetPath: string) => {
      if (guardMobile()) return;
      setSplit((current) => openInSplitState(current, activePath, targetPath));
    },
    [guardMobile],
  );

  const compareWithCurrent = useCallback(
    (path: string, versionHash: string) => {
      if (guardMobile()) return;
      setSplit(compareVersionSplitState(path, versionHash, split.sizes));
    },
    [guardMobile, split.sizes],
  );

  const closeSplit = useCallback(() => {
    setSplit(closeSplitState());
  }, []);

  const navigatePane = useCallback((pane: SplitPaneSide, path: string) => {
    setSplit((current) => navigateSplitPane(current, pane, path));
  }, []);

  const setSizes = useCallback((sizes: [number, number]) => {
    setSplit((current) => (current.enabled ? { ...current, sizes } : current));
  }, []);

  const clearMobileBlocked = useCallback(() => setMobileBlocked(false), []);

  const value = useMemo(
    () => ({
      split,
      isMobile,
      mobileBlocked,
      clearMobileBlocked,
      toggleSplit,
      openInSplit,
      compareWithCurrent,
      closeSplit,
      navigatePane,
      setSizes,
    }),
    [
      split,
      isMobile,
      mobileBlocked,
      clearMobileBlocked,
      toggleSplit,
      openInSplit,
      compareWithCurrent,
      closeSplit,
      navigatePane,
      setSizes,
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

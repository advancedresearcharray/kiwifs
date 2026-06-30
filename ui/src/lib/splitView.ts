/** Split / side-by-side page view state (#426). */

export const SPLIT_VIEW_STORAGE_KEY = "kiwifs-split-view";
export const MOBILE_MAX_WIDTH = 767;
export const DEFAULT_SPLIT_SIZES: [number, number] = [50, 50];

export type SplitPaneSide = "left" | "right";

export type SplitViewState = {
  enabled: boolean;
  leftPath: string;
  rightPath: string;
  /** When set, the right pane shows a historical version of rightPath. */
  rightVersionHash: string | null;
  sizes: [number, number];
};

export type PersistedSplitView = Omit<SplitViewState, "enabled"> & { enabled?: boolean };

export function isMobileViewport(width = typeof window !== "undefined" ? window.innerWidth : 1024): boolean {
  return width <= MOBILE_MAX_WIDTH;
}

export function createSplitState(
  leftPath: string,
  rightPath: string,
  options?: { rightVersionHash?: string | null; sizes?: [number, number] },
): SplitViewState {
  return {
    enabled: true,
    leftPath,
    rightPath,
    rightVersionHash: options?.rightVersionHash ?? null,
    sizes: options?.sizes ?? DEFAULT_SPLIT_SIZES,
  };
}

export function closeSplitState(): SplitViewState {
  return {
    enabled: false,
    leftPath: "",
    rightPath: "",
    rightVersionHash: null,
    sizes: DEFAULT_SPLIT_SIZES,
  };
}

export function toggleSplitState(
  current: SplitViewState,
  activePath: string | null,
): SplitViewState {
  if (current.enabled) return closeSplitState();
  if (!activePath) return current;
  return createSplitState(activePath, activePath);
}

export function openInSplitState(
  current: SplitViewState,
  activePath: string | null,
  targetPath: string,
): SplitViewState {
  const left = current.enabled ? current.leftPath : activePath ?? targetPath;
  return createSplitState(left, targetPath, { sizes: current.sizes });
}

export function compareVersionSplitState(
  path: string,
  versionHash: string,
  sizes: [number, number] = DEFAULT_SPLIT_SIZES,
): SplitViewState {
  return createSplitState(path, path, { rightVersionHash: versionHash, sizes });
}

export function navigateSplitPane(
  state: SplitViewState,
  pane: SplitPaneSide,
  path: string,
): SplitViewState {
  if (!state.enabled) return state;
  if (pane === "left") {
    return { ...state, leftPath: path };
  }
  return { ...state, rightPath: path, rightVersionHash: null };
}

export function loadSplitViewFromSession(): SplitViewState | null {
  if (typeof sessionStorage === "undefined") return null;
  try {
    const raw = sessionStorage.getItem(SPLIT_VIEW_STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as PersistedSplitView;
    if (!parsed.leftPath || !parsed.rightPath) return null;
    return {
      enabled: parsed.enabled !== false,
      leftPath: parsed.leftPath,
      rightPath: parsed.rightPath,
      rightVersionHash: parsed.rightVersionHash ?? null,
      sizes: normalizeSizes(parsed.sizes),
    };
  } catch {
    return null;
  }
}

export function saveSplitViewToSession(state: SplitViewState): void {
  if (typeof sessionStorage === "undefined") return;
  try {
    if (!state.enabled) {
      sessionStorage.removeItem(SPLIT_VIEW_STORAGE_KEY);
      return;
    }
    const payload: PersistedSplitView = {
      enabled: true,
      leftPath: state.leftPath,
      rightPath: state.rightPath,
      rightVersionHash: state.rightVersionHash,
      sizes: state.sizes,
    };
    sessionStorage.setItem(SPLIT_VIEW_STORAGE_KEY, JSON.stringify(payload));
  } catch {
    /* ignore quota / privacy mode */
  }
}

export function normalizeSizes(sizes: unknown): [number, number] {
  if (!Array.isArray(sizes) || sizes.length !== 2) return DEFAULT_SPLIT_SIZES;
  const a = Number(sizes[0]);
  const b = Number(sizes[1]);
  if (!Number.isFinite(a) || !Number.isFinite(b) || a <= 0 || b <= 0) {
    return DEFAULT_SPLIT_SIZES;
  }
  const total = a + b;
  return [Math.round((a / total) * 100), Math.round((b / total) * 100)];
}

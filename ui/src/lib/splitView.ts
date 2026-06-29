/**
 * Split-view state helpers — persisted in sessionStorage per workspace space.
 */

export type SplitPane = "left" | "right";

export type SplitViewPersisted = {
  enabled: boolean;
  leftPath: string;
  rightPath: string;
  leftSize: number;
  rightSize: number;
  leftVersionHash: string | null;
  rightVersionHash: string | null;
};

export const SPLIT_VIEW_MIN_PANE = 20;
export const SPLIT_VIEW_DEFAULT_LEFT = 50;
export const SPLIT_VIEW_DEFAULT_RIGHT = 50;
export const SPLIT_VIEW_MOBILE_MAX = 767;

const STORAGE_PREFIX = "kiwifs-split-view";

export function splitViewStorageKey(space: string): string {
  const normalized = space.trim() || "default";
  return `${STORAGE_PREFIX}:${normalized}`;
}

export function defaultSplitViewState(leftPath: string): SplitViewPersisted {
  return {
    enabled: false,
    leftPath,
    rightPath: leftPath,
    leftSize: SPLIT_VIEW_DEFAULT_LEFT,
    rightSize: SPLIT_VIEW_DEFAULT_RIGHT,
    leftVersionHash: null,
    rightVersionHash: null,
  };
}

export function loadSplitViewState(space: string, fallbackLeftPath: string): SplitViewPersisted {
  if (typeof window === "undefined") {
    return defaultSplitViewState(fallbackLeftPath);
  }
  try {
    const raw = sessionStorage.getItem(splitViewStorageKey(space));
    if (!raw) return defaultSplitViewState(fallbackLeftPath);
    const parsed = JSON.parse(raw) as Partial<SplitViewPersisted>;
    if (!parsed || typeof parsed !== "object") {
      return defaultSplitViewState(fallbackLeftPath);
    }
    const leftPath = typeof parsed.leftPath === "string" && parsed.leftPath ? parsed.leftPath : fallbackLeftPath;
    const rightPath = typeof parsed.rightPath === "string" && parsed.rightPath ? parsed.rightPath : leftPath;
    const leftSize = clampPaneSize(parsed.leftSize ?? SPLIT_VIEW_DEFAULT_LEFT);
    const rightSize = clampPaneSize(parsed.rightSize ?? SPLIT_VIEW_DEFAULT_RIGHT);
    return {
      enabled: Boolean(parsed.enabled),
      leftPath,
      rightPath,
      leftSize,
      rightSize,
      leftVersionHash: typeof parsed.leftVersionHash === "string" ? parsed.leftVersionHash : null,
      rightVersionHash: typeof parsed.rightVersionHash === "string" ? parsed.rightVersionHash : null,
    };
  } catch {
    return defaultSplitViewState(fallbackLeftPath);
  }
}

export function saveSplitViewState(space: string, state: SplitViewPersisted): void {
  if (typeof window === "undefined") return;
  try {
    sessionStorage.setItem(splitViewStorageKey(space), JSON.stringify(state));
  } catch {
    // ignore quota / private mode
  }
}

export function clearSplitViewState(space: string): void {
  if (typeof window === "undefined") return;
  try {
    sessionStorage.removeItem(splitViewStorageKey(space));
  } catch {
    // ignore
  }
}

export function clampPaneSize(size: number): number {
  if (!Number.isFinite(size)) return SPLIT_VIEW_DEFAULT_LEFT;
  return Math.max(SPLIT_VIEW_MIN_PANE, Math.min(100 - SPLIT_VIEW_MIN_PANE, size));
}

export function isSplitViewMobileViewport(): boolean {
  if (typeof window === "undefined") return false;
  return window.innerWidth <= SPLIT_VIEW_MOBILE_MAX;
}

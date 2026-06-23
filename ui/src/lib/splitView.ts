/**
 * Split / side-by-side page view state (issue #426).
 * Persisted in sessionStorage so layout survives refresh within the tab.
 */

export type SplitPaneVersion = {
  path: string;
  versionHash: string | null;
};

export type SplitViewState = {
  enabled: boolean;
  left: SplitPaneVersion;
  right: SplitPaneVersion;
  /** Left pane width as percentage (10–90). */
  leftSize: number;
};

export const SPLIT_VIEW_STORAGE_KEY = "kiwifs-split-view";
export const SPLIT_VIEW_MIN_PANE = 20;
export const SPLIT_VIEW_MAX_PANE = 80;
export const SPLIT_VIEW_DEFAULT_LEFT_SIZE = 50;
export const MOBILE_SPLIT_MAX_WIDTH = 767;

export function clampPaneSize(size: number): number {
  if (!Number.isFinite(size)) return SPLIT_VIEW_DEFAULT_LEFT_SIZE;
  return Math.max(SPLIT_VIEW_MIN_PANE, Math.min(SPLIT_VIEW_MAX_PANE, size));
}

export function defaultSplitViewState(
  leftPath: string,
  rightPath?: string,
): SplitViewState {
  const right = rightPath ?? leftPath;
  return {
    enabled: true,
    left: { path: leftPath, versionHash: null },
    right: { path: right, versionHash: null },
    leftSize: SPLIT_VIEW_DEFAULT_LEFT_SIZE,
  };
}

export function parseSplitViewState(raw: unknown): SplitViewState | null {
  if (!raw || typeof raw !== "object") return null;
  const o = raw as Record<string, unknown>;
  if (typeof o.enabled !== "boolean") return null;

  const readPane = (v: unknown): SplitPaneVersion | null => {
    if (!v || typeof v !== "object") return null;
    const p = v as Record<string, unknown>;
    if (typeof p.path !== "string" || !p.path) return null;
    const versionHash =
      p.versionHash == null
        ? null
        : typeof p.versionHash === "string"
          ? p.versionHash
          : null;
    return { path: p.path, versionHash };
  };

  const left = readPane(o.left);
  const right = readPane(o.right);
  if (!left || !right) return null;

  return {
    enabled: o.enabled,
    left,
    right,
    leftSize: clampPaneSize(Number(o.leftSize)),
  };
}

export function loadSplitViewState(): SplitViewState | null {
  if (typeof sessionStorage === "undefined") return null;
  try {
    const raw = sessionStorage.getItem(SPLIT_VIEW_STORAGE_KEY);
    if (!raw) return null;
    return parseSplitViewState(JSON.parse(raw));
  } catch {
    return null;
  }
}

export function saveSplitViewState(state: SplitViewState | null): void {
  if (typeof sessionStorage === "undefined") return;
  try {
    if (!state || !state.enabled) {
      sessionStorage.removeItem(SPLIT_VIEW_STORAGE_KEY);
      return;
    }
    sessionStorage.setItem(SPLIT_VIEW_STORAGE_KEY, JSON.stringify(state));
  } catch {
    // ignore quota / private mode
  }
}

export function openSplitView(
  current: SplitViewState | null,
  leftPath: string,
  rightPath: string,
  rightVersion?: string | null,
): SplitViewState {
  const leftSize = current?.leftSize ?? SPLIT_VIEW_DEFAULT_LEFT_SIZE;
  return {
    enabled: true,
    left: { path: leftPath, versionHash: null },
    right: { path: rightPath, versionHash: rightVersion ?? null },
    leftSize: clampPaneSize(leftSize),
  };
}

export function closeSplitView(): null {
  saveSplitViewState(null);
  return null;
}

export function toggleSplitView(
  current: SplitViewState | null,
  activePath: string | null,
): SplitViewState | null {
  if (current?.enabled) {
    saveSplitViewState(null);
    return null;
  }
  if (!activePath) return null;
  const next = defaultSplitViewState(activePath);
  saveSplitViewState(next);
  return next;
}

export function updateSplitPanePath(
  state: SplitViewState,
  pane: "left" | "right",
  path: string,
): SplitViewState {
  const next: SplitViewState = {
    ...state,
    left: pane === "left" ? { path, versionHash: null } : state.left,
    right: pane === "right" ? { path, versionHash: null } : state.right,
  };
  saveSplitViewState(next);
  return next;
}

export function isMobileViewport(width = typeof window !== "undefined" ? window.innerWidth : 1024): boolean {
  return width <= MOBILE_SPLIT_MAX_WIDTH;
}

/**
 * Split / side-by-side page view state (sessionStorage-backed).
 */

export type SplitPaneSpec = {
  path: string;
  versionHash?: string | null;
};

export type SplitViewState = {
  enabled: boolean;
  left: SplitPaneSpec | null;
  right: SplitPaneSpec | null;
  /** Left pane width as a percentage (25–75). */
  leftSize: number;
};

export const SPLIT_VIEW_STORAGE_KEY = "kiwifs-split-view";
export const SPLIT_VIEW_MIN_PANE_PERCENT = 25;
export const SPLIT_VIEW_MAX_PANE_PERCENT = 75;
export const SPLIT_VIEW_MOBILE_BREAKPOINT = 768;

const DEFAULT_STATE: SplitViewState = {
  enabled: false,
  left: null,
  right: null,
  leftSize: 50,
};

export function clampPaneSize(size: number): number {
  return Math.max(SPLIT_VIEW_MIN_PANE_PERCENT, Math.min(SPLIT_VIEW_MAX_PANE_PERCENT, size));
}

export function parseSplitViewState(raw: string | null): SplitViewState | null {
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as Partial<SplitViewState>;
    if (!parsed || typeof parsed !== "object") return null;
    const left = parsePaneSpec(parsed.left);
    const right = parsePaneSpec(parsed.right);
    const enabled = Boolean(parsed.enabled && left && right);
    const leftSize = clampPaneSize(typeof parsed.leftSize === "number" ? parsed.leftSize : 50);
    if (!enabled) {
      return { ...DEFAULT_STATE, leftSize };
    }
    return { enabled: true, left, right, leftSize };
  } catch {
    return null;
  }
}

function parsePaneSpec(value: unknown): SplitPaneSpec | null {
  if (!value || typeof value !== "object") return null;
  const path = (value as SplitPaneSpec).path;
  if (typeof path !== "string" || !path.trim()) return null;
  const versionHash = (value as SplitPaneSpec).versionHash;
  return {
    path,
    versionHash: typeof versionHash === "string" && versionHash ? versionHash : null,
  };
}

export function loadSplitViewState(): SplitViewState {
  if (typeof sessionStorage === "undefined") return { ...DEFAULT_STATE };
  try {
    return parseSplitViewState(sessionStorage.getItem(SPLIT_VIEW_STORAGE_KEY)) ?? { ...DEFAULT_STATE };
  } catch {
    return { ...DEFAULT_STATE };
  }
}

export function saveSplitViewState(state: SplitViewState): void {
  if (typeof sessionStorage === "undefined") return;
  try {
    if (!state.enabled || !state.left || !state.right) {
      sessionStorage.removeItem(SPLIT_VIEW_STORAGE_KEY);
      return;
    }
    sessionStorage.setItem(
      SPLIT_VIEW_STORAGE_KEY,
      JSON.stringify({
        enabled: true,
        left: state.left,
        right: state.right,
        leftSize: clampPaneSize(state.leftSize),
      }),
    );
  } catch {
    // ignore quota / private mode errors
  }
}

export function openSplitView(
  state: SplitViewState,
  left: SplitPaneSpec,
  right: SplitPaneSpec,
): SplitViewState {
  return {
    enabled: true,
    left,
    right,
    leftSize: state.leftSize || 50,
  };
}

export function closeSplitView(state: SplitViewState): SplitViewState {
  return {
    ...state,
    enabled: false,
    right: null,
  };
}

export function toggleSplitView(state: SplitViewState, primaryPath: string | null): SplitViewState {
  if (state.enabled) {
    return closeSplitView(state);
  }
  if (!primaryPath) return state;
  const pane: SplitPaneSpec = { path: primaryPath };
  return openSplitView(state, pane, pane);
}

export function isSplitViewMobileViewport(): boolean {
  return typeof window !== "undefined" && window.innerWidth < SPLIT_VIEW_MOBILE_BREAKPOINT;
}

/**
 * Split-view state helpers — persisted in sessionStorage per tab.
 */

export const SPLIT_VIEW_SESSION_KEY = "kiwifs-split-view";

export type SplitPaneVersion = {
  path: string;
  hash: string;
};

export type SplitViewState = {
  enabled: boolean;
  rightPath: string | null;
  rightVersion: SplitPaneVersion | null;
  /** Percent widths for [left, right] panes. */
  sizes: [number, number];
};

export const DEFAULT_SPLIT_SIZES: [number, number] = [50, 50];

export function createSplitViewState(
  overrides: Partial<SplitViewState> = {},
): SplitViewState {
  return {
    enabled: false,
    rightPath: null,
    rightVersion: null,
    sizes: DEFAULT_SPLIT_SIZES,
    ...overrides,
  };
}

export function loadSplitViewState(): SplitViewState | null {
  if (typeof sessionStorage === "undefined") return null;
  try {
    const raw = sessionStorage.getItem(SPLIT_VIEW_SESSION_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Partial<SplitViewState>;
    if (typeof parsed !== "object" || parsed == null) return null;
    const sizes = parsed.sizes;
    const normalizedSizes: [number, number] =
      Array.isArray(sizes)
      && sizes.length === 2
      && typeof sizes[0] === "number"
      && typeof sizes[1] === "number"
      && sizes[0] > 0
      && sizes[1] > 0
        ? [sizes[0], sizes[1]]
        : DEFAULT_SPLIT_SIZES;
    return createSplitViewState({
      enabled: Boolean(parsed.enabled),
      rightPath: typeof parsed.rightPath === "string" ? parsed.rightPath : null,
      rightVersion:
        parsed.rightVersion
        && typeof parsed.rightVersion.path === "string"
        && typeof parsed.rightVersion.hash === "string"
          ? { path: parsed.rightVersion.path, hash: parsed.rightVersion.hash }
          : null,
      sizes: normalizedSizes,
    });
  } catch {
    return null;
  }
}

export function saveSplitViewState(state: SplitViewState): void {
  if (typeof sessionStorage === "undefined") return;
  try {
    sessionStorage.setItem(SPLIT_VIEW_SESSION_KEY, JSON.stringify(state));
  } catch {
    // ignore quota / private mode errors
  }
}

export function clearSplitViewState(): void {
  if (typeof sessionStorage === "undefined") return;
  try {
    sessionStorage.removeItem(SPLIT_VIEW_SESSION_KEY);
  } catch {
    // ignore
  }
}

/** Whether the right pane has content to render. */
export function splitViewHasSecondary(state: SplitViewState): boolean {
  return Boolean(state.rightPath || state.rightVersion);
}

export function openPathInSplit(
  state: SplitViewState,
  rightPath: string,
): SplitViewState {
  return {
    ...state,
    enabled: true,
    rightPath,
    rightVersion: null,
  };
}

export function openVersionInSplit(
  state: SplitViewState,
  version: SplitPaneVersion,
): SplitViewState {
  return {
    ...state,
    enabled: true,
    rightPath: null,
    rightVersion: version,
  };
}

export function toggleSplitView(
  state: SplitViewState,
  activePath: string | null,
): SplitViewState {
  if (state.enabled) {
    return createSplitViewState({ sizes: state.sizes });
  }
  if (!activePath) return state;
  return {
    enabled: true,
    rightPath: activePath,
    rightVersion: null,
    sizes: state.sizes,
  };
}

export function closeSecondaryPane(state: SplitViewState): SplitViewState {
  return {
    ...state,
    rightPath: null,
    rightVersion: null,
  };
}

export function clampSplitSizes(
  leftPercent: number,
  min = 20,
  max = 80,
): [number, number] {
  const clamped = Math.max(min, Math.min(max, leftPercent));
  return [clamped, 100 - clamped];
}

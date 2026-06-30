import {
  isViewRouteAllowed,
  viewFeatureFromPathname,
  type UIFeatureKey,
} from "./uiFeatures";

/** Built-in full-screen views addressable via /view/{name}. */
export type AppViewId =
  | "graph"
  | "bases"
  | "canvas"
  | "whiteboard"
  | "timeline"
  | "calendar"
  | "kanban"
  | "data";

const VIEW_FEATURE_TO_ID: Partial<Record<UIFeatureKey, AppViewId>> = {
  graph: "graph",
  bases: "bases",
  canvas: "canvas",
  whiteboard: "whiteboard",
  timeline: "timeline",
  calendar: "calendar",
  kanban: "kanban",
  data_sources: "data",
};

/** Map a /view/* pathname to a built-in view id, or null when not a view route. */
export function viewIdFromPathname(pathname: string): AppViewId | null {
  const feature = viewFeatureFromPathname(pathname);
  if (!feature) return null;
  return VIEW_FEATURE_TO_ID[feature] ?? null;
}

/** Resolve which view to open from the URL when the feature flag allows it. */
export function shouldOpenViewFromPathname(
  pathname: string,
  features: Record<UIFeatureKey, boolean>,
): AppViewId | null {
  if (!pathname.startsWith("/view/")) return null;
  if (!isViewRouteAllowed(pathname, features)) return null;
  return viewIdFromPathname(pathname);
}

/** Keep /view/* URLs when no page is active (avoid redirecting to /). */
export function shouldPreservePathnameForViewRoute(pathname: string): boolean {
  return pathname.startsWith("/view/");
}

/**
 * Host embed configuration for KiwiFS UI (self-host, cloud shell, iframes).
 *
 * Set before boot:
 *   window.__KIWIFS_CONFIG__ = {
 *     toolbarActions: [{ id: "my-tool", icon: "Wand2", label: "My tool" }],
 *   };
 *
 * Listen for clicks:
 *   window.addEventListener("kiwifs-toolbar-action", (e) => {
 *     if (e.detail.id === "my-tool") { ... }
 *   });
 *
 * Optional active/disabled styling from the host:
 *   window.dispatchEvent(new CustomEvent("kiwifs-host-toolbar-state", {
 *     detail: { "my-tool": { active: true } },
 *   }));
 */

export type KiwiToolbarAction = {
  /** Stable id; emitted in kiwifs-toolbar-action detail */
  id: string;
  /** Lucide icon export name, e.g. "BotMessageSquare" */
  icon: string;
  /** Tooltip + aria-label */
  label: string;
  disabled?: boolean;
};

export type KiwiToolbarActionState = {
  active?: boolean;
  disabled?: boolean;
};

export type KiwiHostConfig = {
  allowedOrigins?: string[];
  toolbarActions?: KiwiToolbarAction[];
};

export const KIWI_TOOLBAR_ACTION_EVENT = "kiwifs-toolbar-action";
export const KIWI_TOOLBAR_STATE_EVENT = "kiwifs-host-toolbar-state";

declare global {
  interface Window {
    __KIWIFS_CONFIG__?: KiwiHostConfig;
  }
}

export function getHostConfig(): KiwiHostConfig {
  if (typeof window === "undefined") return {};
  return window.__KIWIFS_CONFIG__ ?? {};
}

export function getToolbarActions(): KiwiToolbarAction[] {
  return getHostConfig().toolbarActions ?? [];
}

export function dispatchToolbarAction(id: string): void {
  window.dispatchEvent(
    new CustomEvent(KIWI_TOOLBAR_ACTION_EVENT, { detail: { id } }),
  );
}

export type ToolbarActionEvent = CustomEvent<{ id: string }>;
export type ToolbarStateEvent = CustomEvent<Record<string, KiwiToolbarActionState>>;

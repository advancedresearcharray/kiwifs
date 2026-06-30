import { describe, expect, it } from "vitest";
import {
  isKeyboardShortcutsOverlayOpen,
  resolveOverlayDismiss,
  setKeyboardShortcutsOverlayOpen,
  shouldSuppressKeybindingWhileShortcutsOpen,
  type OverlayState,
} from "./overlayDismiss";

const closed: OverlayState = {
  shortcutsOpen: false,
  newOpen: false,
  searchOpen: false,
  graphOpen: false,
  historyOpen: false,
  dataOpen: false,
  basesOpen: false,
  canvasOpen: false,
  whiteboardOpen: false,
  timelineOpen: false,
  kanbanOpen: false,
};

describe("isKeyboardShortcutsOverlayOpen", () => {
  it("tracks open state via setKeyboardShortcutsOverlayOpen", () => {
    setKeyboardShortcutsOverlayOpen(false);
    expect(isKeyboardShortcutsOverlayOpen()).toBe(false);
    setKeyboardShortcutsOverlayOpen(true);
    expect(isKeyboardShortcutsOverlayOpen()).toBe(true);
    setKeyboardShortcutsOverlayOpen(false);
  });
});

describe("shouldSuppressKeybindingWhileShortcutsOpen", () => {
  it("allows close_overlay while shortcuts help is open", () => {
    expect(shouldSuppressKeybindingWhileShortcutsOpen(true, "close_overlay")).toBe(false);
  });

  it("blocks other actions while shortcuts help is open", () => {
    expect(shouldSuppressKeybindingWhileShortcutsOpen(true, "search")).toBe(true);
    expect(shouldSuppressKeybindingWhileShortcutsOpen(true, "new_page")).toBe(true);
  });

  it("does not block when shortcuts help is closed", () => {
    expect(shouldSuppressKeybindingWhileShortcutsOpen(false, "search")).toBe(false);
  });
});

describe("resolveOverlayDismiss", () => {
  it("returns null when no overlays are open", () => {
    expect(resolveOverlayDismiss(closed)).toBeNull();
  });

  it("prefers shortcuts help over other overlays", () => {
    expect(
      resolveOverlayDismiss({ ...closed, shortcutsOpen: true, searchOpen: true, graphOpen: true }),
    ).toBe("shortcuts");
  });

  it("prefers new-page dialog over search and views", () => {
    expect(resolveOverlayDismiss({ ...closed, newOpen: true, searchOpen: true })).toBe("new");
  });

  it("prefers search over full-screen views", () => {
    expect(resolveOverlayDismiss({ ...closed, searchOpen: true, graphOpen: true })).toBe("search");
  });

  it("dismisses full-screen views in stable priority order", () => {
    expect(resolveOverlayDismiss({ ...closed, graphOpen: true, kanbanOpen: true })).toBe("graph");
    expect(resolveOverlayDismiss({ ...closed, historyOpen: true, dataOpen: true })).toBe("history");
    expect(resolveOverlayDismiss({ ...closed, kanbanOpen: true })).toBe("kanban");
  });
});

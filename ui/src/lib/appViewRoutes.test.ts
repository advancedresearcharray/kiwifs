import { describe, expect, it } from "vitest";
import {
  shouldOpenViewFromPathname,
  shouldPreservePathnameForViewRoute,
  viewIdFromPathname,
} from "./appViewRoutes";
import { DEFAULT_UI_FEATURES } from "./uiFeatures";

describe("viewIdFromPathname", () => {
  it("maps /view/calendar to calendar", () => {
    expect(viewIdFromPathname("/view/calendar")).toBe("calendar");
  });

  it("maps /view/data to data", () => {
    expect(viewIdFromPathname("/view/data")).toBe("data");
  });

  it("returns null for page routes", () => {
    expect(viewIdFromPathname("/page/notes/today.md")).toBeNull();
    expect(viewIdFromPathname("/")).toBeNull();
  });
});

describe("shouldOpenViewFromPathname", () => {
  it("opens calendar when feature is enabled", () => {
    expect(
      shouldOpenViewFromPathname("/view/calendar", DEFAULT_UI_FEATURES),
    ).toBe("calendar");
  });

  it("returns null when calendar feature is disabled", () => {
    expect(
      shouldOpenViewFromPathname("/view/calendar", {
        ...DEFAULT_UI_FEATURES,
        calendar: false,
      }),
    ).toBeNull();
  });

  it("returns null for non-view paths", () => {
    expect(shouldOpenViewFromPathname("/", DEFAULT_UI_FEATURES)).toBeNull();
  });
});

describe("shouldPreservePathnameForViewRoute", () => {
  it("preserves /view/calendar without an active page", () => {
    expect(shouldPreservePathnameForViewRoute("/view/calendar")).toBe(true);
  });

  it("does not preserve root or page paths", () => {
    expect(shouldPreservePathnameForViewRoute("/")).toBe(false);
    expect(shouldPreservePathnameForViewRoute("/page/a.md")).toBe(false);
  });
});

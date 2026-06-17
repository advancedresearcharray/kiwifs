import { afterEach, describe, expect, it, vi } from "vitest";
import { api } from "./api";
import { DEFAULT_UI_FEATURES } from "./uiFeatures";
import { useUIConfigStore } from "./uiConfigStore";

describe("uiConfigStore", () => {
  afterEach(() => {
    useUIConfigStore.setState({
      themeLocked: false,
      features: DEFAULT_UI_FEATURES,
      loaded: false,
    });
    vi.restoreAllMocks();
  });

  it("defaults before load", () => {
    expect(useUIConfigStore.getState().themeLocked).toBe(false);
    expect(useUIConfigStore.getState().features).toEqual(DEFAULT_UI_FEATURES);
    expect(useUIConfigStore.getState().loaded).toBe(false);
  });

  it("falls back to defaults when ui-config fetch fails", async () => {
    vi.spyOn(api, "getUIConfig").mockRejectedValue(new Error("network"));

    await useUIConfigStore.getState().load();

    expect(useUIConfigStore.getState().features).toEqual(DEFAULT_UI_FEATURES);
    expect(useUIConfigStore.getState().loaded).toBe(true);
  });

  it("stores feature flags from ui-config", async () => {
    vi.spyOn(api, "getUIConfig").mockResolvedValue({
      themeLocked: false,
      features: { kanban: false, graph: true },
    });

    await useUIConfigStore.getState().load();

    expect(useUIConfigStore.getState().features.kanban).toBe(false);
    expect(useUIConfigStore.getState().features.graph).toBe(true);
    expect(useUIConfigStore.getState().features.canvas).toBe(true);
  });
});

import { create } from "zustand";
import { api } from "./api";
import { DEFAULT_UI_FEATURES, resolveUIFeatures, type UIFeatureKey } from "./uiFeatures";

type UIConfigState = {
  themeLocked: boolean;
  features: Record<UIFeatureKey, boolean>;
  loaded: boolean;
  load: () => Promise<void>;
};

export const useUIConfigStore = create<UIConfigState>((set) => ({
  themeLocked: false,
  features: DEFAULT_UI_FEATURES,
  loaded: false,
  load: async () => {
    try {
      const config = await api.getUIConfig();
      set({
        themeLocked: config.themeLocked === true,
        features: resolveUIFeatures(config.features),
        loaded: true,
      });
    } catch {
      set({
        themeLocked: false,
        features: DEFAULT_UI_FEATURES,
        loaded: true,
      });
    }
  },
}));

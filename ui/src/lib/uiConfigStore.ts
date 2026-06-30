import { create } from "zustand";
import { api } from "./api";
import { DEFAULT_BRANDING, resolveBranding, type BrandingConfig } from "./branding";
import { DEFAULT_UI_FEATURES, resolveUIFeatures, type UIFeatureKey } from "./uiFeatures";

export const THEME_LOCKED_TOOLTIP = "Theme locked by admin";

type UIConfigState = {
  themeLocked: boolean;
  allowedPresets: string[];
  branding: BrandingConfig;
  features: Record<UIFeatureKey, boolean>;
  toolbarViews: string[] | null | undefined;
  loaded: boolean;
  load: () => Promise<void>;
};

export const useUIConfigStore = create<UIConfigState>((set) => ({
  themeLocked: false,
  allowedPresets: [],
  branding: DEFAULT_BRANDING,
  features: DEFAULT_UI_FEATURES,
  toolbarViews: undefined,
  loaded: false,
  load: async () => {
    try {
      const config = await api.getUIConfig();
      set({
        themeLocked: config.themeLocked === true,
        allowedPresets: config.theme?.allowedPresets ?? [],
        branding: resolveBranding(config.branding ?? {}),
        features: resolveUIFeatures(config.features),
        toolbarViews: config.toolbarViews ?? null,
        loaded: true,
      });
    } catch {
      set({
        themeLocked: false,
        allowedPresets: [],
        branding: DEFAULT_BRANDING,
        features: DEFAULT_UI_FEATURES,
        toolbarViews: null,
        loaded: true,
      });
    }
  },
}));

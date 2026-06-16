import { create } from "zustand";
import { api } from "./api";
import { DEFAULT_BRANDING, resolveBranding, type BrandingConfig } from "./branding";

type UIConfigState = {
  themeLocked: boolean;
  branding: BrandingConfig;
  loaded: boolean;
  load: () => Promise<void>;
};

export const useUIConfigStore = create<UIConfigState>((set) => ({
  themeLocked: false,
  branding: DEFAULT_BRANDING,
  loaded: false,
  load: async () => {
    try {
      const config = await api.getUIConfig();
      set({
        themeLocked: config.themeLocked === true,
        branding: resolveBranding(config.branding ?? {}),
        loaded: true,
      });
    } catch {
      set({ themeLocked: false, branding: DEFAULT_BRANDING, loaded: true });
    }
  },
}));

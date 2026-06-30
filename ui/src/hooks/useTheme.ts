import { useCallback, useEffect, useRef, useState } from "react";
import {
  applyKiwiTheme,
  applyKiwiCustomCSS,
  removeKiwiTheme,
  type KiwiThemeOverrides,
} from "../lib/kiwiTheme";
import { api, getCurrentSpace, onSpaceChange } from "../lib/api";
import { guardedThemeAction } from "../lib/themeEditLock";
import { useUIConfigStore } from "../lib/uiConfigStore";
import {
  presets as builtinPresets,
  presetToOverrides,
  findPreset,
  mergePresets,
  filterPresetsWithAllowList,
  resolvePresetName,
  type ThemePreset,
} from "../themes";
import type { UserPreferences } from "../lib/userPreferences";

export type Theme = "light" | "dark";

const LS_THEME = "kiwifs-theme";

function spaceKey(base: string): string {
  const space = getCurrentSpace();
  return space && space !== "default" ? `${base}:${space}` : base;
}

function lsPreset(): string { return spaceKey("kiwifs-preset"); }
function lsCustomTheme(): string { return spaceKey("kiwifs-custom-theme"); }

function readLS(key: string, fallback: string): string {
  try {
    return localStorage.getItem(key) || fallback;
  } catch {
    return fallback;
  }
}

function writeLS(key: string, val: string) {
  try {
    localStorage.setItem(key, val);
  } catch {
    /* ignore */
  }
}

export function getCustomTheme(): KiwiThemeOverrides | null {
  try {
    const raw = localStorage.getItem(lsCustomTheme());
    if (!raw) return null;
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function setCustomTheme(t: KiwiThemeOverrides | null) {
  if (t) {
    writeLS(lsCustomTheme(), JSON.stringify(t));
  } else {
    try {
      localStorage.removeItem(lsCustomTheme());
    } catch {
      /* ignore */
    }
  }
}

function externalThemeAPI(): {
  get: () => Theme;
  toggle: () => void;
  set: (t: Theme) => void;
} | null {
  if (typeof window === "undefined") return null;
  const w = window as unknown as Record<string, unknown>;
  if (
    typeof w.__kiwi_theme_get__ === "function" &&
    typeof w.__kiwi_theme_toggle__ === "function" &&
    typeof w.__kiwi_theme_set__ === "function"
  ) {
    return {
      get: w.__kiwi_theme_get__ as () => Theme,
      toggle: w.__kiwi_theme_toggle__ as () => void,
      set: w.__kiwi_theme_set__ as (t: Theme) => void,
    };
  }
  return null;
}

function workspacePresetFromAPI(
  p: {
    name: string;
    description?: string;
    light: Record<string, string>;
    dark: Record<string, string>;
  },
): ThemePreset {
  return {
    name: p.name,
    description: p.description || "",
    light: p.light,
    dark: p.dark,
  };
}

export function useTheme(options?: {
  serverPrefs?: UserPreferences | null;
  onPresetChange?: (preset: string) => void;
}): {
  theme: Theme;
  toggleTheme: () => void;
  preset: string;
  setPreset: (name: string) => void;
  presets: ThemePreset[];
  presetErrors: Array<{ file: string; error: string }>;
  themeLocked: boolean;
} {
  const themeLocked = useUIConfigStore((s) => s.themeLocked);
  const serverPreset = options?.serverPrefs?.theme;
  const onPresetChange = options?.onPresetChange;
  const [theme, setTheme] = useState<Theme>(() => {
    if (typeof document === "undefined") return "light";
    return document.documentElement.classList.contains("dark") ? "dark" : "light";
  });

  const [preset, setPresetState] = useState(() => serverPreset || readLS(lsPreset(), "Kiwi"));
  const [availablePresets, setAvailablePresets] = useState<ThemePreset[]>(builtinPresets);
  const [presetErrors, setPresetErrors] = useState<Array<{ file: string; error: string }>>([]);
  const presetRef = useRef(preset);
  presetRef.current = preset;

  useEffect(() => {
    if (serverPreset) {
      setPresetState(serverPreset);
    }
  }, [serverPreset]);

  const loadPresets = useCallback(() => {
    api.getThemePresets()
      .then((presetRes) => {
        const workspace = (presetRes.presets || []).map(workspacePresetFromAPI);
        const merged = mergePresets(builtinPresets, workspace);
        const allowed = useUIConfigStore.getState().allowedPresets;
        const filtered = filterPresetsWithAllowList(merged, allowed);
        setAvailablePresets(filtered);
        setPresetErrors(presetRes.errors || []);
        const saved = readLS(lsPreset(), "");
        const active = serverPreset || saved || presetRef.current;
        const resolved = resolvePresetName(active, filtered);
        if (resolved !== active) {
          setPresetState(resolved);
          writeLS(lsPreset(), resolved);
          onPresetChange?.(resolved);
        }
      })
      .catch(() => {
        const allowed = useUIConfigStore.getState().allowedPresets;
        setAvailablePresets(filterPresetsWithAllowList(builtinPresets, allowed));
        setPresetErrors([]);
      });
  }, [serverPreset, onPresetChange]);

  useEffect(() => {
    loadPresets();
    return onSpaceChange(loadPresets);
  }, [loadPresets]);

  const resolvePreset = useCallback(
    (name: string) => findPreset(name, availablePresets),
    [availablePresets],
  );

  useEffect(() => {
    const observer = new MutationObserver(() => {
      const next: Theme = document.documentElement.classList.contains("dark")
        ? "dark"
        : "light";
      setTheme((prev) => (prev !== next ? next : prev));
    });
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    });
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (externalThemeAPI()) return;
    const root = document.documentElement;
    if (theme === "dark") root.classList.add("dark");
    else root.classList.remove("dark");
    writeLS(LS_THEME, theme);
    writeLS("app-theme", theme);
  }, [theme]);

  useEffect(() => {
    const custom = getCustomTheme();
    if (custom) {
      applyKiwiTheme(custom);
      return;
    }

    const saved = readLS(lsPreset(), "");
    if (saved || serverPreset) return;

    api.getTheme().then((data) => {
      const name = data?.preset as string | undefined;
      if (name) {
        const found = findPreset(name, availablePresets);
        if (found) {
          setPresetState(name);
          applyKiwiTheme(presetToOverrides(found));
          return;
        }
      }
      if (data?.light || data?.dark) {
        applyKiwiTheme(data as KiwiThemeOverrides);
      }
    }).catch(() => {});
  }, [availablePresets, serverPreset]);

  useEffect(() => {
    const custom = getCustomTheme();
    if (custom) {
      applyKiwiTheme(custom);
      return;
    }
    const found = resolvePreset(preset);
    if (found) {
      applyKiwiTheme(presetToOverrides(found));
    } else {
      removeKiwiTheme();
    }
  }, [preset, resolvePreset]);

  useEffect(() => {
    api.getCustomCSS().then(applyKiwiCustomCSS).catch(() => {});
  }, []);

  useEffect(() => {
    return onSpaceChange(() => {
      api.getCustomCSS().then(applyKiwiCustomCSS).catch(() => {});
    });
  }, []);

  useEffect(() => {
    return onSpaceChange(() => {
      const custom = getCustomTheme();
      if (custom) {
        applyKiwiTheme(custom);
        return;
      }
      const saved = readLS(lsPreset(), "");
      if (saved) {
        setPresetState(saved);
        const found = findPreset(saved, availablePresets);
        if (found) applyKiwiTheme(presetToOverrides(found));
        else removeKiwiTheme();
        return;
      }
      api.getTheme().then((data) => {
        const name = data?.preset as string | undefined;
        if (name) {
          const found = findPreset(name, availablePresets);
          if (found) {
            setPresetState(name);
            applyKiwiTheme(presetToOverrides(found));
            return;
          }
        }
        if (data?.light || data?.dark) {
          applyKiwiTheme(data as KiwiThemeOverrides);
        } else {
          removeKiwiTheme();
        }
      }).catch(() => {});
    });
  }, [availablePresets]);

  const toggleTheme = useCallback(() => {
    guardedThemeAction(themeLocked, () => {
      const ext = externalThemeAPI();
      if (ext) {
        ext.toggle();
      } else {
        setTheme((t) => (t === "dark" ? "light" : "dark"));
      }
    });
  }, [themeLocked]);

  const setPreset = useCallback((name: string) => {
    guardedThemeAction(themeLocked, () => {
      setCustomTheme(null);
      setPresetState(name);
      writeLS(lsPreset(), name);
      onPresetChange?.(name);
      const found = findPreset(name, availablePresets);
      if (found) {
        api.putTheme({ preset: name, ...presetToOverrides(found) } as unknown as Record<string, unknown>).catch(() => {});
      }
    });
  }, [themeLocked, onPresetChange, availablePresets]);

  return {
    theme,
    toggleTheme,
    preset,
    setPreset,
    presets: availablePresets,
    presetErrors,
    themeLocked,
  };
}

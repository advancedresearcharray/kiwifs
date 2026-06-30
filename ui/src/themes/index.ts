import type { KiwiThemeOverrides } from "../lib/kiwiTheme";

import kiwi from "./kiwi.json";
import neutral from "./neutral.json";
import ocean from "./ocean.json";
import sunset from "./sunset.json";
import forest from "./forest.json";

export interface ThemePreset {
  name: string;
  description: string;
  light: Record<string, string>;
  dark: Record<string, string>;
}

export const presets: ThemePreset[] = [kiwi, neutral, ocean, sunset, forest];

export function presetToOverrides(preset: ThemePreset): KiwiThemeOverrides {
  return { light: preset.light, dark: preset.dark };
}

export function findPreset(name: string, list: ThemePreset[] = presets): ThemePreset | undefined {
  return list.find((p) => p.name.toLowerCase() === name.toLowerCase());
}

/** Merge built-in presets with workspace-defined presets (built-in wins on name clash). */
export function mergePresets(builtin: ThemePreset[], workspace: ThemePreset[]): ThemePreset[] {
  const seen = new Set(builtin.map((p) => p.name.toLowerCase()));
  const merged = [...builtin];
  for (const p of workspace) {
    const key = p.name.toLowerCase();
    if (!seen.has(key)) {
      merged.push(p);
      seen.add(key);
    }
  }
  return merged;
}

/** When allowed is non-empty, keep only presets whose names match (case-insensitive). */
export function filterPresets(list: ThemePreset[], allowed: string[] | undefined): ThemePreset[] {
  if (!allowed || allowed.length === 0) return list;
  const set = new Set(allowed.map((a) => a.trim().toLowerCase()).filter(Boolean));
  if (set.size === 0) return list;
  return list.filter((p) => set.has(p.name.toLowerCase()));
}

/** Apply configured allow-list; empty allow-list keeps all presets. */
export function filterPresetsWithAllowList(list: ThemePreset[], allowed: string[]): ThemePreset[] {
  return filterPresets(list, allowed.length > 0 ? allowed : undefined);
}

/** Pick a valid preset name after filtering; falls back to the first available preset. */
export function resolvePresetName(
  current: string,
  available: ThemePreset[],
): string {
  if (available.length === 0) return current;
  if (findPreset(current, available)) return current;
  return available[0].name;
}

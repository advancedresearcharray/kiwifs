import { describe, expect, it } from "vitest";
import {
  filterPresets,
  filterPresetsWithAllowList,
  findPreset,
  mergePresets,
  presets,
  resolvePresetName,
  type ThemePreset,
} from "./index";

describe("theme presets", () => {
  it("finds presets case-insensitively", () => {
    expect(findPreset("kiwi")?.name).toBe("Kiwi");
    expect(findPreset("OCEAN")?.name).toBe("Ocean");
  });

  it("merges workspace presets after built-ins without duplicates", () => {
    const workspace: ThemePreset[] = [
      {
        name: "Corporate",
        description: "Brand",
        light: { background: "#fff" },
        dark: { background: "#111" },
      },
      {
        name: "Kiwi",
        description: "Should not override built-in",
        light: { background: "#000" },
        dark: { background: "#eee" },
      },
    ];
    const merged = mergePresets(presets, workspace);
    expect(merged).toHaveLength(presets.length + 1);
    expect(findPreset("kiwi", merged)?.light.background).toMatch(/hsl/);
    expect(findPreset("corporate", merged)?.description).toBe("Brand");
  });

  it("filters presets by allowed names", () => {
    const all = mergePresets(presets, [
      {
        name: "Corporate",
        description: "",
        light: { background: "#fff" },
        dark: { background: "#111" },
      },
    ]);
    const filtered = filterPresets(all, ["kiwi", "corporate"]);
    expect(filtered.map((p) => p.name)).toEqual(["Kiwi", "Corporate"]);
    expect(filterPresets(all, [])).toEqual(all);
  });

  it("filterPresetsWithAllowList applies allow-list for API failure fallback", () => {
    const filtered = filterPresetsWithAllowList(presets, ["kiwi", "ocean"]);
    expect(filtered.map((p) => p.name)).toEqual(["Kiwi", "Ocean"]);
    expect(filterPresetsWithAllowList(presets, [])).toEqual(presets);
  });

  it("resolves preset name when current choice is not allowed", () => {
    const available = filterPresets(presets, ["kiwi", "ocean"]);
    expect(resolvePresetName("Forest", available)).toBe("Kiwi");
    expect(resolvePresetName("Ocean", available)).toBe("Ocean");
    expect(resolvePresetName("Missing", [])).toBe("Missing");
  });
});

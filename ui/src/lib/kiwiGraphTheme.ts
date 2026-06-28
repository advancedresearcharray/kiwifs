/**
 * Graph colors are defined in `src/styles/kiwi-theme.css` (--kiwi-graph-*).
 * This module reads the resolved values for WebGL (Sigma).
 */

export type KiwiGraphTheme = {
  defaultNode: string;
  edge: string;
  nodeDim: string;
  edgeGhost: string;
  edgeStrong: string;
  palette: string[];
};

const PALETTE_VAR_NAMES = Array.from(
  { length: 10 },
  (_, i) => `--kiwi-graph-palette-${i}`,
);

const DEFAULTS: KiwiGraphTheme = {
  defaultNode: "#7c8a6e",
  edge: "#c8c8c8",
  nodeDim: "#e0e0e0",
  edgeGhost: "#ececec",
  edgeStrong: "#555555",
  palette: [
    "#5b9e4f", "#4a89c8", "#d97b3e", "#c254a5", "#3db5a6",
    "#c9534e", "#8b6cc1", "#c4a832", "#4eadd4", "#7a8f3e",
  ],
};

function pick(cs: CSSStyleDeclaration, name: string, fallback: string) {
  const v = cs.getPropertyValue(name).trim();
  return v || fallback;
}

export function readKiwiGraphTheme(
  el: Element = document.documentElement,
): KiwiGraphTheme {
  const cs = getComputedStyle(el);
  const pal = PALETTE_VAR_NAMES.map((name, i) =>
    pick(cs, name, DEFAULTS.palette[i] ?? "#808080"),
  );
  return {
    defaultNode: pick(cs, "--kiwi-graph-default-node", DEFAULTS.defaultNode),
    edge: pick(cs, "--kiwi-graph-edge", DEFAULTS.edge),
    nodeDim: pick(cs, "--kiwi-graph-node-dim", DEFAULTS.nodeDim),
    edgeGhost: pick(cs, "--kiwi-graph-edge-ghost", DEFAULTS.edgeGhost),
    edgeStrong: pick(cs, "--kiwi-graph-edge-strong", DEFAULTS.edgeStrong),
    palette: pal,
  };
}

export function colorForGraphCommunity(
  i: number,
  theme: KiwiGraphTheme,
): string {
  if (i < theme.palette.length) return theme.palette[i]!;

  // Use a golden-angle hue walk rather than grayscale fallback. Large vaults can
  // produce dozens of Louvain communities, so falling back to neutral colors
  // makes most clusters look black/gray on the dark graph background. Emit hex
  // instead of CSS Color 4 HSL because Sigma/WebGL color parsers are stricter
  // than the DOM and can render unsupported strings as black.
  const hue = (i * 137.508) % 360;
  const saturation = 0.72;
  const lightness = 0.58;
  const chroma = (1 - Math.abs(2 * lightness - 1)) * saturation;
  const x = chroma * (1 - Math.abs(((hue / 60) % 2) - 1));
  const m = lightness - chroma / 2;
  let r = 0;
  let g = 0;
  let b = 0;

  if (hue < 60) [r, g, b] = [chroma, x, 0];
  else if (hue < 120) [r, g, b] = [x, chroma, 0];
  else if (hue < 180) [r, g, b] = [0, chroma, x];
  else if (hue < 240) [r, g, b] = [0, x, chroma];
  else if (hue < 300) [r, g, b] = [x, 0, chroma];
  else [r, g, b] = [chroma, 0, x];

  const toHex = (v: number) => Math.round((v + m) * 255).toString(16).padStart(2, "0");
  return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
}

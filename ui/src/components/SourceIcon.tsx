/**
 * SourceIcon — brand-colored inline SVG icons for import source types.
 * Uses `simple-icons` for brand logos; hand-drawn SVG for sources without entries.
 * Unknown types use a letter-circle fallback.
 */

import {
  siMarkdown,
  siFirebase,
  siPostgresql,
  siMysql,
  siMongodb,
  siNotion,
  siAirtable,
  siLibreofficecalc,
  siSqlite,
  siJson,
  siYaml,
  siObsidian,
} from "simple-icons";

interface SourceIconInfo {
  path: string;
  hex: string;
  title: string;
}

interface SourceIconDef {
  icon: SourceIconInfo;
  colorOverride?: string;
}

/** Microsoft Excel — spreadsheet grid icon */
const EXCEL_ICON: SourceIconInfo = {
  path: "M23 1.5H6.5c-.28 0-.5.22-.5.5v4H1c-.28 0-.5.22-.5.5v11c0 .28.22.5.5.5h5v4c0 .28.22.5.5.5H23c.28 0 .5-.22.5-.5v-20c0-.28-.22-.5-.5-.5zM7 21.5V18h3v3.5H7zm0-4V14h3v3.5H7zm0-4V10h3v3.5H7zm4 8V18h3.5v3.5H11zm0-4V14h3.5v3.5H11zm0-4V10h3.5v3.5H11zM5.5 17H1.5V7h4v10zM15 21.5V18h7.5v3.5H15zm0-4V14h7.5v3.5H15zm0-4V10h7.5v3.5H15zm-8-4V2.5h15.5V9.5H7z",
  hex: "217346",
  title: "Microsoft Excel",
};

const SOURCE_ICON_MAP: Record<string, SourceIconDef> = {
  // Builtin file sources
  markdown:        { icon: siMarkdown,        colorOverride: "5B8DEE" },
  obsidian:        { icon: siObsidian },
  csv:             { icon: siLibreofficecalc },
  json:            { icon: siJson,            colorOverride: "5B8DEE" },
  jsonl:           { icon: siJson,            colorOverride: "5B8DEE" },
  yaml:            { icon: siYaml },
  excel:           { icon: EXCEL_ICON },
  sqlite:          { icon: siSqlite },

  // Native network
  postgres:        { icon: siPostgresql },
  mysql:           { icon: siMysql },
  mongodb:         { icon: siMongodb },

  // Airbyte-powered
  firestore:       { icon: siFirebase },
  "firebase-rtdb": { icon: siFirebase,        colorOverride: "FFA000" },
  notion:          { icon: siNotion,          colorOverride: "787774" },
  airtable:        { icon: siAirtable },
};

/**
 * Renders a brand-colored inline SVG icon for a data source type.
 * Falls back to a neutral gray circle with the first letter for unknown types.
 */
export function SourceIcon({
  source,
  size = 24,
  className,
}: {
  source: string;
  size?: number;
  className?: string;
}) {
  const def = SOURCE_ICON_MAP[source];

  if (!def) {
    return (
      <svg
        className={className}
        style={{ width: size, height: size }}
        viewBox="0 0 24 24"
        xmlns="http://www.w3.org/2000/svg"
      >
        <circle cx="12" cy="12" r="12" fill="#6B7280" opacity="0.2" />
        <text x="12" y="16.5" textAnchor="middle" fontSize="13" fontWeight="600" fill="#6B7280">
          {source.charAt(0).toUpperCase()}
        </text>
      </svg>
    );
  }

  const color = `#${def.colorOverride || def.icon.hex}`;

  return (
    <svg
      className={className}
      style={{ width: size, height: size }}
      role="img"
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
    >
      <title>{def.icon.title}</title>
      <path d={def.icon.path} fill={color} />
    </svg>
  );
}

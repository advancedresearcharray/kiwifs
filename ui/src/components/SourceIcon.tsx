/**
 * SourceIcon — brand-colored inline SVG icons for data source types.
 *
 * Uses the `simple-icons` npm package (same approach as apps/web's file-icon.tsx)
 * to render crisp, scalable brand logos as inline <svg> elements — no CDN calls,
 * no image loading, instant render, works offline.
 *
 * Covers all 19 source types supported by the kiwifs backend:
 *   API-supported (11): markdown, postgres, mysql, firestore, sqlite, mongodb,
 *                       csv, json, jsonl, notion, airtable
 *   CLI-only (8):       yaml, excel, gsheets, obsidian, confluence, dynamodb,
 *                       redis, elasticsearch
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
  siGooglesheets,
  siObsidian,
  siConfluence,
  siRedis,
  siElasticsearch,
} from "simple-icons";

interface SourceIconInfo {
  path: string;
  hex: string;
  title: string;
}

interface SourceIconDef {
  icon: SourceIconInfo;
  /** Override the simple-icons default hex color (without #) */
  colorOverride?: string;
}

// DynamoDB and Excel have no simple-icons entry. We provide hand-drawn SVG
// paths so they fit the same inline-SVG pattern as the rest.

/** AWS DynamoDB — simplified table icon */
const DYNAMODB_ICON: SourceIconInfo = {
  path: "M12 2L2 7v10l10 5 10-5V7L12 2zm0 2.18L19.35 7.5 12 10.82 4.65 7.5 12 4.18zM3.5 8.35l8 4v8.3l-8-4v-8.3zm17 0v8.3l-8 4v-8.3l8-4z",
  hex: "4053D6",
  title: "Amazon DynamoDB",
};

/** Microsoft Excel — spreadsheet grid icon */
const EXCEL_ICON: SourceIconInfo = {
  path: "M23 1.5H6.5c-.28 0-.5.22-.5.5v4H1c-.28 0-.5.22-.5.5v11c0 .28.22.5.5.5h5v4c0 .28.22.5.5.5H23c.28 0 .5-.22.5-.5v-20c0-.28-.22-.5-.5-.5zM7 21.5V18h3v3.5H7zm0-4V14h3v3.5H7zm0-4V10h3v3.5H7zm4 8V18h3.5v3.5H11zm0-4V14h3.5v3.5H11zm0-4V10h3.5v3.5H11zM5.5 17H1.5V7h4v10zM15 21.5V18h7.5v3.5H15zm0-4V14h7.5v3.5H15zm0-4V10h7.5v3.5H15zm-8-4V2.5h15.5V9.5H7z",
  hex: "217346",
  title: "Microsoft Excel",
};

const SOURCE_ICON_MAP: Record<string, SourceIconDef> = {
  // ── API-supported (11) ────────────────────────────────────────────
  markdown:      { icon: siMarkdown,        colorOverride: "5B8DEE" },
  firestore:     { icon: siFirebase },
  postgres:      { icon: siPostgresql },
  mysql:         { icon: siMysql },
  mongodb:       { icon: siMongodb },
  notion:        { icon: siNotion,          colorOverride: "787774" },
  airtable:      { icon: siAirtable },
  csv:           { icon: siLibreofficecalc },
  sqlite:        { icon: siSqlite },
  json:          { icon: siJson,            colorOverride: "5B8DEE" },
  jsonl:         { icon: siJson,            colorOverride: "5B8DEE" },

  // ── CLI-only (8) ─────────────────────────────────────────────────
  yaml:          { icon: siYaml },
  excel:         { icon: EXCEL_ICON },
  gsheets:       { icon: siGooglesheets },
  obsidian:      { icon: siObsidian },
  confluence:    { icon: siConfluence },
  dynamodb:      { icon: DYNAMODB_ICON },
  redis:         { icon: siRedis },
  elasticsearch: { icon: siElasticsearch },
};

/**
 * Renders a brand-colored inline SVG icon for a data source type.
 *
 * Falls back to a neutral gray circle with the first letter when the
 * source type has no registered icon.
 */
export function SourceIcon({
  source,
  size = 24,
  className,
}: {
  /** Source type key, e.g. "postgres", "firestore" */
  source: string;
  /** Icon size in pixels */
  size?: number;
  className?: string;
}) {
  const def = SOURCE_ICON_MAP[source];

  // Unknown source type — first-letter fallback
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

// Kanban UI helpers — tag colors, priority styling, avatar generation.

// ---------------------------------------------------------------------------
// Tag color palette
// ---------------------------------------------------------------------------

/** Well-known tag → color presets (lowercase match). */
const TAG_PRESETS: Record<string, { bg: string; fg: string }> = {
  bug: { bg: "#FECACA", fg: "#991B1B" },
  fix: { bg: "#FECACA", fg: "#991B1B" },
  hotfix: { bg: "#FECACA", fg: "#991B1B" },
  feature: { bg: "#BFDBFE", fg: "#1E3A5F" },
  feat: { bg: "#BFDBFE", fg: "#1E3A5F" },
  enhancement: { bg: "#BFDBFE", fg: "#1E3A5F" },
  improvement: { bg: "#C7D2FE", fg: "#312E81" },
  docs: { bg: "#DDD6FE", fg: "#4C1D95" },
  documentation: { bg: "#DDD6FE", fg: "#4C1D95" },
  test: { bg: "#FDE68A", fg: "#78350F" },
  testing: { bg: "#FDE68A", fg: "#78350F" },
  refactor: { bg: "#E0E7FF", fg: "#3730A3" },
  chore: { bg: "#E5E7EB", fg: "#374151" },
  design: { bg: "#FBCFE8", fg: "#831843" },
  ui: { bg: "#FBCFE8", fg: "#831843" },
  ux: { bg: "#FBCFE8", fg: "#831843" },
  backend: { bg: "#D1FAE5", fg: "#065F46" },
  frontend: { bg: "#CFFAFE", fg: "#155E75" },
  api: { bg: "#D1FAE5", fg: "#065F46" },
  security: { bg: "#FEE2E2", fg: "#7F1D1D" },
  performance: { bg: "#FEF3C7", fg: "#78350F" },
  perf: { bg: "#FEF3C7", fg: "#78350F" },
  urgent: { bg: "#FEE2E2", fg: "#991B1B" },
  critical: { bg: "#FEE2E2", fg: "#991B1B" },
  blocked: { bg: "#FEE2E2", fg: "#991B1B" },
  wip: { bg: "#FEF9C3", fg: "#713F12" },
  "in-progress": { bg: "#FEF9C3", fg: "#713F12" },
  draft: { bg: "#F3F4F6", fg: "#4B5563" },
  review: { bg: "#EDE9FE", fg: "#5B21B6" },
  approved: { bg: "#D1FAE5", fg: "#065F46" },
  rejected: { bg: "#FEE2E2", fg: "#991B1B" },
  question: { bg: "#DBEAFE", fg: "#1E40AF" },
  help: { bg: "#DBEAFE", fg: "#1E40AF" },
  idea: { bg: "#FEF3C7", fg: "#92400E" },
  research: { bg: "#E0E7FF", fg: "#3730A3" },
  spike: { bg: "#E0E7FF", fg: "#3730A3" },
  devops: { bg: "#CCFBF1", fg: "#134E4A" },
  infra: { bg: "#CCFBF1", fg: "#134E4A" },
  ci: { bg: "#CCFBF1", fg: "#134E4A" },
  "breaking-change": { bg: "#FEE2E2", fg: "#991B1B" },
  deprecated: { bg: "#F3F4F6", fg: "#6B7280" },
  v1: { bg: "#DBEAFE", fg: "#1E40AF" },
  v2: { bg: "#C7D2FE", fg: "#4338CA" },
  mvp: { bg: "#D1FAE5", fg: "#065F46" },
};

/** Hashed fallback palette for tags without a preset. */
const HASH_PALETTE: { bg: string; fg: string }[] = [
  { bg: "#DBEAFE", fg: "#1E40AF" }, // blue
  { bg: "#D1FAE5", fg: "#065F46" }, // emerald
  { bg: "#EDE9FE", fg: "#5B21B6" }, // violet
  { bg: "#FEF3C7", fg: "#92400E" }, // amber
  { bg: "#FBCFE8", fg: "#831843" }, // pink
  { bg: "#CCFBF1", fg: "#134E4A" }, // teal
  { bg: "#FEE2E2", fg: "#991B1B" }, // rose
  { bg: "#E0E7FF", fg: "#3730A3" }, // indigo
  { bg: "#CFFAFE", fg: "#155E75" }, // cyan
  { bg: "#FDE68A", fg: "#78350F" }, // yellow
  { bg: "#D9F99D", fg: "#365314" }, // lime
  { bg: "#C4B5FD", fg: "#4C1D95" }, // purple
];

function hashString(s: string): number {
  let hash = 0;
  for (let i = 0; i < s.length; i++) {
    hash = (hash * 31 + s.charCodeAt(i)) | 0;
  }
  return Math.abs(hash);
}

export function tagColor(tag: string): { bg: string; fg: string } {
  const lower = tag.toLowerCase().trim();
  if (TAG_PRESETS[lower]) return TAG_PRESETS[lower];
  return HASH_PALETTE[hashString(lower) % HASH_PALETTE.length]!;
}

// ---------------------------------------------------------------------------
// Priority config
// ---------------------------------------------------------------------------

export type PriorityLevel = "critical" | "high" | "medium" | "low" | "none";

const PRIORITY_MAP: Record<string, PriorityLevel> = {
  critical: "critical",
  urgent: "critical",
  p0: "critical",
  high: "high",
  p1: "high",
  important: "high",
  medium: "medium",
  normal: "medium",
  p2: "medium",
  low: "low",
  minor: "low",
  p3: "low",
  trivial: "low",
  none: "none",
  p4: "none",
};

export type PriorityStyle = {
  label: string;
  dotColor: string;
  textColor: string;
  bgColor: string;
};

const PRIORITY_STYLES: Record<PriorityLevel, PriorityStyle> = {
  critical: {
    label: "Critical",
    dotColor: "#DC2626",
    textColor: "#991B1B",
    bgColor: "#FEE2E2",
  },
  high: {
    label: "High",
    dotColor: "#EA580C",
    textColor: "#9A3412",
    bgColor: "#FFF7ED",
  },
  medium: {
    label: "Medium",
    dotColor: "#CA8A04",
    textColor: "#854D0E",
    bgColor: "#FEFCE8",
  },
  low: {
    label: "Low",
    dotColor: "#16A34A",
    textColor: "#166534",
    bgColor: "#F0FDF4",
  },
  none: {
    label: "None",
    dotColor: "#9CA3AF",
    textColor: "#6B7280",
    bgColor: "#F9FAFB",
  },
};

export function parsePriority(raw?: string): PriorityLevel | null {
  if (!raw) return null;
  const key = raw.toLowerCase().trim();
  return PRIORITY_MAP[key] ?? null;
}

export function priorityStyle(level: PriorityLevel): PriorityStyle {
  return PRIORITY_STYLES[level];
}

// ---------------------------------------------------------------------------
// Avatar helpers
// ---------------------------------------------------------------------------

const AVATAR_PALETTE = [
  "#6366F1", // indigo
  "#8B5CF6", // violet
  "#EC4899", // pink
  "#EF4444", // red
  "#F97316", // orange
  "#EAB308", // yellow
  "#22C55E", // green
  "#14B8A6", // teal
  "#06B6D4", // cyan
  "#3B82F6", // blue
];

export function authorInitials(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length === 0 || parts[0] === "") return "?";
  if (parts.length === 1) return parts[0]!.charAt(0).toUpperCase();
  return (parts[0]!.charAt(0) + parts[parts.length - 1]!.charAt(0)).toUpperCase();
}

export function authorColor(name: string): string {
  return AVATAR_PALETTE[hashString(name) % AVATAR_PALETTE.length]!;
}

// ---------------------------------------------------------------------------
// Due-date helpers
// ---------------------------------------------------------------------------

export type DueStatus = "overdue" | "due-soon" | "upcoming" | "no-date";

/**
 * Parse an ISO date string and classify how urgent it is relative to now.
 * "due-soon" = within 2 days. "overdue" = in the past.
 */
export function dueStatus(dueStr?: string): { status: DueStatus; date: Date | null } {
  if (!dueStr) return { status: "no-date", date: null };
  const date = new Date(dueStr);
  if (isNaN(date.getTime())) return { status: "no-date", date: null };

  const now = new Date();
  // Strip time portion for date comparison.
  const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const dueDay = new Date(date.getFullYear(), date.getMonth(), date.getDate());
  const diffMs = dueDay.getTime() - todayStart.getTime();
  const diffDays = diffMs / (1000 * 60 * 60 * 24);

  if (diffDays < 0) return { status: "overdue", date };
  if (diffDays <= 2) return { status: "due-soon", date };
  return { status: "upcoming", date };
}

export function dueStatusColor(status: DueStatus): { text: string; bg: string } {
  switch (status) {
    case "overdue":
      return { text: "#DC2626", bg: "#FEE2E2" };
    case "due-soon":
      return { text: "#D97706", bg: "#FEF3C7" };
    case "upcoming":
      return { text: "#6B7280", bg: "transparent" };
    case "no-date":
      return { text: "#9CA3AF", bg: "transparent" };
  }
}

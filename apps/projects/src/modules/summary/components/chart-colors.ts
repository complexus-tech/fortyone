export const SUMMARY_CHART_PRIMARY = "var(--color-primary)";
export const SUMMARY_CHART_GRID = "var(--color-border)";
export const SUMMARY_CHART_SURFACE = "var(--color-surface)";
export const SUMMARY_CHART_CURSOR = "var(--color-state-hover)";

export const SUMMARY_STATUS_COLORS = [
  SUMMARY_CHART_PRIMARY,
  "var(--color-info)",
  "var(--color-success)",
  "var(--color-warning)",
  "var(--color-secondary)",
  "var(--color-danger)",
  "var(--color-text-muted)",
] as const;

const SUMMARY_PRIORITY_COLORS: Readonly<Record<string, string>> = {
  urgent: "var(--color-danger)",
  high: "var(--color-warning)",
  medium: "var(--color-success)",
  low: "var(--color-info)",
  "no priority": "var(--color-text-muted)",
};

export const getSummaryPriorityColor = (priority: string) =>
  SUMMARY_PRIORITY_COLORS[priority.trim().toLowerCase()] ??
  SUMMARY_CHART_PRIMARY;

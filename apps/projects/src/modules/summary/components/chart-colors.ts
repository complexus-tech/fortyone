export const SUMMARY_CHART_PRIMARY = "var(--color-primary)";
export const SUMMARY_CHART_GRID = "var(--color-border)";
export const SUMMARY_CHART_SURFACE = "var(--color-surface)";
export const SUMMARY_CHART_CURSOR = "var(--color-state-hover)";

export const SUMMARY_STATUS_COLORS = [
  SUMMARY_CHART_PRIMARY,
  "color-mix(in oklab, var(--color-primary) 88%, black)",
  "color-mix(in oklab, var(--color-primary) 62%, var(--color-surface))",
  "color-mix(in oklab, var(--color-primary) 48%, var(--color-surface))",
  "color-mix(in oklab, var(--color-primary) 36%, var(--color-surface))",
  "color-mix(in oklab, var(--color-primary) 26%, var(--color-surface))",
  "color-mix(in oklab, var(--color-primary) 18%, var(--color-surface))",
] as const;

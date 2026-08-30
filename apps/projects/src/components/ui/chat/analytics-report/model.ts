import { format } from "date-fns";

export type AnalyticsReportOutput = Record<string, unknown>;

export type ChartRow = Record<string, string | number | null | undefined>;

export type Metric = {
  label: string;
  value: string | number;
};

export type SprintBurndownPoint = {
  date: string;
  ideal: number;
  remaining: number;
};

export const COLORS = {
  muted: "#94A3B8",
  primary: "#6366F1",
  success: "#22C55E",
  warning: "#F59E0B",
};

export const isRecord = (value: unknown): value is Record<string, unknown> =>
  Boolean(value) && typeof value === "object" && !Array.isArray(value);

export const asRows = (value: unknown): ChartRow[] =>
  Array.isArray(value) ? (value as ChartRow[]) : [];

export const asRecord = (value: unknown): Record<string, unknown> =>
  isRecord(value) ? value : {};

export const asSingleSprintBurndown = (value: unknown): SprintBurndownPoint[] =>
  asRows(value).flatMap((row) => {
    const date = typeof row.date === "string" ? row.date : "";
    const ideal = Number(row.ideal);
    const remaining = Number(row.remaining);
    const parsedDate = new Date(date);

    if (
      !date ||
      Number.isNaN(parsedDate.getTime()) ||
      !Number.isFinite(ideal) ||
      !Number.isFinite(remaining)
    ) {
      return [];
    }

    return [{ date, ideal, remaining }];
  });

export const asWorkingDays = (value: unknown): number[] | undefined => {
  if (!Array.isArray(value)) return undefined;

  const workingDays = value.filter(
    (day): day is number =>
      typeof day === "number" && Number.isInteger(day) && day >= 1 && day <= 7,
  );

  return workingDays.length ? workingDays : undefined;
};

export const completionRate = (completed: unknown, total: unknown) => {
  const completedNumber = Number(completed ?? 0);
  const totalNumber = Number(total ?? 0);
  if (!totalNumber) return "0%";
  return `${Math.round((completedNumber / totalNumber) * 100)}%`;
};

export const ratioPercent = (value: unknown) =>
  `${Math.round(Number(value ?? 0) * 100)}%`;

export const hasPositiveMetric = (
  record: Record<string, unknown>,
  keys: string[],
) => keys.some((key) => Number(record[key] ?? 0) > 0);

export const humanizeLabel = (value: unknown) =>
  String(value ?? "")
    .replaceAll(/[_-]+/g, " ")
    .replaceAll(/\b\w/g, (character) => character.toUpperCase());

export const progressSummary = (completed: unknown, total: unknown) =>
  `${completionRate(completed, total)} · ${Number(completed ?? 0)} of ${Number(
    total ?? 0,
  )}`;

export const formatChartDate = (value: unknown) => {
  const date = new Date(String(value));
  return Number.isNaN(date.getTime()) ? String(value) : format(date, "MMM d");
};

export const formatCategoryTick = (value: unknown, maxLength = 15) => {
  const label = String(value ?? "");
  return label.length > maxLength ? `${label.slice(0, maxLength - 1)}…` : label;
};

export type TimeNeededUnit = "minutes" | "hours";

export const DEFAULT_TIME_NEEDED_MINUTES = 60;
export const TIME_NEEDED_PRESETS = [15, 30, 60, 120, 240, 480] as const;
export const MAX_TIME_NEEDED_MINUTES = 40 * 60;

const normalizePositiveMinutes = (value: number | null | undefined) => {
  if (
    !Number.isFinite(value) ||
    !value ||
    value <= 0 ||
    value > MAX_TIME_NEEDED_MINUTES
  )
    return null;
  return Math.round(value);
};

export const formatTimeNeeded = (
  minutes: number | null | undefined,
  mode: "compact" | "full" = "compact",
) => {
  const normalizedMinutes = normalizePositiveMinutes(minutes);
  if (!normalizedMinutes) {
    return mode === "full" ? "No time needed" : "Time needed";
  }

  const hours = Math.floor(normalizedMinutes / 60);
  const remainingMinutes = normalizedMinutes % 60;

  if (mode === "compact") {
    if (!hours) return `${remainingMinutes}m`;
    if (!remainingMinutes) return `${hours}h`;
    return `${hours}h ${remainingMinutes}m`;
  }

  const parts: string[] = [];
  if (hours) parts.push(`${hours} ${hours === 1 ? "hour" : "hours"}`);
  if (remainingMinutes) {
    parts.push(
      `${remainingMinutes} ${remainingMinutes === 1 ? "minute" : "minutes"}`,
    );
  }
  return parts.join(" ");
};

export const parseTimeNeededInput = (input: string, unit: TimeNeededUnit) => {
  const value = Number(input.trim());
  if (!Number.isFinite(value) || value <= 0) return null;

  const minutes = unit === "hours" ? value * 60 : value;
  const roundedMinutes = Math.max(1, Math.round(minutes));
  return roundedMinutes <= MAX_TIME_NEEDED_MINUTES ? roundedMinutes : null;
};

export const normalizeTimeNeeded = ({
  estimatedDurationMinutes,
  minimumFocusBlockMinutes,
}: {
  estimatedDurationMinutes: number | null | undefined;
  minimumFocusBlockMinutes: number | null | undefined;
}) => {
  const duration = normalizePositiveMinutes(estimatedDurationMinutes);
  const focusBlock = normalizePositiveMinutes(minimumFocusBlockMinutes);

  return {
    estimatedDurationMinutes: duration,
    minimumFocusBlockMinutes:
      duration && focusBlock && focusBlock <= duration ? focusBlock : null,
  };
};

export const normalizeTimeNeededPatch = (
  estimatedDurationMinutes: number | null | undefined,
  minimumFocusBlockMinutes: number | null | undefined,
) => ({
  estimatedDurationMinutes,
  minimumFocusBlockMinutes:
    estimatedDurationMinutes === null ? null : minimumFocusBlockMinutes,
});

import type { ImportTask } from "./schema";
import {
  isValidImportDurationMinutes,
  isValidImportEstimateValue,
} from "./schema";

export const getCanonicalImportEffort = (task: ImportTask) => {
  const rawEstimate =
    task.canonical?.estimateValue === undefined
      ? task.estimateValue
      : task.canonical.estimateValue;
  const rawDuration =
    task.canonical?.estimatedDurationMinutes === undefined
      ? task.estimatedDurationMinutes
      : task.canonical.estimatedDurationMinutes;
  const rawFocus =
    task.canonical?.minimumFocusBlockMinutes === undefined
      ? task.minimumFocusBlockMinutes
      : task.canonical.minimumFocusBlockMinutes;
  const estimateValue = isValidImportEstimateValue(rawEstimate)
    ? rawEstimate
    : undefined;
  const estimatedDurationMinutes = isValidImportDurationMinutes(rawDuration)
    ? rawDuration
    : undefined;
  const minimumFocusBlockMinutes =
    isValidImportDurationMinutes(rawFocus) &&
    estimatedDurationMinutes !== undefined &&
    rawFocus <= estimatedDurationMinutes
      ? rawFocus
      : undefined;
  const unsupported =
    task.canonical &&
    ((rawEstimate !== null && estimateValue === undefined) ||
      (rawDuration !== null && estimatedDurationMinutes === undefined) ||
      (rawFocus !== null && minimumFocusBlockMinutes === undefined));
  return {
    estimateValue,
    estimatedDurationMinutes,
    minimumFocusBlockMinutes,
    unsupported: Boolean(unsupported),
  };
};

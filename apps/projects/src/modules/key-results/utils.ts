import type { KeyResult } from "@/modules/objectives/types";

const numberFormatter = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 2,
});

export const getKeyResultProgress = (
  keyResult: Pick<
    KeyResult,
    "measurementType" | "startValue" | "currentValue" | "targetValue"
  >,
) => {
  if (keyResult.measurementType === "boolean") {
    return keyResult.currentValue === 1 ? 100 : 0;
  }

  const targetChange = keyResult.targetValue - keyResult.startValue;
  if (targetChange === 0) {
    return keyResult.currentValue === keyResult.targetValue ? 100 : 0;
  }

  const currentChange = keyResult.currentValue - keyResult.startValue;
  const progress = Math.round((currentChange / targetChange) * 100);

  return Math.min(Math.max(progress, 0), 100);
};

export const formatKeyResultValue = (
  value: number,
  measurementType: KeyResult["measurementType"],
) => {
  if (measurementType === "boolean") {
    return value === 1 ? "Complete" : "Incomplete";
  }

  const formattedValue = numberFormatter.format(value);
  return measurementType === "percentage"
    ? `${formattedValue}%`
    : formattedValue;
};

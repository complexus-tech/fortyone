import type { MeasureType, NewObjectiveKeyResult, Objective } from "./types";
import type { GeneratedKeyResult } from "./schemas/key-result-generation";

const DATE_ONLY_PATTERN = /^\d{4}-\d{2}-\d{2}/;

const getDateOnly = (value?: string | null) => {
  const date = value?.match(DATE_ONLY_PATTERN)?.[0];
  if (!date) return null;

  const parsed = new Date(`${date}T00:00:00.000Z`);
  return Number.isNaN(parsed.getTime()) ||
    parsed.toISOString().slice(0, 10) !== date
    ? null
    : date;
};

const clampDate = (date: string, minimum: string, maximum: string) => {
  if (date < minimum) return minimum;
  if (date > maximum) return maximum;
  return date;
};

const clampMeasurementValue = (value: number, type: MeasureType) => {
  if (type === "boolean") return value >= 1 ? 1 : 0;
  if (type === "percentage") return Math.min(100, Math.max(0, value));
  return value;
};

export const toKeyResultCreateInput = (
  generated: GeneratedKeyResult,
  objective: Objective,
  currentDate = new Date().toISOString().slice(0, 10),
): NewObjectiveKeyResult => {
  const fallbackStartDate = getDateOnly(objective.startDate) ?? currentDate;
  const fallbackEndDate = getDateOnly(objective.endDate) ?? fallbackStartDate;
  const objectiveStartDate = fallbackStartDate;
  const objectiveEndDate =
    fallbackEndDate < objectiveStartDate ? objectiveStartDate : fallbackEndDate;
  const startDate = clampDate(
    getDateOnly(generated.startDate) ?? objectiveStartDate,
    objectiveStartDate,
    objectiveEndDate,
  );
  const endDate = clampDate(
    getDateOnly(generated.endDate) ?? objectiveEndDate,
    startDate,
    objectiveEndDate,
  );
  const startValue = clampMeasurementValue(
    generated.startValue,
    generated.measurementType,
  );
  const targetValue = clampMeasurementValue(
    generated.targetValue,
    generated.measurementType,
  );

  return {
    contributors: [],
    currentValue: startValue,
    endDate,
    lead: null,
    measurementType: generated.measurementType,
    name: generated.name.trim(),
    objectiveId: objective.id,
    startDate,
    startValue,
    targetValue,
  };
};

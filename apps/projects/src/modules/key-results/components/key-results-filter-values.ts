import type { KeyResultFilters } from "../types";
import type { KeyResultsMember } from "./key-results-member";

export type KeyResultsFilterField =
  | "leadIds"
  | "endDate"
  | "measurementTypes"
  | "teamIds"
  | "objectiveIds";

export type KeyResultsMeasurementType = NonNullable<
  KeyResultFilters["measurementTypes"]
>[number];

export const KEY_RESULT_MEASUREMENT_OPTIONS = [
  { label: "Percentage", value: "percentage" },
  { label: "Number", value: "number" },
  { label: "Complete / incomplete", value: "boolean" },
] as const satisfies readonly {
  label: string;
  value: KeyResultsMeasurementType;
}[];

export const toggleKeyResultsFilterValue = (
  values: readonly string[] | undefined,
  value: string,
) => {
  const selected = values ?? [];

  return selected.includes(value)
    ? selected.filter((item) => item !== value)
    : [...selected, value];
};

export const normalizeKeyResultsFilterValues = <T>(values: T[]) =>
  values.length > 0 ? values : undefined;

export const getKeyResultsMemberName = (member: KeyResultsMember) =>
  member.fullName.trim() || member.username || member.email || "Unknown user";

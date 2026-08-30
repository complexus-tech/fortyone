import type {
  StoriesFilter,
  StoriesFilterOperatorField,
} from "../stories-filter-types";
import { getStoriesFilterOperator } from "../stories-filter-types";
import type { OperatorOption, StoriesFilterField } from "./types";

export const EMPTY_FILTER_FIELDS: readonly StoriesFilterField[] = [];

const CONTENT_OPERATOR_OPTIONS = [
  { label: "contains", value: "contains" },
  { label: "does not contain", value: "doesNotContain" },
] as const satisfies readonly OperatorOption[];

const MULTI_VALUE_OPERATOR_OPTIONS = [
  { label: "is any of", value: "isAnyOf" },
  { label: "is not any of", value: "isNotAnyOf" },
] as const satisfies readonly OperatorOption[];

const SINGLE_VALUE_OPERATOR_OPTIONS = [
  { label: "is", value: "is" },
  { label: "is not", value: "isNot" },
] as const satisfies readonly OperatorOption[];

const DATE_OPERATOR_OPTIONS = [
  { label: "is", value: "is" },
  { label: "is on or before", value: "isOnOrBefore" },
  { label: "is on or after", value: "isOnOrAfter" },
  { label: "is not", value: "isNot" },
] as const satisfies readonly OperatorOption[];

const ASSIGNEE_PRESENCE_OPERATOR_OPTIONS = [
  { label: "is", value: "isEmpty" },
  { label: "is not", value: "isNotEmpty" },
] as const satisfies readonly OperatorOption[];

const FILTER_OPERATOR_FIELDS = new Set<StoriesFilterField>([
  "contentContains",
  "statusIds",
  "assigneeIds",
  "reporterIds",
  "priorities",
  "teamIds",
  "sprintIds",
  "labelIds",
  "estimateValues",
  "objectiveId",
  "startDate",
  "endDate",
  "hasNoAssignee",
]);

export const isFilterOperatorField = (
  field: StoriesFilterField,
): field is StoriesFilterOperatorField => FILTER_OPERATOR_FIELDS.has(field);

export const getOperatorOptions = (field: StoriesFilterOperatorField) => {
  if (field === "contentContains") return CONTENT_OPERATOR_OPTIONS;
  if (field === "startDate" || field === "endDate") {
    return DATE_OPERATOR_OPTIONS;
  }
  if (field === "objectiveId") return SINGLE_VALUE_OPERATOR_OPTIONS;
  if (field === "hasNoAssignee") {
    return ASSIGNEE_PRESENCE_OPERATOR_OPTIONS;
  }
  return MULTI_VALUE_OPERATOR_OPTIONS;
};

export const getOperatorConfig = (
  filters: StoriesFilter,
  field: StoriesFilterOperatorField,
) => {
  const operator = getStoriesFilterOperator(filters, field);
  const operatorOptions = getOperatorOptions(field);

  return {
    operator:
      operatorOptions.find((option) => option.value === operator)?.label ??
      operatorOptions[0].label,
    operatorOptions,
  };
};

export const getNames = (
  ids: string[] | null | undefined,
  labelsById: ReadonlyMap<string, string>,
) => {
  if (!ids?.length) return "";
  return ids.map((id) => labelsById.get(id) ?? id).join(", ");
};

export const normalizeArrayFilter = <T>(values: T[]) =>
  values.length > 0 ? values : null;

export const getPluralLabel = (
  count: number,
  singular: string,
  plural: string,
) => `${count} ${count === 1 ? singular : plural}`;

export const getEditorContentClassName = (field: StoriesFilterField) => {
  if (field === "contentContains") return "w-80 overflow-hidden py-2";
  if (field === "objectiveId") return "w-96 overflow-hidden py-2";
  if (field === "assigneeIds" || field === "reporterIds") {
    return "w-80 overflow-hidden py-2";
  }
  if (field === "startDate" || field === "endDate") {
    return "w-auto overflow-hidden py-2";
  }
  return "w-64 overflow-hidden py-2";
};

export const shouldFetchNextPage = (
  target: Pick<HTMLDivElement, "clientHeight" | "scrollHeight" | "scrollTop">,
  hasNextPage: boolean,
  isFetchingNextPage: boolean,
) =>
  target.scrollHeight - target.scrollTop - target.clientHeight <= 80 &&
  hasNextPage &&
  !isFetchingNextPage;

export const removeStoriesFilterField = (
  filters: StoriesFilter,
  field: StoriesFilterField,
): StoriesFilter => {
  if (field === "assignedToMe" || field === "createdByMe") {
    return { ...filters, [field]: false };
  }

  if (isFilterOperatorField(field)) {
    return {
      ...filters,
      [field]: null,
      operators: { ...filters.operators, [field]: undefined },
    };
  }

  return { ...filters, [field]: null };
};

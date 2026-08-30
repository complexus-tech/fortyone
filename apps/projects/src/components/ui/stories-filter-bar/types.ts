import type { ReactNode } from "react";
import type {
  StoriesFilter,
  StoriesFilterOperator,
} from "../stories-filter-types";

export type StoriesFilterField =
  | "contentContains"
  | "statusIds"
  | "assigneeIds"
  | "reporterIds"
  | "priorities"
  | "teamIds"
  | "sprintIds"
  | "labelIds"
  | "estimateValues"
  | "objectiveId"
  | "keyResultId"
  | "startDate"
  | "endDate"
  | "assignedToMe"
  | "createdByMe"
  | "hasNoAssignee";

export type SetStoriesFilters = (value: StoriesFilter) => void;

export type StoriesFilterEditorProps = {
  filters: StoriesFilter;
  setFilters: SetStoriesFilters;
};

export type OperatorOption = {
  label: string;
  value: StoriesFilterOperator;
};

export type FilterChip = {
  field: StoriesFilterField;
  icon?: ReactNode;
  label: string;
  operator: string;
  operatorOptions?: readonly OperatorOption[];
  value: ReactNode;
};

export type FilterOption = {
  field: StoriesFilterField;
  icon: ReactNode;
  label: string;
};

export type StoriesFilterBarProps = {
  filters: StoriesFilter;
  hiddenFields?: readonly StoriesFilterField[];
  resetFilters: () => void;
  setFilters: SetStoriesFilters;
  showWhenEmpty?: boolean;
};

export type UserChipSummary = {
  avatarUrl: string | null;
  id: string;
  name: string;
  username: string;
};

export type LabelChipSummary = {
  color: string;
  id: string;
  name: string;
};

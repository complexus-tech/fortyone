export type StoriesFilterOperator =
  | "contains"
  | "doesNotContain"
  | "isAnyOf"
  | "isNotAnyOf"
  | "is"
  | "isNot"
  | "isOnOrBefore"
  | "isOnOrAfter"
  | "isEmpty"
  | "isNotEmpty";

export type StoriesFilterOperatorField =
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
  | "startDate"
  | "endDate"
  | "hasNoAssignee";

export type StoriesFilter = {
  statusIds: string[] | null;
  assigneeIds: string[] | null;
  reporterIds: string[] | null;
  priorities: string[] | null;
  teamIds: string[] | null;
  sprintIds: string[] | null;
  labelIds: string[] | null;
  estimateValues: number[] | null;
  parentId: string | null;
  objectiveId: string | null;
  epicId: string | null;
  keyResultId: string | null;
  contentContains?: string | null;
  startDate?: string | null;
  endDate?: string | null;
  hasNoAssignee: boolean | null;
  hasBlockedBy: boolean | null;
  assignedToMe: boolean;
  createdByMe: boolean;
  completedAfter?: string | null;
  completedBefore?: string | null;
  isCompleted?: boolean | null;
  isNotCompleted?: boolean | null;
  operators?: Partial<
    Record<StoriesFilterOperatorField, StoriesFilterOperator>
  >;
};

export const DEFAULT_STORIES_FILTER: StoriesFilter = {
  statusIds: null,
  assigneeIds: null,
  reporterIds: null,
  priorities: null,
  teamIds: null,
  sprintIds: null,
  labelIds: null,
  estimateValues: null,
  parentId: null,
  objectiveId: null,
  epicId: null,
  keyResultId: null,
  contentContains: null,
  startDate: null,
  endDate: null,
  hasNoAssignee: null,
  hasBlockedBy: null,
  assignedToMe: false,
  createdByMe: false,
  operators: {},
};

const DEFAULT_FILTER_OPERATORS: Record<
  StoriesFilterOperatorField,
  StoriesFilterOperator
> = {
  contentContains: "contains",
  statusIds: "isAnyOf",
  assigneeIds: "isAnyOf",
  reporterIds: "isAnyOf",
  priorities: "isAnyOf",
  teamIds: "isAnyOf",
  sprintIds: "isAnyOf",
  labelIds: "isAnyOf",
  estimateValues: "isAnyOf",
  objectiveId: "is",
  startDate: "is",
  endDate: "is",
  hasNoAssignee: "isEmpty",
};

export const getStoriesFilterOperator = (
  filters: StoriesFilter,
  field: StoriesFilterOperatorField,
) => filters.operators?.[field] ?? DEFAULT_FILTER_OPERATORS[field];

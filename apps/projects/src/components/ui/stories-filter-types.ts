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
};

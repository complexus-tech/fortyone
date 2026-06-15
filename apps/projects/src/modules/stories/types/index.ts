import type { StateCategory } from "@/types/states";
import type { UserSummary } from "@/types";

export type StoryPriority =
  | "No Priority"
  | "Urgent"
  | "High"
  | "Medium"
  | "Low";

export type StoryTeamSummary = {
  id: string;
  name: string;
  code: string;
};

export type StoryObjectiveSummary = {
  id: string;
  name: string;
  description: string | null;
};

export type StorySprintSummary = {
  id: string;
  name: string;
  goal: string | null;
  startDate: string;
  endDate: string;
};

export type Story = {
  id: string;
  title: string;
  estimateLabel: string | null;
  estimateValue: number | null;
  estimateScheme: "points" | "hours" | "tshirt" | "ideal_days";
  description?: string;
  statusId: string;
  sprintId: string | null;
  sprint?: StorySprintSummary | null;
  objectiveId: string | null;
  objective?: StoryObjectiveSummary | null;
  keyResultId: string | null;
  teamId: string;
  team?: StoryTeamSummary | null;
  workspaceId: string;
  assigneeId: string | null;
  assignee?: UserSummary | null;
  reporterId: string;
  reporter?: UserSummary | null;
  epicId: string | null;
  sequenceId: number;
  priority: StoryPriority;
  startDate: string | null;
  endDate: string | null;
  createdAt: string;
  updatedAt: string;
  completedAt: string | null;
  deletedAt: string | null;
  archivedAt: string | null;
  labels: string[] | null;
  subStories: Story[];
};

export type StoryActivity = {
  id: string;
  storyId: string;
  userId: string;
  user: UserSummary;
  type: "update" | "create" | "link";
  field: string;
  currentValue: string;
  oldValue?: unknown;
  newValue?: unknown;
  reason?: string | null;
  createdAt: string;
};

export type StoryGroup = {
  key: string;
  loadedCount: number;
  totalCount: number;
  hasMore: boolean;
  stories: Story[];
  nextPage: number;
};

export type StoryFilters = {
  statusIds?: string[] | null;
  categories?: StateCategory[];
  assigneeIds?: string[] | null;
  reporterIds?: string[] | null;
  titleContains?: string | null;
  priorities?: string[] | null;
  teamIds?: string[] | null;
  sprintIds?: string[] | null;
  labelIds?: string[] | null;
  estimateValues?: number[] | null;
  parentId?: string | null;
  objectiveId?: string | null;
  epicId?: string | null;
  hasNoAssignee?: boolean | null;
  hasBlockedBy?: boolean | null;
  assignedToMe?: boolean;
  createdByMe?: boolean;
  createdAfter?: string;
  createdBefore?: string;
  updatedAfter?: string;
  updatedBefore?: string;
  startDateAfter?: string;
  startDateBefore?: string;
  deadlineAfter?: string;
  deadlineBefore?: string;
  includeArchived?: boolean;
  includeDeleted?: boolean;
  showSubStories?: boolean;
  completedAfter?: string;
  completedBefore?: string;
};

export type GroupedStoriesResponse = {
  groups: StoryGroup[];
  meta: {
    totalGroups: number;
    filters: StoryFilters;
    groupBy: "priority" | "status" | "assignee" | "none";
    orderBy: "created" | "updated" | "deadline" | "priority";
    orderDirection: "asc" | "desc";
  };
};

export type GroupStoriesResponse = {
  groupKey: string;
  stories: Story[];
  pagination: {
    page: number;
    pageSize: number;
    hasMore: boolean;
    nextPage: number;
  };
  filters: StoryFilters;
  orderBy: "created" | "updated" | "deadline" | "priority";
  orderDirection: "asc" | "desc";
};

export type GroupedStoryParams = {
  groupBy: "priority" | "status" | "assignee" | "none";
  orderBy?: "created" | "updated" | "deadline" | "priority";
  orderDirection?: "asc" | "desc";
  teamIds?: string[];
  categories?: StateCategory[];
  assignedToMe?: boolean;
  createdByMe?: boolean;
  storiesPerGroup?: number;
  statusIds?: string[];
  assigneeIds?: string[];
  reporterIds?: string[];
  titleContains?: string;
  priorities?: string[];
  sprintIds?: string[];
  labelIds?: string[];
  estimateValues?: number[];
  parentId?: string;
  objectiveId?: string;
  epicId?: string;
  hasNoAssignee?: boolean;
  hasBlockedBy?: boolean;
  createdAfter?: string;
  createdBefore?: string;
  updatedAfter?: string;
  updatedBefore?: string;
  startDateAfter?: string;
  startDateBefore?: string;
  deadlineAfter?: string;
  deadlineBefore?: string;
  includeArchived?: boolean;
  showSubStories?: boolean;
  completedAfter?: string;
  completedBefore?: string;
  includeDeleted?: boolean;
};

export type GroupStoryParams = {
  groupKey: string;
  groupBy: "priority" | "status" | "assignee" | "none";
  orderBy?: "created" | "updated" | "deadline" | "priority";
  orderDirection?: "asc" | "desc";
  page?: number;
  pageSize?: number;
  assignedToMe?: boolean;
  createdByMe?: boolean;
  statusIds?: string[];
  categories?: StateCategory[];
  assigneeIds?: string[];
  reporterIds?: string[];
  titleContains?: string;
  priorities?: string[];
  teamIds?: string[];
  sprintIds?: string[];
  labelIds?: string[];
  estimateValues?: number[];
  parentId?: string;
  objectiveId?: string;
  epicId?: string;
  hasNoAssignee?: boolean;
  hasBlockedBy?: boolean;
  createdAfter?: string;
  createdBefore?: string;
  updatedAfter?: string;
  updatedBefore?: string;
  startDateAfter?: string;
  startDateBefore?: string;
  deadlineAfter?: string;
  deadlineBefore?: string;
  includeArchived?: boolean;
  showSubStories?: boolean;
  includeDeleted?: boolean;
  completedAfter?: string;
  completedBefore?: string;
};

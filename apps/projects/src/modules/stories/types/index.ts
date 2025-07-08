import type { StateCategory } from "@/types/states";

export type StoryPriority =
  | "No Priority"
  | "Urgent"
  | "High"
  | "Medium"
  | "Low";

export type Story = {
  id: string;
  title: string;
  description?: string;
  statusId: string;
  sprintId: string | null;
  objectiveId: string | null;
  keyResultId: string | null;
  teamId: string;
  workspaceId: string;
  assigneeId: string | null;
  reporterId: string;
  epicId: string | null;
  sequenceId: number;
  priority: StoryPriority;
  startDate: string | null;
  endDate: string | null;
  createdAt: string;
  updatedAt: string;
  labels: string[] | null;
  subStories: Story[];
};

export type StoryActivity = {
  id: string;
  storyId: string;
  userId: string;
  type: "update" | "create";
  field: string;
  currentValue: string;
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
  priorities?: string[] | null;
  teamIds?: string[] | null;
  sprintIds?: string[] | null;
  labelIds?: string[] | null;
  parentId?: string | null;
  objectiveId?: string | null;
  epicId?: string | null;
  hasNoAssignee?: boolean | null;
  assignedToMe?: boolean;
  createdByMe?: boolean;
  createdAfter?: string;
  createdBefore?: string;
  updatedAfter?: string;
  updatedBefore?: string;
  deadlineAfter?: string;
  deadlineBefore?: string;
  includeArchived?: boolean;
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
  priorities?: string[];
  sprintIds?: string[];
  labelIds?: string[];
  parentId?: string;
  objectiveId?: string;
  epicId?: string;
  hasNoAssignee?: boolean;
  createdAfter?: string;
  createdBefore?: string;
  updatedAfter?: string;
  updatedBefore?: string;
  deadlineAfter?: string;
  deadlineBefore?: string;
  includeArchived?: boolean;
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
  priorities?: string[];
  teamIds?: string[];
  sprintIds?: string[];
  labelIds?: string[];
  parentId?: string;
  objectiveId?: string;
  epicId?: string;
  hasNoAssignee?: boolean;
  createdAfter?: string;
  createdBefore?: string;
  updatedAfter?: string;
  updatedBefore?: string;
  deadlineAfter?: string;
  deadlineBefore?: string;
  includeArchived?: boolean;
};

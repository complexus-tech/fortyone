import type { Story, StoryFilters, StoryGroup } from "@/modules/stories/types";

export type MyWorkTab = "all" | "assigned" | "created";

export type MyWorkViewOptions = {
  groupBy: "status" | "priority" | "assignee";
  orderBy: "created" | "updated" | "deadline" | "priority";
  orderDirection: "asc" | "desc";
};

export type GroupedStoryParams = {
  groupBy: "priority" | "status" | "assignee" | "none";
  orderBy?: "created" | "updated" | "deadline" | "priority";
  orderDirection?: "asc" | "desc";
  assignedToMe?: boolean;
  createdByMe?: boolean;
  teamIds?: string[];
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
  completedAfter?: string;
  completedBefore?: string;
  includeDeleted?: boolean;
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

export type MyWorkSection = {
  title: string;
  color?: string;
  data: Story[];
  key: string;
  totalCount: number;
  loadedCount: number;
  hasMore: boolean;
};

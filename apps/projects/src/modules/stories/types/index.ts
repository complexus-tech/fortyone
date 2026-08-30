import type { Story } from "@/shared/story/types";
import type { StateCategory } from "@/types/states";

export type {
  AutoSchedulingStatus,
  DetailedStory,
  NewStory,
  Story,
  StoryActivity,
  StoryAssociation,
  StoryAssociationType,
  StoryAttachment,
  StoryObjectiveSummary,
  StoryPriority,
  StorySprintSummary,
  StoryTeamSummary,
  StoryUpdate,
} from "@/shared/story/types";

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
  excludedStatusIds?: string[] | null;
  categories?: StateCategory[];
  assigneeIds?: string[] | null;
  excludedAssigneeIds?: string[] | null;
  reporterIds?: string[] | null;
  excludedReporterIds?: string[] | null;
  titleContains?: string | null;
  titleNotContains?: string | null;
  priorities?: string[] | null;
  excludedPriorities?: string[] | null;
  teamIds?: string[] | null;
  excludedTeamIds?: string[] | null;
  sprintIds?: string[] | null;
  excludedSprintIds?: string[] | null;
  labelIds?: string[] | null;
  excludedLabelIds?: string[] | null;
  estimateValues?: number[] | null;
  excludedEstimateValues?: number[] | null;
  parentId?: string | null;
  objectiveId?: string | null;
  excludedObjectiveId?: string | null;
  keyResultId?: string | null;
  epicId?: string | null;
  hasNoAssignee?: boolean | null;
  hasAssignee?: boolean | null;
  hasBlockedBy?: boolean | null;
  assignedToMe?: boolean;
  collaboratingWithMe?: boolean;
  createdByMe?: boolean;
  createdAfter?: string;
  createdBefore?: string;
  updatedAfter?: string;
  updatedBefore?: string;
  startDateAfter?: string;
  startDateBefore?: string;
  startDateNot?: string;
  deadlineAfter?: string;
  deadlineBefore?: string;
  deadlineNot?: string;
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
  excludedTeamIds?: string[];
  categories?: StateCategory[];
  assignedToMe?: boolean;
  collaboratingWithMe?: boolean;
  createdByMe?: boolean;
  storiesPerGroup?: number;
  statusIds?: string[];
  excludedStatusIds?: string[];
  assigneeIds?: string[];
  excludedAssigneeIds?: string[];
  reporterIds?: string[];
  excludedReporterIds?: string[];
  titleContains?: string;
  titleNotContains?: string;
  priorities?: string[];
  excludedPriorities?: string[];
  sprintIds?: string[];
  excludedSprintIds?: string[];
  labelIds?: string[];
  excludedLabelIds?: string[];
  estimateValues?: number[];
  excludedEstimateValues?: number[];
  parentId?: string;
  objectiveId?: string;
  excludedObjectiveId?: string;
  keyResultId?: string;
  epicId?: string;
  hasNoAssignee?: boolean;
  hasAssignee?: boolean;
  hasBlockedBy?: boolean;
  createdAfter?: string;
  createdBefore?: string;
  updatedAfter?: string;
  updatedBefore?: string;
  startDateAfter?: string;
  startDateBefore?: string;
  startDateNot?: string;
  deadlineAfter?: string;
  deadlineBefore?: string;
  deadlineNot?: string;
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
  keyResultId?: string;
  page?: number;
  pageSize?: number;
  assignedToMe?: boolean;
  collaboratingWithMe?: boolean;
  createdByMe?: boolean;
  statusIds?: string[];
  excludedStatusIds?: string[];
  categories?: StateCategory[];
  assigneeIds?: string[];
  excludedAssigneeIds?: string[];
  reporterIds?: string[];
  excludedReporterIds?: string[];
  titleContains?: string;
  titleNotContains?: string;
  priorities?: string[];
  excludedPriorities?: string[];
  teamIds?: string[];
  excludedTeamIds?: string[];
  sprintIds?: string[];
  excludedSprintIds?: string[];
  labelIds?: string[];
  excludedLabelIds?: string[];
  estimateValues?: number[];
  excludedEstimateValues?: number[];
  parentId?: string;
  objectiveId?: string;
  excludedObjectiveId?: string;
  epicId?: string;
  hasNoAssignee?: boolean;
  hasAssignee?: boolean;
  hasBlockedBy?: boolean;
  createdAfter?: string;
  createdBefore?: string;
  updatedAfter?: string;
  updatedBefore?: string;
  startDateAfter?: string;
  startDateBefore?: string;
  startDateNot?: string;
  deadlineAfter?: string;
  deadlineBefore?: string;
  deadlineNot?: string;
  includeArchived?: boolean;
  showSubStories?: boolean;
  includeDeleted?: boolean;
  completedAfter?: string;
  completedBefore?: string;
};

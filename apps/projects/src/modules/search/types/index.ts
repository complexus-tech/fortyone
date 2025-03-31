import type { Objective } from "@/modules/objectives/types";
import type { Story, StoryPriority } from "@/modules/stories/types";

export type SearchQueryParams = {
  type?: "all" | "stories" | "objectives";
  query?: string;
  teamId?: string;
  assigneeId?: string;
  labelId?: string;
  statusId?: string;
  priority?: StoryPriority;
  sortBy?: "relevance" | "updated" | "created";
  page?: number;
  pageSize?: number;
};

export type SearchResponse = {
  stories: Story[];
  objectives: Objective[];
  totalStories: number;
  totalObjectives: number;
  totalPages: number;
  page: number;
  pageSize: number;
};

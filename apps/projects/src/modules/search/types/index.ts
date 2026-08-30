import type { Objective } from "@/shared/objectives/types";
import type { Story, StoryPriority } from "@/shared/story/types";
import type { StateCategory } from "@/types/states";

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

export type SearchStory = Story & {
  statusName?: string | null;
  statusColor?: string | null;
  statusCategory?: StateCategory | null;
  teamName: string;
  teamCode: string;
  assigneeFullName?: string | null;
  assigneeUsername?: string | null;
};

export type SearchObjective = Objective & {
  leadUser: string | null;
  leadFullName?: string | null;
  leadUsername?: string | null;
  teamName: string;
  teamCode: string;
};

export type SearchResponse = {
  stories: SearchStory[];
  objectives: SearchObjective[];
  totalStories: number;
  totalObjectives: number;
  totalPages: number;
  page: number;
  pageSize: number;
};

export type SimilarStory = {
  id: string;
  sequenceId: number;
  title: string;
  teamId: string;
  statusId?: string | null;
  assigneeId?: string | null;
  priority?: StoryPriority;
  confidence: number;
};

import type { Dispatch, SetStateAction } from "react";
import type { TeamStoryFilters } from "@/modules/teams/stories/team-story-filters";
import type { StoryPriority } from "@/modules/stories/types";
import type { StatusCategory } from "@/types/statuses";

export type StoryFilterFacet =
  | "status"
  | "priority"
  | "assignee"
  | "sprint"
  | "objective";
export type StoryFiltersSheetProps = {
  isOpen: boolean;
  onClose: () => void;
  filters: TeamStoryFilters;
  onChange: Dispatch<SetStateAction<TeamStoryFilters>>;
  teamId?: string;
  initialFacet?: StoryFilterFacet;
  allowAssignee?: boolean;
};
export type StoryFilterOption = {
  id: string;
  label: string;
  description?: string;
  color?: string;
  statusCategory?: StatusCategory;
  priority?: StoryPriority;
};
export type StoryFilterSection = {
  id: StoryFilterFacet;
  label: string;
  value: string;
  selectedIds: string[];
  options: StoryFilterOption[];
  loading: boolean;
  error: Error | null;
  retry: () => void;
};

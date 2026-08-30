import type { StoryPriority } from "@/shared/story/types";

/**
 * Stable objective resource contract shared by cross-feature read models.
 * Feature-owned commands, analytics, and workflow types remain in Objectives.
 */
export type ObjectiveHealth = "On Track" | "At Risk" | "Off Track" | null;

export type ObjectiveScheduleStatus =
  | "on_track"
  | "at_risk"
  | "no_target"
  | "no_schedule";

export type Objective = {
  id: string;
  sequenceId: number;
  name: string;
  description: string;
  shortSummary: string | null;
  leadUser: string;
  teamId: string;
  workspaceId: string;
  startDate: string;
  endDate: string;
  isPrivate: boolean;
  createdAt: string;
  updatedAt: string;
  createdBy: string;
  statusId: string;
  keyResultCount: number;
  priority?: StoryPriority;
  health: ObjectiveHealth;
  color: string;
  forecastStartDate: string | null;
  forecastEndDate: string | null;
  scheduleStatus: ObjectiveScheduleStatus;
  forecastDaysDelta: number;
  forecastCauseStory: {
    id: string;
    sequenceId: number;
    title: string;
    source: "calendar" | "planning";
  } | null;
  stats?: {
    total: number;
    cancelled: number;
    completed: number;
    started: number;
    unstarted: number;
    backlog: number;
  };
};

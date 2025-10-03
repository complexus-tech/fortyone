import type { StatusCategory } from "@/types/statuses";
import type { StoryPriority } from "@/modules/stories/types";

export type ObjectiveHealth = "On Track" | "At Risk" | "Off Track" | null;

export type Objective = {
  id: string;
  name: string;
  description: string;
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
  priority?: StoryPriority;
  health: ObjectiveHealth;
  stats?: {
    total: number;
    cancelled: number;
    completed: number;
    started: number;
    unstarted: number;
    backlog: number;
  };
};

export type ObjectiveStatus = {
  id: string;
  name: string;
  category: StatusCategory;
  isDefault: boolean;
  color: string;
  orderIndex: number;
  workspaceId: string;
  createdAt: string;
  updatedAt: string;
};

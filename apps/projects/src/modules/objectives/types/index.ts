import type { StoryPriority } from "@/modules/stories/types";
import type { StateCategory } from "@/types/states";

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

export type MeasureType = "percentage" | "number" | "boolean";

export type KeyResult = {
  id: string;
  name: string;
  measurementType: MeasureType;
  objectiveId: string;
  startValue: number;
  targetValue: number;
  currentValue: number;
  createdAt: string;
  updatedAt: string;
  createdBy: string;
  lastUpdatedBy: string;
};

export type NewKeyResult = {
  name: string;
  measurementType: MeasureType;
  startValue: number;
  targetValue: number;
  currentValue: number;
};

export type KeyResultUpdate = Partial<Omit<NewKeyResult, "measurementType">>;

export type NewObjectiveKeyResult = NewKeyResult & {
  objectiveId: string;
};

export type NewObjective = {
  name: string;
  description?: string;
  leadUser?: string | null;
  teamId: string;
  startDate?: string | null;
  endDate?: string | null;
  isPrivate?: boolean;
  statusId: string;
  priority?: StoryPriority;
  keyResults?: NewKeyResult[];
};

export type ObjectiveUpdate = Partial<Omit<NewObjective, "keyResults">> & {
  health?: ObjectiveHealth;
};

export type ObjectiveStatus = {
  id: string;
  name: string;
  category: StateCategory;
  isDefault: boolean;
  color: string;
  orderIndex: number;
  workspaceId: string;
  createdAt: string;
  updatedAt: string;
};

export type ObjectiveAnalytics = {
  objectiveId: string;
  priorityBreakdown: {
    priority: string;
    count: number;
  }[];
  progressBreakdown: {
    total: number;
    completed: number;
    inProgress: number;
    todo: number;
    blocked: number;
    cancelled: number;
  };
  teamAllocation: {
    memberId: string;
    username: string;
    avatarUrl: string;
    assigned: number;
    completed: number;
  }[];
  progressChart: {
    date: string;
    completed: number;
    inProgress: number;
    total: number;
  }[];
};

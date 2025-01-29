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
  stats: {
    total: number;
    cancelled: number;
    completed: number;
    started: number;
    unstarted: number;
    backlog: number;
  };
};

export type KeyResult = {
  id: string;
  name: string;
  measurementType: "number" | "percentage" | "boolean";
  objectiveId: string;
  startValue: number;
  targetValue: number;
  createdAt: string;
  updatedAt: string;
};

export type NewKeyResult = {
  name: string;
  measurementType: "percentage" | "number" | "boolean";
  startValue: number;
  targetValue: number;
};

export type NewObjective = {
  name: string;
  description?: string;
  leadUserId?: string;
  teamId: string;
  startDate?: string;
  endDate?: string;
  isPrivate: boolean;
  statusId: string;
  priority?: string;
  keyResults?: NewKeyResult[];
};

export type ObjectiveUpdate = Partial<{
  name: string;
  description: string;
  leadUser: string;
  startDate: string;
  endDate: string;
  isPrivate: boolean;
  statusId: string;
  priority: string;
}>;

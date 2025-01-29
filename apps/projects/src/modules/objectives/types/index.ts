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

export type NewKeyResults = {
  name: string;
  measurementType: "percentage" | "number" | "boolean";
  startValue: number;
  targetValue: number;
}[];

export type NewKeyResult = NewKeyResults[number] & {
  objectiveId: string;
};

export type NewObjective = {
  name: string;
  description?: string;
  leadUser?: string;
  teamId: string;
  startDate?: string;
  endDate?: string;
  isPrivate: boolean;
  statusId: string;
  priority?: string;
  keyResults?: NewKeyResults;
};

export type ObjectiveUpdate = Partial<Omit<NewObjective, "keyResults">>;

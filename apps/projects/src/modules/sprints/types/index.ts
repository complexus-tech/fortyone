export type Sprint = {
  id: string;
  name: string;
  goal: string;
  objectiveId: string;
  teamId: string;
  workspaceId: string;
  startDate: string;
  endDate: string;
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

export type NewSprint = {
  name: string;
  goal?: string;
  objectiveId?: string | null;
  teamId: string;
  startDate: string;
  endDate: string;
};

export type UpdateSprint = Partial<Omit<NewSprint, "teamId">>;

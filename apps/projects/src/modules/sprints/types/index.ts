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

export type SprintDetails = Omit<Sprint, "stats">;

export type NewSprint = {
  name: string;
  goal?: string;
  objectiveId?: string | null;
  teamId: string;
  startDate: string;
  endDate: string;
};

export type UpdateSprint = Partial<Omit<NewSprint, "teamId">>;

export type SprintAnalytics = {
  sprintId: string;
  overview: {
    completionPercentage: number;
    daysElapsed: number;
    daysRemaining: number;
    status: string;
  };
  storyBreakdown: {
    total: number;
    completed: number;
    inProgress: number;
    todo: number;
    blocked: number;
    cancelled: number;
  };
  burndown: {
    date: string;
    remaining: number;
    ideal: number;
  }[];
  teamAllocation: {
    memberId: string;
    username: string;
    avatarUrl: string;
    assigned: number;
    completed: number;
  }[];
};

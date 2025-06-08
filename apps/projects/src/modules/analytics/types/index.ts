// Common filter types
export type AnalyticsFilters = {
  teamIds?: string[];
  startDate?: string;
  endDate?: string;
  sprintIds?: string[];
  objectiveIds?: string[];
};

// Workspace Overview Types
export type WorkspaceOverviewMetrics = {
  totalStories: number;
  completedStories: number;
  activeObjectives: number;
  activeSprints: number;
  totalTeamMembers: number;
};

export type CompletionTrendPoint = {
  date: string;
  completed: number;
  total: number;
};

export type VelocityTrendPoint = {
  period: string;
  velocity: number;
};

export type WorkspaceOverview = {
  workspaceId: string;
  reportDate: string;
  filters: AnalyticsFilters;
  metrics: WorkspaceOverviewMetrics;
  completionTrend: CompletionTrendPoint[];
  velocityTrend: VelocityTrendPoint[];
};

// Story Analytics Types
export type StatusBreakdownItem = {
  statusName: string;
  count: number;
  teamId: string;
};

export type PriorityDistributionItem = {
  priority: string;
  count: number;
};

export type CompletionByTeamItem = {
  teamId: string;
  teamName: string;
  completed: number;
  total: number;
};

export type BurndownPoint = {
  date: string;
  remaining: number;
};

export type StoryAnalytics = {
  statusBreakdown: StatusBreakdownItem[];
  priorityDistribution: PriorityDistributionItem[];
  completionByTeam: CompletionByTeamItem[];
  burndown: BurndownPoint[];
};

// Objective Progress Types
export type HealthDistributionItem = {
  status: string;
  count: number;
};

export type ObjectiveStatusBreakdownItem = {
  statusName: string;
  count: number;
};

export type KeyResultProgressItem = {
  objectiveId: string;
  objectiveName: string;
  completed: number;
  total: number;
  avgProgress: number;
};

export type ProgressByTeamItem = {
  teamId: string;
  teamName: string;
  objectives: number;
  completed: number;
};

export type ObjectiveProgress = {
  healthDistribution: HealthDistributionItem[];
  statusBreakdown: ObjectiveStatusBreakdownItem[];
  keyResultsProgress: KeyResultProgressItem[];
  progressByTeam: ProgressByTeamItem[];
};

// Team Performance Types
export type TeamWorkloadItem = {
  teamId: string;
  teamName: string;
  assigned: number;
  completed: number;
  capacity: number;
};

export type MemberContributionItem = {
  userId: string;
  username: string;
  avatarUrl: string;
  teamId: string;
  completed: number;
  assigned: number;
};

export type VelocityByTeamItem = {
  teamId: string;
  teamName: string;
  week1: number;
  week2: number;
  week3: number;
  average: number;
};

export type WorkloadTrendPoint = {
  date: string;
  assigned: number;
  completed: number;
};

export type TeamPerformance = {
  teamWorkload: TeamWorkloadItem[];
  memberContributions: MemberContributionItem[];
  velocityByTeam: VelocityByTeamItem[];
  workloadTrend: WorkloadTrendPoint[];
};

// Sprint Analytics Types
export type SprintProgressItem = {
  sprintId: string;
  sprintName: string;
  teamId: string;
  completed: number;
  total: number;
  status: string;
};

export type CombinedBurndownPoint = {
  date: string;
  planned: number;
  actual: number;
};

export type TeamAllocationItem = {
  teamId: string;
  teamName: string;
  activeSprints: number;
  totalStories: number;
  completedStories: number;
};

export type SprintHealthItem = {
  status: string;
  count: number;
};

export type SprintAnalytics = {
  sprintProgress: SprintProgressItem[];
  combinedBurndown: CombinedBurndownPoint[];
  teamAllocation: TeamAllocationItem[];
  sprintHealth: SprintHealthItem[];
};

// Timeline Trends Types
export type StoryCompletionPoint = {
  date: string;
  completed: number;
  created: number;
};

export type ObjectiveProgressPoint = {
  date: string;
  totalObjectives: number;
  completedObjectives: number;
};

export type TeamVelocityPoint = {
  date: string;
  teamId: string;
  velocity: number;
};

export type KeyMetricsTrendPoint = {
  date: string;
  activeUsers: number;
  storiesPerDay: number;
  avgCycleTime: number;
};

export type TimelineTrends = {
  storyCompletion: StoryCompletionPoint[];
  objectiveProgress: ObjectiveProgressPoint[];
  teamVelocity: TeamVelocityPoint[];
  keyMetricsTrend: KeyMetricsTrendPoint[];
};

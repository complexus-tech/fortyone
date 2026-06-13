// Common filter types
export type AnalyticsFilters = {
  teamIds?: string[];
  assigneeIds?: string[];
  startDate?: string;
  endDate?: string;
  sprintIds?: string[];
  objectiveIds?: string[];
};

export type WorkloadSummary = {
  totalOpenStories: number;
  totalEstimate: number;
  overdueStories: number;
  urgentStories: number;
  highPriorityStories: number;
  unestimatedStories: number;
  unassignedStories: number;
};

export type MemberWorkload = {
  userId: string;
  fullName: string;
  username: string;
  avatarUrl: string;
  openStories: number;
  startedStories: number;
  pausedStories: number;
  completedStories: number;
  overdueStories: number;
  urgentStories: number;
  highPriorityStories: number;
  unestimatedStories: number;
  estimateTotal: number;
};

export type TeamWorkloadSummary = {
  teamId: string;
  teamName: string;
  teamCode: string;
  openStories: number;
  estimateTotal: number;
  overdueStories: number;
  unassignedStories: number;
  unestimatedStories: number;
};

export type UnassignedWorkload = {
  stories: number;
  estimateTotal: number;
  overdueStories: number;
  urgentStories: number;
  highPriorityStories: number;
  unestimatedStories: number;
};

export type WorkloadRisks = {
  overloadedMembers: MemberWorkload[];
  overdueMembers: MemberWorkload[];
  unassignedStories: number;
  unestimatedStories: number;
  highPriorityStories: number;
};

export type WorkloadAnalysis = {
  summary: WorkloadSummary;
  members: MemberWorkload[];
  teams: TeamWorkloadSummary[];
  unassigned: UnassignedWorkload;
  risks: WorkloadRisks;
};

export type PulseRiskSeverity = "high" | "medium" | "low";

export type PulseRiskKind =
  | "overdue_stories"
  | "blocked_stories"
  | "overloaded_members"
  | "at_risk_sprints"
  | "at_risk_objectives"
  | "pending_requests"
  | "unassigned_stories";

export type PulseSummary = {
  openStories: number;
  overdueStories: number;
  blockedStories: number;
  atRiskSprints: number;
  atRiskObjectives: number;
  pendingRequests: number;
  overloadedMembers: number;
};

export type PulseStoryHealth = {
  openStories: number;
  startedStories: number;
  pausedStories: number;
  completedStories: number;
  cancelledStories: number;
  blockedStories: number;
  overdueStories: number;
  urgentStories: number;
  highPriorityStories: number;
  unassignedStories: number;
  unestimatedStories: number;
};

export type PulseSprintHealth = {
  activeSprints: number;
  upcomingSprints: number;
  completedSprints: number;
  atRiskSprints: number;
  overdueSprints: number;
  unestimatedStories: number;
};

export type PulseObjectiveHealth = {
  activeObjectives: number;
  atRiskObjectives: number;
  offTrackObjectives: number;
  overdueObjectives: number;
  objectivesDueSoon: number;
};

export type PulseRequestHealth = {
  pendingRequests: number;
  urgentRequests: number;
  highRequests: number;
  gitHubRequests: number;
  slackRequests: number;
  intercomRequests: number;
  staleRequests: number;
};

export type PulseRisk = {
  kind: PulseRiskKind;
  severity: PulseRiskSeverity;
  title: string;
  description: string;
  count: number;
};

export type PulseReport = {
  workspaceId: string;
  reportDate: string;
  filters: AnalyticsFilters;
  summary: PulseSummary;
  stories: PulseStoryHealth;
  sprints: PulseSprintHealth;
  objectives: PulseObjectiveHealth;
  requests: PulseRequestHealth;
  workload: WorkloadAnalysis;
  risks: PulseRisk[];
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

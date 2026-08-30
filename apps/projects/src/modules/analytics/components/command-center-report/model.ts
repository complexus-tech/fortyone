import type { State } from "@/types/states";
import type {
  AnalyticsFilters,
  MemberWorkload,
  RequestProviderPerformance,
  TeamWorkloadSummary,
  WorkspaceCommandCenterReport,
  WorkspaceEngagementCount,
} from "../../types";

export type ChartBreakdownRow = {
  color?: string;
  label: string;
  value: number;
};

export type CompletionTrendChartRow = {
  completed: number;
  date: string;
  total: number;
};

export type MemberWorkloadChartRow = {
  estimate: number;
  name: string;
  open: number;
  overdue: number;
};

export type ProgressChartRow = {
  completed: number;
  label: string;
  remaining: number;
};

export type ProviderChartRow = {
  accepted: number;
  acceptedShare: number;
  pending: number;
  pendingShare: number;
  provider: string;
  stale: number;
  staleShare: number;
  total: number;
};

const statusFallbackColors: Partial<Record<State["category"], string>> = {
  backlog: "#A855F7",
  cancelled: "#F43F5E",
  completed: "#22C55E",
  paused: "#EAB308",
  started: "#6366F1",
  unstarted: "#06B6D4",
};

const priorityColors: Record<string, string> = {
  high: "#EAB308",
  low: "#06B6D4",
  medium: "#22C55E",
  "no priority": "#94A3B8",
  urgent: "#EA6060",
};

export const chartPalette = {
  danger: "#EA6060",
  info: "#06B6D4",
  muted: "#94A3B8",
  navy: "#002F61",
  primary: "#6366F1",
  rose: "#F43F5E",
  success: "#22C55E",
  violet: "#A855F7",
  warning: "#EAB308",
};

const numberFormatter = new Intl.NumberFormat();
const percentFormatter = new Intl.NumberFormat(undefined, {
  maximumFractionDigits: 0,
  style: "percent",
});

export const formatNumber = (value?: number) =>
  numberFormatter.format(value ?? 0);

export const formatPercent = (value: number) => percentFormatter.format(value);

export const completionRate = (completed: number, total: number) =>
  total > 0 ? completed / total : 0;

export const titleCase = (value: string) =>
  value
    .split(/[\s_-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");

export const formatShortDate = (value: string) => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;

  return new Intl.DateTimeFormat(undefined, {
    day: "numeric",
    month: "short",
  }).format(date);
};

export const getDisplayName = (member: {
  fullName?: string | null;
  username: string;
}) => member.fullName?.trim() || member.username;

export const getGridStroke = (resolvedTheme?: string) =>
  resolvedTheme === "dark" ? "#2A2A2A" : "#E5E7EB";

export const getCursorFill = (resolvedTheme?: string) =>
  resolvedTheme === "dark"
    ? "rgba(255, 255, 255, 0.03)"
    : "rgba(2, 6, 23, 0.025)";

export const truncateChartLabel = (value: string, maxLength = 18) =>
  value.length > maxLength ? `${value.slice(0, maxLength - 1)}…` : value;

export const buildFilterSignature = (filters: AnalyticsFilters) =>
  [
    filters.startDate ?? "",
    filters.endDate ?? "",
    filters.teamIds?.join(",") ?? "",
    filters.assigneeIds?.join(",") ?? "",
    filters.sprintIds?.join(",") ?? "",
    filters.objectiveIds?.join(",") ?? "",
  ].join("|");

export const buildCompletionTrendChartData = (
  completionTrend: WorkspaceCommandCenterReport["overview"]["completionTrend"],
): CompletionTrendChartRow[] =>
  completionTrend.map((point) => ({
    completed: point.completed,
    date: formatShortDate(point.date),
    total: point.total,
  }));

export const buildMemberWorkloadChartData = (
  members: MemberWorkload[],
): MemberWorkloadChartRow[] =>
  members.slice(0, 8).map((member) => ({
    estimate: member.estimateTotal,
    name: getDisplayName(member),
    open: member.openStories,
    overdue: member.overdueStories,
  }));

export const buildStatusBreakdownData = (
  statusBreakdown: WorkspaceCommandCenterReport["stories"]["statusBreakdown"],
  statuses: Pick<State, "category" | "color" | "name" | "teamId">[],
): ChartBreakdownRow[] =>
  statusBreakdown.slice(0, 8).map((item) => {
    const status =
      statuses.find(
        (candidate) =>
          candidate.name.toLowerCase() === item.statusName.toLowerCase() &&
          (!item.teamId || candidate.teamId === item.teamId),
      ) ??
      statuses.find(
        (candidate) =>
          candidate.name.toLowerCase() === item.statusName.toLowerCase(),
      );
    const fallbackColor = status?.category
      ? statusFallbackColors[status.category]
      : undefined;

    return {
      color: status?.color ?? fallbackColor ?? chartPalette.primary,
      label: item.statusName,
      value: item.count,
    };
  });

export const buildPriorityDistributionData = (
  priorityDistribution: WorkspaceCommandCenterReport["stories"]["priorityDistribution"],
): ChartBreakdownRow[] =>
  priorityDistribution.map((item) => ({
    color: priorityColors[item.priority.toLowerCase()] ?? chartPalette.primary,
    label: titleCase(item.priority),
    value: item.count,
  }));

export const buildObjectiveProgressChartData = (
  objectives: WorkspaceCommandCenterReport["objectives"]["keyResultsProgress"],
): ProgressChartRow[] =>
  objectives.slice(0, 8).map((objective) => ({
    completed: objective.completed,
    label: objective.objectiveName,
    remaining: Math.max(objective.total - objective.completed, 0),
  }));

export const buildSprintProgressChartData = (
  sprints: WorkspaceCommandCenterReport["sprints"]["sprintProgress"],
): ProgressChartRow[] =>
  sprints.slice(0, 8).map((sprint) => ({
    completed: sprint.completed,
    label: sprint.sprintName,
    remaining: Math.max(sprint.total - sprint.completed, 0),
  }));

export const buildProviderChartData = (
  providers: RequestProviderPerformance[],
): ProviderChartRow[] =>
  providers
    .slice()
    .sort((left, right) => right.totalRequests - left.totalRequests)
    .slice(0, 10)
    .map((provider) => {
      const total = Math.max(provider.totalRequests, 1);

      return {
        accepted: provider.acceptedRequests,
        acceptedShare: (provider.acceptedRequests / total) * 100,
        pending: provider.pendingRequests,
        pendingShare: (provider.pendingRequests / total) * 100,
        provider: titleCase(provider.provider),
        stale: provider.staleRequests,
        staleShare: (provider.staleRequests / total) * 100,
        total: provider.totalRequests,
      };
    });

export const buildTeamWorkloadChartData = (
  teams: TeamWorkloadSummary[],
): ChartBreakdownRow[] =>
  teams.slice(0, 8).map((team) => ({
    color:
      team.overdueStories > 0 ? chartPalette.warning : chartPalette.primary,
    label: team.teamName,
    value: team.openStories,
  }));

export const buildEngagementCountChartData = (
  items: WorkspaceEngagementCount[],
): ChartBreakdownRow[] =>
  items.slice(0, 8).map((item) => ({
    label: titleCase(item.name),
    value: item.count,
  }));

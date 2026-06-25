"use client";

import type { ReactNode } from "react";
import { useEffect, useMemo } from "react";
import Link from "next/link";
import { useTheme } from "next-themes";
import {
  Bar,
  BarChart,
  CartesianGrid,
  Cell,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  type TooltipProps,
} from "recharts";
import {
  Avatar,
  Badge,
  Box,
  Button,
  Flex,
  Skeleton,
  Tabs,
  Text,
  Wrapper,
} from "ui";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { useStatuses } from "@/lib/hooks/statuses";
import { useAppliedFilters } from "@/modules/analytics/hooks/filters";
import { useCommandCenterReport } from "@/modules/analytics/hooks/command-center-report";
import { useWorkspaceAnalyticsEvent } from "@/modules/analytics/hooks/workspace-analytics-event";
import type {
  MemberWorkload,
  RequestProviderPerformance,
  TeamWorkloadSummary,
  WorkspaceCommandCenterReport,
  WorkspaceEngagementCount,
  WorkspaceEngagementUser,
} from "@/modules/analytics/types";
import { Filters } from "./filters";

type ChartRow = Record<string, string | number>;

type MetricCardProps = {
  label: string;
  value: string;
  description: string;
  accent?: string | null;
};

const chartPalette = {
  primary: "#6366F1",
  success: "#22C55E",
  warning: "#EAB308",
  danger: "#EA6060",
  info: "#06B6D4",
  violet: "#A855F7",
  rose: "#F43F5E",
  navy: "#002F61",
  muted: "#94A3B8",
};

const numberFormatter = new Intl.NumberFormat();
const percentFormatter = new Intl.NumberFormat(undefined, {
  maximumFractionDigits: 0,
  style: "percent",
});

const formatNumber = (value?: number) => numberFormatter.format(value ?? 0);

const formatPercent = (value: number) => percentFormatter.format(value);

const completionRate = (completed: number, total: number) =>
  total > 0 ? completed / total : 0;

const titleCase = (value: string) =>
  value
    .split(/[\s_-]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");

const formatShortDate = (value: string) => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;

  return new Intl.DateTimeFormat(undefined, {
    day: "numeric",
    month: "short",
  }).format(date);
};

const getDisplayName = (member: {
  fullName?: string | null;
  username: string;
}) => member.fullName?.trim() || member.username;

const getGridStroke = (resolvedTheme?: string) =>
  resolvedTheme === "dark" ? "#2A2A2A" : "#E5E7EB";

const getCursorFill = (resolvedTheme?: string) =>
  resolvedTheme === "dark"
    ? "rgba(255, 255, 255, 0.03)"
    : "rgba(2, 6, 23, 0.025)";

const truncateChartLabel = (value: string, maxLength = 18) =>
  value.length > maxLength ? `${value.slice(0, maxLength - 1)}…` : value;

type ProviderChartRow = {
  accepted: number;
  acceptedShare: number;
  pending: number;
  pendingShare: number;
  provider: string;
  stale: number;
  staleShare: number;
  total: number;
};

type ProgressChartRow = {
  completed: number;
  label: string;
  remaining: number;
};

type ProgressSegmentShapeProps = {
  fill?: string;
  height?: number;
  payload?: ProgressChartRow;
  segment: "completed" | "remaining";
  width?: number;
  x?: number;
  y?: number;
};

const statusFallbackColors: Partial<Record<string, string>> = {
  backlog: chartPalette.violet,
  cancelled: chartPalette.rose,
  completed: chartPalette.success,
  paused: chartPalette.warning,
  started: chartPalette.primary,
  unstarted: chartPalette.info,
};

const priorityColors: Record<string, string> = {
  high: chartPalette.warning,
  low: chartPalette.info,
  medium: chartPalette.success,
  "no priority": chartPalette.muted,
  urgent: chartPalette.danger,
};

const ReportCard = ({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) => {
  return (
    <Wrapper className={["h-full", className].filter(Boolean).join(" ")}>
      {children}
    </Wrapper>
  );
};

const SectionTitle = ({
  children,
  description,
}: {
  children: ReactNode;
  description?: string;
}) => {
  return (
    <Box>
      <Text className="text-lg font-semibold">{children}</Text>
      {description ? (
        <Text className="text-foreground/70 mt-1 text-[0.92rem] leading-5">
          {description}
        </Text>
      ) : null}
    </Box>
  );
};

const MetricCard = ({ label, value, description, accent }: MetricCardProps) => {
  return (
    <Wrapper className="px-3 py-3 md:px-5 md:py-4">
      <Flex align="center" className="gap-4" justify="between">
        <Text className="text-2xl antialiased" fontWeight="semibold">
          {value}
        </Text>
        {accent ? (
          <Text className="text-success shrink-0 text-base font-medium">
            {accent}
          </Text>
        ) : null}
      </Flex>
      <Text className="mt-2 opacity-80" color="muted">
        {label}
      </Text>
      <Text
        className="text-foreground/70 mt-1 truncate text-[0.9rem] leading-5"
        title={description}
      >
        {description}
      </Text>
    </Wrapper>
  );
};

const MiniMetric = ({
  label,
  value,
  description,
}: {
  label: string;
  value: string;
  description: string;
}) => {
  return (
    <Wrapper className="py-3">
      <Text className="text-[1.45rem] leading-none font-semibold">{value}</Text>
      <Text className="mt-2 font-medium">{label}</Text>
      <Text className="text-foreground/70 mt-1 text-[0.9rem] leading-5">
        {description}
      </Text>
    </Wrapper>
  );
};

const ChartTooltip = ({
  active,
  label,
  payload,
}: TooltipProps<number, string>) => {
  if (!active || !payload?.length) {
    return null;
  }

  return (
    <Box className="border-border dark:bg-surface-elevated z-50 min-w-36 rounded-lg border bg-white px-3 py-3 text-[0.95rem] shadow-lg">
      <Text className="mb-2 font-medium">{label}</Text>
      <Box className="space-y-1">
        {payload.map((entry) => (
          <Flex align="center" className="gap-2" key={String(entry.dataKey)}>
            <Box
              className="size-2.5 rounded-sm"
              style={{ backgroundColor: entry.color }}
            />
            <Text color="muted">
              {entry.name ?? entry.dataKey}: {formatNumber(Number(entry.value))}
            </Text>
          </Flex>
        ))}
      </Box>
    </Box>
  );
};

const EmptyState = ({ children }: { children: ReactNode }) => {
  return (
    <Flex
      align="center"
      className="bg-surface-muted/40 min-h-28 rounded-lg px-4 py-6 text-center"
      justify="center"
    >
      <Text color="muted">{children}</Text>
    </Flex>
  );
};

const CommandCenterSkeleton = () => {
  return (
    <Box className="pt-3 pb-5">
      <Flex className="mb-6 flex-col gap-3 @3xl:flex-row @3xl:items-end @3xl:justify-between">
        <Box>
          <Skeleton className="mb-3 h-8 w-72" />
          <Skeleton className="h-5 w-full max-w-xl" />
        </Box>
        <Skeleton className="h-10 w-80" />
      </Flex>
      <Box className="mb-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 8 }).map((_, index) => (
          <Skeleton className="h-32" key={index} />
        ))}
      </Box>
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <Skeleton className="h-[28rem]" />
        <Skeleton className="h-[28rem]" />
      </Box>
    </Box>
  );
};

const WorkloadMemberRow = ({ member }: { member: MemberWorkload }) => {
  const { withWorkspace } = useWorkspacePath();
  const displayName = getDisplayName(member);

  return (
    <Link href={withWorkspace(`/profile/${member.userId}`)}>
      <Flex
        align="center"
        className="border-border hover:bg-surface-muted/60 gap-3 border-b-[0.5px] px-1 py-2.5 transition last:border-b-0"
      >
        <Avatar name={displayName} size="md" src={member.avatarUrl} />
        <Box className="min-w-0 flex-1">
          <Flex align="center" className="gap-2">
            <Text className="truncate" fontWeight="medium">
              {displayName}
            </Text>
            {member.teamAiRoleTitle ? (
              <Badge
                className="shrink-0 bg-transparent"
                color="tertiary"
                rounded="full"
                size="sm"
                variant="outline"
              >
                {member.teamAiRoleTitle}
              </Badge>
            ) : null}
          </Flex>
          <Flex align="center" className="mt-1 flex-wrap gap-x-3 gap-y-1">
            <Text color="muted">{formatNumber(member.openStories)} open</Text>
            <Text color="muted">
              {formatNumber(member.estimateTotal)} estimate
            </Text>
            <Text color="muted">
              {formatNumber(member.completedStories)} completed
            </Text>
          </Flex>
        </Box>
        <Box className="hidden min-w-28 text-right md:block">
          <Text className="font-semibold">
            {formatNumber(member.overdueStories)}
          </Text>
          <Text color="muted">overdue</Text>
        </Box>
        <Box className="hidden min-w-28 text-right @4xl:block">
          <Text className="font-semibold">
            {formatNumber(member.unestimatedStories)}
          </Text>
          <Text color="muted">unestimated</Text>
        </Box>
      </Flex>
    </Link>
  );
};

const TeamWorkloadRow = ({ team }: { team: TeamWorkloadSummary }) => {
  const { withWorkspace } = useWorkspacePath();

  return (
    <Link href={withWorkspace(`/teams/${team.teamId}`)}>
      <Flex
        align="center"
        className="border-border hover:bg-surface-muted/60 gap-3 border-b-[0.5px] px-1 py-2.5 transition last:border-b-0"
        justify="between"
      >
        <Box className="min-w-0">
          <Flex align="center" className="gap-2">
            <Text className="truncate" fontWeight="medium">
              {team.teamName}
            </Text>
            <Badge
              className="shrink-0 bg-transparent"
              color="tertiary"
              rounded="full"
              size="sm"
              variant="outline"
            >
              {team.teamCode}
            </Badge>
          </Flex>
          <Text className="mt-1" color="muted">
            {formatNumber(team.estimateTotal)} estimate /{" "}
            {formatNumber(team.unassignedStories)} unassigned
          </Text>
        </Box>
        <Box className="text-right">
          <Text className="text-lg" fontWeight="semibold">
            {formatNumber(team.openStories)}
          </Text>
          <Text color="muted">open</Text>
        </Box>
      </Flex>
    </Link>
  );
};

const DeliveryChart = ({
  report,
}: {
  report: WorkspaceCommandCenterReport;
}) => {
  const { resolvedTheme } = useTheme();
  const gridStroke = getGridStroke(resolvedTheme);
  const data = report.overview.completionTrend.map((point) => ({
    completed: point.completed,
    date: formatShortDate(point.date),
    total: point.total,
  }));

  if (!data.length) {
    return <EmptyState>No completion trend data is available.</EmptyState>;
  }

  return (
    <ResponsiveContainer height={300} width="100%">
      <LineChart
        data={data}
        margin={{ top: 8, right: 12, left: -18, bottom: 4 }}
      >
        <CartesianGrid
          stroke={gridStroke}
          strokeDasharray="3 3"
          vertical={false}
        />
        <XAxis dataKey="date" tick={{ fontSize: 12 }} tickLine={false} />
        <YAxis axisLine={false} tick={{ fontSize: 12 }} tickLine={false} />
        <Tooltip
          content={<ChartTooltip />}
          cursor={{ stroke: getCursorFill(resolvedTheme), strokeWidth: 1 }}
        />
        <Line
          dataKey="total"
          dot={false}
          name="Total"
          stroke={chartPalette.primary}
          strokeWidth={2.5}
          type="monotone"
        />
        <Line
          dataKey="completed"
          dot={false}
          name="Completed"
          stroke={chartPalette.success}
          strokeWidth={2.5}
          type="monotone"
        />
      </LineChart>
    </ResponsiveContainer>
  );
};

const WorkloadChart = ({
  report,
}: {
  report: WorkspaceCommandCenterReport;
}) => {
  const { resolvedTheme } = useTheme();
  const gridStroke = getGridStroke(resolvedTheme);
  const data: ChartRow[] = report.workload.members
    .slice(0, 8)
    .map((member) => ({
      estimate: member.estimateTotal,
      name: getDisplayName(member),
      open: member.openStories,
      overdue: member.overdueStories,
    }));

  if (!data.length) {
    return <EmptyState>No member workload chart data is available.</EmptyState>;
  }

  return (
    <ResponsiveContainer height={300} width="100%">
      <BarChart
        data={data}
        margin={{ top: 8, right: 12, left: -18, bottom: 4 }}
      >
        <CartesianGrid
          stroke={gridStroke}
          strokeDasharray="3 3"
          vertical={false}
        />
        <XAxis dataKey="name" tick={{ fontSize: 12 }} tickLine={false} />
        <YAxis axisLine={false} tick={{ fontSize: 12 }} tickLine={false} />
        <Tooltip
          content={<ChartTooltip />}
          cursor={{ fill: getCursorFill(resolvedTheme) }}
        />
        <Bar
          dataKey="open"
          fill={chartPalette.primary}
          name="Open"
          radius={[4, 4, 0, 0]}
        />
        <Bar
          dataKey="overdue"
          fill={chartPalette.danger}
          name="Overdue"
          radius={[4, 4, 0, 0]}
        />
      </BarChart>
    </ResponsiveContainer>
  );
};

const HorizontalBreakdownChart = ({
  color,
  data,
  emptyText,
  height = 300,
  labelWidth = 112,
}: {
  color: string;
  data: { label: string; value: number; color?: string }[];
  emptyText: string;
  height?: number;
  labelWidth?: number;
}) => {
  const { resolvedTheme } = useTheme();
  const gridStroke = getGridStroke(resolvedTheme);
  const chartData = data.map((item) => ({
    color: item.color ?? color,
    name: item.label,
    value: item.value,
  }));

  if (!chartData.length) {
    return <EmptyState>{emptyText}</EmptyState>;
  }

  return (
    <ResponsiveContainer height={height} width="100%">
      <BarChart
        data={chartData}
        layout="vertical"
        margin={{ top: 4, right: 20, left: -18, bottom: 4 }}
      >
        <CartesianGrid
          horizontal={false}
          stroke={gridStroke}
          strokeDasharray="3 3"
        />
        <XAxis
          axisLine={false}
          tick={{ fontSize: 12 }}
          tickLine={false}
          type="number"
        />
        <YAxis
          axisLine={false}
          dataKey="name"
          tick={{ fontSize: 12 }}
          tickFormatter={(value: string) => truncateChartLabel(value, 20)}
          tickLine={false}
          type="category"
          width={labelWidth}
        />
        <Tooltip
          content={<ChartTooltip />}
          cursor={{ fill: getCursorFill(resolvedTheme) }}
        />
        <Bar dataKey="value" name="Count" radius={[0, 4, 4, 0]}>
          {chartData.map((entry) => (
            <Cell fill={entry.color} key={entry.name} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
};

const ProgressComparisonChart = ({
  data,
  emptyText,
  height = 220,
}: {
  data: ProgressChartRow[];
  emptyText: string;
  height?: number;
}) => {
  const { resolvedTheme } = useTheme();
  const gridStroke = getGridStroke(resolvedTheme);

  if (!data.length) {
    return <EmptyState>{emptyText}</EmptyState>;
  }

  return (
    <ResponsiveContainer height={height} width="100%">
      <BarChart
        data={data}
        layout="vertical"
        margin={{ top: 4, right: 20, left: -18, bottom: 4 }}
      >
        <CartesianGrid
          horizontal={false}
          stroke={gridStroke}
          strokeDasharray="3 3"
        />
        <XAxis
          axisLine={false}
          tick={{ fontSize: 12 }}
          tickLine={false}
          type="number"
        />
        <YAxis
          axisLine={false}
          dataKey="label"
          tick={{ fontSize: 12 }}
          tickFormatter={(value: string) => truncateChartLabel(value, 20)}
          tickLine={false}
          type="category"
          width={120}
        />
        <Tooltip
          content={<ChartTooltip />}
          cursor={{ fill: getCursorFill(resolvedTheme) }}
        />
        <Bar
          dataKey="completed"
          fill={chartPalette.success}
          name="Completed"
          shape={
            <ProgressSegmentShape
              fill={chartPalette.success}
              segment="completed"
            />
          }
          stackId="progress"
        />
        <Bar
          dataKey="remaining"
          fill={chartPalette.primary}
          name="Remaining"
          shape={
            <ProgressSegmentShape
              fill={chartPalette.primary}
              segment="remaining"
            />
          }
          stackId="progress"
        />
      </BarChart>
    </ResponsiveContainer>
  );
};

const ProgressSegmentShape = ({
  fill = chartPalette.primary,
  height = 0,
  payload,
  segment,
  width = 0,
  x = 0,
  y = 0,
}: ProgressSegmentShapeProps) => {
  if (width <= 0 || height <= 0) {
    return null;
  }

  const shouldRoundRight =
    segment === "remaining" || (payload?.remaining ?? 0) <= 0;

  if (!shouldRoundRight) {
    return <rect fill={fill} height={height} width={width} x={x} y={y} />;
  }

  const radius = Math.min(4, width / 2, height / 2);
  const right = x + width;
  const bottom = y + height;

  return (
    <path
      d={[
        `M ${x} ${y}`,
        `H ${right - radius}`,
        `Q ${right} ${y} ${right} ${y + radius}`,
        `V ${bottom - radius}`,
        `Q ${right} ${bottom} ${right - radius} ${bottom}`,
        `H ${x}`,
        "Z",
      ].join(" ")}
      fill={fill}
    />
  );
};

const getPriorityColor = (priority: string) => {
  return priorityColors[priority.toLowerCase()] ?? chartPalette.primary;
};

const useFlowBreakdownData = (report: WorkspaceCommandCenterReport) => {
  const { data: statuses = [] } = useStatuses();
  const statusData = report.stories.statusBreakdown.slice(0, 8).map((item) => {
    const status =
      statuses.find(
        (status) =>
          status.name.toLowerCase() === item.statusName.toLowerCase() &&
          (!item.teamId || status.teamId === item.teamId),
      ) ??
      statuses.find(
        (status) => status.name.toLowerCase() === item.statusName.toLowerCase(),
      );

    return {
      color:
        status?.color ??
        statusFallbackColors[status?.category ?? ""] ??
        chartPalette.primary,
      label: item.statusName,
      value: item.count,
    };
  });
  const priorityData = report.stories.priorityDistribution.map((item) => ({
    color: getPriorityColor(item.priority),
    label: titleCase(item.priority),
    value: item.count,
  }));

  return { priorityData, statusData };
};

const StatusBreakdownCard = ({
  statusData,
}: {
  statusData: { label: string; value: number; color?: string }[];
}) => {
  return (
    <ReportCard>
      <SectionTitle description="Current work state across the selected filters.">
        Flow breakdown
      </SectionTitle>
      <Box className="mt-5">
        <HorizontalBreakdownChart
          color={chartPalette.primary}
          data={statusData}
          emptyText="No status breakdown is available."
          height={330}
          labelWidth={122}
        />
      </Box>
    </ReportCard>
  );
};

const PriorityDistributionCard = ({
  priorityData,
}: {
  priorityData: { label: string; value: number; color?: string }[];
}) => {
  return (
    <ReportCard>
      <SectionTitle description="Priority concentration across open and recently completed work.">
        Priority distribution
      </SectionTitle>
      <Box className="mt-5">
        <HorizontalBreakdownChart
          color={chartPalette.warning}
          data={priorityData}
          emptyText="No priority distribution is available."
          height={330}
          labelWidth={106}
        />
      </Box>
    </ReportCard>
  );
};

const FlowBreakdownSection = ({
  report,
}: {
  report: WorkspaceCommandCenterReport;
}) => {
  const { priorityData, statusData } = useFlowBreakdownData(report);

  return (
    <Box className="grid gap-5 @6xl:grid-cols-2">
      <StatusBreakdownCard statusData={statusData} />
      <PriorityDistributionCard priorityData={priorityData} />
    </Box>
  );
};

const PlanningSection = ({
  report,
  sprintTermPlural,
  objectiveTermPlural,
}: {
  report: WorkspaceCommandCenterReport;
  sprintTermPlural: string;
  objectiveTermPlural: string;
}) => {
  const objectives = report.objectives.keyResultsProgress.slice(0, 8);
  const sprints = report.sprints.sprintProgress.slice(0, 8);
  const objectiveChartData = objectives.map((objective) => ({
    completed: objective.completed,
    label: objective.objectiveName,
    remaining: Math.max(objective.total - objective.completed, 0),
  }));
  const sprintChartData = sprints.map((sprint) => ({
    completed: sprint.completed,
    label: sprint.sprintName,
    remaining: Math.max(sprint.total - sprint.completed, 0),
  }));

  return (
    <Box className="grid gap-5 @6xl:grid-cols-2">
      <ReportCard>
        <SectionTitle
          description={`Progress against active ${objectiveTermPlural} and their key results.`}
        >
          {titleCase(objectiveTermPlural)} progress
        </SectionTitle>
        <Box className="mt-5">
          <ProgressComparisonChart
            data={objectiveChartData}
            emptyText={`No ${objectiveTermPlural} visible for this filter.`}
          />
        </Box>
      </ReportCard>

      <ReportCard>
        <SectionTitle
          description={`Progress and health across active ${sprintTermPlural}.`}
        >
          {titleCase(sprintTermPlural)} progress
        </SectionTitle>
        <Box className="mt-5">
          <ProgressComparisonChart
            data={sprintChartData}
            emptyText={`No ${sprintTermPlural} visible for this filter.`}
          />
        </Box>
      </ReportCard>
    </Box>
  );
};

const ProviderChart = ({
  providers,
}: {
  providers: RequestProviderPerformance[];
}) => {
  const { resolvedTheme } = useTheme();
  const gridStroke = getGridStroke(resolvedTheme);
  const data: ProviderChartRow[] = providers
    .slice()
    .sort((a, b) => b.totalRequests - a.totalRequests)
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

  if (!data.length) {
    return <EmptyState>No request source data is available.</EmptyState>;
  }

  return (
    <ResponsiveContainer
      height={Math.max(300, data.length * 42 + 90)}
      width="100%"
    >
      <BarChart
        data={data}
        layout="vertical"
        margin={{ top: 8, right: 18, left: -12, bottom: 4 }}
      >
        <CartesianGrid
          horizontal={false}
          stroke={gridStroke}
          strokeDasharray="3 3"
        />
        <XAxis
          axisLine={false}
          domain={[0, 100]}
          tick={{ fontSize: 12 }}
          tickFormatter={(value: number) => `${value}%`}
          tickLine={false}
          type="number"
        />
        <YAxis
          axisLine={false}
          dataKey="provider"
          tick={{ fontSize: 12 }}
          tickFormatter={(value: string) => truncateChartLabel(value, 18)}
          tickLine={false}
          type="category"
          width={96}
        />
        <Tooltip
          content={<ProviderTooltip />}
          cursor={{ fill: getCursorFill(resolvedTheme) }}
        />
        <Bar
          dataKey="acceptedShare"
          fill={chartPalette.success}
          name="Accepted"
          radius={[0, 0, 0, 0]}
          stackId="requests"
        />
        <Bar
          dataKey="pendingShare"
          fill={chartPalette.warning}
          name="Pending"
          radius={[0, 0, 0, 0]}
          stackId="requests"
        />
        <Bar
          dataKey="staleShare"
          fill={chartPalette.danger}
          name="Stale"
          radius={[0, 4, 4, 0]}
          stackId="requests"
        />
      </BarChart>
    </ResponsiveContainer>
  );
};

const ProviderTooltip = ({ active, payload }: TooltipProps<number, string>) => {
  const row = payload?.[0]?.payload as ProviderChartRow | undefined;

  if (!active || !row) {
    return null;
  }

  const items = [
    {
      color: chartPalette.success,
      count: row.accepted,
      label: "Accepted",
      share: row.acceptedShare,
    },
    {
      color: chartPalette.warning,
      count: row.pending,
      label: "Pending",
      share: row.pendingShare,
    },
    {
      color: chartPalette.danger,
      count: row.stale,
      label: "Stale",
      share: row.staleShare,
    },
  ];

  return (
    <Box className="border-border dark:bg-surface-elevated z-50 min-w-44 rounded-lg border bg-white px-3 py-3 text-[0.95rem] shadow-lg">
      <Flex align="center" className="mb-2 gap-2" justify="between">
        <Text className="font-medium">{row.provider}</Text>
        <Text color="muted">{formatNumber(row.total)} total</Text>
      </Flex>
      <Box className="space-y-1">
        {items.map((item) => (
          <Flex align="center" className="gap-2" key={item.label}>
            <Box
              className="size-2.5 rounded-sm"
              style={{ backgroundColor: item.color }}
            />
            <Text color="muted">
              {item.label}: {formatNumber(item.count)} ({Math.round(item.share)}
              %)
            </Text>
          </Flex>
        ))}
      </Box>
    </Box>
  );
};

const ProviderLegend = () => {
  const items = [
    { color: chartPalette.success, label: "Accepted" },
    { color: chartPalette.warning, label: "Pending" },
    { color: chartPalette.danger, label: "Stale" },
  ];

  return (
    <Flex align="center" className="mt-3 flex-wrap gap-x-5 gap-y-2">
      {items.map((item) => (
        <Flex align="center" className="gap-2" key={item.label}>
          <Box
            className="size-2.5 rounded-full"
            style={{ backgroundColor: item.color }}
          />
          <Text className="text-foreground/80 text-[0.92rem]">
            {item.label}
          </Text>
        </Flex>
      ))}
    </Flex>
  );
};

const TeamWorkloadChart = ({ teams }: { teams: TeamWorkloadSummary[] }) => {
  return (
    <HorizontalBreakdownChart
      color={chartPalette.primary}
      data={teams.slice(0, 8).map((team) => ({
        color:
          team.overdueStories > 0 ? chartPalette.warning : chartPalette.primary,
        label: team.teamName,
        value: team.openStories,
      }))}
      emptyText="No team workload is visible for these filters."
      height={260}
      labelWidth={106}
    />
  );
};

const EngagementCountChart = ({
  color,
  emptyText,
  items,
}: {
  color: string;
  emptyText: string;
  items: WorkspaceEngagementCount[];
}) => {
  return (
    <HorizontalBreakdownChart
      color={color}
      data={items.slice(0, 8).map((item) => ({
        label: titleCase(item.name),
        value: item.count,
      }))}
      emptyText={emptyText}
      height={300}
    />
  );
};

const EngagementUserRow = ({ user }: { user: WorkspaceEngagementUser }) => {
  const { withWorkspace } = useWorkspacePath();
  const displayName = getDisplayName(user);

  return (
    <Link href={withWorkspace(`/profile/${user.userId}`)}>
      <Flex
        align="center"
        className="border-border hover:bg-surface-muted/60 gap-3 border-b-[0.5px] px-1 py-3 transition last:border-b-0"
      >
        <Avatar name={displayName} size="sm" src={user.avatarUrl} />
        <Box className="min-w-0 flex-1">
          <Text className="truncate" fontWeight="medium">
            {displayName}
          </Text>
          <Text color="muted">{user.username}</Text>
        </Box>
        <Badge
          className="bg-transparent"
          color="info"
          rounded="full"
          size="sm"
          variant="outline"
        >
          {formatNumber(user.events)}
        </Badge>
      </Flex>
    </Link>
  );
};

const RiskSummary = ({
  report,
  storyTermPlural,
}: {
  report: WorkspaceCommandCenterReport;
  storyTermPlural: string;
}) => {
  const risks = report.pulse.risks.slice(0, 5);

  return (
    <ReportCard>
      <SectionTitle
        description={`Highest-signal risks across ${storyTermPlural}, requests, workload, objectives, and sprints.`}
      >
        What needs attention
      </SectionTitle>
      {risks.length ? (
        <Box className="mt-4">
          {risks.map((risk) => (
            <Box
              className="border-border border-b-[0.5px] px-1 py-3.5 last:border-b-0"
              key={risk.kind}
            >
              <Flex align="start" className="gap-3" justify="between">
                <Box className="min-w-0">
                  <Text fontWeight="medium">{risk.title}</Text>
                  <Text className="mt-1 leading-5" color="muted">
                    {risk.description}
                  </Text>
                </Box>
                <Badge
                  className="shrink-0 bg-transparent"
                  color={risk.severity === "high" ? "danger" : "tertiary"}
                  rounded="sm"
                  size="sm"
                  variant="outline"
                >
                  {formatNumber(risk.count)}
                </Badge>
              </Flex>
            </Box>
          ))}
        </Box>
      ) : (
        <Box className="mt-5">
          <EmptyState>No active risks are currently detected.</EmptyState>
        </Box>
      )}
    </ReportCard>
  );
};

const OverviewReadout = ({
  report,
  storyTermPlural,
  sprintTermPlural,
  objectiveTermPlural,
}: {
  report: WorkspaceCommandCenterReport;
  storyTermPlural: string;
  sprintTermPlural: string;
  objectiveTermPlural: string;
}) => {
  const topProvider = report.requests.providers.at(0);

  return (
    <Box>
      <SectionTitle description="Fast answers for the operating questions that usually come next.">
        Operating readout
      </SectionTitle>
      <Box className="mt-5 grid gap-3 md:grid-cols-2 @6xl:grid-cols-4">
        <MiniMetric
          description="Work with no owner."
          label={`Unassigned ${storyTermPlural}`}
          value={formatNumber(report.workload.summary.unassignedStories)}
        />
        <MiniMetric
          description="Work without estimates."
          label={`Unestimated ${storyTermPlural}`}
          value={formatNumber(report.workload.summary.unestimatedStories)}
        />
        <MiniMetric
          description="Urgent or high-priority work."
          label="Priority pressure"
          value={formatNumber(
            report.workload.summary.urgentStories +
              report.workload.summary.highPriorityStories,
          )}
        />
        <MiniMetric
          description={
            topProvider
              ? `${formatNumber(topProvider.totalRequests)} from ${titleCase(topProvider.provider)}`
              : "No connected source activity."
          }
          label="Top request source"
          value={topProvider ? titleCase(topProvider.provider) : "None"}
        />
        <MiniMetric
          description={`${formatNumber(report.pulse.objectives.offTrackObjectives)} off track.`}
          label={`${titleCase(objectiveTermPlural)} at risk`}
          value={formatNumber(report.pulse.objectives.atRiskObjectives)}
        />
        <MiniMetric
          description={`${formatNumber(report.pulse.sprints.overdueSprints)} overdue.`}
          label={`${titleCase(sprintTermPlural)} at risk`}
          value={formatNumber(report.pulse.sprints.atRiskSprints)}
        />
        <MiniMetric
          description="Requests not handled yet."
          label="Pending requests"
          value={formatNumber(report.requests.pendingRequests)}
        />
        <MiniMetric
          description="Requests aging without action."
          label="Stale requests"
          value={formatNumber(
            report.requests.providers.reduce(
              (sum, provider) => sum + provider.staleRequests,
              0,
            ),
          )}
        />
      </Box>
    </Box>
  );
};

const OverviewTab = ({
  report,
  storyTermPlural,
  sprintTermPlural,
  objectiveTermPlural,
}: {
  report: WorkspaceCommandCenterReport;
  storyTermPlural: string;
  sprintTermPlural: string;
  objectiveTermPlural: string;
}) => {
  const { priorityData } = useFlowBreakdownData(report);

  return (
    <Box className="space-y-5">
      <OverviewReadout
        objectiveTermPlural={objectiveTermPlural}
        report={report}
        sprintTermPlural={sprintTermPlural}
        storyTermPlural={storyTermPlural}
      />
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <PriorityDistributionCard priorityData={priorityData} />
        <ReportCard>
          <SectionTitle description="Whether open work is spread across the team or concentrated with a few people.">
            Workload concentration
          </SectionTitle>
          <Box className="mt-5">
            <WorkloadChart report={report} />
          </Box>
        </ReportCard>
      </Box>
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <RiskSummary report={report} storyTermPlural={storyTermPlural} />
        <ReportCard>
          <SectionTitle description="How creation and completion are moving across the selected time window.">
            Delivery trend
          </SectionTitle>
          <Box className="mt-5">
            <DeliveryChart report={report} />
          </Box>
        </ReportCard>
      </Box>
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <ReportCard>
          <SectionTitle description="Which connected sources create work and how much is still waiting.">
            Request source performance
          </SectionTitle>
          <Box className="mt-5">
            <ProviderChart providers={report.requests.providers} />
          </Box>
          <ProviderLegend />
        </ReportCard>
      </Box>
    </Box>
  );
};

const WorkloadTab = ({ report }: { report: WorkspaceCommandCenterReport }) => {
  const topMembers = report.workload.members.slice(0, 10);
  const topTeams = report.workload.teams.slice(0, 10);

  return (
    <Box className="space-y-5">
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <ReportCard>
          <SectionTitle description="Open and overdue work by assignee.">
            Workload by person
          </SectionTitle>
          <Box className="mt-5">
            <WorkloadChart report={report} />
          </Box>
        </ReportCard>
        <ReportCard>
          <SectionTitle description="Open work concentration across teams.">
            Workload by team
          </SectionTitle>
          <Box className="mt-5">
            <TeamWorkloadChart teams={topTeams} />
          </Box>
        </ReportCard>
      </Box>
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <ReportCard>
          <SectionTitle description="People with the most open work, estimate load, overdue work, and unestimated work.">
            Workload by person
          </SectionTitle>
          <Box className="mt-3">
            {topMembers.length ? (
              topMembers.map((member) => (
                <WorkloadMemberRow key={member.userId} member={member} />
              ))
            ) : (
              <EmptyState>
                No assigned work is visible for these filters.
              </EmptyState>
            )}
          </Box>
        </ReportCard>

        <ReportCard>
          <SectionTitle description="Team load, ownership gaps, and estimate concentration.">
            Workload by team
          </SectionTitle>
          <Box className="mt-3">
            {topTeams.length ? (
              topTeams.map((team) => (
                <TeamWorkloadRow key={team.teamId} team={team} />
              ))
            ) : (
              <EmptyState>
                No team workload is visible for these filters.
              </EmptyState>
            )}
          </Box>
        </ReportCard>
      </Box>
    </Box>
  );
};

const FlowTab = ({ report }: { report: WorkspaceCommandCenterReport }) => {
  return (
    <Box className="space-y-5">
      <FlowBreakdownSection report={report} />
      <ReportCard>
        <SectionTitle description="Completion and creation movement for the current window.">
          Delivery trend
        </SectionTitle>
        <Box className="mt-5">
          <DeliveryChart report={report} />
        </Box>
      </ReportCard>
    </Box>
  );
};

const PlanningTab = ({
  report,
  sprintTermPlural,
  objectiveTermPlural,
}: {
  report: WorkspaceCommandCenterReport;
  sprintTermPlural: string;
  objectiveTermPlural: string;
}) => {
  return (
    <PlanningSection
      objectiveTermPlural={objectiveTermPlural}
      report={report}
      sprintTermPlural={sprintTermPlural}
    />
  );
};

const EngagementTab = ({
  report,
}: {
  report: WorkspaceCommandCenterReport;
}) => {
  return (
    <Box className="space-y-5">
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <ReportCard>
          <SectionTitle description="Provider mix, acceptance, and stale requests across connected sources.">
            Request sources
          </SectionTitle>
          <Box className="mt-5">
            <ProviderChart providers={report.requests.providers} />
          </Box>
          <ProviderLegend />
        </ReportCard>

        <ReportCard>
          <SectionTitle description="First-party workspace events captured for analytics and later questions.">
            Workspace engagement
          </SectionTitle>
          <Box className="mt-5 grid grid-cols-2 gap-3">
            <MiniMetric
              description="Total captured activity."
              label="Tracked events"
              value={formatNumber(report.engagement.totalEvents)}
            />
            <MiniMetric
              description="People with tracked activity."
              label="Active users"
              value={formatNumber(report.engagement.uniqueUsers)}
            />
          </Box>
          {report.engagement.topUsers.length ? (
            <Box className="mt-5">
              <Text className="mb-2" fontWeight="medium">
                Most active people
              </Text>
              {report.engagement.topUsers.slice(0, 6).map((user) => (
                <EngagementUserRow key={user.userId} user={user} />
              ))}
            </Box>
          ) : null}
        </ReportCard>
      </Box>

      <Box className="grid gap-5 @6xl:grid-cols-2">
        <ReportCard>
          <SectionTitle description="Which tracked events are happening most often.">
            Event mix
          </SectionTitle>
          <Box className="mt-5">
            <EngagementCountChart
              color={chartPalette.info}
              emptyText="No event names have been tracked yet."
              items={report.engagement.eventsByName}
            />
          </Box>
        </ReportCard>
        <ReportCard>
          <SectionTitle description="Where tracked activity is happening in the product.">
            Surface mix
          </SectionTitle>
          <Box className="mt-5">
            <EngagementCountChart
              color={chartPalette.violet}
              emptyText="No surfaces have been tracked yet."
              items={report.engagement.eventsBySurface}
            />
          </Box>
        </ReportCard>
      </Box>
    </Box>
  );
};

export const CommandCenterReport = () => {
  const filters = useAppliedFilters();
  const { getTermDisplay } = useTerminology();
  const { trackEvent } = useWorkspaceAnalyticsEvent();
  const {
    data: report,
    isError,
    isFetching,
    isPending,
    refetch,
  } = useCommandCenterReport(filters);
  const filterSignature = useMemo(
    () =>
      [
        filters.startDate ?? "",
        filters.endDate ?? "",
        filters.teamIds?.join(",") ?? "",
        filters.assigneeIds?.join(",") ?? "",
        filters.sprintIds?.join(",") ?? "",
        filters.objectiveIds?.join(",") ?? "",
      ].join("|"),
    [
      filters.assigneeIds,
      filters.endDate,
      filters.objectiveIds,
      filters.sprintIds,
      filters.startDate,
      filters.teamIds,
    ],
  );

  useEffect(() => {
    trackEvent({
      eventName: "analytics_command_center_viewed",
      properties: {
        hasFilters: Boolean(filterSignature.replaceAll("|", "")),
      },
      surface: "analytics_command_center",
    });
  }, [filterSignature, trackEvent]);

  if (isPending) {
    return <CommandCenterSkeleton />;
  }

  if (isError) {
    return (
      <ReportCard className="mt-3">
        <Text className="mb-1" fontSize="lg" fontWeight="medium">
          Analytics are unavailable
        </Text>
        <Text color="muted">
          The detailed workspace report could not be loaded right now.
        </Text>
        <Button className="mt-4" onClick={() => void refetch()}>
          Try again
        </Button>
      </ReportCard>
    );
  }

  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const sprintTermPlural = getTermDisplay("sprintTerm", { variant: "plural" });
  const objectiveTermPlural = getTermDisplay("objectiveTerm", {
    variant: "plural",
  });
  const completion = completionRate(
    report.overview.metrics.completedStories,
    report.overview.metrics.totalStories,
  );
  const overloadedMembers = report.workload.risks.overloadedMembers.length;
  const activeRiskCount = report.pulse.risks.length;
  const acceptedRequestRate = completionRate(
    report.requests.acceptedRequests,
    report.requests.totalRequests,
  );

  return (
    <Box className="pt-3 pb-5">
      {report.sectionErrors.length > 0 ? (
        <ReportCard className="border-warning/35 bg-warning/5 dark:bg-warning/10 mb-5">
          <Text fontWeight="medium">Some analytics sections are delayed</Text>
          <Text className="mt-1 leading-5" color="muted">
            {report.sectionErrors
              .map((sectionError) => titleCase(sectionError.section))
              .join(", ")}
          </Text>
        </ReportCard>
      ) : null}

      <Box className="mb-6">
        <Flex className="flex-col items-start justify-between gap-3 @4xl:flex-row @4xl:items-center">
          <Flex align="center" className="gap-2">
            <Text
              as="h2"
              className="text-2xl @3xl:text-3xl"
              fontWeight="medium"
            >
              Analytics
            </Text>
            {isFetching ? (
              <Badge
                className="bg-transparent"
                color="tertiary"
                rounded="full"
                size="sm"
                variant="outline"
              >
                Refreshing
              </Badge>
            ) : null}
          </Flex>
          <Filters />
        </Flex>
        <Text className="mt-2 max-w-3xl leading-6" color="muted" fontSize="lg">
          Detailed workload, delivery, planning, source, and engagement reports
          for the selected workspace.
        </Text>
      </Box>

      <Box className="mb-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard
          description={`${formatNumber(report.overview.metrics.completedStories)} completed in period`}
          label={`Open ${storyTermPlural}`}
          value={formatNumber(report.workload.summary.totalOpenStories)}
        />
        <MetricCard
          accent={`${formatNumber(report.overview.metrics.totalStories)} total`}
          description="Across tracked work"
          label="Completion rate"
          value={formatPercent(completion)}
        />
        <MetricCard
          description={`${formatNumber(report.pulse.summary.overdueStories)} overdue / ${formatNumber(report.pulse.summary.blockedStories)} blocked`}
          label="Active risks"
          value={formatNumber(activeRiskCount)}
        />
        <MetricCard
          description={`${formatNumber(report.workload.summary.totalEstimate)} estimate loaded`}
          label="Overloaded people"
          value={formatNumber(overloadedMembers)}
        />
        <MetricCard
          description={`${formatNumber(report.requests.pendingRequests)} pending, ${formatNumber(report.requests.declinedRequests)} declined`}
          label="Request intake"
          value={formatNumber(report.requests.totalRequests)}
        />
        <MetricCard
          accent={formatPercent(acceptedRequestRate)}
          description="Accepted requests"
          label="Request acceptance"
          value={formatNumber(report.requests.acceptedRequests)}
        />
        <MetricCard
          description={`${formatNumber(report.engagement.uniqueUsers)} active people`}
          label="Tracked events"
          value={formatNumber(report.engagement.totalEvents)}
        />
        <MetricCard
          description="Tracked people"
          label="Active users"
          value={formatNumber(report.engagement.uniqueUsers)}
        />
      </Box>

      <Tabs defaultValue="overview">
        <Tabs.List className="mx-0 mb-5 md:mx-0">
          <Tabs.Tab value="overview">Overview</Tabs.Tab>
          <Tabs.Tab value="workload">Workload</Tabs.Tab>
          <Tabs.Tab value="flow">Flow</Tabs.Tab>
          <Tabs.Tab value="planning">Planning</Tabs.Tab>
          <Tabs.Tab value="engagement">Engagement</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="overview">
          <OverviewTab
            objectiveTermPlural={objectiveTermPlural}
            report={report}
            sprintTermPlural={sprintTermPlural}
            storyTermPlural={storyTermPlural}
          />
        </Tabs.Panel>
        <Tabs.Panel value="workload">
          <WorkloadTab report={report} />
        </Tabs.Panel>
        <Tabs.Panel value="flow">
          <FlowTab report={report} />
        </Tabs.Panel>
        <Tabs.Panel value="planning">
          <PlanningTab
            objectiveTermPlural={objectiveTermPlural}
            report={report}
            sprintTermPlural={sprintTermPlural}
          />
        </Tabs.Panel>
        <Tabs.Panel value="engagement">
          <EngagementTab report={report} />
        </Tabs.Panel>
      </Tabs>
    </Box>
  );
};

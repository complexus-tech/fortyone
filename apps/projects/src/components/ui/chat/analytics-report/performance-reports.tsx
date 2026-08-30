import { Box, Flex, Text } from "ui";
import type { AnalyticsReportOutput } from "./model";
import { asRecord, asRows, COLORS, completionRate } from "./model";
import {
  ChartSection,
  CompactBarChart,
  CompactLineChart,
  KeyValueList,
  MetricGrid,
  RiskList,
} from "./primitives";

type ReportProps = {
  output: AnalyticsReportOutput;
  title: string;
};

export const WorkspacePerformanceReport = ({
  output,
  storyTermPlural,
  title,
}: ReportProps & { storyTermPlural: string }) => {
  const overview = asRecord(output.overview);
  const metrics = asRecord(overview.metrics);

  return (
    <Box className="mt-3 space-y-3">
      <Text className="text-xl font-semibold">{title}</Text>
      <MetricGrid
        metrics={[
          {
            label: storyTermPlural,
            value: Number(metrics.totalStories ?? 0),
          },
          {
            label: "Completed",
            value: Number(metrics.completedStories ?? 0),
          },
          {
            label: "Completion",
            value: completionRate(
              metrics.completedStories,
              metrics.totalStories,
            ),
          },
          {
            label: "Objectives",
            value: Number(metrics.activeObjectives ?? 0),
          },
          { label: "Sprints", value: Number(metrics.activeSprints ?? 0) },
          { label: "Members", value: Number(metrics.totalTeamMembers ?? 0) },
        ]}
      />
      <ChartSection title="Completion trend">
        <CompactLineChart
          data={asRows(overview.completionTrend)}
          lines={[
            { key: "completed", color: COLORS.success, name: "Completed" },
            { key: "total", color: COLORS.primary, name: "Total" },
          ]}
          xKey="date"
        />
      </ChartSection>
    </Box>
  );
};

export const PulseReport = ({ output, title }: ReportProps) => {
  const report = asRecord(output.report);
  const summary = asRecord(report.summary);
  const workload = asRecord(report.workload);
  const members = [...asRows(workload.members)]
    .sort(
      (firstMember, secondMember) =>
        Number(secondMember.openStories ?? 0) -
        Number(firstMember.openStories ?? 0),
    )
    .slice(0, 5);

  return (
    <Box className="mt-3 space-y-3">
      <Text className="text-xl font-semibold">{title}</Text>
      <MetricGrid
        metrics={[
          {
            label: "Open work",
            value: Number(summary.openStories ?? 0),
          },
          {
            label: "Overdue",
            value: Number(summary.overdueStories ?? 0),
          },
          {
            label: "Blocked",
            value: Number(summary.blockedStories ?? 0),
          },
          {
            label: "At-risk sprints",
            value: Number(summary.atRiskSprints ?? 0),
          },
          {
            label: "At-risk objectives",
            value: Number(summary.atRiskObjectives ?? 0),
          },
          {
            label: "Pending requests",
            value: Number(summary.pendingRequests ?? 0),
          },
        ]}
      />
      <ChartSection
        description="Shows the five members carrying the most open work."
        title="Highest workload"
      >
        <CompactBarChart
          bars={[
            { key: "openStories", color: COLORS.primary, name: "Open" },
            { key: "overdueStories", color: "#EF4444", name: "Overdue" },
          ]}
          data={members}
          maxLabelLength={12}
          xKey="username"
        />
      </ChartSection>
      <ChartSection
        description="The most important delivery issues that need attention."
        title="Active risks"
      >
        <RiskList risks={asRows(report.risks)} />
      </ChartSection>
    </Box>
  );
};

export const WorkloadAnalysisReport = ({ output, title }: ReportProps) => {
  const analysis = asRecord(output.analysis);
  const summary = asRecord(analysis.summary);
  const risks = asRecord(analysis.risks);
  const members = [...asRows(analysis.members)]
    .sort(
      (firstMember, secondMember) =>
        Number(secondMember.openStories ?? 0) -
        Number(firstMember.openStories ?? 0),
    )
    .slice(0, 5);

  return (
    <Box className="mt-3 space-y-3">
      <Text className="text-xl font-semibold">{title}</Text>
      <MetricGrid
        metrics={[
          {
            label: "Open work",
            value: Number(summary.totalOpenStories ?? 0),
          },
          {
            label: "Complexity",
            value: Number(summary.totalEstimate ?? 0),
          },
          {
            label: "Overdue",
            value: Number(summary.overdueStories ?? 0),
          },
          {
            label: "Urgent",
            value: Number(summary.urgentStories ?? 0),
          },
          {
            label: "Unassigned",
            value: Number(summary.unassignedStories ?? 0),
          },
          {
            label: "No complexity",
            value: Number(summary.unestimatedStories ?? 0),
          },
        ]}
      />
      <ChartSection
        description="Shows the five members carrying the most open work."
        title="Workload by member"
      >
        <CompactBarChart
          bars={[
            { key: "openStories", color: COLORS.primary, name: "Open" },
            { key: "overdueStories", color: "#EF4444", name: "Overdue" },
          ]}
          data={members}
          maxLabelLength={12}
          xKey="username"
        />
      </ChartSection>
      <ChartSection title="Workload risks">
        <KeyValueList
          rows={[
            {
              label: "Overloaded members",
              value: asRows(risks.overloadedMembers).length,
            },
            {
              label: "Members with overdue work",
              value: asRows(risks.overdueMembers).length,
            },
            {
              label: "High-priority work",
              value: Number(risks.highPriorityStories ?? 0),
            },
          ]}
        />
      </ChartSection>
    </Box>
  );
};

export const StoryPerformanceReport = ({ output, title }: ReportProps) => {
  const analytics = asRecord(output.analytics);

  return (
    <Box className="mt-3 space-y-3">
      <Text className="text-xl font-semibold">{title}</Text>
      <ChartSection title="Status breakdown">
        <CompactBarChart
          bars={[{ key: "count", color: COLORS.primary }]}
          data={asRows(analytics.statusBreakdown)}
          xKey="statusName"
        />
      </ChartSection>
      <ChartSection title="Completion by team">
        <CompactBarChart
          bars={[
            { key: "completed", color: COLORS.success, name: "Completed" },
            { key: "total", color: COLORS.muted, name: "Total" },
          ]}
          data={asRows(analytics.completionByTeam)}
          xKey="teamName"
        />
      </ChartSection>
    </Box>
  );
};

export const ObjectiveProgressReport = ({ output, title }: ReportProps) => {
  const progress = asRecord(output.progress);

  return (
    <Box className="mt-3 space-y-3">
      <Text className="text-xl font-semibold">{title}</Text>
      <ChartSection title="Key-result progress">
        <CompactBarChart
          bars={[
            {
              key: "avgProgress",
              color: COLORS.primary,
              name: "Avg progress",
            },
          ]}
          data={asRows(progress.keyResultsProgress)}
          xKey="objectiveName"
        />
      </ChartSection>
      <ChartSection title="Progress by team">
        <CompactBarChart
          bars={[
            { key: "completed", color: COLORS.success, name: "Completed" },
            { key: "objectives", color: COLORS.muted, name: "Objectives" },
          ]}
          data={asRows(progress.progressByTeam)}
          xKey="teamName"
        />
      </ChartSection>
    </Box>
  );
};

export const TeamPerformanceReport = ({ output, title }: ReportProps) => {
  const performance = asRecord(output.performance);
  const focusMember = asRecord(output.focusMember);

  return (
    <Box className="mt-3 space-y-3">
      <Flex align="center" justify="between">
        <Text className="text-xl font-semibold">{title}</Text>
        {focusMember.userId ? (
          <Text className="text-muted text-base">
            {completionRate(focusMember.completed, focusMember.assigned)}{" "}
            complete
          </Text>
        ) : null}
      </Flex>
      <ChartSection title="Team workload">
        <CompactBarChart
          bars={[
            { key: "completed", color: COLORS.success, name: "Completed" },
            { key: "assigned", color: COLORS.primary, name: "Assigned" },
            { key: "capacity", color: COLORS.warning, name: "Capacity" },
          ]}
          data={asRows(performance.teamWorkload)}
          xKey="teamName"
        />
      </ChartSection>
      <ChartSection title="Member contributions">
        <CompactBarChart
          bars={[
            { key: "completed", color: COLORS.success, name: "Completed" },
            { key: "assigned", color: COLORS.primary, name: "Assigned" },
          ]}
          data={asRows(performance.memberContributions).slice(0, 5)}
          maxLabelLength={12}
          xKey="username"
        />
      </ChartSection>
    </Box>
  );
};

export const SprintPerformanceReport = ({ output, title }: ReportProps) => {
  const analytics = asRecord(output.analytics);

  return (
    <Box className="mt-3 space-y-3">
      <Text className="text-xl font-semibold">{title}</Text>
      <ChartSection title="Sprint progress">
        <CompactBarChart
          bars={[
            { key: "completed", color: COLORS.success, name: "Completed" },
            { key: "total", color: COLORS.muted, name: "Total" },
          ]}
          data={asRows(analytics.sprintProgress)}
          xKey="sprintName"
        />
      </ChartSection>
      <ChartSection title="Combined burndown">
        <CompactLineChart
          data={asRows(analytics.combinedBurndown)}
          lines={[
            { key: "planned", color: COLORS.muted, name: "Planned" },
            { key: "actual", color: COLORS.primary, name: "Actual" },
          ]}
          xKey="date"
        />
      </ChartSection>
    </Box>
  );
};

export const TimelineTrendsReport = ({
  output,
  storyTermCapitalized,
  title,
}: ReportProps & { storyTermCapitalized: string }) => {
  const trends = asRecord(output.trends);

  return (
    <Box className="mt-3 space-y-3">
      <Text className="text-xl font-semibold">{title}</Text>
      <ChartSection title={`${storyTermCapitalized} completion`}>
        <CompactLineChart
          data={asRows(trends.storyCompletion)}
          lines={[
            { key: "completed", color: COLORS.success, name: "Completed" },
            { key: "created", color: COLORS.primary, name: "Created" },
          ]}
          xKey="date"
        />
      </ChartSection>
      <ChartSection title="Key metrics">
        <CompactLineChart
          data={asRows(trends.keyMetricsTrend)}
          lines={[
            {
              key: "activeUsers",
              color: COLORS.primary,
              name: "Active users",
            },
            {
              key: "storiesPerDay",
              color: COLORS.success,
              name: "Stories/day",
            },
          ]}
          xKey="date"
        />
      </ChartSection>
    </Box>
  );
};

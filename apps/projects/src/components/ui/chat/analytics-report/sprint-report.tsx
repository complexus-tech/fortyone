import { Box, Text } from "ui";
import { BurndownChart } from "@/components/ui/burndown-chart";
import type { AnalyticsReportOutput } from "./model";
import {
  asRecord,
  asRows,
  asSingleSprintBurndown,
  asWorkingDays,
  COLORS,
} from "./model";
import {
  ChartSection,
  CompactBarChart,
  EmptyChart,
  MetricGrid,
} from "./primitives";

export const SingleSprintAnalyticsReport = ({
  output,
  title,
}: {
  output: AnalyticsReportOutput;
  title: string;
}) => {
  const analytics = asRecord(output.analytics ?? output.analyticsReport);
  const overview = asRecord(analytics.overview);
  const storyBreakdown = asRecord(analytics.storyBreakdown);
  const burndown = asSingleSprintBurndown(analytics.burndown);
  const teamAllocation = [...asRows(analytics.teamAllocation)]
    .filter(
      (member) =>
        Number(member.assigned ?? 0) > 0 || Number(member.completed ?? 0) > 0,
    )
    .sort((firstMember, secondMember) => {
      const assignedDifference =
        Number(secondMember.assigned ?? 0) - Number(firstMember.assigned ?? 0);
      if (assignedDifference) return assignedDifference;

      return (
        Number(secondMember.completed ?? 0) - Number(firstMember.completed ?? 0)
      );
    })
    .slice(0, 5);

  return (
    <Box className="mt-3 space-y-3">
      <Text className="text-xl font-semibold">{title}</Text>
      <MetricGrid
        metrics={[
          {
            label: "Completion",
            value: `${Math.round(Number(overview.completionPercentage ?? 0))}%`,
          },
          {
            label: "Remaining",
            value: Number(overview.daysRemaining ?? 0),
          },
          { label: "Total", value: Number(storyBreakdown.total ?? 0) },
          {
            label: "Completed",
            value: Number(storyBreakdown.completed ?? 0),
          },
          {
            label: "In progress",
            value: Number(storyBreakdown.inProgress ?? 0),
          },
          { label: "Blocked", value: Number(storyBreakdown.blocked ?? 0) },
        ]}
      />
      <ChartSection
        description="Tracks remaining work against the ideal sprint pace."
        title="Burndown"
      >
        {burndown.length ? (
          <BurndownChart
            burndownData={burndown}
            className="h-52"
            workingDays={asWorkingDays(analytics.workingDays)}
          />
        ) : (
          <EmptyChart />
        )}
      </ChartSection>
      <ChartSection
        description="Shows completed and assigned work for each team member."
        title="Team allocation"
      >
        <CompactBarChart
          bars={[
            { key: "completed", color: COLORS.success, name: "Completed" },
            { key: "assigned", color: COLORS.primary, name: "Assigned" },
          ]}
          data={teamAllocation}
          maxLabelLength={12}
          xKey="username"
        />
      </ChartSection>
    </Box>
  );
};

import { useMemo } from "react";
import { Box } from "ui";
import { useStatuses } from "@/lib/hooks/statuses";
import type { WorkspaceCommandCenterReport } from "../../types";
import {
  DeliveryChart,
  HorizontalBreakdownChart,
  ProgressComparisonChart,
} from "./charts";
import {
  buildCompletionTrendChartData,
  buildObjectiveProgressChartData,
  buildPriorityDistributionData,
  buildSprintProgressChartData,
  buildStatusBreakdownData,
  chartPalette,
  titleCase,
  summarizeCycleTime,
  formatNumber,
} from "./model";
import type { ChartBreakdownRow } from "./model";
import { ReportCard, SectionTitle, MetricCard } from "./primitives";

const cycleDaysFormatter = new Intl.NumberFormat(undefined, {
  maximumFractionDigits: 1,
});

export const useFlowBreakdownData = (report: WorkspaceCommandCenterReport) => {
  const { data: statuses = [] } = useStatuses();

  return useMemo(
    () => ({
      priorityData: buildPriorityDistributionData(
        report.stories.priorityDistribution,
      ),
      statusData: buildStatusBreakdownData(
        report.stories.statusBreakdown,
        statuses,
      ),
    }),
    [
      report.stories.priorityDistribution,
      report.stories.statusBreakdown,
      statuses,
    ],
  );
};

const StatusBreakdownCard = ({
  statusData,
}: {
  statusData: ChartBreakdownRow[];
}) => {
  return (
    <ReportCard>
      <SectionTitle description="Status of work created in the selected period.">
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

export const PriorityDistributionCard = ({
  priorityData,
}: {
  priorityData: ChartBreakdownRow[];
}) => {
  return (
    <ReportCard>
      <SectionTitle description="Priority of work created in the selected period.">
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

export const FlowTab = ({
  report,
}: {
  report: WorkspaceCommandCenterReport;
}) => {
  const cycleTime = summarizeCycleTime(report.trends.keyMetricsTrend);
  const completed = report.overview.metrics.completedInPeriod;
  return (
    <Box className="space-y-5">
      <Box className="grid gap-5 md:grid-cols-2">
        <MetricCard
          description={`${formatNumber(cycleTime.samples)} completions with recorded start history`}
          label="Average cycle time"
          value={
            cycleTime.averageDays === null
              ? "Unavailable"
              : `${cycleDaysFormatter.format(cycleTime.averageDays)} days`
          }
        />
        <MetricCard
          description="Includes work created before this period"
          label="Completed in period"
          value={
            completed === undefined ? "Unavailable" : formatNumber(completed)
          }
        />
      </Box>
      <FlowBreakdownSection report={report} />
      <ReportCard>
        <SectionTitle description="Completion and creation movement for the current window.">
          Delivery trend
        </SectionTitle>
        <Box className="mt-5">
          <DeliveryChart
            data={buildCompletionTrendChartData(
              report.overview.completionTrend,
              report.overview.filters,
            )}
          />
        </Box>
      </ReportCard>
    </Box>
  );
};

export const PlanningTab = ({
  objectiveTermPlural,
  report,
  sprintTermPlural,
}: {
  objectiveTermPlural: string;
  report: WorkspaceCommandCenterReport;
  sprintTermPlural: string;
}) => {
  return (
    <Box className="grid gap-5 @6xl:grid-cols-2">
      <ReportCard>
        <SectionTitle
          description={`Key result progress for ${objectiveTermPlural} created in the selected period. Uses the team and ${objectiveTermPlural} filters.`}
        >
          {titleCase(objectiveTermPlural)} progress
        </SectionTitle>
        <Box className="mt-5">
          <ProgressComparisonChart
            data={buildObjectiveProgressChartData(
              report.objectives.keyResultsProgress,
            )}
            emptyText={`No ${objectiveTermPlural} visible for this filter.`}
          />
        </Box>
      </ReportCard>

      <ReportCard>
        <SectionTitle
          description={`Completed and remaining work in ${sprintTermPlural} created in the selected period. Uses the team and ${sprintTermPlural} filters.`}
        >
          {titleCase(sprintTermPlural)} progress
        </SectionTitle>
        <Box className="mt-5">
          <ProgressComparisonChart
            data={buildSprintProgressChartData(report.sprints.sprintProgress)}
            emptyText={`No ${sprintTermPlural} visible for this filter.`}
          />
        </Box>
      </ReportCard>
    </Box>
  );
};

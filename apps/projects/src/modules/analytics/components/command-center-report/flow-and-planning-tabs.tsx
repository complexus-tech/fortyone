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
} from "./model";
import type { ChartBreakdownRow } from "./model";
import { ReportCard, SectionTitle } from "./primitives";

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

export const PriorityDistributionCard = ({
  priorityData,
}: {
  priorityData: ChartBreakdownRow[];
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

export const FlowTab = ({
  report,
}: {
  report: WorkspaceCommandCenterReport;
}) => {
  return (
    <Box className="space-y-5">
      <FlowBreakdownSection report={report} />
      <ReportCard>
        <SectionTitle description="Completion and creation movement for the current window.">
          Delivery trend
        </SectionTitle>
        <Box className="mt-5">
          <DeliveryChart
            data={buildCompletionTrendChartData(
              report.overview.completionTrend,
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
          description={`Progress against active ${objectiveTermPlural} and their key results.`}
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
          description={`Progress and health across active ${sprintTermPlural}.`}
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

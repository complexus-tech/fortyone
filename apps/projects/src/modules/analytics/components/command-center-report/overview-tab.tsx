import { Badge, Box, Flex, Text } from "ui";
import type { WorkspaceCommandCenterReport } from "../../types";
import { DeliveryChart, WorkloadChart } from "./charts";
import {
  buildCompletionTrendChartData,
  buildMemberWorkloadChartData,
  buildProviderChartData,
  formatNumber,
  titleCase,
} from "./model";
import { ProviderChart, ProviderLegend } from "./provider-chart";
import { EmptyState, ReportCard, SectionTitle } from "./primitives";
import {
  PriorityDistributionCard,
  useFlowBreakdownData,
} from "./flow-and-planning-tabs";

const RiskSummary = ({
  objectiveTermPlural,
  report,
  sprintTermPlural,
  storyTermPlural,
}: {
  objectiveTermPlural: string;
  report: WorkspaceCommandCenterReport;
  sprintTermPlural: string;
  storyTermPlural: string;
}) => {
  const risks = report.pulse.risks.slice(0, 5);
  const displayRiskText = (value: string) =>
    value
      .replace(/\bStories\b/g, titleCase(storyTermPlural))
      .replace(/\bstories\b/g, storyTermPlural)
      .replace(/open-story/g, "unfinished work")
      .replace(/\bObjectives\b/g, titleCase(objectiveTermPlural))
      .replace(/\bobjectives\b/g, objectiveTermPlural)
      .replace(/\bsprints\b/g, sprintTermPlural);

  return (
    <ReportCard>
      <SectionTitle
        description={`Risks affecting ${storyTermPlural} and team delivery.`}
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
                  <Text fontWeight="medium">{displayRiskText(risk.title)}</Text>
                  <Text className="mt-1 leading-5" color="muted">
                    {displayRiskText(risk.description)}
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
  objectiveTermPlural,
  report,
  sprintTermPlural,
  storyTermPlural,
}: {
  objectiveTermPlural: string;
  report: WorkspaceCommandCenterReport;
  sprintTermPlural: string;
  storyTermPlural: string;
}) => {
  const topProvider = report.requests.providers.at(0);
  const rows = [
    {
      label: `Unassigned ${storyTermPlural}`,
      value: formatNumber(report.workload.summary.unassignedStories),
    },
    {
      label: "Without complexity",
      value: formatNumber(report.workload.summary.unestimatedStories),
    },
    {
      label: "Urgent or high priority",
      value: formatNumber(
        report.workload.summary.urgentStories +
          report.workload.summary.highPriorityStories,
      ),
    },
    {
      label: `${titleCase(objectiveTermPlural)} at risk`,
      value: formatNumber(report.pulse.summary.atRiskObjectives),
    },
    {
      label: `Overdue ${objectiveTermPlural}`,
      value: formatNumber(report.pulse.objectives.overdueObjectives),
    },
    {
      label: `${titleCase(objectiveTermPlural)} due within 7 days`,
      value: formatNumber(report.pulse.objectives.objectivesDueSoon),
    },
    {
      label: `${titleCase(sprintTermPlural)} at risk`,
      value: formatNumber(report.pulse.sprints.atRiskSprints),
    },
    {
      label: `Overdue ${sprintTermPlural}`,
      value: formatNumber(report.pulse.sprints.overdueSprints),
    },
    {
      label: "Pending requests",
      value: formatNumber(report.requests.pendingRequests),
    },
    {
      label: "Stale requests",
      value: formatNumber(
        report.requests.providers.reduce(
          (sum, provider) => sum + provider.staleRequests,
          0,
        ),
      ),
    },
    {
      label: "Leading request source",
      value: topProvider ? titleCase(topProvider.provider) : "No activity",
    },
  ];
  return (
    <ReportCard>
      <SectionTitle description="Ownership, planning, and request health.">
        Workspace health
      </SectionTitle>
      <Box className="mt-4 grid gap-x-8 md:grid-cols-2">
        {rows.map((row) => (
          <Flex
            align="center"
            className="border-border gap-4 border-b-[0.5px] py-3"
            justify="between"
            key={row.label}
          >
            <Text color="muted">{row.label}</Text>
            <Text className="shrink-0 tabular-nums" fontWeight="medium">
              {row.value}
            </Text>
          </Flex>
        ))}
      </Box>
    </ReportCard>
  );
};

export const OverviewTab = ({
  objectiveTermPlural,
  report,
  sprintTermPlural,
  storyTermPlural,
}: {
  objectiveTermPlural: string;
  report: WorkspaceCommandCenterReport;
  sprintTermPlural: string;
  storyTermPlural: string;
}) => {
  const { priorityData } = useFlowBreakdownData(report);
  const memberChartData = buildMemberWorkloadChartData(report.workload.members);
  const deliveryChartData = buildCompletionTrendChartData(
    report.overview.completionTrend,
    report.overview.filters,
  );
  const providerChartData = buildProviderChartData(report.requests.providers);

  return (
    <Box className="space-y-6">
      <Box className="grid gap-5 @6xl:grid-cols-3">
        <ReportCard className="@6xl:col-span-2">
          <SectionTitle description="Work created and completed over time.">
            Delivery trend
          </SectionTitle>
          <Box className="mt-6">
            <DeliveryChart data={deliveryChartData} />
          </Box>
        </ReportCard>
        <RiskSummary
          objectiveTermPlural={objectiveTermPlural}
          report={report}
          sprintTermPlural={sprintTermPlural}
          storyTermPlural={storyTermPlural}
        />
      </Box>
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <PriorityDistributionCard priorityData={priorityData} />
        <ReportCard>
          <SectionTitle description="Open and overdue work by person.">
            Workload distribution
          </SectionTitle>
          <Box className="mt-6">
            <WorkloadChart data={memberChartData} />
          </Box>
        </ReportCard>
      </Box>
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <OverviewReadout
          objectiveTermPlural={objectiveTermPlural}
          report={report}
          sprintTermPlural={sprintTermPlural}
          storyTermPlural={storyTermPlural}
        />
        <ReportCard>
          <SectionTitle description="Intake and acceptance across connected sources.">
            Request sources
          </SectionTitle>
          <Box className="mt-6">
            <ProviderChart data={providerChartData} />
          </Box>
          {providerChartData.length ? <ProviderLegend /> : null}
        </ReportCard>
      </Box>
    </Box>
  );
};

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
import { EmptyState, MiniMetric, ReportCard, SectionTitle } from "./primitives";
import {
  PriorityDistributionCard,
  useFlowBreakdownData,
} from "./flow-and-planning-tabs";

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
          description="Work without a complexity value."
          label={`${storyTermPlural} without complexity`}
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
  );
  const providerChartData = buildProviderChartData(report.requests.providers);

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
            <WorkloadChart data={memberChartData} />
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
            <DeliveryChart data={deliveryChartData} />
          </Box>
        </ReportCard>
      </Box>
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <ReportCard>
          <SectionTitle description="Which connected sources create work and how much is still waiting.">
            Request source performance
          </SectionTitle>
          <Box className="mt-5">
            <ProviderChart data={providerChartData} />
          </Box>
          <ProviderLegend />
        </ReportCard>
      </Box>
    </Box>
  );
};

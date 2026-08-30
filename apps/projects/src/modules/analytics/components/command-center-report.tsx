"use client";

import { useEffect } from "react";
import { Badge, Box, Button, Flex, Tabs, Text } from "ui";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useCommandCenterReport } from "../hooks/command-center-report";
import { useAppliedFilters } from "../hooks/filters";
import { useWorkspaceAnalyticsEvent } from "../hooks/workspace-analytics-event";
import { EngagementTab } from "./command-center-report/engagement-tab";
import {
  FlowTab,
  PlanningTab,
} from "./command-center-report/flow-and-planning-tabs";
import {
  buildFilterSignature,
  completionRate,
  formatNumber,
  formatPercent,
  titleCase,
} from "./command-center-report/model";
import {
  CommandCenterSkeleton,
  MetricCard,
  ReportCard,
} from "./command-center-report/primitives";
import { OverviewTab } from "./command-center-report/overview-tab";
import { WorkloadTab } from "./command-center-report/workload-tab";
import { Filters } from "./filters";

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
  const filterSignature = buildFilterSignature(filters);

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
          description={`${formatNumber(report.workload.summary.totalEstimate)} total complexity`}
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

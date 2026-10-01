"use client";

import { useEffect } from "react";
import { Badge, Box, Button, Tabs, Text } from "ui";
import { useTerminology } from "@/hooks/use-terminology-display";
import { CustomFieldReportPanel } from "@/modules/custom-fields/public/report";
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

  return (
    <Box className="min-w-0 space-y-6 pb-6">
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

      {isFetching ? (
        <Badge
          className="bg-transparent"
          color="tertiary"
          role="status"
          variant="outline"
        >
          Refreshing reports…
        </Badge>
      ) : null}
      <Tabs.Panel className="space-y-6" value="overview">
        <Box className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <MetricCard
            description="Current unfinished work"
            label={`Open ${storyTermPlural}`}
            value={formatNumber(report.workload.summary.totalOpenStories)}
          />
          <MetricCard
            description="Finished in the selected period"
            label="Completed"
            value={formatNumber(
              report.overview.metrics.completedInPeriod ??
                report.overview.metrics.completedStories,
            )}
          />
          <MetricCard
            accent={`${formatNumber(report.overview.metrics.totalStories)} created`}
            description="Among work created in this period"
            label="Completion rate"
            value={formatPercent(completion)}
          />
          <MetricCard
            description={`${formatNumber(report.pulse.summary.blockedStories)} blocked · needs attention`}
            label="Overdue work"
            value={formatNumber(report.pulse.summary.overdueStories)}
          />
        </Box>
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
      <Tabs.Panel value="custom-fields">
        <CustomFieldReportPanel />
      </Tabs.Panel>
    </Box>
  );
};

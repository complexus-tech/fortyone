import { Box, Text } from "ui";
import type { AnalyticsReportOutput } from "./model";
import {
  asRecord,
  asRows,
  COLORS,
  completionRate,
  hasPositiveMetric,
  humanizeLabel,
  progressSummary,
  ratioPercent,
} from "./model";
import {
  ChartSection,
  CompactBarChart,
  HealthAndProgressSection,
  KeyValueList,
  MetricGrid,
  RiskList,
} from "./primitives";

export const WorkspaceCommandCenterReport = ({
  output,
  title,
}: {
  output: AnalyticsReportOutput;
  title: string;
}) => {
  const report = asRecord(output.report);
  const overview = asRecord(report.overview);
  const metrics = asRecord(overview.metrics);
  const pulse = asRecord(report.pulse);
  const pulseSummary = asRecord(pulse.summary);
  const storyHealth = asRecord(pulse.stories);
  const sprintHealth = asRecord(pulse.sprints);
  const objectiveHealth = asRecord(pulse.objectives);
  const workload = asRecord(report.workload);
  const workloadSummary = asRecord(workload.summary);
  const sprints = asRecord(report.sprints);
  const objectives = asRecord(report.objectives);
  const requests = asRecord(report.requests);
  const engagement = asRecord(report.engagement);
  const members = [...asRows(workload.members)]
    .sort(
      (firstMember, secondMember) =>
        Number(secondMember.openStories ?? 0) -
        Number(firstMember.openStories ?? 0),
    )
    .slice(0, 8)
    .map((member) => ({
      ...member,
      memberName: String(member.fullName ?? member.username ?? "Team member"),
    }));
  const providers = asRows(requests.providers);
  const risks = asRows(pulse.risks);
  const sprintProgress = asRows(sprints.sprintProgress).filter(
    (sprint) => typeof sprint.sprintName === "string" && sprint.sprintName,
  );
  const objectiveProgress = asRows(objectives.keyResultsProgress).filter(
    (objective) =>
      typeof objective.objectiveName === "string" && objective.objectiveName,
  );
  const engagementEvents = asRows(engagement.eventsByName);
  const unavailableSections = asRows(report.sectionErrors).flatMap(
    (sectionError) => {
      const section = humanizeLabel(sectionError.section);
      return section ? [section] : [];
    },
  );
  const hasWorkload =
    members.length > 0 ||
    hasPositiveMetric(workloadSummary, [
      "totalOpenStories",
      "totalEstimate",
      "unassignedStories",
      "unestimatedStories",
    ]);
  const hasSprintData =
    sprintProgress.length > 0 ||
    hasPositiveMetric(sprintHealth, [
      "activeSprints",
      "upcomingSprints",
      "completedSprints",
      "atRiskSprints",
      "overdueSprints",
    ]);
  const hasObjectiveData =
    objectiveProgress.length > 0 ||
    hasPositiveMetric(objectiveHealth, [
      "activeObjectives",
      "atRiskObjectives",
      "offTrackObjectives",
      "overdueObjectives",
      "objectivesDueSoon",
    ]);
  const hasRequestData =
    providers.length > 0 ||
    hasPositiveMetric(requests, [
      "totalRequests",
      "pendingRequests",
      "acceptedRequests",
      "declinedRequests",
    ]);
  const hasEngagementData =
    engagementEvents.length > 0 ||
    hasPositiveMetric(engagement, ["totalEvents", "uniqueUsers"]);

  return (
    <Box className="mt-3 space-y-3">
      <Text className="text-xl font-semibold">{title}</Text>
      {unavailableSections.length ? (
        <Text className="border-warning/25 bg-warning/5 text-foreground/70 dark:bg-warning/10 rounded-lg border px-3 py-2 text-base dark:text-white/65">
          Some analytics are temporarily unavailable:{" "}
          {unavailableSections.join(", ")}.
        </Text>
      ) : null}
      <MetricGrid
        metrics={[
          {
            label: "Open work",
            value: Number(workloadSummary.totalOpenStories ?? 0),
          },
          {
            label: "Completion",
            value: completionRate(
              metrics.completedStories,
              metrics.totalStories,
            ),
          },
          {
            label: "Overdue",
            value: Number(pulseSummary.overdueStories ?? 0),
          },
          {
            label: "Blocked",
            value: Number(pulseSummary.blockedStories ?? 0),
          },
          {
            label: "Active risks",
            value: risks.length,
          },
          {
            label: "Overloaded",
            value: Number(pulseSummary.overloadedMembers ?? 0),
          },
          {
            label: "At-risk sprints",
            value: Number(pulseSummary.atRiskSprints ?? 0),
          },
          {
            label: "At-risk objectives",
            value: Number(pulseSummary.atRiskObjectives ?? 0),
          },
        ]}
      />

      <ChartSection
        description="Completion and work that is currently moving or waiting."
        title="Delivery health"
      >
        <KeyValueList
          rows={[
            {
              label: "Delivered",
              value: `${Number(metrics.completedStories ?? 0)} of ${Number(
                metrics.totalStories ?? 0,
              )} (${completionRate(
                metrics.completedStories,
                metrics.totalStories,
              )})`,
            },
            {
              label: "In progress",
              value: Number(storyHealth.startedStories ?? 0),
            },
            {
              label: "Paused",
              value: Number(storyHealth.pausedStories ?? 0),
            },
            {
              label: "Unassigned",
              value: Number(storyHealth.unassignedStories ?? 0),
            },
          ]}
        />
      </ChartSection>

      {hasWorkload ? (
        <ChartSection
          description="Ownership, complexity, and the people carrying the most open work."
          title="Workload"
        >
          <Box className="space-y-3">
            <KeyValueList
              rows={[
                {
                  label: "Total complexity",
                  value: Number(workloadSummary.totalEstimate ?? 0),
                },
                {
                  label: "Unassigned work",
                  value: Number(workloadSummary.unassignedStories ?? 0),
                },
                {
                  label: "Without complexity",
                  value: Number(workloadSummary.unestimatedStories ?? 0),
                },
              ]}
            />
            {members.length ? (
              <CompactBarChart
                bars={[
                  {
                    key: "openStories",
                    color: COLORS.primary,
                    name: "Open",
                  },
                  {
                    key: "overdueStories",
                    color: "#EF4444",
                    name: "Overdue",
                  },
                ]}
                data={members}
                maxLabelLength={12}
                xKey="memberName"
              />
            ) : null}
          </Box>
        </ChartSection>
      ) : null}

      {hasSprintData || hasObjectiveData ? (
        <Box className="grid gap-4 md:grid-cols-2">
          {hasSprintData ? (
            <HealthAndProgressSection
              description="Current sprint exposure and delivery progress."
              healthRows={[
                {
                  label: "Active",
                  value: Number(sprintHealth.activeSprints ?? 0),
                },
                {
                  label: "At risk",
                  value: Number(sprintHealth.atRiskSprints ?? 0),
                },
                {
                  label: "Overdue",
                  value: Number(sprintHealth.overdueSprints ?? 0),
                },
              ]}
              progressRows={sprintProgress.slice(0, 5).map((sprint) => ({
                key: String(sprint.sprintId ?? sprint.sprintName),
                label: String(sprint.sprintName),
                value: `${progressSummary(
                  sprint.completed,
                  sprint.total,
                )} · ${humanizeLabel(sprint.status)}`,
              }))}
              title="Sprint health and progress"
            />
          ) : null}
          {hasObjectiveData ? (
            <HealthAndProgressSection
              description="Objective exposure and average key-result progress."
              healthRows={[
                {
                  label: "Active",
                  value: Number(objectiveHealth.activeObjectives ?? 0),
                },
                {
                  label: "At risk",
                  value: Number(objectiveHealth.atRiskObjectives ?? 0),
                },
                {
                  label: "Off track",
                  value: Number(objectiveHealth.offTrackObjectives ?? 0),
                },
                {
                  label: "Overdue",
                  value: Number(objectiveHealth.overdueObjectives ?? 0),
                },
              ]}
              progressRows={objectiveProgress.slice(0, 5).map((objective) => ({
                key: String(objective.objectiveId ?? objective.objectiveName),
                label: String(objective.objectiveName),
                value: `${Math.round(
                  Number(objective.avgProgress ?? 0),
                )}% average · ${Number(
                  objective.completed ?? 0,
                )} of ${Number(objective.total ?? 0)} key results`,
              }))}
              title="Objective health and progress"
            />
          ) : null}
        </Box>
      ) : null}

      {risks.length ? (
        <ChartSection
          description="The highest-signal delivery issues that need attention."
          title="Active risks"
        >
          <RiskList risks={risks} />
        </ChartSection>
      ) : null}

      {hasRequestData || hasEngagementData ? (
        <Box className="grid gap-4 md:grid-cols-2">
          {hasRequestData ? (
            <ChartSection title="Request sources">
              <KeyValueList
                rows={[
                  {
                    key: "all-requests",
                    label: "All requests",
                    value: Number(requests.totalRequests ?? 0),
                  },
                  {
                    key: "pending-requests",
                    label: "Pending",
                    value: Number(requests.pendingRequests ?? 0),
                  },
                  ...providers.slice(0, 4).map((provider) => ({
                    key: `provider-${String(provider.provider ?? "source")}`,
                    label: humanizeLabel(provider.provider || "Source"),
                    value: `${Number(
                      provider.totalRequests ?? 0,
                    )} requests · ${ratioPercent(
                      provider.acceptanceRate,
                    )} accepted`,
                  })),
                ]}
              />
            </ChartSection>
          ) : null}

          {hasEngagementData ? (
            <ChartSection title="Engagement">
              <KeyValueList
                rows={[
                  {
                    key: "tracked-events",
                    label: "Tracked events",
                    value: Number(engagement.totalEvents ?? 0),
                  },
                  {
                    key: "active-people",
                    label: "Active people",
                    value: Number(engagement.uniqueUsers ?? 0),
                  },
                  ...engagementEvents.slice(0, 4).map((event) => ({
                    key: `event-${String(event.name ?? "event")}`,
                    label: humanizeLabel(event.name || "Event"),
                    value: Number(event.count ?? 0),
                  })),
                ]}
              />
            </ChartSection>
          ) : null}
        </Box>
      ) : null}
    </Box>
  );
};

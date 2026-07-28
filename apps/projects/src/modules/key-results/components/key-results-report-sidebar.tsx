"use client";

import { differenceInCalendarDays } from "date-fns";
import { OKRIcon } from "icons";
import Link from "next/link";
import {
  Cell,
  Pie,
  PieChart,
  ResponsiveContainer,
  Tooltip as ChartTooltip,
} from "recharts";
import type { TooltipProps } from "recharts";
import { Avatar, Box, Divider, Flex, ProgressBar, Tabs, Text } from "ui";
import { RowWrapper, TeamColor } from "@/components/ui";
import { useTerminology, useWorkspacePath } from "@/hooks";
import type { Member } from "@/types";
import { hexToRgba } from "@/utils";
import type { KeyResultWithTeam } from "../types";
import { getKeyResultProgress } from "../utils";

const getDisplayName = (member?: Member) =>
  member?.fullName.trim() || member?.username || "Unassigned";

const healthChartColor = (status: string) => {
  switch (status.toLowerCase()) {
    case "on track":
    case "completed":
      return "#22c55e";
    case "at risk":
      return "#eab308";
    case "off track":
    case "behind":
      return "#ea6060";
    default:
      return "#6b7280";
  }
};

const getObjectiveHealth = (progress: number) => {
  if (progress >= 100) {
    return { color: "#22c55e", label: "Completed", textColor: "text-success" };
  }
  if (progress >= 70) {
    return { color: "#22c55e", label: "On track", textColor: "text-success" };
  }
  if (progress >= 40) {
    return { color: "#eab308", label: "At risk", textColor: "text-warning" };
  }
  return { color: "#ea6060", label: "Behind", textColor: "text-danger" };
};

const DonutTooltip = ({ active, payload }: TooltipProps<number, string>) => {
  const { getTermDisplay } = useTerminology();

  if (!active || !payload?.length) return null;

  return (
    <Box className="border-border/60 bg-surface-elevated/90 rounded-lg border-[0.5px] px-3 py-2 backdrop-blur">
      <Text fontWeight="medium">{payload[0].name}</Text>
      <Text color="muted">
        {payload[0].value}{" "}
        {getTermDisplay("keyResultTerm", {
          variant: payload[0].value === 1 ? "singular" : "plural",
        })}
      </Text>
    </Box>
  );
};

const DonutChart = ({
  data,
  total,
}: {
  data: { color: string; count: number; label: string }[];
  total: number;
}) => (
  <Box className="relative h-36 w-full">
    <Box className="pointer-events-none absolute inset-0 z-1 flex flex-col items-center justify-center">
      <Text fontSize="xl" fontWeight="medium">
        {total}
      </Text>
      <Text color="muted">Total</Text>
    </Box>
    <ResponsiveContainer height="100%" width="100%">
      <PieChart margin={{ bottom: 4, left: 4, right: 4, top: 4 }}>
        <Pie
          cornerRadius={3}
          data={data.filter(({ count }) => count > 0)}
          dataKey="count"
          innerRadius={46}
          nameKey="label"
          outerRadius={62}
          paddingAngle={2}
          stroke="none"
        >
          {data
            .filter(({ count }) => count > 0)
            .map((item) => (
              <Cell fill={item.color} key={item.label} stroke="none" />
            ))}
        </Pie>
        <ChartTooltip content={<DonutTooltip />} />
      </PieChart>
    </ResponsiveContainer>
  </Box>
);

const ChartLegend = ({
  data,
}: {
  data: { color: string; count: number; label: string }[];
}) => (
  <Box className="mt-2 grid grid-cols-2 gap-x-5 gap-y-2">
    {data.map((item) => (
      <Flex align="center" className="min-w-0 gap-2" key={item.label}>
        <Box
          className="size-2.5 shrink-0 rounded-full"
          style={{ backgroundColor: item.color }}
        />
        <Text className="min-w-0 flex-1 truncate" color="muted">
          {item.label}
        </Text>
        <Text className="shrink-0" fontWeight="medium">
          {item.count}
        </Text>
      </Flex>
    ))}
  </Box>
);

const HealthDistributionChart = ({
  data,
  total,
}: {
  data: { color: string; count: number; label: string }[];
  total: number;
}) => (
  <Box>
    <Flex className="bg-surface-muted h-3 w-full overflow-hidden rounded-full">
      {data.map((item) =>
        item.count > 0 ? (
          <Box
            key={item.label}
            style={{
              backgroundColor: item.color,
              width: `${(item.count / Math.max(total, 1)) * 100}%`,
            }}
          />
        ) : null,
      )}
    </Flex>
    <ChartLegend data={data} />
  </Box>
);

export const KeyResultsReportSidebar = ({
  keyResults,
  memberById,
  teamColorById,
  totalCount,
}: {
  keyResults: KeyResultWithTeam[];
  memberById: ReadonlyMap<string, Member>;
  teamColorById: ReadonlyMap<string, string>;
  totalCount: number;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const keyResultSingular = getTermDisplay("keyResultTerm");
  const keyResultPlural = getTermDisplay("keyResultTerm", {
    variant: "plural",
  });
  const objectiveSingular = getTermDisplay("objectiveTerm");
  const objectivePlural = getTermDisplay("objectiveTerm", {
    variant: "plural",
  });
  const objectiveStats = new Map<
    string,
    {
      completed: number;
      objectiveId: string;
      objectiveName: string;
      progressTotal: number;
      teamCode: string;
      teamId: string;
      total: number;
    }
  >();
  let completed = 0;
  let progressTotal = 0;

  for (const keyResult of keyResults) {
    const progress = getKeyResultProgress(keyResult);
    const objective = objectiveStats.get(keyResult.objectiveId) ?? {
      completed: 0,
      objectiveId: keyResult.objectiveId,
      objectiveName: keyResult.objectiveName,
      progressTotal: 0,
      teamCode: keyResult.teamCode,
      teamId: keyResult.teamId,
      total: 0,
    };
    objective.total += 1;
    objective.progressTotal += progress;
    progressTotal += progress;
    if (progress >= 100) {
      objective.completed += 1;
      completed += 1;
    }
    objectiveStats.set(keyResult.objectiveId, objective);
  }

  const objectiveProgress = Array.from(
    objectiveStats.values(),
    (objective) => ({
      ...objective,
      avgProgress: Math.round(
        objective.progressTotal / Math.max(objective.total, 1),
      ),
    }),
  );
  const reportTotal = keyResults.length;
  const overallProgress =
    reportTotal === 0 ? 0 : Math.round(progressTotal / reportTotal);
  const overallHealth = getObjectiveHealth(overallProgress);
  const objectiveTeamIds = new Map(
    keyResults.map((keyResult) => [keyResult.objectiveId, keyResult.teamId]),
  );
  const leadStats = new Map<string, { count: number; progressTotal: number }>();

  for (const keyResult of keyResults) {
    const leadId = keyResult.lead ?? "unassigned";
    const current = leadStats.get(leadId) ?? {
      count: 0,
      progressTotal: 0,
    };
    current.count += 1;
    current.progressTotal += getKeyResultProgress(keyResult);
    leadStats.set(leadId, current);
  }

  const leads = Array.from(leadStats, ([leadId, stats]) => ({
    averageProgress: Math.round(stats.progressTotal / stats.count),
    count: stats.count,
    lead: leadId === "unassigned" ? undefined : memberById.get(leadId),
    leadId,
  })).sort((a, b) => b.count - a.count);
  const objectiveHealth = objectiveProgress.reduce(
    (summary, objective) => {
      const health = getObjectiveHealth(objective.avgProgress).label;
      if (health === "Completed") summary.completed += 1;
      else if (health === "On track") summary.onTrack += 1;
      else if (health === "At risk") summary.atRisk += 1;
      else summary.behind += 1;
      return summary;
    },
    { atRisk: 0, behind: 0, completed: 0, onTrack: 0 },
  );
  const healthTotal = objectiveProgress.length;
  const today = new Date();
  const deliveryHealth = keyResults.reduce(
    (summary, keyResult) => {
      const progress = getKeyResultProgress(keyResult);
      if (progress >= 100) {
        summary.completed += 1;
        return summary;
      }

      const daysRemaining = differenceInCalendarDays(
        new Date(keyResult.endDate),
        today,
      );
      if (daysRemaining < 0) summary.overdue += 1;
      else if (daysRemaining <= 14) summary.dueSoon += 1;
      else summary.later += 1;
      return summary;
    },
    { completed: 0, dueSoon: 0, later: 0, overdue: 0 },
  );
  const deliveryTotal = Object.values(deliveryHealth).reduce(
    (total, count) => total + count,
    0,
  );
  const deliverySegments = [
    { color: "#ea6060", count: deliveryHealth.overdue, label: "Overdue" },
    { color: "#eab308", count: deliveryHealth.dueSoon, label: "Due soon" },
    { color: "#6366f1", count: deliveryHealth.later, label: "Later" },
    {
      color: "#22c55e",
      count: deliveryHealth.completed,
      label: "Completed",
    },
  ];
  const objectiveHealthSegments = [
    {
      color: healthChartColor("Completed"),
      count: objectiveHealth.completed,
      label: "Completed",
    },
    {
      color: healthChartColor("On track"),
      count: objectiveHealth.onTrack,
      label: "On track",
    },
    {
      color: healthChartColor("At risk"),
      count: objectiveHealth.atRisk,
      label: "At risk",
    },
    {
      color: healthChartColor("Behind"),
      count: objectiveHealth.behind,
      label: "Behind",
    },
  ];

  return (
    <Box className="bg-surface-muted/30 bg-surface/60 min-h-full overflow-x-hidden pt-6 pb-48">
      <Box className="px-6">
        <Text className="flex items-center gap-1.5" fontSize="lg">
          <OKRIcon className="h-[1.25rem]" />
          {getTermDisplay("keyResultTerm", { capitalize: true })} insights
        </Text>
        <Box className="mt-5">
          <Text fontSize="4xl" fontWeight="medium">
            {overallProgress}%
          </Text>
          <Text color="muted" fontSize="lg">
            Overall progress
          </Text>
        </Box>
        <Flex align="center" className="mt-4 w-full gap-3">
          <ProgressBar
            className="h-2 min-w-0 flex-1"
            progress={overallProgress}
          />
          <Text
            className="shrink-0 rounded-md border px-2.5 py-1"
            fontWeight="medium"
            style={{
              backgroundColor: hexToRgba(overallHealth.color, 0.1),
              borderColor: hexToRgba(overallHealth.color, 0.2),
              color: overallHealth.color,
            }}
          >
            {completed} of {reportTotal} complete
          </Text>
        </Flex>
      </Box>

      <Divider className="my-6" />

      <Box className="px-6">
        <Flex align="center" className="mb-3" justify="between">
          <Text>Linked {objectiveSingular} health</Text>
          <Text color="muted">
            {healthTotal}{" "}
            {healthTotal === 1 ? objectiveSingular : objectivePlural}
          </Text>
        </Flex>
        <HealthDistributionChart
          data={objectiveHealthSegments}
          total={healthTotal}
        />
      </Box>

      <Divider className="my-6" />

      <Box className="px-6">
        <Flex align="center" className="mb-3" justify="between">
          <Text>Delivery health</Text>
          <Text color="muted">
            {deliveryTotal}{" "}
            {deliveryTotal === 1 ? keyResultSingular : keyResultPlural}
          </Text>
        </Flex>
        <DonutChart data={deliverySegments} total={deliveryTotal} />
        <ChartLegend data={deliverySegments} />
      </Box>

      <Divider className="my-6" />

      <Box className="px-6">
        <Tabs defaultValue="objectives">
          <Tabs.List className="mx-0 mb-3 md:mx-0">
            <Tabs.Tab value="objectives">
              {getTermDisplay("objectiveTerm", {
                variant: "plural",
                capitalize: true,
              })}
            </Tabs.Tab>
            <Tabs.Tab value="leads">Leads</Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="objectives">
            {objectiveProgress
              .toSorted((a, b) => a.avgProgress - b.avgProgress)
              .slice(0, 8)
              .map((objective) => {
                const teamId = objectiveTeamIds.get(objective.objectiveId);
                const teamColor = teamColorById.get(objective.teamId);
                const health = getObjectiveHealth(objective.avgProgress);
                const content = (
                  <Flex align="center" className="w-full gap-3">
                    <Flex align="center" className="min-w-0 flex-1 gap-2">
                      <Flex
                        align="center"
                        className="shrink-0 gap-1.5 rounded-md border px-2 py-1"
                        style={{
                          backgroundColor: hexToRgba(teamColor, 0.1),
                          borderColor: hexToRgba(teamColor, 0.2),
                        }}
                      >
                        <TeamColor color={teamColor} />
                        <Text>{objective.teamCode}</Text>
                      </Flex>
                      <Box className="min-w-0">
                        <Text className="truncate">
                          {objective.objectiveName}
                        </Text>
                        <Text className={health.textColor}>{health.label}</Text>
                      </Box>
                    </Flex>
                    <Flex align="center" className="w-24 shrink-0 gap-2">
                      <ProgressBar
                        className="min-w-0 flex-1"
                        progress={objective.avgProgress}
                      />
                      <Text color="muted">
                        {Math.round(objective.avgProgress)}%
                      </Text>
                    </Flex>
                  </Flex>
                );

                return (
                  <RowWrapper
                    className="px-1 py-2.5 md:px-0"
                    key={objective.objectiveId}
                  >
                    {teamId ? (
                      <Link
                        className="w-full hover:opacity-90"
                        href={withWorkspace(
                          `/teams/${teamId}/objectives/${objective.objectiveId}`,
                        )}
                      >
                        {content}
                      </Link>
                    ) : (
                      content
                    )}
                  </RowWrapper>
                );
              })}
          </Tabs.Panel>
          <Tabs.Panel value="leads">
            {leads
              .slice(0, 8)
              .map(({ averageProgress, count, lead, leadId }) => (
                <RowWrapper className="gap-3 px-1 py-2.5 md:px-0" key={leadId}>
                  <Flex align="center" className="min-w-0" gap={2}>
                    <Avatar
                      name={lead ? getDisplayName(lead) : undefined}
                      size="xs"
                      src={lead?.avatarUrl}
                    />
                    <Box className="min-w-0">
                      <Text className="truncate">{getDisplayName(lead)}</Text>
                      <Text color="muted">
                        {count}{" "}
                        {count === 1 ? keyResultSingular : keyResultPlural}
                      </Text>
                    </Box>
                  </Flex>
                  <Flex align="center" className="w-24 shrink-0" gap={2}>
                    <ProgressBar progress={averageProgress} />
                    <Text color="muted">{averageProgress}%</Text>
                  </Flex>
                </RowWrapper>
              ))}
            {keyResults.length < totalCount ? (
              <Text className="mt-3" color="muted">
                Lead coverage reflects the {keyResults.length} loaded{" "}
                {keyResults.length === 1 ? keyResultSingular : keyResultPlural}.
              </Text>
            ) : null}
          </Tabs.Panel>
        </Tabs>
      </Box>
    </Box>
  );
};

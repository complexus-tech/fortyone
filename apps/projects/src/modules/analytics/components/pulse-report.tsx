"use client";

import type { ReactNode } from "react";
import Link from "next/link";
import { useTheme } from "next-themes";
import {
  Bar,
  BarChart,
  Cell,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { Avatar, Badge, Box, Button, Flex, Skeleton, Text } from "ui";
import { useTerminology, useWorkspacePath } from "@/hooks";
import { usePulseReport } from "@/modules/analytics/hooks/pulse-report";
import type {
  MemberWorkload,
  PulseReport,
  PulseRisk,
  PulseRiskSeverity,
} from "@/modules/analytics/types";

const severityClasses: Record<PulseRiskSeverity, string> = {
  high: "text-danger dark:text-danger",
  medium: "text-warning dark:text-warning",
  low: "text-info dark:text-info",
};

const formatNumber = (value?: number) =>
  new Intl.NumberFormat().format(value ?? 0);

const applyTerminology = (
  value: string,
  storyTermPlural: string,
  sprintTermPlural: string,
  objectiveTermPlural: string,
) => {
  return value
    .replaceAll("Stories", titleCase(storyTermPlural))
    .replaceAll("stories", storyTermPlural)
    .replaceAll("Sprints", titleCase(sprintTermPlural))
    .replaceAll("sprints", sprintTermPlural)
    .replaceAll("Objectives", titleCase(objectiveTermPlural))
    .replaceAll("objectives", objectiveTermPlural);
};

const titleCase = (value: string) =>
  value.charAt(0).toUpperCase() + value.slice(1);

const ReportCard = ({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) => {
  return (
    <Box
      className={[
        "border-border/70 dark:bg-surface h-full rounded-lg border bg-white px-5 py-5",
        className,
      ]
        .filter(Boolean)
        .join(" ")}
    >
      {children}
    </Box>
  );
};

const getRiskHref = ({
  report,
  risk,
  withWorkspace,
}: {
  report: PulseReport;
  risk: PulseRisk;
  withWorkspace: (path: string) => string;
}) => {
  const firstOverloadedMember = report.workload.risks.overloadedMembers.at(0);
  const firstTeam = report.workload.teams.at(0);

  switch (risk.kind) {
    case "overdue_stories":
      return withWorkspace("/my-work?tab=all&overdue=true");
    case "blocked_stories":
      return withWorkspace("/my-work?tab=all&category=paused");
    case "overloaded_members":
      return firstOverloadedMember
        ? withWorkspace(`/profile/${firstOverloadedMember.userId}`)
        : withWorkspace("/my-work?tab=all");
    case "at_risk_sprints":
      return withWorkspace("/sprints");
    case "at_risk_objectives":
      return withWorkspace("/roadmaps");
    case "pending_requests":
      return firstTeam
        ? withWorkspace(`/teams/${firstTeam.teamId}/requests`)
        : withWorkspace("/analytics?tab=pulse");
    case "unassigned_stories":
      return withWorkspace("/my-work?tab=all");
  }
};

const getRiskDescription = (risk: PulseRisk) => {
  if (risk.kind === "unassigned_stories") {
    return "Open stories that do not have an owner yet.";
  }

  return risk.description;
};

const MetricCard = ({
  title,
  count,
  href,
}: {
  title: string;
  count: number;
  href: string;
}) => {
  return (
    <Link href={href}>
      <ReportCard className="hover:bg-surface-muted/60 px-5 py-4 transition">
        <Text className="text-foreground/80 text-[0.95rem]">{title}</Text>
        <Text className="mt-2 text-[2rem] leading-none font-semibold">
          {formatNumber(count)}
        </Text>
      </ReportCard>
    </Link>
  );
};

const RiskCard = ({
  href,
  risk,
  storyTermPlural,
  sprintTermPlural,
  objectiveTermPlural,
}: {
  href: string;
  risk: PulseRisk;
  storyTermPlural: string;
  sprintTermPlural: string;
  objectiveTermPlural: string;
}) => {
  return (
    <Link href={href}>
      <Flex
        align="center"
        className="border-border/70 hover:bg-surface-muted/60 dark:bg-surface h-full gap-3 rounded-lg border bg-white px-4 py-4 transition"
      >
        <Box className="min-w-0 flex-1">
          <Flex align="center" className="gap-2" justify="between">
            <Flex align="center" className="min-w-0 gap-2">
              <Text
                className={["line-clamp-1", severityClasses[risk.severity]]
                  .filter(Boolean)
                  .join(" ")}
                fontWeight="medium"
              >
                {applyTerminology(
                  risk.title,
                  storyTermPlural,
                  sprintTermPlural,
                  objectiveTermPlural,
                )}
              </Text>
              <Badge
                className="shrink-0 bg-transparent"
                color="tertiary"
                rounded="full"
                size="sm"
              >
                {formatNumber(risk.count)}
              </Badge>
            </Flex>
          </Flex>
          <Text className="mt-1 leading-5" color="muted">
            {applyTerminology(
              getRiskDescription(risk),
              storyTermPlural,
              sprintTermPlural,
              objectiveTermPlural,
            )}
          </Text>
        </Box>
      </Flex>
    </Link>
  );
};

const WorkloadRow = ({
  href,
  member,
}: {
  href: string;
  member: MemberWorkload;
}) => {
  const displayName = member.fullName.trim() || member.username;

  return (
    <Link href={href}>
      <Flex
        align="center"
        className="border-border hover:bg-surface-muted/60 gap-3 border-b-[0.5px] px-2 py-3.5 transition last:border-b-0"
      >
        <Avatar name={displayName} size="md" src={member.avatarUrl} />
        <Box className="min-w-0 flex-1">
          <Text className="line-clamp-1" fontWeight="medium">
            {displayName}
          </Text>
          <Text color="muted">
            {formatNumber(member.openStories)} open /{" "}
            {formatNumber(member.estimateTotal)} estimate
          </Text>
        </Box>
        <Flex align="center" className="gap-2">
          <Badge
            className="bg-transparent text-[0.95rem]"
            color={member.overdueStories > 0 ? "danger" : "tertiary"}
            rounded="md"
            size="md"
            variant={member.overdueStories > 0 ? "solid" : "outline"}
          >
            {formatNumber(member.overdueStories)} overdue
          </Badge>
        </Flex>
      </Flex>
    </Link>
  );
};

const PulseSkeleton = () => {
  return (
    <Box className="py-3">
      <Skeleton className="mb-3 h-8 w-72" />
      <Skeleton className="mb-5 h-5 w-full max-w-xl" />
      <Box className="mb-4 grid grid-cols-2 gap-3 @3xl:grid-cols-3 @4xl:grid-cols-4 @7xl:grid-cols-5">
        {Array.from({ length: 5 }).map((_, index) => (
          <Skeleton className="h-28" key={index} />
        ))}
      </Box>
      <Box className="grid gap-4 @5xl:grid-cols-2">
        <Skeleton className="h-80" />
        <Skeleton className="h-80" />
      </Box>
    </Box>
  );
};

export const PulseReportPanel = () => {
  const { resolvedTheme } = useTheme();
  const { withWorkspace } = useWorkspacePath();
  const { getTermDisplay } = useTerminology();
  const {
    data: report,
    isPending,
    isError,
    refetch,
    isFetching,
  } = usePulseReport();

  if (isPending) {
    return <PulseSkeleton />;
  }

  if (isError) {
    return (
      <ReportCard className="mt-3">
        <Text className="mb-1" fontSize="lg" fontWeight="medium">
          Workspace pulse is unavailable
        </Text>
        <Text color="muted">
          The workspace health report could not be loaded right now.
        </Text>
        <Button className="mt-4" onClick={() => void refetch()}>
          Try again
        </Button>
      </ReportCard>
    );
  }

  const storyTerm = getTermDisplay("storyTerm");
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const sprintTermPlural = getTermDisplay("sprintTerm", { variant: "plural" });
  const objectiveTermPlural = getTermDisplay("objectiveTerm", {
    variant: "plural",
  });
  const storyHealthData = [
    { label: "Started", value: report.stories.startedStories, fill: "#6366F1" },
    { label: "Paused", value: report.stories.pausedStories, fill: "#eab308" },
    { label: "Overdue", value: report.stories.overdueStories, fill: "#EA6060" },
    { label: "Blocked", value: report.stories.blockedStories, fill: "#f43f5e" },
    {
      label: "Unassigned",
      value: report.stories.unassignedStories,
      fill: "#06b6d4",
    },
  ];
  const workloadData = report.workload.members.slice(0, 6).map((member) => ({
    name: member.fullName || member.username,
    open: member.openStories,
    overdue: member.overdueStories,
  }));
  const overloadedMembers = report.workload.risks.overloadedMembers.slice(0, 5);
  const gridStroke = resolvedTheme === "dark" ? "#222" : "#E0E0E0";
  const tooltipStyle = {
    backgroundColor:
      resolvedTheme === "dark"
        ? "rgba(24, 24, 27, 0.96)"
        : "rgba(255, 255, 255, 0.96)",
    border: `1px solid ${gridStroke}`,
    borderRadius: 8,
  };

  return (
    <Box className="pt-3 pb-4">
      <Flex className="flex flex-col items-start justify-between gap-3 @3xl:flex-row @3xl:items-center">
        <Box>
          <Text
            as="h2"
            className="mb-1 text-2xl @3xl:text-3xl"
            fontWeight="medium"
          >
            Workspace health
          </Text>
          <Text color="muted" fontSize="lg">
            Workspace health across {storyTermPlural}, workload,{" "}
            {sprintTermPlural}, {objectiveTermPlural}, and requests.
          </Text>
        </Box>
        <Button
          color="tertiary"
          disabled={isFetching}
          onClick={() => void refetch()}
          variant="outline"
        >
          Refresh
        </Button>
      </Flex>

      <Box className="mt-3 mb-4 grid grid-cols-2 gap-3 @3xl:grid-cols-3 @4xl:grid-cols-4 @7xl:grid-cols-5 @7xl:gap-4">
        <MetricCard
          count={report.summary.openStories}
          href={withWorkspace("/my-work?tab=all")}
          title={`Open ${storyTermPlural}`}
        />
        <MetricCard
          count={report.summary.overdueStories}
          href={withWorkspace("/my-work?tab=all&overdue=true")}
          title={`Overdue ${storyTermPlural}`}
        />
        <MetricCard
          count={report.summary.blockedStories}
          href={withWorkspace("/my-work?tab=all&category=paused")}
          title={`Blocked ${storyTermPlural}`}
        />
        <MetricCard
          count={report.summary.atRiskSprints}
          href={withWorkspace("/sprints")}
          title={`At-risk ${sprintTermPlural}`}
        />
        <MetricCard
          count={report.summary.atRiskObjectives}
          href={withWorkspace("/roadmaps")}
          title={`At-risk ${objectiveTermPlural}`}
        />
      </Box>

      <Box className="my-4 grid grid-cols-1 gap-4 @5xl:grid-cols-5">
        <ReportCard className="@5xl:col-span-3">
          <Box className="mb-6">
            <Flex align="center" className="mb-1 gap-2">
              <Text fontSize="lg">Needs attention</Text>
              <Badge
                className="bg-transparent"
                color="tertiary"
                rounded="full"
                size="sm"
              >
                {formatNumber(report.risks.length)}
              </Badge>
            </Flex>
            <Text color="muted">
              Click any item to open the screen where that risk can be handled.
            </Text>
          </Box>
          <Box className="grid grid-cols-1 gap-3 @3xl:grid-cols-2">
            {report.risks.length > 0 ? (
              report.risks.map((risk) => (
                <RiskCard
                  href={getRiskHref({ report, risk, withWorkspace })}
                  key={risk.kind}
                  objectiveTermPlural={objectiveTermPlural}
                  risk={risk}
                  sprintTermPlural={sprintTermPlural}
                  storyTermPlural={storyTermPlural}
                />
              ))
            ) : (
              <Box className="border-border bg-surface-muted rounded-lg border-[0.5px] px-4 py-6 text-center @3xl:col-span-2">
                <Text fontWeight="medium">No active risks</Text>
                <Text className="mt-1" color="muted">
                  Nothing needs immediate attention right now.
                </Text>
              </Box>
            )}
          </Box>
        </ReportCard>

        <ReportCard className="@5xl:col-span-2">
          <Box className="mb-6">
            <Text className="mb-1" fontSize="lg">
              Workload pressure
            </Text>
            <Text color="muted">Members above workload thresholds.</Text>
          </Box>
          {overloadedMembers.length > 0 ? (
            overloadedMembers.map((member) => (
              <WorkloadRow
                href={withWorkspace(`/profile/${member.userId}`)}
                key={member.userId}
                member={member}
              />
            ))
          ) : (
            <Text color="muted">No overloaded members right now.</Text>
          )}
        </ReportCard>
      </Box>

      <Box className="my-4 grid grid-cols-1 gap-4 @5xl:grid-cols-2">
        <ReportCard>
          <Box className="mb-6">
            <Text className="mb-1" fontSize="lg">
              {titleCase(storyTerm)} health
            </Text>
            <Text color="muted">Current work that may need intervention.</Text>
          </Box>
          <ResponsiveContainer height={260} width="100%">
            <BarChart
              data={storyHealthData}
              margin={{ top: 20, right: 10, left: -35, bottom: 0 }}
            >
              <CartesianGrid
                stroke={gridStroke}
                strokeDasharray="3 3"
                vertical={false}
              />
              <XAxis
                axisLine={{ stroke: gridStroke }}
                dataKey="label"
                tick={{ fontSize: 12 }}
              />
              <YAxis
                axisLine={false}
                tick={{ fontSize: 12 }}
                tickLine={false}
              />
              <Tooltip
                contentStyle={tooltipStyle}
                cursor={{ fill: "transparent" }}
              />
              <Bar barSize={35} dataKey="value" radius={[6, 6, 0, 0]}>
                {storyHealthData.map((entry) => (
                  <Cell fill={entry.fill} key={entry.label} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </ReportCard>

        <ReportCard>
          <Box className="mb-6">
            <Text className="mb-1" fontSize="lg">
              Team load
            </Text>
            <Text color="muted">Open and overdue work by assignee.</Text>
          </Box>
          <ResponsiveContainer height={260} width="100%">
            <BarChart
              data={workloadData}
              margin={{ top: 20, right: 10, left: -35, bottom: 0 }}
            >
              <CartesianGrid
                stroke={gridStroke}
                strokeDasharray="3 3"
                vertical={false}
              />
              <XAxis
                axisLine={{ stroke: gridStroke }}
                dataKey="name"
                tick={{ fontSize: 12 }}
              />
              <YAxis
                axisLine={false}
                tick={{ fontSize: 12 }}
                tickLine={false}
              />
              <Tooltip
                contentStyle={tooltipStyle}
                cursor={{ fill: "transparent" }}
              />
              <Bar
                barSize={28}
                dataKey="open"
                fill="#6366F1"
                radius={[6, 6, 0, 0]}
              />
              <Bar
                barSize={28}
                dataKey="overdue"
                fill="#EA6060"
                radius={[6, 6, 0, 0]}
              />
            </BarChart>
          </ResponsiveContainer>
        </ReportCard>
      </Box>
    </Box>
  );
};

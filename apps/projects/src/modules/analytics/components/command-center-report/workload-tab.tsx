import Link from "next/link";
import { Avatar, Badge, Box, Flex, Text } from "ui";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type {
  MemberWorkload,
  TeamWorkloadSummary,
  WorkspaceCommandCenterReport,
} from "../../types";
import { HorizontalBreakdownChart, WorkloadChart } from "./charts";
import {
  buildMemberWorkloadChartData,
  buildTeamWorkloadChartData,
  chartPalette,
  formatNumber,
  getDisplayName,
} from "./model";
import { EmptyState, ReportCard, SectionTitle } from "./primitives";

const WorkloadMemberRow = ({ member }: { member: MemberWorkload }) => {
  const { withWorkspace } = useWorkspacePath();
  const displayName = getDisplayName(member);

  return (
    <Link href={withWorkspace(`/profile/${member.userId}`)}>
      <Flex
        align="center"
        className="border-border hover:bg-surface-muted/60 gap-3 border-b-[0.5px] px-1 py-2.5 transition last:border-b-0"
      >
        <Avatar name={displayName} size="md" src={member.avatarUrl} />
        <Box className="min-w-0 flex-1">
          <Flex align="center" className="gap-2">
            <Text className="truncate" fontWeight="medium">
              {displayName}
            </Text>
            {member.teamAiRoleTitle ? (
              <Badge
                className="shrink-0 bg-transparent"
                color="tertiary"
                rounded="full"
                size="sm"
                variant="outline"
              >
                {member.teamAiRoleTitle}
              </Badge>
            ) : null}
          </Flex>
          <Flex align="center" className="mt-1 flex-wrap gap-x-3 gap-y-1">
            <Text color="muted">{formatNumber(member.openStories)} open</Text>
            <Text color="muted">
              {formatNumber(member.estimateTotal)} complexity
            </Text>
            <Text color="muted">
              {formatNumber(member.completedStories)} completed
            </Text>
          </Flex>
        </Box>
        <Box className="hidden min-w-28 text-right md:block">
          <Text className="font-semibold">
            {formatNumber(member.overdueStories)}
          </Text>
          <Text color="muted">overdue</Text>
        </Box>
        <Box className="hidden min-w-28 text-right @4xl:block">
          <Text className="font-semibold">
            {formatNumber(member.unestimatedStories)}
          </Text>
          <Text color="muted">no complexity</Text>
        </Box>
      </Flex>
    </Link>
  );
};

const TeamWorkloadRow = ({ team }: { team: TeamWorkloadSummary }) => {
  const { withWorkspace } = useWorkspacePath();

  return (
    <Link href={withWorkspace(`/teams/${team.teamId}`)}>
      <Flex
        align="center"
        className="border-border hover:bg-surface-muted/60 gap-3 border-b-[0.5px] px-1 py-2.5 transition last:border-b-0"
        justify="between"
      >
        <Box className="min-w-0">
          <Flex align="center" className="gap-2">
            <Text className="truncate" fontWeight="medium">
              {team.teamName}
            </Text>
            <Badge
              className="shrink-0 bg-transparent"
              color="tertiary"
              rounded="full"
              size="sm"
              variant="outline"
            >
              {team.teamCode}
            </Badge>
          </Flex>
          <Text className="mt-1" color="muted">
            {formatNumber(team.estimateTotal)} complexity /{" "}
            {formatNumber(team.unassignedStories)} unassigned
          </Text>
        </Box>
        <Box className="text-right">
          <Text className="text-lg" fontWeight="semibold">
            {formatNumber(team.openStories)}
          </Text>
          <Text color="muted">open</Text>
        </Box>
      </Flex>
    </Link>
  );
};

export const WorkloadTab = ({
  report,
}: {
  report: WorkspaceCommandCenterReport;
}) => {
  const topMembers = report.workload.members.slice(0, 10);
  const topTeams = report.workload.teams.slice(0, 10);
  const memberChartData = buildMemberWorkloadChartData(report.workload.members);
  const teamChartData = buildTeamWorkloadChartData(topTeams);

  return (
    <Box className="space-y-5">
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <ReportCard>
          <SectionTitle description="Open and overdue work by assignee.">
            Workload by person
          </SectionTitle>
          <Box className="mt-5">
            <WorkloadChart data={memberChartData} />
          </Box>
        </ReportCard>
        <ReportCard>
          <SectionTitle description="Open work concentration across teams.">
            Workload by team
          </SectionTitle>
          <Box className="mt-5">
            <HorizontalBreakdownChart
              color={chartPalette.primary}
              data={teamChartData}
              emptyText="No team workload is visible for these filters."
              height={260}
              labelWidth={106}
            />
          </Box>
        </ReportCard>
      </Box>
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <ReportCard>
          <SectionTitle description="People with the most open work, relative complexity, overdue work, and work without complexity.">
            Workload by person
          </SectionTitle>
          <Box className="mt-3">
            {topMembers.length ? (
              topMembers.map((member) => (
                <WorkloadMemberRow key={member.userId} member={member} />
              ))
            ) : (
              <EmptyState>
                No assigned work is visible for these filters.
              </EmptyState>
            )}
          </Box>
        </ReportCard>

        <ReportCard>
          <SectionTitle description="Team load, ownership gaps, and complexity concentration.">
            Workload by team
          </SectionTitle>
          <Box className="mt-3">
            {topTeams.length ? (
              topTeams.map((team) => (
                <TeamWorkloadRow key={team.teamId} team={team} />
              ))
            ) : (
              <EmptyState>
                No team workload is visible for these filters.
              </EmptyState>
            )}
          </Box>
        </ReportCard>
      </Box>
    </Box>
  );
};

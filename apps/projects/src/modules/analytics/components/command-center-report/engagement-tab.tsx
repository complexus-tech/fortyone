import Link from "next/link";
import { Avatar, Badge, Box, Flex, Text } from "ui";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type {
  WorkspaceCommandCenterReport,
  WorkspaceEngagementUser,
} from "../../types";
import { HorizontalBreakdownChart } from "./charts";
import {
  buildEngagementCountChartData,
  buildProviderChartData,
  chartPalette,
  formatNumber,
  getDisplayName,
} from "./model";
import { ProviderChart, ProviderLegend } from "./provider-chart";
import { MiniMetric, ReportCard, SectionTitle } from "./primitives";

const EngagementUserRow = ({ user }: { user: WorkspaceEngagementUser }) => {
  const { withWorkspace } = useWorkspacePath();
  const displayName = getDisplayName(user);

  return (
    <Link href={withWorkspace(`/profile/${user.userId}`)}>
      <Flex
        align="center"
        className="border-border hover:bg-surface-muted/60 gap-3 border-b-[0.5px] px-1 py-3 transition last:border-b-0"
      >
        <Avatar name={displayName} size="sm" src={user.avatarUrl} />
        <Box className="min-w-0 flex-1">
          <Text className="truncate" fontWeight="medium">
            {displayName}
          </Text>
          <Text color="muted">{user.username}</Text>
        </Box>
        <Badge
          className="bg-transparent"
          color="info"
          rounded="full"
          size="sm"
          variant="outline"
        >
          {formatNumber(user.events)}
        </Badge>
      </Flex>
    </Link>
  );
};

export const EngagementTab = ({
  report,
}: {
  report: WorkspaceCommandCenterReport;
}) => {
  const providerChartData = buildProviderChartData(report.requests.providers);
  const eventChartData = buildEngagementCountChartData(
    report.engagement.eventsByName,
  );
  const surfaceChartData = buildEngagementCountChartData(
    report.engagement.eventsBySurface,
  );

  return (
    <Box className="space-y-5">
      <Box className="grid gap-5 @6xl:grid-cols-2">
        <ReportCard>
          <SectionTitle description="Provider mix, acceptance, and stale requests across connected sources.">
            Request sources
          </SectionTitle>
          <Box className="mt-5">
            <ProviderChart data={providerChartData} />
          </Box>
          <ProviderLegend />
        </ReportCard>

        <ReportCard>
          <SectionTitle description="First-party workspace events captured for analytics and later questions.">
            Workspace engagement
          </SectionTitle>
          <Box className="mt-5 grid grid-cols-2 gap-3">
            <MiniMetric
              description="Total captured activity."
              label="Tracked events"
              value={formatNumber(report.engagement.totalEvents)}
            />
            <MiniMetric
              description="People with tracked activity."
              label="Active users"
              value={formatNumber(report.engagement.uniqueUsers)}
            />
          </Box>
          {report.engagement.topUsers.length ? (
            <Box className="mt-5">
              <Text className="mb-2" fontWeight="medium">
                Most active people
              </Text>
              {report.engagement.topUsers.slice(0, 6).map((user) => (
                <EngagementUserRow key={user.userId} user={user} />
              ))}
            </Box>
          ) : null}
        </ReportCard>
      </Box>

      <Box className="grid gap-5 @6xl:grid-cols-2">
        <ReportCard>
          <SectionTitle description="Which tracked events are happening most often.">
            Event mix
          </SectionTitle>
          <Box className="mt-5">
            <HorizontalBreakdownChart
              color={chartPalette.info}
              data={eventChartData}
              emptyText="No event names have been tracked yet."
              height={300}
            />
          </Box>
        </ReportCard>
        <ReportCard>
          <SectionTitle description="Where tracked activity is happening in the product.">
            Surface mix
          </SectionTitle>
          <Box className="mt-5">
            <HorizontalBreakdownChart
              color={chartPalette.violet}
              data={surfaceChartData}
              emptyText="No surfaces have been tracked yet."
              height={300}
            />
          </Box>
        </ReportCard>
      </Box>
    </Box>
  );
};

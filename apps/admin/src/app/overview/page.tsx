import Link from "next/link";
import { AnalyticsIcon, HistoryIcon, UserIcon, WorkspaceIcon } from "icons";
import { Avatar, Badge, Box, Button, Flex, Table, Text } from "ui";
import {
  getAuditLogs,
  getDashboardSummary,
  getWorkspaces,
} from "@/lib/admin-api";
import {
  formatCount,
  formatDate,
  formatDateTime,
  formatTrialState,
  humanizeKey,
} from "@/lib/format";
import { MetricCard } from "@/components/metric-card";
import { PageHeader } from "@/components/page-header";
import { WorkspaceStatusBadge } from "@/components/status-badge";

export default async function OverviewPage() {
  const [summary, expiringWorkspaces, auditLogs] = await Promise.all([
    getDashboardSummary(),
    getWorkspaces({ status: "expired", limit: 6 }),
    getAuditLogs({ limit: 6 }),
  ]);

  return (
    <Box>
      <PageHeader
        description="Monitor platform health, billing posture, access state, and high-risk admin activity."
        eyebrow="Internal operations"
        icon={<AnalyticsIcon className="h-[1.1rem]" />}
        title="Admin overview"
      />

      <Box className="space-y-5 p-5 md:p-7">
        <Box className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <MetricCard
            detail={`${formatCount(summary.activeTrials)} active trials, ${formatCount(summary.expiredTrials)} expired`}
            icon={<WorkspaceIcon />}
            label="Workspaces"
            value={formatCount(summary.totalWorkspaces)}
          />
          <MetricCard
            detail={`${formatCount(summary.internalUsers)} internal users`}
            icon={<UserIcon />}
            label="Users"
            value={formatCount(summary.totalUsers)}
          />
          <MetricCard
            detail={`${formatCount(summary.paidWorkspaces)} paid workspaces`}
            icon={<AnalyticsIcon />}
            label="Active subscriptions"
            value={formatCount(summary.activeSubscriptions)}
          />
          <MetricCard
            detail={`${formatCount(summary.slackInstallations)} Slack, ${formatCount(summary.githubInstallations)} GitHub`}
            icon={<HistoryIcon />}
            label="Recent admin logs"
            value={formatCount(summary.recentAdminAuditLogs)}
          />
        </Box>

        <Box className="grid gap-5 xl:grid-cols-[minmax(0,1.15fr)_minmax(420px,0.85fr)]">
          <Box className="border-border bg-surface rounded-lg border-[0.5px]">
            <Flex
              align="center"
              className="border-border border-b-[0.5px] px-4 py-3"
              justify="between"
            >
              <Box>
                <Text fontWeight="semibold">Expired trials</Text>
                <Text className="mt-1 text-[0.95rem]" color="muted">
                  Workspaces that may need commercial or support follow-up.
                </Text>
              </Box>
              <Button
                color="tertiary"
                href="/workspaces?status=expired"
                size="sm"
                variant="naked"
              >
                Open list
              </Button>
            </Flex>
            <Box className="overflow-x-auto">
              <Table color="light" variant="bordered">
                <Table.Head>
                  <Table.Tr>
                    <Table.Th>Workspace</Table.Th>
                    <Table.Th>Status</Table.Th>
                    <Table.Th>Trial</Table.Th>
                    <Table.Th>Members</Table.Th>
                  </Table.Tr>
                </Table.Head>
                <Table.Body>
                  {expiringWorkspaces.items.length > 0 ? (
                    expiringWorkspaces.items.map((workspace) => (
                      <Table.Tr key={workspace.id}>
                        <Table.Td className="min-w-72 whitespace-nowrap">
                          <Flex align="center" className="gap-2">
                            <Avatar
                              name={workspace.name}
                              src={workspace.avatarUrl}
                            />
                            <Link
                              className="hover:text-primary line-clamp-1"
                              href={`/workspaces/${workspace.id}`}
                            >
                              {workspace.name}
                            </Link>
                            <Text
                              as="span"
                              className="line-clamp-1 text-[0.95rem]"
                              color="muted"
                            >
                              /{workspace.slug}
                            </Text>
                          </Flex>
                        </Table.Td>
                        <Table.Td>
                          <WorkspaceStatusBadge workspace={workspace} />
                        </Table.Td>
                        <Table.Td className="min-w-52 whitespace-nowrap">
                          <Text>
                            {formatDate(workspace.trialEndsOn)}
                            <Text
                              as="span"
                              className="ml-1 text-[0.95rem]"
                              color="muted"
                            >
                              · {formatTrialState(workspace.trialEndsOn)}
                            </Text>
                          </Text>
                        </Table.Td>
                        <Table.Td className="whitespace-nowrap">
                          {workspace.memberCount}
                        </Table.Td>
                      </Table.Tr>
                    ))
                  ) : (
                    <Table.Tr>
                      <Table.Td className="py-8 text-center" colSpan={4}>
                        <Text color="muted">No expired trials right now.</Text>
                      </Table.Td>
                    </Table.Tr>
                  )}
                </Table.Body>
              </Table>
            </Box>
          </Box>

          <Box className="border-border bg-surface rounded-lg border-[0.5px]">
            <Flex
              align="center"
              className="border-border border-b-[0.5px] px-4 py-3"
              justify="between"
            >
              <Box>
                <Text fontWeight="semibold">Recent audit activity</Text>
                <Text className="mt-1 text-[0.95rem]" color="muted">
                  Administrative changes recorded by the platform.
                </Text>
              </Box>
              <Button color="tertiary" href="/audit" size="sm" variant="naked">
                Audit log
              </Button>
            </Flex>
            <Box className="divide-border divide-y">
              {auditLogs.items.length > 0 ? (
                auditLogs.items.map((entry) => (
                  <Box className="px-4 py-3" key={entry.id}>
                    <Flex align="center" justify="between">
                      <Text fontWeight="semibold">
                        {humanizeKey(entry.action)}
                      </Text>
                      <Badge color="tertiary">{entry.targetType}</Badge>
                    </Flex>
                    <Text className="mt-1 text-[0.95rem]" color="muted">
                      {entry.actorName || "Unknown actor"} ·{" "}
                      {formatDateTime(entry.createdAt)}
                    </Text>
                    {entry.workspaceName ? (
                      <Text className="mt-1 text-[0.95rem]" color="muted">
                        {entry.workspaceName}
                      </Text>
                    ) : null}
                  </Box>
                ))
              ) : (
                <Box className="px-4 py-8 text-center">
                  <Text color="muted">No admin changes recorded yet.</Text>
                </Box>
              )}
            </Box>
          </Box>
        </Box>
      </Box>
    </Box>
  );
}

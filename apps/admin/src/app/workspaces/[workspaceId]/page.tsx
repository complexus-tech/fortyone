import Link from "next/link";
import type { ReactNode } from "react";
import { GitHubIcon, SlackIcon, UserIcon, WorkspaceIcon } from "icons";
import { Badge, Box, Flex, Table, Text } from "ui";
import { getAuditLogs, getWorkspace } from "@/lib/admin-api";
import {
  formatCount,
  formatDate,
  formatDateTime,
  formatTrialState,
  formatValue,
  humanizeKey,
} from "@/lib/format";
import { MetricCard } from "@/components/metric-card";
import { PageHeader } from "@/components/page-header";
import { TrialExtensionDialog } from "@/components/trial-extension-dialog";
import {
  UserStatusBadge,
  WorkspaceStatusBadge,
} from "@/components/status-badge";

export default async function WorkspaceDetailPage({
  params,
}: {
  params: Promise<{ workspaceId: string }>;
}) {
  const { workspaceId } = await params;
  const [overview, auditLogs] = await Promise.all([
    getWorkspace(workspaceId),
    getAuditLogs({ workspaceId, limit: 8 }),
  ]);
  const { workspace, members } = overview;

  return (
    <Box>
      <PageHeader
        actions={<TrialExtensionDialog workspace={workspace} />}
        description={`${workspace.slug} · Created ${formatDate(workspace.createdAt)}`}
        eyebrow="Workspace"
        title={workspace.name}
      />

      <Box className="space-y-5 p-5 md:p-7">
        <Box className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <MetricCard
            detail={`${formatCount(workspace.teamCount)} teams, ${formatCount(workspace.storyCount)} active stories`}
            icon={<WorkspaceIcon />}
            label="Workspace activity"
            value={`${formatCount(workspace.memberCount)} members`}
          />
          <MetricCard
            detail={formatTrialState(workspace.trialEndsOn)}
            icon={<UserIcon />}
            label="Trial ends"
            value={formatDate(workspace.trialEndsOn)}
          />
          <MetricCard
            detail={workspace.subscriptionStatus ?? "No active Stripe status"}
            label="Plan"
            value={workspace.subscriptionTier ?? "Free"}
          />
          <MetricCard
            detail={`Last opened ${formatDateTime(workspace.lastAccessedAt)}`}
            label="Status"
            value={workspace.deletedAt ? "Deleted" : "Active"}
          />
        </Box>

        <Box className="grid gap-5 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
          <Box className="border-border bg-surface rounded-xl border-[0.5px]">
            <Box className="border-border border-b-[0.5px] px-4 py-3">
              <Text fontWeight="semibold">Workspace details</Text>
              <Text className="mt-1 text-[0.92rem]" color="muted">
                Operational state and connected systems.
              </Text>
            </Box>
            <Box className="divide-border divide-y">
              <DetailRow label="Status">
                <WorkspaceStatusBadge workspace={workspace} />
              </DetailRow>
              <DetailRow label="Creator">
                <Text>
                  {workspace.createdByName ||
                    workspace.createdByEmail ||
                    "Unknown"}
                </Text>
                {workspace.createdByEmail ? (
                  <Text className="mt-0.5 text-[0.92rem]" color="muted">
                    {workspace.createdByEmail}
                  </Text>
                ) : null}
              </DetailRow>
              <DetailRow label="Stripe customer">
                <Text>{workspace.stripeCustomerId ?? "Not linked"}</Text>
              </DetailRow>
              <DetailRow label="Stripe subscription">
                <Text>{workspace.stripeSubscriptionId ?? "Not linked"}</Text>
              </DetailRow>
              <DetailRow label="Integrations">
                <Flex align="center" className="gap-2">
                  <Badge
                    color={workspace.slackInstalled ? "success" : "tertiary"}
                    rounded="full"
                    size="sm"
                    variant={workspace.slackInstalled ? "outline" : "solid"}
                  >
                    <SlackIcon className="h-3.5" />
                    Slack
                  </Badge>
                  <Badge
                    color={workspace.gitHubInstalled ? "success" : "tertiary"}
                    rounded="full"
                    size="sm"
                    variant={workspace.gitHubInstalled ? "outline" : "solid"}
                  >
                    <GitHubIcon className="h-3.5" />
                    GitHub
                  </Badge>
                </Flex>
              </DetailRow>
            </Box>
          </Box>

          <Box className="border-border bg-surface overflow-hidden rounded-xl border-[0.5px]">
            <Box className="border-border border-b-[0.5px] px-4 py-3">
              <Text fontWeight="semibold">Members</Text>
              <Text className="mt-1 text-[0.92rem]" color="muted">
                Users with access to this workspace.
              </Text>
            </Box>
            <Box className="overflow-x-auto">
              <Table>
                <Table.Head>
                  <Table.Tr>
                    <Table.Th>User</Table.Th>
                    <Table.Th>Role</Table.Th>
                    <Table.Th>Status</Table.Th>
                    <Table.Th>Joined</Table.Th>
                  </Table.Tr>
                </Table.Head>
                <Table.Body>
                  {members.map((member) => (
                    <Table.Tr key={member.userId}>
                      <Table.Td>
                        <Link
                          className="hover:text-primary line-clamp-1"
                          href={`/users/${member.userId}`}
                        >
                          {member.fullName || member.email}
                        </Link>
                        <Text className="mt-0.5 text-[0.92rem]" color="muted">
                          {member.email}
                        </Text>
                      </Table.Td>
                      <Table.Td className="capitalize">{member.role}</Table.Td>
                      <Table.Td>
                        <UserStatusBadge
                          isActive
                          isInternal={member.isInternal}
                        />
                      </Table.Td>
                      <Table.Td>{formatDate(member.joinedAt)}</Table.Td>
                    </Table.Tr>
                  ))}
                </Table.Body>
              </Table>
            </Box>
          </Box>
        </Box>

        <Box className="border-border bg-surface overflow-hidden rounded-xl border-[0.5px]">
          <Box className="border-border border-b-[0.5px] px-4 py-3">
            <Text fontWeight="semibold">Workspace audit history</Text>
            <Text className="mt-1 text-[0.92rem]" color="muted">
              Admin actions related to this workspace.
            </Text>
          </Box>
          <Box className="overflow-x-auto">
            <Table>
              <Table.Head>
                <Table.Tr>
                  <Table.Th>Action</Table.Th>
                  <Table.Th>Actor</Table.Th>
                  <Table.Th>Change</Table.Th>
                  <Table.Th>Reason</Table.Th>
                  <Table.Th>Time</Table.Th>
                </Table.Tr>
              </Table.Head>
              <Table.Body>
                {auditLogs.items.length > 0 ? (
                  auditLogs.items.map((entry) => (
                    <Table.Tr key={entry.id}>
                      <Table.Td>{humanizeKey(entry.action)}</Table.Td>
                      <Table.Td>{entry.actorName || entry.actorEmail}</Table.Td>
                      <Table.Td className="min-w-52">
                        <Text>
                          {formatValue(entry.oldValue)} -&gt;{" "}
                          {formatValue(entry.newValue)}
                        </Text>
                        {entry.fieldName ? (
                          <Text className="mt-0.5 text-[0.92rem]" color="muted">
                            {entry.fieldName}
                          </Text>
                        ) : null}
                      </Table.Td>
                      <Table.Td className="max-w-80">
                        <Text className="line-clamp-2">
                          {entry.reason || "No reason provided"}
                        </Text>
                      </Table.Td>
                      <Table.Td>{formatDateTime(entry.createdAt)}</Table.Td>
                    </Table.Tr>
                  ))
                ) : (
                  <Table.Tr>
                    <Table.Td className="py-10 text-center" colSpan={5}>
                      <Text color="muted">
                        No admin audit entries for this workspace.
                      </Text>
                    </Table.Td>
                  </Table.Tr>
                )}
              </Table.Body>
            </Table>
          </Box>
        </Box>
      </Box>
    </Box>
  );
}

const DetailRow = ({
  children,
  label,
}: {
  children: ReactNode;
  label: string;
}) => {
  return (
    <Flex align="start" className="gap-4 px-4 py-3" justify="between">
      <Text className="text-[0.92rem]" color="muted">
        {label}
      </Text>
      <Box className="max-w-[70%] text-right">{children}</Box>
    </Flex>
  );
};

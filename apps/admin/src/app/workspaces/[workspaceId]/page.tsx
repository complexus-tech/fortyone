import Link from "next/link";
import type { ReactNode } from "react";
import { GitHubIcon, SlackIcon, UserIcon, WorkspaceIcon } from "icons";
import { Avatar, Badge, Box, Flex, Table, Text } from "ui";
import { getAuditLogs, getWorkspace } from "@/lib/admin-api";
import {
  formatAuditValue,
  formatCount,
  formatDate,
  formatDateTime,
  formatTrialState,
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
        description={`${workspace.slug} · Created ${formatDate(workspace.createdAt)}`}
        eyebrow="Workspace"
        icon={
          <Avatar
            className="h-5 text-[0.7rem]"
            name={workspace.name}
            src={workspace.avatarUrl}
          />
        }
        parentHref="/workspaces"
        title={workspace.name}
        titleActions={<TrialExtensionDialog workspace={workspace} />}
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
          <Box className="border-border bg-surface rounded-lg border-[0.5px]">
            <Box className="border-border border-b-[0.5px] px-4 py-3">
              <Text fontWeight="semibold">Workspace details</Text>
              <Text className="mt-1 text-[0.95rem]" color="muted">
                Operational state and connected systems.
              </Text>
            </Box>
            <Box className="divide-border divide-y">
              <DetailRow label="Status">
                <WorkspaceStatusBadge workspace={workspace} />
              </DetailRow>
              <DetailRow label="Creator">
                <Text
                  className="truncate"
                  title={[workspace.createdByName, workspace.createdByEmail]
                    .filter(Boolean)
                    .join(" - ")}
                >
                  {workspace.createdByName ||
                    workspace.createdByEmail ||
                    "Unknown"}
                  {workspace.createdByName && workspace.createdByEmail ? (
                    <Text
                      as="span"
                      className="ml-1 text-[0.95rem]"
                      color="muted"
                    >
                      · {workspace.createdByEmail}
                    </Text>
                  ) : null}
                </Text>
              </DetailRow>
              <DetailRow label="Stripe customer">
                <Text>{workspace.stripeCustomerId ?? "Not linked"}</Text>
              </DetailRow>
              <DetailRow label="Stripe subscription">
                <Text>{workspace.stripeSubscriptionId ?? "Not linked"}</Text>
              </DetailRow>
              <DetailRow label="Integrations">
                <Flex align="center" className="gap-2">
                  <Badge color="tertiary">
                    <SlackIcon className="h-4" />
                    {workspace.slackInstalled ? "Slack" : "No Slack"}
                  </Badge>
                  <Badge color="tertiary">
                    <GitHubIcon className="h-4" />
                    {workspace.githubInstalled ? "GitHub" : "No GitHub"}
                  </Badge>
                </Flex>
              </DetailRow>
            </Box>
          </Box>

          <Box className="border-border bg-surface overflow-hidden rounded-lg border-[0.5px]">
            <Box className="border-border border-b-[0.5px] px-4 py-3">
              <Text fontWeight="semibold">Members</Text>
              <Text className="mt-1 text-[0.95rem]" color="muted">
                Users with access to this workspace.
              </Text>
            </Box>
            <Box className="overflow-x-auto">
              <Table color="light" variant="bordered">
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
                      <Table.Td className="min-w-80 whitespace-nowrap">
                        <Flex align="center" className="gap-2">
                          <Link
                            className="hover:text-primary line-clamp-1"
                            href={`/users/${member.userId}`}
                          >
                            {member.fullName || "Unknown user"}
                          </Link>
                        </Flex>
                      </Table.Td>
                      <Table.Td className="whitespace-nowrap capitalize">
                        {member.role}
                      </Table.Td>
                      <Table.Td>
                        <UserStatusBadge
                          isActive
                          isInternal={member.isInternal}
                        />
                      </Table.Td>
                      <Table.Td className="whitespace-nowrap">
                        {formatDate(member.joinedAt)}
                      </Table.Td>
                    </Table.Tr>
                  ))}
                </Table.Body>
              </Table>
            </Box>
          </Box>
        </Box>

        <Box className="border-border bg-surface overflow-hidden rounded-lg border-[0.5px]">
          <Box className="border-border border-b-[0.5px] px-4 py-3">
            <Text fontWeight="semibold">Workspace audit history</Text>
            <Text className="mt-1 text-[0.95rem]" color="muted">
              Admin actions related to this workspace.
            </Text>
          </Box>
          <Box className="overflow-x-auto">
            <Table color="light" variant="bordered">
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
                      <Table.Td className="min-w-48 whitespace-nowrap">
                        {humanizeKey(entry.action)}
                      </Table.Td>
                      <Table.Td className="min-w-64 whitespace-nowrap">
                        {entry.actorName || "Unknown actor"}
                      </Table.Td>
                      <Table.Td className="min-w-72 whitespace-nowrap">
                        <Text>
                          <Text as="span" fontWeight="semibold">
                            {entry.fieldName
                              ? humanizeKey(entry.fieldName)
                              : "Change"}
                          </Text>
                          <Text
                            as="span"
                            className="ml-1 text-[0.95rem]"
                            color="muted"
                          >
                            {formatAuditValue(entry.oldValue)} -&gt;{" "}
                            {formatAuditValue(entry.newValue)}
                          </Text>
                        </Text>
                      </Table.Td>
                      <Table.Td className="max-w-80">
                        <Text className="line-clamp-2">
                          {entry.reason || "No reason provided"}
                        </Text>
                      </Table.Td>
                      <Table.Td className="whitespace-nowrap">
                        {formatDateTime(entry.createdAt)}
                      </Table.Td>
                    </Table.Tr>
                  ))
                ) : (
                  <Table.Tr>
                    <Table.Td
                      className="h-36 text-center align-middle"
                      colSpan={5}
                    >
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
      <Text className="text-[0.95rem]" color="muted">
        {label}
      </Text>
      <Box className="max-w-[70%] text-right">{children}</Box>
    </Flex>
  );
};

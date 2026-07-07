import Link from "next/link";
import type { ReactNode } from "react";
import { UserIcon, WorkspaceIcon } from "icons";
import { Avatar, Box, Flex, Table, Text } from "ui";
import { getUser } from "@/lib/admin-api";
import { formatCount, formatDate, formatDateTime } from "@/lib/format";
import { MetricCard } from "@/components/metric-card";
import { PageHeader } from "@/components/page-header";
import { UserStatusBadge } from "@/components/status-badge";

export default async function UserDetailPage({
  params,
}: {
  params: Promise<{ userId: string }>;
}) {
  const { userId } = await params;
  const overview = await getUser(userId);
  const { user, memberships } = overview;

  return (
    <Box>
      <PageHeader
        description={`${user.email} · Joined ${formatDate(user.createdAt)}`}
        eyebrow="User"
        icon={
          <Avatar
            className="h-5 text-[0.7rem]"
            name={user.fullName || user.username}
            src={user.avatarUrl}
          />
        }
        parentHref="/users"
        title={user.fullName || user.username}
      />

      <Box className="space-y-5 p-5 md:p-7">
        <Box className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
          <MetricCard
            detail={
              user.isInternal
                ? "Internal admin-capable account"
                : "Customer account"
            }
            icon={<UserIcon />}
            label="Account status"
            value={user.isActive ? "Active" : "Inactive"}
          />
          <MetricCard
            detail="Workspace memberships"
            icon={<WorkspaceIcon />}
            label="Access"
            value={formatCount(user.workspaceCount)}
          />
          <MetricCard
            detail={user.lastUsedWorkspace ?? "No workspace recorded"}
            label="Last login"
            value={formatDate(user.lastLoginAt)}
          />
          <MetricCard
            detail={
              user.githubUsername
                ? `GitHub: ${user.githubUsername}`
                : "No GitHub username"
            }
            label="Identity"
            value={user.username}
          />
        </Box>

        <Box className="grid gap-5 xl:grid-cols-[minmax(0,0.85fr)_minmax(0,1.15fr)]">
          <Box className="border-border bg-surface rounded-lg border-[0.5px]">
            <Box className="border-border border-b-[0.5px] px-4 py-3">
              <Text fontWeight="semibold">Account details</Text>
              <Text className="mt-1 text-[0.95rem]" color="muted">
                Core user metadata and access flags.
              </Text>
            </Box>
            <Box className="divide-border divide-y">
              <DetailRow label="Status">
                <UserStatusBadge
                  isActive={user.isActive}
                  isInternal={user.isInternal}
                />
              </DetailRow>
              <DetailRow label="Email">
                <Text>{user.email}</Text>
              </DetailRow>
              <DetailRow label="Username">
                <Text>{user.username}</Text>
              </DetailRow>
              <DetailRow label="Last login">
                <Text>{formatDateTime(user.lastLoginAt)}</Text>
              </DetailRow>
              <DetailRow label="Last workspace">
                <Text>{user.lastUsedWorkspace ?? "Not available"}</Text>
              </DetailRow>
              <DetailRow label="Updated">
                <Text>{formatDateTime(user.updatedAt)}</Text>
              </DetailRow>
            </Box>
          </Box>

          <Box className="border-border overflow-hidden rounded-lg border-[0.5px]">
            <Box className="border-border border-b-[0.5px] px-4 py-3">
              <Text fontWeight="semibold">Workspace memberships</Text>
              <Text className="mt-1 text-[0.95rem]" color="muted">
                Workspaces this user can access.
              </Text>
            </Box>
            <Box className="overflow-x-auto">
              <Table color="light" variant="bordered">
                <Table.Head>
                  <Table.Tr>
                    <Table.Th>Workspace</Table.Th>
                    <Table.Th>Role</Table.Th>
                    <Table.Th>Joined</Table.Th>
                  </Table.Tr>
                </Table.Head>
                <Table.Body>
                  {memberships.length > 0 ? (
                    memberships.map((membership) => (
                      <Table.Tr key={membership.workspaceId}>
                        <Table.Td className="min-w-72 whitespace-nowrap">
                          <Flex align="center" className="gap-2">
                            <Link
                              className="hover:text-primary line-clamp-1"
                              href={`/workspaces/${membership.workspaceId}`}
                            >
                              {membership.workspaceName}
                            </Link>
                            <Text
                              as="span"
                              className="line-clamp-1 text-[0.95rem]"
                              color="muted"
                            >
                              /{membership.workspaceSlug}
                            </Text>
                          </Flex>
                        </Table.Td>
                        <Table.Td className="whitespace-nowrap capitalize">
                          {membership.role}
                        </Table.Td>
                        <Table.Td className="whitespace-nowrap">
                          {formatDate(membership.joinedAt)}
                        </Table.Td>
                      </Table.Tr>
                    ))
                  ) : (
                    <Table.Tr>
                      <Table.Td className="py-10 text-center" colSpan={3}>
                        <Text color="muted">
                          This user has no workspace memberships.
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

import Link from "next/link";
import { Box, Table, Text } from "ui";
import { getUsers } from "@/lib/admin-api";
import { formatDate, formatDateTime } from "@/lib/format";
import { PageHeader } from "@/components/page-header";
import { PaginationControls } from "@/components/pagination-controls";
import { SearchToolbar } from "@/components/search-toolbar";
import { UserStatusBadge } from "@/components/status-badge";

export default async function UsersPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string; q?: string }>;
}) {
  const params = await searchParams;
  const users = await getUsers({
    page: params.page ?? 1,
    q: params.q,
    limit: 25,
  });

  return (
    <Box>
      <PageHeader
        description="Search users, inspect account status, and review workspace access."
        eyebrow="Platform"
        title="Users"
      />

      <Box className="space-y-4 p-5 md:p-7">
        <SearchToolbar
          defaultQuery={params.q}
          placeholder="Search by name, username, or email"
        />

        <Box className="border-border bg-surface overflow-hidden rounded-xl border-[0.5px]">
          <Box className="overflow-x-auto">
            <Table>
              <Table.Head>
                <Table.Tr>
                  <Table.Th>User</Table.Th>
                  <Table.Th>Status</Table.Th>
                  <Table.Th>Workspaces</Table.Th>
                  <Table.Th>Last used workspace</Table.Th>
                  <Table.Th>Last login</Table.Th>
                  <Table.Th>Created</Table.Th>
                </Table.Tr>
              </Table.Head>
              <Table.Body>
                {users.items.length > 0 ? (
                  users.items.map((user) => (
                    <Table.Tr key={user.id}>
                      <Table.Td className="min-w-64">
                        <Link
                          className="hover:text-primary line-clamp-1"
                          href={`/users/${user.id}`}
                        >
                          {user.fullName || user.username}
                        </Link>
                        <Text className="mt-0.5 text-[0.92rem]" color="muted">
                          {user.email}
                        </Text>
                        {user.gitHubUsername ? (
                          <Text className="mt-0.5 text-[0.92rem]" color="muted">
                            GitHub: {user.gitHubUsername}
                          </Text>
                        ) : null}
                      </Table.Td>
                      <Table.Td>
                        <UserStatusBadge
                          isActive={user.isActive}
                          isInternal={user.isInternal}
                        />
                      </Table.Td>
                      <Table.Td>{user.workspaceCount}</Table.Td>
                      <Table.Td className="min-w-48">
                        {user.lastUsedWorkspace ?? "Not available"}
                      </Table.Td>
                      <Table.Td>{formatDateTime(user.lastLoginAt)}</Table.Td>
                      <Table.Td>{formatDate(user.createdAt)}</Table.Td>
                    </Table.Tr>
                  ))
                ) : (
                  <Table.Tr>
                    <Table.Td className="py-10 text-center" colSpan={6}>
                      <Text color="muted">No users match this search.</Text>
                    </Table.Td>
                  </Table.Tr>
                )}
              </Table.Body>
            </Table>
          </Box>
          <PaginationControls
            pagination={users.pagination}
            params={{ q: params.q }}
            pathname="/users"
          />
        </Box>
      </Box>
    </Box>
  );
}

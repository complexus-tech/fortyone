import Link from "next/link";
import { GitHubIcon, SlackIcon } from "icons";
import { Badge, Box, Flex, Table, Text } from "ui";
import { getWorkspaces } from "@/lib/admin-api";
import { formatDate, formatDateTime, formatTrialState } from "@/lib/format";
import { PageHeader } from "@/components/page-header";
import { PaginationControls } from "@/components/pagination-controls";
import { SearchToolbar } from "@/components/search-toolbar";
import { WorkspaceStatusBadge } from "@/components/status-badge";

const statusOptions = [
  { label: "All statuses", value: "" },
  { label: "Active", value: "active" },
  { label: "Trialing", value: "trialing" },
  { label: "Expired", value: "expired" },
  { label: "Paid", value: "paid" },
  { label: "Deleted", value: "deleted" },
];

export default async function WorkspacesPage({
  searchParams,
}: {
  searchParams: Promise<{ page?: string; q?: string; status?: string }>;
}) {
  const params = await searchParams;
  const workspaces = await getWorkspaces({
    page: params.page ?? 1,
    q: params.q,
    status: params.status,
    limit: 25,
  });

  return (
    <Box>
      <PageHeader
        description="Inspect workspace health, billing posture, integrations, trial windows, and account ownership."
        eyebrow="Platform"
        title="Workspaces"
      />

      <Box className="space-y-4 p-5 md:p-7">
        <SearchToolbar
          defaultQuery={params.q}
          defaultStatus={params.status}
          placeholder="Search by workspace, slug, or creator email"
          statusOptions={statusOptions}
        />

        <Box className="border-border bg-surface overflow-hidden rounded-lg border-[0.5px]">
          <Box className="overflow-x-auto">
            <Table>
              <Table.Head>
                <Table.Tr>
                  <Table.Th>Workspace</Table.Th>
                  <Table.Th>Status</Table.Th>
                  <Table.Th>Trial</Table.Th>
                  <Table.Th>Plan</Table.Th>
                  <Table.Th>Activity</Table.Th>
                  <Table.Th>Integrations</Table.Th>
                </Table.Tr>
              </Table.Head>
              <Table.Body>
                {workspaces.items.length > 0 ? (
                  workspaces.items.map((workspace) => (
                    <Table.Tr key={workspace.id}>
                      <Table.Td className="min-w-64">
                        <Link
                          className="hover:text-primary line-clamp-1"
                          href={`/workspaces/${workspace.id}`}
                        >
                          {workspace.name}
                        </Link>
                        <Text className="mt-0.5 text-[0.92rem]" color="muted">
                          {workspace.slug}
                        </Text>
                        {workspace.createdByEmail ? (
                          <Text
                            className="mt-0.5 line-clamp-1 text-[0.92rem]"
                            color="muted"
                          >
                            Created by{" "}
                            {workspace.createdByName ||
                              workspace.createdByEmail}
                          </Text>
                        ) : null}
                      </Table.Td>
                      <Table.Td>
                        <WorkspaceStatusBadge workspace={workspace} />
                      </Table.Td>
                      <Table.Td className="min-w-36">
                        <Text>{formatDate(workspace.trialEndsOn)}</Text>
                        <Text className="mt-0.5 text-[0.92rem]" color="muted">
                          {formatTrialState(workspace.trialEndsOn)}
                        </Text>
                      </Table.Td>
                      <Table.Td className="min-w-36">
                        <Text className="capitalize">
                          {workspace.subscriptionTier ?? "free"}
                        </Text>
                        <Text className="mt-0.5 text-[0.92rem]" color="muted">
                          {workspace.subscriptionStatus ?? "No subscription"}
                        </Text>
                      </Table.Td>
                      <Table.Td className="min-w-44">
                        <Text>
                          {workspace.memberCount} members ·{" "}
                          {workspace.teamCount} teams
                        </Text>
                        <Text className="mt-0.5 text-[0.92rem]" color="muted">
                          Last opened {formatDateTime(workspace.lastAccessedAt)}
                        </Text>
                      </Table.Td>
                      <Table.Td>
                        <Flex align="center" className="gap-2">
                          <Badge
                            color={
                              workspace.slackInstalled ? "success" : "tertiary"
                            }
                            rounded="full"
                            size="sm"
                            variant={
                              workspace.slackInstalled ? "outline" : "solid"
                            }
                          >
                            <SlackIcon className="h-3.5" />
                            Slack
                          </Badge>
                          <Badge
                            color={
                              workspace.gitHubInstalled ? "success" : "tertiary"
                            }
                            rounded="full"
                            size="sm"
                            variant={
                              workspace.gitHubInstalled ? "outline" : "solid"
                            }
                          >
                            <GitHubIcon className="h-3.5" />
                            GitHub
                          </Badge>
                        </Flex>
                      </Table.Td>
                    </Table.Tr>
                  ))
                ) : (
                  <Table.Tr>
                    <Table.Td className="py-10 text-center" colSpan={6}>
                      <Text color="muted">No workspaces match this view.</Text>
                    </Table.Td>
                  </Table.Tr>
                )}
              </Table.Body>
            </Table>
          </Box>
          <PaginationControls
            pagination={workspaces.pagination}
            params={{ q: params.q, status: params.status }}
            pathname="/workspaces"
          />
        </Box>
      </Box>
    </Box>
  );
}

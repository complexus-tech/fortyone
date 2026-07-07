import Link from "next/link";
import { GitHubIcon, SlackIcon, WorkspaceIcon } from "icons";
import { Avatar, Badge, Box, Flex, Table, Text } from "ui";
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
        icon={<WorkspaceIcon className="h-[1.1rem]" />}
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
            <Table color="light" variant="bordered">
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
                      <Table.Td className="min-w-80 whitespace-nowrap">
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
                      <Table.Td className="min-w-48 whitespace-nowrap">
                        <Text>
                          <Text as="span" className="capitalize">
                            {workspace.subscriptionTier ?? "free"}
                          </Text>
                          <Text
                            as="span"
                            className="ml-1 text-[0.95rem]"
                            color="muted"
                          >
                            ·{" "}
                            {workspace.subscriptionStatus ?? "No subscription"}
                          </Text>
                        </Text>
                      </Table.Td>
                      <Table.Td className="min-w-72 whitespace-nowrap">
                        <Text>
                          {workspace.memberCount} members ·{" "}
                          {workspace.teamCount} teams
                          <Text
                            as="span"
                            className="ml-1 text-[0.95rem]"
                            color="muted"
                          >
                            · Last opened{" "}
                            {formatDateTime(workspace.lastAccessedAt)}
                          </Text>
                        </Text>
                      </Table.Td>
                      <Table.Td className="whitespace-nowrap">
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

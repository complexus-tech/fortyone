import Link from "next/link";
import type { ReactNode } from "react";
import {
  GitHubIcon,
  GoogleCalendarIcon,
  ImageIcon,
  LinkIcon,
  SlackIcon,
} from "icons";
import { Avatar, Box, Flex, Table, Text } from "ui";
import { IntegrationFilterToolbar } from "@/components/integration-filter-toolbar";
import { IntegrationStatusBadge } from "@/components/integration-status-badge";
import { PageHeader } from "@/components/page-header";
import { PaginationControls } from "@/components/pagination-controls";
import { getIntegrationWorkspaces } from "@/lib/admin-api";
import { formatDateTime } from "@/lib/format";
import type { IntegrationProvider, IntegrationSummary } from "@/lib/types";

const providerMeta: Record<
  IntegrationProvider,
  { label: string; icon: ReactNode }
> = {
  slack: { label: "Slack", icon: <SlackIcon className="h-4" /> },
  github: { label: "GitHub", icon: <GitHubIcon className="h-4" /> },
  calendar: {
    label: "Calendar",
    icon: <GoogleCalendarIcon className="h-4" />,
  },
  figma: { label: "Figma", icon: <ImageIcon className="h-4" /> },
};

const IntegrationCell = ({
  integration,
}: {
  integration: IntegrationSummary;
}) => {
  const meta = providerMeta[integration.provider];
  const hasMappings = integration.provider !== "figma";

  return (
    <Box className="min-w-44 space-y-1.5">
      <Flex align="center" className="gap-1.5">
        {meta.icon}
        <Text fontWeight="semibold">{meta.label}</Text>
      </Flex>
      <IntegrationStatusBadge state={integration.state} />
      {integration.accountLabel ? (
        <Text className="max-w-48 truncate text-[0.9rem]" color="muted">
          {integration.accountLabel}
        </Text>
      ) : null}
      {hasMappings && integration.state !== "not_connected" ? (
        <Text className="text-[0.9rem]" color="muted">
          {integration.mappingCount} mapped · {integration.unmappedMemberCount}{" "}
          missing
        </Text>
      ) : null}
      {integration.lastSyncedAt ? (
        <Text className="text-[0.85rem]" color="muted">
          Updated {formatDateTime(integration.lastSyncedAt)}
        </Text>
      ) : null}
    </Box>
  );
};

export default async function IntegrationsPage({
  searchParams,
}: {
  searchParams: Promise<{
    page?: string;
    provider?: string;
    q?: string;
    status?: string;
  }>;
}) {
  const params = await searchParams;
  const result = await getIntegrationWorkspaces({
    page: params.page ?? 1,
    limit: 25,
    provider: params.provider,
    q: params.q,
    status: params.status,
  });

  return (
    <Box>
      <PageHeader
        description="Inspect connected services, synchronization health, and member identity coverage across customer workspaces."
        eyebrow="Platform"
        icon={<LinkIcon className="h-[1.1rem]" />}
        title="Integrations"
      />

      <Box className="space-y-4 p-5 md:p-7">
        <IntegrationFilterToolbar
          defaultProvider={params.provider}
          defaultQuery={params.q}
          defaultStatus={params.status}
        />

        <Box className="border-border overflow-hidden rounded-lg border-[0.5px]">
          <Box className="overflow-x-auto">
            <Table color="light" variant="bordered">
              <Table.Head>
                <Table.Tr>
                  <Table.Th>Workspace</Table.Th>
                  <Table.Th>Slack</Table.Th>
                  <Table.Th>GitHub</Table.Th>
                  <Table.Th>Calendar</Table.Th>
                  <Table.Th>Figma</Table.Th>
                </Table.Tr>
              </Table.Head>
              <Table.Body>
                {result.items.length > 0 ? (
                  result.items.map((workspace) => {
                    const integrations = new Map(
                      workspace.integrations.map((integration) => [
                        integration.provider,
                        integration,
                      ]),
                    );
                    return (
                      <Table.Tr key={workspace.workspaceId}>
                        <Table.Td className="min-w-72 align-top">
                          <Flex align="center" className="gap-2">
                            <Avatar
                              name={workspace.workspaceName}
                              src={workspace.workspaceAvatar}
                            />
                            <Box className="min-w-0">
                              <Link
                                className="hover:text-primary line-clamp-1"
                                href={`/workspaces/${workspace.workspaceId}`}
                              >
                                {workspace.workspaceName}
                              </Link>
                              <Text
                                className="truncate text-[0.9rem]"
                                color="muted"
                              >
                                /{workspace.workspaceSlug} ·{" "}
                                {workspace.memberCount} members ·{" "}
                                {workspace.subscriptionTier ?? "free"}
                              </Text>
                            </Box>
                          </Flex>
                        </Table.Td>
                        {(
                          ["slack", "github", "calendar", "figma"] as const
                        ).map((provider) => {
                          const integration = integrations.get(provider);
                          return (
                            <Table.Td className="align-top" key={provider}>
                              {integration ? (
                                <IntegrationCell integration={integration} />
                              ) : null}
                            </Table.Td>
                          );
                        })}
                      </Table.Tr>
                    );
                  })
                ) : (
                  <Table.Tr>
                    <Table.Td className="py-12 text-center" colSpan={5}>
                      <Text color="muted">
                        No workspaces match this integration view.
                      </Text>
                    </Table.Td>
                  </Table.Tr>
                )}
              </Table.Body>
            </Table>
          </Box>
          <PaginationControls
            pagination={result.pagination}
            params={{
              provider: params.provider,
              q: params.q,
              status: params.status,
            }}
            pathname="/integrations"
          />
        </Box>
      </Box>
    </Box>
  );
}

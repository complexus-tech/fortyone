import Link from "next/link";
import type { ReactNode } from "react";
import { GitHubIcon, GoogleCalendarIcon, ImageIcon, SlackIcon } from "icons";
import { Badge, Box, Flex, Table, Text } from "ui";
import { IntegrationStatusBadge } from "@/components/integration-status-badge";
import { formatDateTime, humanizeKey } from "@/lib/format";
import type {
  IntegrationProvider,
  IntegrationSummary,
  WorkspaceIntegrations,
} from "@/lib/types";

const providerMeta: Record<
  IntegrationProvider,
  { icon: ReactNode; label: string }
> = {
  slack: { icon: <SlackIcon className="h-4" />, label: "Slack" },
  github: { icon: <GitHubIcon className="h-4" />, label: "GitHub" },
  calendar: {
    icon: <GoogleCalendarIcon className="h-4" />,
    label: "Calendar",
  },
  figma: { icon: <ImageIcon className="h-4" />, label: "Figma" },
};

const SummaryCard = ({ summary }: { summary: IntegrationSummary }) => {
  const meta = providerMeta[summary.provider];
  return (
    <Box className="border-border bg-surface rounded-lg border-[0.5px] p-4">
      <Flex align="center" className="gap-2" justify="between">
        <Flex align="center" className="gap-2">
          {meta.icon}
          <Text fontWeight="semibold">{meta.label}</Text>
        </Flex>
        <IntegrationStatusBadge state={summary.state} />
      </Flex>
      <Text className="mt-3 truncate" fontWeight="semibold">
        {summary.accountLabel ??
          (summary.connectionCount > 0
            ? `${summary.connectionCount} connected`
            : "Not connected")}
      </Text>
      <Text className="mt-1 text-[0.9rem]" color="muted">
        {summary.provider === "figma"
          ? `${summary.connectionCount} workspace connection`
          : `${summary.mappingCount} members mapped · ${summary.unmappedMemberCount} missing`}
      </Text>
      {summary.lastSyncedAt ? (
        <Text className="mt-1 text-[0.85rem]" color="muted">
          Last activity {formatDateTime(summary.lastSyncedAt)}
        </Text>
      ) : null}
    </Box>
  );
};

const NotLinked = () => (
  <Text className="text-[0.9rem]" color="muted">
    Not linked
  </Text>
);

export const WorkspaceIntegrationsPanel = ({
  integrations,
}: {
  integrations: WorkspaceIntegrations;
}) => (
  <Box className="space-y-5">
    <Box>
      <Flex align="end" className="mb-3 gap-4" justify="between">
        <Box>
          <Text fontWeight="semibold">Integrations</Text>
          <Text className="mt-1 text-[0.95rem]" color="muted">
            Connection health, synchronization coverage, and external member
            identities.
          </Text>
        </Box>
        <Badge color="tertiary">Read only</Badge>
      </Flex>
      <Box className="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        {integrations.summaries.map((summary) => (
          <SummaryCard key={summary.provider} summary={summary} />
        ))}
      </Box>
    </Box>

    <Box className="border-border overflow-hidden rounded-lg border-[0.5px]">
      <Box className="border-border border-b-[0.5px] px-4 py-3">
        <Text fontWeight="semibold">Connection details</Text>
        <Text className="mt-1 text-[0.95rem]" color="muted">
          Operational metadata only. Credentials and provider content are never
          exposed.
        </Text>
      </Box>
      <Box className="overflow-x-auto">
        <Table color="light" variant="bordered">
          <Table.Head>
            <Table.Tr>
              <Table.Th>Provider</Table.Th>
              <Table.Th>Account</Table.Th>
              <Table.Th>State</Table.Th>
              <Table.Th>Configuration</Table.Th>
              <Table.Th>Connected by</Table.Th>
              <Table.Th>Last activity</Table.Th>
            </Table.Tr>
          </Table.Head>
          <Table.Body>
            {integrations.slack ? (
              <Table.Tr>
                <Table.Td>
                  <Flex align="center" className="gap-2">
                    <SlackIcon className="h-4" />
                    Slack
                  </Flex>
                </Table.Td>
                <Table.Td className="min-w-52">
                  {integrations.slack.teamName}
                  <Text className="text-[0.9rem]" color="muted">
                    {integrations.slack.teamDomain}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <IntegrationStatusBadge state="connected" />
                </Table.Td>
                <Table.Td className="min-w-56">
                  {integrations.slack.channelCount} channels ·{" "}
                  {integrations.slack.channelMappingCount} team mappings
                </Table.Td>
                <Table.Td className="min-w-56">
                  {integrations.slack.installedByName ??
                    integrations.slack.installedByEmail ??
                    "Unknown"}
                </Table.Td>
                <Table.Td className="whitespace-nowrap">
                  {formatDateTime(integrations.slack.lastSyncedAt)}
                </Table.Td>
              </Table.Tr>
            ) : null}
            {integrations.github.map((installation) => (
              <Table.Tr key={installation.id}>
                <Table.Td>
                  <Flex align="center" className="gap-2">
                    <GitHubIcon className="h-4" />
                    GitHub
                  </Flex>
                </Table.Td>
                <Table.Td className="min-w-52">
                  {installation.accountLogin}
                  <Text className="text-[0.9rem] capitalize" color="muted">
                    {installation.accountType}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <IntegrationStatusBadge state={installation.state} />
                </Table.Td>
                <Table.Td className="min-w-56">
                  {installation.repositoryCount} repositories ·{" "}
                  {installation.teamMappingCount} team mappings
                  <Text className="text-[0.9rem] capitalize" color="muted">
                    {installation.repositorySelection} repositories
                  </Text>
                </Table.Td>
                <Table.Td className="min-w-56">
                  {installation.installedByName ??
                    installation.installedByEmail ??
                    "Unknown"}
                </Table.Td>
                <Table.Td className="whitespace-nowrap">
                  {formatDateTime(installation.lastSyncedAt)}
                </Table.Td>
              </Table.Tr>
            ))}
            {integrations.calendar.map((connection) => (
              <Table.Tr key={connection.id}>
                <Table.Td>
                  <Flex align="center" className="gap-2">
                    <GoogleCalendarIcon className="h-4" />
                    Calendar
                  </Flex>
                </Table.Td>
                <Table.Td className="min-w-52">
                  {connection.connectedEmail}
                  <Text className="text-[0.9rem] capitalize" color="muted">
                    {connection.provider}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <IntegrationStatusBadge state={connection.state} />
                </Table.Td>
                <Table.Td>Member calendar</Table.Td>
                <Table.Td className="min-w-56">
                  {connection.userName || connection.userEmail}
                </Table.Td>
                <Table.Td className="whitespace-nowrap">
                  {formatDateTime(connection.lastSyncedAt)}
                </Table.Td>
              </Table.Tr>
            ))}
            {integrations.figma ? (
              <Table.Tr>
                <Table.Td>
                  <Flex align="center" className="gap-2">
                    <ImageIcon className="h-4" />
                    Figma
                  </Flex>
                </Table.Td>
                <Table.Td className="min-w-52">
                  {integrations.figma.accountLabel}
                </Table.Td>
                <Table.Td>
                  <IntegrationStatusBadge state={integrations.figma.state} />
                </Table.Td>
                <Table.Td className="min-w-56">
                  {integrations.figma.linkedFileCount} linked files ·{" "}
                  {integrations.figma.webhookCount} webhooks
                </Table.Td>
                <Table.Td className="min-w-56">
                  {integrations.figma.connectedByName ??
                    integrations.figma.connectedByEmail}
                </Table.Td>
                <Table.Td className="whitespace-nowrap">
                  Expires {formatDateTime(integrations.figma.expiresAt)}
                </Table.Td>
              </Table.Tr>
            ) : null}
            {!integrations.slack &&
            integrations.github.length === 0 &&
            integrations.calendar.length === 0 &&
            !integrations.figma ? (
              <Table.Tr>
                <Table.Td className="py-10 text-center" colSpan={6}>
                  <Text color="muted">
                    This workspace has no active integrations.
                  </Text>
                </Table.Td>
              </Table.Tr>
            ) : null}
          </Table.Body>
        </Table>
      </Box>
    </Box>

    <Box className="border-border overflow-hidden rounded-lg border-[0.5px]">
      <Box className="border-border border-b-[0.5px] px-4 py-3">
        <Text fontWeight="semibold">Member mappings</Text>
        <Text className="mt-1 text-[0.95rem]" color="muted">
          External identities associated with each current workspace member.
        </Text>
      </Box>
      <Box className="overflow-x-auto">
        <Table color="light" variant="bordered">
          <Table.Head>
            <Table.Tr>
              <Table.Th>Member</Table.Th>
              <Table.Th>Slack</Table.Th>
              <Table.Th>GitHub</Table.Th>
              <Table.Th>Calendar</Table.Th>
            </Table.Tr>
          </Table.Head>
          <Table.Body>
            {integrations.memberMappings.map((member) => (
              <Table.Tr key={member.userId}>
                <Table.Td className="min-w-72">
                  <Link
                    className="hover:text-primary"
                    href={`/users/${member.userId}`}
                  >
                    {member.name || member.email}
                  </Link>
                  <Text className="text-[0.9rem] capitalize" color="muted">
                    {member.role} · {member.email}
                  </Text>
                </Table.Td>
                <Table.Td className="min-w-48">
                  {member.slackUserId ? (
                    <>
                      <Text>{member.slackUserId}</Text>
                      <Text className="text-[0.9rem]" color="muted">
                        {humanizeKey(member.slackLinkedVia ?? "linked")} ·{" "}
                        {formatDateTime(member.slackLinkedAt)}
                      </Text>
                    </>
                  ) : (
                    <NotLinked />
                  )}
                </Table.Td>
                <Table.Td className="min-w-44">
                  {member.githubUsername ? (
                    <Text>@{member.githubUsername}</Text>
                  ) : (
                    <NotLinked />
                  )}
                </Table.Td>
                <Table.Td className="min-w-56">
                  {member.calendarEmail ? (
                    <>
                      <Text>{member.calendarEmail}</Text>
                      <Text className="text-[0.9rem] capitalize" color="muted">
                        {member.calendarProvider} ·{" "}
                        {humanizeKey(member.calendarState ?? "connected")}
                      </Text>
                    </>
                  ) : (
                    <NotLinked />
                  )}
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Body>
        </Table>
      </Box>
    </Box>
  </Box>
);

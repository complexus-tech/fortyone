import Link from "next/link";
import { HistoryIcon, InfoIcon } from "icons";
import { Box, Flex, Table, Text, Tooltip } from "ui";
import { getAuditLogs } from "@/lib/admin-api";
import { formatAuditValue, formatDateTime, humanizeKey } from "@/lib/format";
import { PageHeader } from "@/components/page-header";
import { PaginationControls } from "@/components/pagination-controls";
import { SearchToolbar } from "@/components/search-toolbar";

const targetTypeOptions = [
  { label: "All targets", value: "" },
  { label: "Workspace", value: "workspace" },
  { label: "User", value: "user" },
  { label: "Subscription", value: "subscription" },
  { label: "System", value: "system" },
];

export default async function AuditPage({
  searchParams,
}: {
  searchParams: Promise<{
    page?: string;
    targetType?: string;
    workspaceId?: string;
  }>;
}) {
  const params = await searchParams;
  const auditLogs = await getAuditLogs({
    page: params.page ?? 1,
    targetType: params.targetType,
    workspaceId: params.workspaceId,
    limit: 30,
  });

  return (
    <Box>
      <PageHeader
        description="Review internal administrative actions and the reason attached to each write."
        eyebrow="Governance"
        icon={<HistoryIcon className="h-[1.1rem]" />}
        title="Audit log"
      />

      <Box className="space-y-4 p-5 md:p-7">
        <SearchToolbar
          defaultFilter={params.targetType}
          defaultQuery={params.workspaceId}
          filterName="targetType"
          filterOptions={targetTypeOptions}
          inputName="workspaceId"
          pathname="/audit"
          placeholder="Search by workspace UUID"
        />

        <Box className="border-border bg-surface overflow-hidden rounded-lg border-[0.5px]">
          <Box className="overflow-x-auto">
            <Table color="light" variant="bordered">
              <Table.Head>
                <Table.Tr>
                  <Table.Th>Actor</Table.Th>
                  <Table.Th>Target</Table.Th>
                  <Table.Th>Workspace</Table.Th>
                  <Table.Th>Action</Table.Th>
                  <Table.Th>Change</Table.Th>
                  <Table.Th>Reason</Table.Th>
                  <Table.Th>Time</Table.Th>
                </Table.Tr>
              </Table.Head>
              <Table.Body>
                {auditLogs.items.length > 0 ? (
                  auditLogs.items.map((entry) => (
                    <Table.Tr key={entry.id}>
                      <Table.Td className="min-w-56 whitespace-nowrap">
                        <Text>{entry.actorName || "Unknown actor"}</Text>
                      </Table.Td>
                      <Table.Td className="min-w-40 whitespace-nowrap">
                        <Text className="capitalize">{entry.targetType}</Text>
                      </Table.Td>
                      <Table.Td className="min-w-72 whitespace-nowrap">
                        {entry.workspaceId ? (
                          <Link
                            className="hover:text-primary line-clamp-1"
                            href={`/workspaces/${entry.workspaceId}`}
                          >
                            {entry.workspaceName || entry.workspaceId}
                          </Link>
                        ) : (
                          <Text color="muted">Not available</Text>
                        )}
                      </Table.Td>
                      <Table.Td className="min-w-56 whitespace-nowrap">
                        <Text>
                          <Text as="span" fontWeight="semibold">
                            {humanizeKey(entry.action)}
                          </Text>
                        </Text>
                      </Table.Td>
                      <Table.Td className="min-w-56 whitespace-nowrap">
                        <Flex align="center" className="gap-2">
                          <Text fontWeight="semibold">
                            {entry.fieldName
                              ? humanizeKey(entry.fieldName)
                              : "Change"}
                          </Text>
                          <Tooltip
                            title={
                              <Box>
                                <Text className="text-[0.9rem]" color="muted">
                                  From
                                </Text>
                                <Text>{formatAuditValue(entry.oldValue)}</Text>
                                <Text
                                  className="mt-2 text-[0.9rem]"
                                  color="muted"
                                >
                                  To
                                </Text>
                                <Text>{formatAuditValue(entry.newValue)}</Text>
                              </Box>
                            }
                          >
                            <span className="text-text-muted inline-flex cursor-help">
                              <InfoIcon className="h-4" />
                            </span>
                          </Tooltip>
                        </Flex>
                      </Table.Td>
                      <Table.Td className="min-w-72">
                        <Text className="line-clamp-2">
                          {entry.reason || "No reason provided"}
                        </Text>
                      </Table.Td>
                      <Table.Td className="min-w-44">
                        {formatDateTime(entry.createdAt)}
                      </Table.Td>
                    </Table.Tr>
                  ))
                ) : (
                  <Table.Tr>
                    <Table.Td className="py-10 text-center" colSpan={7}>
                      <Text color="muted">
                        No audit entries match this filter.
                      </Text>
                    </Table.Td>
                  </Table.Tr>
                )}
              </Table.Body>
            </Table>
          </Box>
          <PaginationControls
            pagination={auditLogs.pagination}
            params={{
              targetType: params.targetType,
              workspaceId: params.workspaceId,
            }}
            pathname="/audit"
          />
        </Box>
      </Box>
    </Box>
  );
}

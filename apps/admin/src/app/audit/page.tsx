import Link from "next/link";
import { HistoryIcon, SearchIcon } from "icons";
import { Badge, Box, Button, Flex, Input, Table, Text } from "ui";
import { getAuditLogs } from "@/lib/admin-api";
import { formatAuditValue, formatDateTime, humanizeKey } from "@/lib/format";
import { PageHeader } from "@/components/page-header";
import { PaginationControls } from "@/components/pagination-controls";

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
        <form>
          <Flex align="end" className="flex-wrap gap-2">
            <label className="block w-full md:w-44 md:shrink-0">
              <Text className="mb-[0.35rem] block">Filters</Text>
              <select
                className="border-input bg-surface focus-visible:ring-ring h-[2.8rem] w-full rounded-lg border px-3 outline-none focus-visible:ring-2"
                defaultValue={params.targetType ?? ""}
                name="targetType"
              >
                {targetTypeOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
            <Box className="w-full min-w-0 md:w-80 md:shrink-0">
              <Input
                className="md:pl-10"
                defaultValue={params.workspaceId}
                label="Workspace ID"
                leftIcon={<SearchIcon className="h-4" />}
                name="workspaceId"
                placeholder="Optional workspace UUID"
              />
            </Box>
            <Button color="tertiary" type="submit">
              Apply
            </Button>
            {params.targetType || params.workspaceId ? (
              <Button color="tertiary" href="/audit" variant="naked">
                Clear
              </Button>
            ) : null}
          </Flex>
        </form>

        <Box className="border-border bg-surface overflow-hidden rounded-lg border-[0.5px]">
          <Box className="overflow-x-auto">
            <Table color="light" variant="bordered">
              <Table.Head>
                <Table.Tr>
                  <Table.Th>Action</Table.Th>
                  <Table.Th>Actor</Table.Th>
                  <Table.Th>Target</Table.Th>
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
                        <Text>
                          <Text as="span" fontWeight="semibold">
                            {humanizeKey(entry.action)}
                          </Text>
                        </Text>
                      </Table.Td>
                      <Table.Td className="min-w-56 whitespace-nowrap">
                        <Text>{entry.actorName || "Unknown actor"}</Text>
                      </Table.Td>
                      <Table.Td className="min-w-72 whitespace-nowrap">
                        <Flex align="center" className="gap-2">
                          <Badge color="tertiary">
                            {entry.targetType}
                          </Badge>
                          {entry.workspaceId ? (
                            <Link
                              className="hover:text-primary line-clamp-1"
                              href={`/workspaces/${entry.workspaceId}`}
                            >
                              {entry.workspaceName || entry.workspaceId}
                            </Link>
                          ) : null}
                        </Flex>
                      </Table.Td>
                      <Table.Td className="min-w-80 whitespace-nowrap">
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
                    <Table.Td className="py-10 text-center" colSpan={6}>
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

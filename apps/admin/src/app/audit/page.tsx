import Link from "next/link";
import { HistoryIcon } from "icons";
import { Badge, Box, Button, Flex, Input, Table, Text } from "ui";
import { getAuditLogs } from "@/lib/admin-api";
import { formatDateTime, formatValue, humanizeKey } from "@/lib/format";
import { MetricCard } from "@/components/metric-card";
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
        title="Audit log"
      />

      <Box className="space-y-4 p-5 md:p-7">
        <Box className="grid gap-3 md:grid-cols-3">
          <MetricCard
            detail="Rows matching the current filter"
            icon={<HistoryIcon />}
            label="Audit entries"
            value={String(auditLogs.pagination.total)}
          />
        </Box>

        <Box className="border-border bg-surface/70 rounded-xl border-[0.5px] p-3">
          <form>
            <Flex align="end" className="gap-2">
              <label className="block w-52 shrink-0">
                <Text className="mb-[0.35rem] block">Target type</Text>
                <select
                  className="border-input bg-surface focus-visible:ring-ring h-[2.8rem] w-full rounded-xl border px-3 outline-none focus-visible:ring-2"
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
              <Box className="min-w-0 flex-1">
                <Input
                  defaultValue={params.workspaceId}
                  label="Workspace ID"
                  name="workspaceId"
                  placeholder="Optional workspace UUID"
                  rounded="lg"
                />
              </Box>
              <Button color="tertiary" rounded="lg" type="submit">
                Apply
              </Button>
              {params.targetType || params.workspaceId ? (
                <Button
                  color="tertiary"
                  href="/audit"
                  rounded="lg"
                  variant="naked"
                >
                  Clear
                </Button>
              ) : null}
            </Flex>
          </form>
        </Box>

        <Box className="border-border bg-surface overflow-hidden rounded-xl border-[0.5px]">
          <Box className="overflow-x-auto">
            <Table>
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
                      <Table.Td className="min-w-48">
                        <Text fontWeight="semibold">
                          {humanizeKey(entry.action)}
                        </Text>
                        {entry.fieldName ? (
                          <Text className="mt-0.5 text-[0.92rem]" color="muted">
                            {entry.fieldName}
                          </Text>
                        ) : null}
                      </Table.Td>
                      <Table.Td className="min-w-48">
                        <Text>{entry.actorName || entry.actorEmail}</Text>
                        <Text className="mt-0.5 text-[0.92rem]" color="muted">
                          {entry.actorEmail}
                        </Text>
                      </Table.Td>
                      <Table.Td className="min-w-48">
                        <Badge color="tertiary" rounded="full" size="sm">
                          {entry.targetType}
                        </Badge>
                        {entry.workspaceId ? (
                          <Link
                            className="hover:text-primary mt-1 block"
                            href={`/workspaces/${entry.workspaceId}`}
                          >
                            {entry.workspaceName || entry.workspaceId}
                          </Link>
                        ) : null}
                      </Table.Td>
                      <Table.Td className="min-w-72">
                        <Text className="line-clamp-2">
                          {formatValue(entry.oldValue)} -&gt;{" "}
                          {formatValue(entry.newValue)}
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

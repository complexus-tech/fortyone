"use client";

import { useState } from "react";
import { ArrowDownIcon, PreferencesIcon } from "icons";
import { cn } from "lib";
import {
  Badge,
  Box,
  Button,
  Checkbox,
  Collapsible,
  Flex,
  Input,
  Select,
  Text,
} from "ui";
import { useJoinedTeams } from "@/modules/teams/public/client";
import { useTeamStatuses } from "@/lib/hooks/statuses";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { useCustomFieldReport, useTeamCustomFields } from "./hooks";
import type { CustomFieldReport, CustomFieldReportInput } from "./types";
import { formatExactDecimal } from "./value-utils";
import { CustomFieldIcon } from "./icons";

const AGGREGATIONS: {
  value: CustomFieldReport["aggregation"];
  label: string;
}[] = [
  { value: "sum", label: "Sum" },
  { value: "average", label: "Average" },
  { value: "min", label: "Minimum" },
  { value: "max", label: "Maximum" },
  { value: "count", label: "Count with a value" },
];
const GROUPS: { value: CustomFieldReport["groupBy"]; label: string }[] = [
  { value: "none", label: "All matching items" },
  { value: "status", label: "Status" },
  { value: "assignee", label: "Assignee" },
  { value: "month", label: "Month" },
];
const toggleId = (ids: string[], id: string) =>
  ids.includes(id) ? ids.filter((value) => value !== id) : [...ids, id];

const ReportResults = ({ report }: { report: CustomFieldReport }) => {
  const coverage = report.totalCount
    ? Math.round((report.valuedCount / report.totalCount) * 100)
    : 0;
  const currency = report.aggregation === "count" ? null : report.currency;
  return (
    <Box className="mt-6">
      <Flex align="center" className="mb-4 gap-3" justify="between" wrap>
        <Text as="h3" fontSize="lg" fontWeight="medium">
          {report.field.name}
        </Text>
        <Badge color="tertiary" variant="outline">
          {coverage}% value coverage
        </Badge>
      </Flex>
      <Flex className="mb-4 gap-x-5 gap-y-2" wrap>
        <Text color="muted">
          {report.totalCount.toLocaleString()} matching items
        </Text>
        <Text color="muted">
          {report.valuedCount.toLocaleString()} with a value
        </Text>
        <Text color="muted">
          {report.missingCount.toLocaleString()} missing a value
        </Text>
      </Flex>
      {report.rows.length ? (
        <Box className="border-border overflow-x-auto rounded-xl border">
          <table className="w-full text-left text-base">
            <thead className="bg-surface-muted">
              <tr>
                <th className="px-4 py-3 font-medium" scope="col">
                  {
                    GROUPS.find((group) => group.value === report.groupBy)
                      ?.label
                  }
                </th>
                <th className="px-4 py-3 text-right font-medium" scope="col">
                  {
                    AGGREGATIONS.find(
                      (aggregation) => aggregation.value === report.aggregation,
                    )?.label
                  }
                  {currency ? ` (${currency})` : ""}
                </th>
                <th className="px-4 py-3 text-right font-medium" scope="col">
                  With a value
                </th>
              </tr>
            </thead>
            <tbody className="divide-border divide-y">
              {report.rows.map((row) => (
                <tr key={row.key}>
                  <th className="px-4 py-4 font-normal" scope="row">
                    {row.label}
                  </th>
                  <td className="px-4 py-4 text-right font-medium whitespace-nowrap tabular-nums">
                    {row.value === "" ? (
                      <Text as="span" color="muted" fontWeight="normal">
                        No value
                      </Text>
                    ) : (
                      formatExactDecimal(row.value, currency)
                    )}
                  </td>
                  <td className="px-4 py-4 text-right tabular-nums">
                    {row.count.toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Box>
      ) : (
        <Box className="bg-surface-muted/40 rounded-xl p-6">
          <Text>No values match this report.</Text>
          <Text className="mt-2" color="muted">
            Adjust the filters or add values to the matching work.
          </Text>
        </Box>
      )}
      <Text className="mt-4 leading-6" color="muted">
        Missing values are excluded. Zero is included.{" "}
        {currency ? `Amounts use ${currency}.` : ""}
      </Text>
    </Box>
  );
};

export const CustomFieldReportPanel = () => {
  const {
    data: teams = [],
    isPending: teamsPending,
    isError: teamsError,
    refetch: refetchTeams,
  } = useJoinedTeams();
  const [selectedTeam, setSelectedTeam] = useState("");
  const teamId = selectedTeam || teams.at(0)?.id || "";
  const fieldsQuery = useTeamCustomFields(teamId);
  const statusesQuery = useTeamStatuses(teamId);
  const peopleQuery = useTeamMembers(teamId);
  const statuses = statusesQuery.data ?? [];
  const people = peopleQuery.data ?? [];
  const [selectedField, setSelectedField] = useState("");
  const [aggregation, setAggregation] =
    useState<CustomFieldReport["aggregation"]>("sum");
  const [groupBy, setGroupBy] =
    useState<CustomFieldReport["groupBy"]>("status");
  const [dateBasis, setDateBasis] = useState("created");
  const [startDate, setStartDate] = useState("");
  const [endDate, setEndDate] = useState("");
  const [statusIds, setStatusIds] = useState<string[]>([]);
  const [assigneeIds, setAssigneeIds] = useState<string[]>([]);
  const [validation, setValidation] = useState<string | null>(null);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const mutation = useCustomFieldReport();
  const fields = fieldsQuery.data ?? [];
  const field =
    fields.find((candidate) => candidate.id === selectedField) ??
    fields.find((candidate) => !candidate.archivedAt) ??
    fields.at(0);
  const numeric = field?.type === "number" || field?.type === "money";
  const input: CustomFieldReportInput = {
    fieldId: field?.id ?? "",
    aggregation: numeric ? aggregation : "count",
    groupBy,
    dateBasis,
    statusIds,
    assigneeIds,
    ...(startDate ? { startDate } : {}),
    ...(endDate ? { endDate } : {}),
  };
  const reportCurrent =
    mutation.variables &&
    JSON.stringify(mutation.variables) === JSON.stringify(input);
  const filterCount =
    statusIds.length +
    assigneeIds.length +
    Number(Boolean(startDate)) +
    Number(Boolean(endDate)) +
    Number(dateBasis !== "created");
  const clearFilters = () => {
    setStatusIds([]);
    setAssigneeIds([]);
    setStartDate("");
    setEndDate("");
    setDateBasis("created");
    setValidation(null);
  };
  const runReport = () => {
    setValidation(null);
    if (startDate && endDate && startDate > endDate) {
      setValidation("The end date must be on or after the start date.");
      setFiltersOpen(true);
      return;
    }
    mutation.mutate(input);
  };
  const changeTeam = (value: string) => {
    setSelectedTeam(value);
    setSelectedField("");
    setStatusIds([]);
    setAssigneeIds([]);
    setDateBasis("created");
    setValidation(null);
    mutation.reset();
  };
  return (
    <Box className="border-border bg-surface rounded-2xl border p-5 md:p-6">
      <Text as="h2" fontSize="lg" fontWeight="medium">
        Field reports
      </Text>
      <Text className="mt-2 max-w-3xl leading-6" color="muted">
        Totals and averages across matching work.
      </Text>
      {teamsPending ? (
        <Text className="mt-5" color="muted">
          Loading teams…
        </Text>
      ) : null}
      {teamsError ? (
        <Box className="mt-5">
          <Text color="danger" role="alert">
            Teams could not be loaded.
          </Text>
          <Button
            className="mt-2"
            color="tertiary"
            onClick={() => void refetchTeams()}
            variant="outline"
          >
            Try again
          </Button>
        </Box>
      ) : null}
      {!teamsPending && !teamsError && teams.length === 0 ? (
        <Text className="mt-5" color="muted">
          Join a team to report on its custom fields.
        </Text>
      ) : null}
      {!teamsPending && !teamsError && teams.length > 0 ? (
        <>
          <Box className="mt-6 grid gap-5 md:grid-cols-2 xl:grid-cols-4">
            <Box>
              <Text className="mb-2" fontWeight="medium">
                Team
              </Text>
              <Select
                disabled={mutation.isPending}
                onValueChange={changeTeam}
                value={teamId}
              >
                <Select.Trigger
                  aria-label="Report team"
                  className="h-11 w-full text-base"
                >
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  {teams.map((team) => (
                    <Select.Option
                      className="text-base"
                      key={team.id}
                      value={team.id}
                    >
                      {team.name}
                    </Select.Option>
                  ))}
                </Select.Content>
              </Select>
            </Box>
            <Box>
              <Text className="mb-2" fontWeight="medium">
                Field
              </Text>
              <Select
                disabled={mutation.isPending || !field}
                onValueChange={setSelectedField}
                value={field?.id ?? ""}
              >
                <Select.Trigger
                  aria-label="Report field"
                  className="h-11 w-full text-base"
                >
                  <Select.Input placeholder="Choose a field" />
                </Select.Trigger>
                <Select.Content>
                  {fields.map((candidate) => (
                    <Select.Option
                      className="text-base"
                      key={candidate.id}
                      value={candidate.id}
                    >
                      <span className="flex items-center gap-2">
                        <CustomFieldIcon
                          className="h-4 w-auto shrink-0"
                          field={candidate}
                        />
                        {candidate.name}
                        {candidate.currency ? ` (${candidate.currency})` : ""}
                        {candidate.archivedAt ? " (archived)" : ""}
                      </span>
                    </Select.Option>
                  ))}
                </Select.Content>
              </Select>
            </Box>
            <Box>
              <Text className="mb-2" fontWeight="medium">
                Measure
              </Text>
              <Select
                disabled={mutation.isPending || !field}
                onValueChange={(value) => {
                  setAggregation(value as CustomFieldReport["aggregation"]);
                }}
                value={input.aggregation}
              >
                <Select.Trigger
                  aria-label="Report measure"
                  className="h-11 w-full text-base"
                >
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  {AGGREGATIONS.filter(
                    (item) => numeric || item.value === "count",
                  ).map((item) => (
                    <Select.Option
                      className="text-base"
                      key={item.value}
                      value={item.value}
                    >
                      {item.label}
                    </Select.Option>
                  ))}
                </Select.Content>
              </Select>
            </Box>
            <Box>
              <Text className="mb-2" fontWeight="medium">
                Group by
              </Text>
              <Select
                disabled={mutation.isPending}
                onValueChange={(value) => {
                  setGroupBy(value as CustomFieldReport["groupBy"]);
                }}
                value={groupBy}
              >
                <Select.Trigger
                  aria-label="Report grouping"
                  className="h-11 w-full text-base"
                >
                  <Select.Input />
                </Select.Trigger>
                <Select.Content>
                  {GROUPS.map((group) => (
                    <Select.Option
                      className="text-base"
                      key={group.value}
                      value={group.value}
                    >
                      {group.label}
                    </Select.Option>
                  ))}
                </Select.Content>
              </Select>
            </Box>
          </Box>
          <Collapsible
            className="mt-5"
            onOpenChange={setFiltersOpen}
            open={filtersOpen}
          >
            <Flex align="center" gap={3}>
              <Collapsible.Trigger asChild>
                <Button
                  color="tertiary"
                  leftIcon={<PreferencesIcon className="h-4 w-auto" />}
                  rightIcon={
                    <ArrowDownIcon
                      className={cn("h-3.5 w-auto transition-transform", {
                        "rotate-180": filtersOpen,
                      })}
                    />
                  }
                  size="sm"
                  variant="outline"
                >
                  Filters{filterCount ? ` (${filterCount})` : ""}
                </Button>
              </Collapsible.Trigger>
              {filterCount ? (
                <Button
                  color="tertiary"
                  disabled={mutation.isPending}
                  onClick={clearFilters}
                  size="sm"
                  variant="naked"
                >
                  Clear filters
                </Button>
              ) : null}
            </Flex>
            <Collapsible.Content className="border-border mt-4 border-t pt-5">
              <Box className="grid gap-5 md:grid-cols-3">
                <Box>
                  <Text className="mb-2" fontWeight="medium">
                    Date basis
                  </Text>
                  <Select
                    disabled={mutation.isPending}
                    onValueChange={setDateBasis}
                    value={dateBasis}
                  >
                    <Select.Trigger
                      aria-label="Report date basis"
                      className="h-11 w-full text-base"
                    >
                      <Select.Input />
                    </Select.Trigger>
                    <Select.Content>
                      <Select.Option className="text-base" value="created">
                        Created date
                      </Select.Option>
                      <Select.Option className="text-base" value="completed">
                        Completed date
                      </Select.Option>
                      {fields
                        .filter((candidate) => candidate.type === "date")
                        .map((candidate) => (
                          <Select.Option
                            className="text-base"
                            key={candidate.id}
                            value={candidate.id}
                          >
                            {candidate.name}
                            {candidate.currency
                              ? ` (${candidate.currency})`
                              : ""}
                            {candidate.archivedAt ? " (archived)" : ""}
                          </Select.Option>
                        ))}
                    </Select.Content>
                  </Select>
                </Box>
                <Input
                  className="h-11 text-base"
                  disabled={mutation.isPending}
                  label="From"
                  labelClassName="mb-2 font-medium"
                  onChange={(event) => {
                    setStartDate(event.target.value);
                  }}
                  type="date"
                  value={startDate}
                />
                <Input
                  className="h-11 text-base"
                  disabled={mutation.isPending}
                  label="Through"
                  labelClassName="mb-2 font-medium"
                  min={startDate || undefined}
                  onChange={(event) => {
                    setEndDate(event.target.value);
                  }}
                  type="date"
                  value={endDate}
                />
              </Box>
              <Box className="mt-4 grid gap-5 md:grid-cols-2">
                <Box>
                  <Text className="mb-3" fontWeight="medium">
                    Statuses
                  </Text>
                  {statusesQuery.isPending ? (
                    <Text color="muted" role="status">
                      Loading statuses…
                    </Text>
                  ) : null}
                  {statusesQuery.isError ? (
                    <Box>
                      <Text color="danger" role="alert">
                        Statuses could not be loaded.
                      </Text>
                      <Button
                        className="mt-2"
                        color="tertiary"
                        onClick={() => void statusesQuery.refetch()}
                        variant="outline"
                      >
                        Retry statuses
                      </Button>
                    </Box>
                  ) : null}
                  {!statusesQuery.isPending &&
                  !statusesQuery.isError &&
                  !statuses.length ? (
                    <Text color="muted">This team has no statuses.</Text>
                  ) : null}
                  <Box className="flex flex-wrap gap-x-5 gap-y-3">
                    {statuses.map((status) => (
                      <label
                        className="flex items-center gap-2"
                        key={status.id}
                      >
                        <Checkbox
                          checked={statusIds.includes(status.id)}
                          disabled={
                            mutation.isPending ||
                            statusesQuery.isError ||
                            statusesQuery.isPending
                          }
                          onCheckedChange={() => {
                            setStatusIds((current) =>
                              toggleId(current, status.id),
                            );
                          }}
                        />
                        {status.name}
                      </label>
                    ))}
                  </Box>
                </Box>
                <Box>
                  <Text className="mb-3" fontWeight="medium">
                    People
                  </Text>
                  {peopleQuery.isPending ? (
                    <Text color="muted" role="status">
                      Loading people…
                    </Text>
                  ) : null}
                  {peopleQuery.isError ? (
                    <Box>
                      <Text color="danger" role="alert">
                        People could not be loaded.
                      </Text>
                      <Button
                        className="mt-2"
                        color="tertiary"
                        onClick={() => void peopleQuery.refetch()}
                        variant="outline"
                      >
                        Retry people
                      </Button>
                    </Box>
                  ) : null}
                  {!peopleQuery.isPending &&
                  !peopleQuery.isError &&
                  !people.some(
                    (person) => person.isActive && !person.isSystem,
                  ) ? (
                    <Text color="muted">This team has no active people.</Text>
                  ) : null}
                  <Box className="flex max-h-48 flex-wrap gap-x-5 gap-y-3 overflow-y-auto">
                    {people
                      .filter((person) => person.isActive && !person.isSystem)
                      .map((person) => (
                        <label
                          className="flex items-center gap-2"
                          key={person.id}
                        >
                          <Checkbox
                            checked={assigneeIds.includes(person.id)}
                            disabled={
                              mutation.isPending ||
                              peopleQuery.isError ||
                              peopleQuery.isPending
                            }
                            onCheckedChange={() => {
                              setAssigneeIds((current) =>
                                toggleId(current, person.id),
                              );
                            }}
                          />
                          {person.fullName || person.username}
                        </label>
                      ))}
                  </Box>
                </Box>
              </Box>
              <Text className="mt-4 leading-6" color="muted">
                Leave filters empty to include all work.
              </Text>
            </Collapsible.Content>
          </Collapsible>
          {fieldsQuery.isError ? (
            <Box className="mt-4">
              <Text color="danger" role="alert">
                Fields could not be loaded.
              </Text>
              <Button
                className="mt-2"
                color="tertiary"
                onClick={() => void fieldsQuery.refetch()}
                variant="outline"
              >
                Try again
              </Button>
            </Box>
          ) : null}
          {fieldsQuery.isPending ? (
            <Text className="mt-4" color="muted">
              Loading fields…
            </Text>
          ) : null}
          {!fieldsQuery.isPending &&
          !fieldsQuery.isError &&
          fields.length === 0 ? (
            <Text className="mt-4 leading-6" color="muted">
              Create custom fields in this team’s settings to start reporting.
            </Text>
          ) : null}
          {validation || mutation.error ? (
            <Text className="mt-4" color="danger" role="alert">
              {validation ?? mutation.error?.message}
            </Text>
          ) : null}
          <Flex align="center" className="mt-5 gap-3" wrap>
            <Button
              disabled={!field || fieldsQuery.isError || mutation.isPending}
              onClick={runReport}
            >
              {mutation.isPending ? "Building report…" : "Build report"}
            </Button>
            {mutation.data && !reportCurrent ? (
              <Text color="muted">
                Build the report to apply these changes.
              </Text>
            ) : null}
          </Flex>
          {mutation.data && reportCurrent ? (
            <ReportResults report={mutation.data} />
          ) : null}
        </>
      ) : null}
    </Box>
  );
};

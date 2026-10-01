"use client";

import { useState } from "react";
import { z } from "zod";
import { Box, Button, Flex, Input, Select, Text } from "ui";
import { useMembers } from "@/lib/hooks/members";
import { SectionHeader } from "@/modules/settings/components/section-header";
import type { AuditFilters } from "./types";
import { useDownloadAudit, useSecurityAudit } from "./hooks";

export const SecurityAudit = () => {
  const [actorId, setActorId] = useState("");
  const [resourceType, setResourceType] = useState("");
  const [resourceId, setResourceId] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [filters, setFilters] = useState<AuditFilters>({});
  const [filterError, setFilterError] = useState("");
  const { data: members = [] } = useMembers();
  const names = new Map(
    members.map((member) => [member.id, member.fullName || member.email]),
  );
  const query = useSecurityAudit(filters);
  const download = useDownloadAudit();
  const events = query.data?.pages.flatMap((page) => page.items) ?? [];
  const applyFilters = () => {
    if (
      (resourceId && !z.string().uuid().safeParse(resourceId).success) ||
      (from && to && from > to)
    ) {
      setFilterError(
        "Use a valid resource ID and an end date on or after the start date.",
      );
      return;
    }
    const end = to ? new Date(`${to}T00:00:00Z`) : null;
    end?.setUTCDate(end.getUTCDate() + 1);
    setFilterError("");
    setFilters({
      ...(actorId ? { actorId } : {}),
      ...(resourceType.trim() ? { resourceType: resourceType.trim() } : {}),
      ...(resourceId ? { resourceId } : {}),
      ...(from ? { from: new Date(`${from}T00:00:00Z`).toISOString() } : {}),
      ...(end ? { to: end.toISOString() } : {}),
    });
  };
  return (
    <Box className="border-border bg-surface overflow-hidden rounded-2xl border">
      <SectionHeader
        description="Trace access and administrative changes."
        title="Workspace audit log"
      />
      <form
        className="space-y-4 px-6 pt-5 pb-5"
        onSubmit={(event) => {
          event.preventDefault();
          applyFilters();
        }}
      >
        <Box className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          <Box>
            <label className="mb-2 block" htmlFor="security-audit-actor">
              Actor
            </label>
            <Select
              onValueChange={(value) => {
                setActorId(value === "all" ? "" : value);
              }}
              value={actorId || "all"}
            >
              <Select.Trigger
                className="h-11 text-base"
                id="security-audit-actor"
              >
                <Select.Input />
              </Select.Trigger>
              <Select.Content>
                <Select.Option className="text-base" value="all">
                  All actors
                </Select.Option>
                {members.map((member) => (
                  <Select.Option
                    className="text-base"
                    key={member.id}
                    value={member.id}
                  >
                    {member.fullName || member.email}
                  </Select.Option>
                ))}
              </Select.Content>
            </Select>
          </Box>
          <Input
            label="Resource type"
            maxLength={64}
            onChange={(event) => {
              setResourceType(event.target.value);
            }}
            placeholder="workspace, story, browser_session"
            value={resourceType}
          />
          <Input
            label="Resource ID"
            onChange={(event) => {
              setResourceId(event.target.value);
            }}
            placeholder="Optional UUID"
            value={resourceId}
          />
          <Input
            label="From date (UTC)"
            onChange={(event) => {
              setFrom(event.target.value);
            }}
            type="date"
            value={from}
          />
          <Input
            label="Through date (UTC)"
            onChange={(event) => {
              setTo(event.target.value);
            }}
            type="date"
            value={to}
          />
        </Box>
        <Flex className="gap-3" wrap>
          <Button type="submit">Apply filters</Button>
          <Button
            color="tertiary"
            disabled={download.isPending || query.isPending || query.isError}
            loading={download.isPending}
            onClick={() => {
              download.mutate(filters);
            }}
            type="button"
            variant="outline"
          >
            {download.isPending ? "Exporting audit…" : "Export filtered CSV"}
          </Button>
        </Flex>
        {filterError ? (
          <Text color="danger" role="alert">
            {filterError}
          </Text>
        ) : null}
        {download.isError ? (
          <Text color="danger" role="alert">
            {download.error.message}
          </Text>
        ) : null}
      </form>
      {query.isPending ? (
        <Text aria-live="polite" className="px-6 pt-5 pb-6" color="muted">
          Loading audit events…
        </Text>
      ) : null}
      {query.isError ? (
        <Box className="px-6 pt-5 pb-6">
          <Text color="danger" role="alert">
            Audit events could not be loaded.
          </Text>
          <Button
            className="mt-3"
            color="tertiary"
            onClick={() => void query.refetch()}
            variant="outline"
          >
            Try again
          </Button>
        </Box>
      ) : null}
      {!query.isPending && !query.isError && events.length === 0 ? (
        <Text className="px-6 pt-5 pb-6" color="muted">
          No events match these filters.
        </Text>
      ) : null}
      <Box className="divide-border border-border divide-y border-t">
        {events.map((event) => (
          <Box className="space-y-2 p-5" key={`${event.source}:${event.id}`}>
            <Flex align="start" className="gap-3" justify="between" wrap>
              <Text className="break-words" fontWeight="medium">
                {event.operation.replaceAll("_", " ")}
              </Text>
              <Text color="muted">
                {new Date(event.createdAt).toLocaleString()}
              </Text>
            </Flex>
            <Text className="break-all" color="muted">
              {event.actorId
                ? names.get(event.actorId) ?? event.actorId
                : event.actorType}{" "}
              · {event.resourceType}
              {event.resourceId ? ` · ${event.resourceId}` : ""}
            </Text>
            <details className="text-text-muted">
              <summary className="cursor-pointer">View event details</summary>
              <pre className="bg-surface-elevated mt-3 max-h-72 overflow-auto rounded-lg p-4 font-mono text-base break-words whitespace-pre-wrap">
                {JSON.stringify(event.metadata, null, 2)}
              </pre>
            </details>
          </Box>
        ))}
      </Box>
      {query.hasNextPage ? (
        <Flex className="p-5" justify="center">
          <Button
            color="tertiary"
            disabled={query.isFetchingNextPage}
            loading={query.isFetchingNextPage}
            onClick={() => void query.fetchNextPage()}
            variant="outline"
          >
            Load more events
          </Button>
        </Flex>
      ) : null}
    </Box>
  );
};

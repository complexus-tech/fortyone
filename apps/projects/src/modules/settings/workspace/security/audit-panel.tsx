"use client";

import { useId, useState } from "react";
import { z } from "zod";
import { Box, Button, Flex, Popover, Select, Text } from "ui";
import {
  ArrowDownIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  FilterIcon,
} from "icons";
import { useMembers } from "@/lib/hooks/members";
import { SectionHeader } from "@/modules/settings/components/section-header";
import type { AuditFilters } from "./types";
import { useDownloadAudit, useSecurityAudit } from "./hooks";
import { SecurityInput } from "./security-input";
import { AuditEvents } from "./audit-events";

export const SecurityAudit = () => {
  const actorInputId = useId();
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [actorId, setActorId] = useState("");
  const [resourceType, setResourceType] = useState("");
  const [resourceId, setResourceId] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [filters, setFilters] = useState<AuditFilters>({});
  const [pageCursors, setPageCursors] = useState([""]);
  const [filterError, setFilterError] = useState("");
  const { data: members = [] } = useMembers();
  const names = new Map(
    members.map((member) => [
      member.id,
      member.username || member.fullName || member.email,
    ]),
  );
  const query = useSecurityAudit(filters, pageCursors[pageCursors.length - 1]);
  const download = useDownloadAudit();
  const events = query.data?.items ?? [];
  const filtersCount = Object.keys(filters).length;
  const clearFilters = () => {
    setActorId("");
    setResourceType("");
    setResourceId("");
    setFrom("");
    setTo("");
    setFilterError("");
    setFilters({});
    setPageCursors([""]);
    setFiltersOpen(false);
  };
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
    setPageCursors([""]);
    setFilters({
      ...(actorId ? { actorId } : {}),
      ...(resourceType.trim() ? { resourceType: resourceType.trim() } : {}),
      ...(resourceId ? { resourceId } : {}),
      ...(from ? { from: new Date(`${from}T00:00:00Z`).toISOString() } : {}),
      ...(end ? { to: end.toISOString() } : {}),
    });
    setFiltersOpen(false);
  };
  return (
    <Box className="border-border bg-surface overflow-hidden rounded-2xl border">
      <SectionHeader
        action={
          <Flex align="center" gap={2} wrap>
            <Button
              color="tertiary"
              disabled={download.isPending || query.isPending || query.isError}
              loading={download.isPending}
              loadingText="Exporting audit…"
              onClick={() => {
                download.mutate(filters);
              }}
              size="sm"
              variant="outline"
            >
              Export filtered CSV
            </Button>
            <Popover onOpenChange={setFiltersOpen} open={filtersOpen}>
              <Popover.Trigger asChild>
                <Button
                  aria-label={
                    filtersCount
                      ? `Filters, ${filtersCount} applied`
                      : "Filters"
                  }
                  color="tertiary"
                  leftIcon={<FilterIcon aria-hidden className="h-4 w-auto" />}
                  rightIcon={
                    <ArrowDownIcon aria-hidden className="h-3.5 w-auto" />
                  }
                  size="sm"
                  variant="outline"
                >
                  Filters{filtersCount ? ` (${filtersCount})` : ""}
                </Button>
              </Popover.Trigger>
              <Popover.Content
                align="end"
                aria-label="Audit log filters"
                className="max-h-[calc(var(--radix-popover-content-available-height)-1rem)] w-[min(32rem,calc(100vw-2rem))] overflow-y-auto px-4 py-4"
              >
                <form
                  className="space-y-4"
                  onSubmit={(event) => {
                    event.preventDefault();
                    applyFilters();
                  }}
                >
                  <Text as="h4" fontWeight="medium">
                    Audit log filters
                  </Text>
                  <Box className="grid items-end gap-4 sm:grid-cols-2">
                    <Box className="min-w-0">
                      <label
                        className="mb-[0.35rem] block"
                        htmlFor={actorInputId}
                      >
                        Actor
                      </label>
                      <Select
                        onValueChange={(value) => {
                          setActorId(value === "all" ? "" : value);
                        }}
                        value={actorId || "all"}
                      >
                        <Select.Trigger className="text-base" id={actorInputId}>
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
                              {member.username ||
                                member.fullName ||
                                member.email}
                            </Select.Option>
                          ))}
                        </Select.Content>
                      </Select>
                    </Box>
                    <SecurityInput
                      label="Resource type"
                      maxLength={64}
                      onChange={(event) => {
                        setResourceType(event.target.value);
                      }}
                      placeholder="workspace, story, browser_session"
                      value={resourceType}
                    />
                    <SecurityInput
                      label="Resource ID"
                      onChange={(event) => {
                        setResourceId(event.target.value);
                      }}
                      placeholder="Optional UUID"
                      value={resourceId}
                    />
                    <SecurityInput
                      label="From date (UTC)"
                      onChange={(event) => {
                        setFrom(event.target.value);
                      }}
                      type="date"
                      value={from}
                    />
                    <SecurityInput
                      label="Through date (UTC)"
                      onChange={(event) => {
                        setTo(event.target.value);
                      }}
                      type="date"
                      value={to}
                    />
                  </Box>
                  {filterError ? (
                    <Text color="danger" role="alert">
                      {filterError}
                    </Text>
                  ) : null}
                  <Flex className="gap-3" wrap>
                    <Button size="sm" type="submit">
                      Apply filters
                    </Button>
                    <Button
                      color="tertiary"
                      onClick={clearFilters}
                      size="sm"
                      type="button"
                      variant="naked"
                    >
                      Clear filters
                    </Button>
                  </Flex>
                </form>
              </Popover.Content>
            </Popover>
          </Flex>
        }
        description="Review access and administrative changes, newest first. Times are shown in UTC."
        title="Workspace audit log"
      />
      {download.isError ? (
        <Text className="px-6 py-4" color="danger" role="alert">
          {download.error.message}
        </Text>
      ) : null}
      {query.isPending ? (
        <Text aria-live="polite" className="px-6 py-4" color="muted">
          Loading audit events…
        </Text>
      ) : null}
      {query.isError ? (
        <Box className="px-6 py-4">
          <Text color="danger" role="alert">
            Audit events could not be loaded.
          </Text>
          <Flex className="mt-3 gap-3" wrap>
            <Button
              color="tertiary"
              onClick={() => void query.refetch()}
              variant="outline"
            >
              Try again
            </Button>
            {pageCursors.length > 1 ? (
              <Button
                color="tertiary"
                onClick={() => {
                  setPageCursors([""]);
                }}
                variant="naked"
              >
                Back to newest events
              </Button>
            ) : null}
          </Flex>
        </Box>
      ) : null}
      {!query.isPending && !query.isError && events.length === 0 ? (
        <Text className="px-6 py-4" color="muted">
          No events match these filters.
        </Text>
      ) : null}
      {events.length > 0 ? <AuditEvents events={events} names={names} /> : null}
      {query.data || pageCursors.length > 1 ? (
        <Flex
          align="center"
          className="border-border gap-3 border-t-[0.5px] px-6 py-4"
          justify="between"
          wrap
        >
          <Text aria-live="polite" color="muted">
            Page {pageCursors.length}
            {query.data
              ? ` · ${events.length} ${events.length === 1 ? "event" : "events"}`
              : ""}
            {query.isFetching && !query.isPending ? " · Updating…" : ""}
          </Text>
          <Flex align="center" gap={2}>
            <Button
              aria-label="Previous page"
              color="tertiary"
              disabled={pageCursors.length === 1 || query.isFetching}
              leftIcon={<ChevronLeftIcon className="h-4 w-auto" />}
              onClick={() => {
                setPageCursors((current) =>
                  current.length > 1 ? current.slice(0, -1) : current,
                );
              }}
              variant="outline"
            >
              Previous
            </Button>
            <Button
              aria-label="Next page"
              color="tertiary"
              disabled={!query.data?.nextCursor || query.isFetching}
              onClick={() => {
                const cursor = query.data?.nextCursor;
                if (cursor)
                  setPageCursors((current) =>
                    current[current.length - 1] === cursor
                      ? current
                      : [...current, cursor],
                  );
              }}
              rightIcon={<ChevronRightIcon className="h-4 w-auto" />}
              variant="outline"
            >
              Next
            </Button>
          </Flex>
        </Flex>
      ) : null}
    </Box>
  );
};

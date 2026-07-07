"use client";

import Link from "next/link";
import { DownloadIcon, FilterIcon, SearchIcon } from "icons";
import { Box, Button, Flex, Input, Menu } from "ui";
import { getAuditLogExportUrl } from "@/lib/admin-api";

type FilterOption = {
  label: string;
  value: string;
};

const buildHref = ({
  action,
  actor,
  from,
  pathname,
  query,
  targetType,
  to,
  workspaceId,
}: {
  action?: string;
  actor?: string;
  from?: string;
  pathname: string;
  query?: string;
  targetType?: string;
  to?: string;
  workspaceId?: string;
}) => {
  const params = new URLSearchParams();

  if (query) params.set("q", query);
  if (targetType) params.set("targetType", targetType);
  if (action) params.set("action", action);
  if (actor) params.set("actor", actor);
  if (from) params.set("from", from);
  if (to) params.set("to", to);
  if (workspaceId) params.set("workspaceId", workspaceId);

  const search = params.toString();
  return search ? `${pathname}?${search}` : pathname;
};

export const AuditFilterToolbar = ({
  action,
  actor,
  from,
  query,
  targetType,
  targetTypeOptions,
  to,
  workspaceId,
}: {
  action?: string;
  actor?: string;
  from?: string;
  query?: string;
  targetType?: string;
  targetTypeOptions: FilterOption[];
  to?: string;
  workspaceId?: string;
}) => {
  const hasFilters = Boolean(
    query || targetType || action || actor || from || to || workspaceId,
  );

  return (
    <form action="/audit">
      <Flex align="center" className="flex-wrap gap-2">
        <Box className="w-full min-w-0 md:w-[30rem] md:shrink-0">
          <Input
            className="md:pl-10"
            defaultValue={query}
            leftIcon={<SearchIcon className="h-4" />}
            name="q"
            placeholder="Search action, reason, workspace, or target ID"
          />
          {targetType ? (
            <input name="targetType" type="hidden" value={targetType} />
          ) : null}
          {workspaceId ? (
            <input name="workspaceId" type="hidden" value={workspaceId} />
          ) : null}
        </Box>
        <Box className="w-full md:w-56">
          <Input defaultValue={actor} name="actor" placeholder="Actor" />
        </Box>
        <Box className="w-full md:w-64">
          <Input defaultValue={action} name="action" placeholder="Action" />
        </Box>
        <Box className="w-[10.5rem]">
          <Input defaultValue={from} name="from" type="date" />
        </Box>
        <Box className="w-[10.5rem]">
          <Input defaultValue={to} name="to" type="date" />
        </Box>
        <Menu>
          <Menu.Button>
            <Button active={Boolean(targetType)} color="tertiary" type="button">
              <FilterIcon className="h-4" />
              Filters
            </Button>
          </Menu.Button>
          <Menu.Items align="start" className="w-52">
            <Menu.Group>
              {targetTypeOptions.map((option) => (
                <Menu.Item
                  active={(targetType ?? "") === option.value}
                  asChild
                  key={option.value}
                >
                  <Link
                    href={buildHref({
                      action,
                      actor,
                      from,
                      pathname: "/audit",
                      query,
                      targetType: option.value,
                      to,
                      workspaceId,
                    })}
                  >
                    {option.label}
                  </Link>
                </Menu.Item>
              ))}
            </Menu.Group>
          </Menu.Items>
        </Menu>
        <Button
          color="tertiary"
          href={getAuditLogExportUrl({
            action,
            actor,
            from,
            q: query,
            targetType,
            to,
            workspaceId,
          })}
          prefetch={false}
          size="sm"
          target="_blank"
        >
          <DownloadIcon className="h-4" />
          Export
        </Button>
        {hasFilters ? (
          <Button color="tertiary" href="/audit" variant="naked">
            Clear
          </Button>
        ) : null}
      </Flex>
    </form>
  );
};

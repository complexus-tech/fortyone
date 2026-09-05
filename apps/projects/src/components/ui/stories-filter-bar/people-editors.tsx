import { useState, type ComponentProps, type UIEvent } from "react";
import { Avatar, Box, Command, Divider, Flex, Text } from "ui";
import { CheckIcon } from "icons";
import { MenuLoadingSkeleton } from "../menu-loading-skeleton";
import { PriorityIcon } from "../priority-icon";
import { StoryStatusIcon } from "../story-status-icon";
import { normalizeArrayFilter, shouldFetchNextPage } from "./filter-model";
import type { StoriesFilterEditorProps } from "./types";

type StoryPriority = NonNullable<
  ComponentProps<typeof PriorityIcon>["priority"]
>;

export type FilterStatusOption = {
  id: string;
  name: string;
};

export type FilterMemberOption = {
  avatarUrl: string | null;
  email: string;
  fullName: string;
  id: string;
  username: string;
};

const PRIORITIES: readonly StoryPriority[] = [
  "Urgent",
  "High",
  "Medium",
  "Low",
  "No Priority",
];

export const StatusEditor = ({
  filters,
  setFilters,
  statuses,
}: StoriesFilterEditorProps & { statuses: FilterStatusOption[] }) => {
  const [query, setQuery] = useState("");
  const filteredStatuses = statuses.filter((status) =>
    status.name.toLowerCase().includes(query.toLowerCase()),
  );

  const toggleStatus = (statusId: string) => {
    const selected = filters.statusIds ?? [];
    const statusIds = selected.includes(statusId)
      ? selected.filter((id) => id !== statusId)
      : [...selected, statusId];
    setFilters({ ...filters, statusIds: normalizeArrayFilter(statusIds) });
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={setQuery}
        placeholder="Search status..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.List className="w-full border-0 bg-transparent p-0 shadow-none backdrop-blur-none dark:bg-transparent">
        <Command.Empty className="px-3 py-2 text-left text-base">
          <Text color="muted">No statuses found.</Text>
        </Command.Empty>
        <Command.Group className="max-h-80 overflow-y-auto">
          {filteredStatuses.map((status, idx) => (
            <Command.Item
              active={Boolean(filters.statusIds?.includes(status.id))}
              className="justify-between gap-4"
              key={status.id}
              onSelect={() => {
                toggleStatus(status.id);
              }}
              value={status.name}
            >
              <Box className="grid min-w-0 flex-1 grid-cols-[16px_minmax(0,1fr)] items-center">
                <span className="min-w-0">
                  <StoryStatusIcon statusId={status.id} />
                </span>
                <Text className="max-w-[22ch] truncate">{status.name}</Text>
              </Box>
              <Flex align="center" className="shrink-0" gap={2}>
                {filters.statusIds?.includes(status.id) ? (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                ) : null}
                <Text color="muted">{idx}</Text>
              </Flex>
            </Command.Item>
          ))}
        </Command.Group>
      </Command.List>
    </Command>
  );
};

export const PeopleEditor = ({
  field,
  filters,
  hasNextPage,
  isFetchingNextPage,
  isLoading,
  members,
  onFetchNextPage,
  onQueryChange,
  query,
  setFilters,
}: StoriesFilterEditorProps & {
  field: "assigneeIds" | "reporterIds";
  hasNextPage: boolean;
  isFetchingNextPage: boolean;
  isLoading: boolean;
  members: FilterMemberOption[];
  onFetchNextPage: () => void;
  onQueryChange: (query: string) => void;
  query: string;
}) => {
  const toggleMember = (memberId: string) => {
    const selected = filters[field] ?? [];
    const memberIds = selected.includes(memberId)
      ? selected.filter((id) => id !== memberId)
      : [...selected, memberId];
    setFilters({ ...filters, [field]: normalizeArrayFilter(memberIds) });
  };

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    if (
      shouldFetchNextPage(event.currentTarget, hasNextPage, isFetchingNextPage)
    ) {
      onFetchNextPage();
    }
  };

  return (
    <Command>
      <Command.Input
        autoFocus
        onValueChange={onQueryChange}
        placeholder="Search people..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.List className="w-full border-0 bg-transparent p-0 shadow-none backdrop-blur-none dark:bg-transparent">
        {!isLoading ? (
          <Command.Empty className="px-3 py-2 text-left text-base">
            <Text color="muted">No user found.</Text>
          </Command.Empty>
        ) : null}
        <Command.Group
          className="max-h-80 overflow-y-auto md:max-h-100"
          onScroll={handleScroll}
        >
          {isLoading ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton avatar rows={5} />
            </Command.Loading>
          ) : null}
          {field === "assigneeIds" ? (
            <Command.Item
              active={Boolean(filters.hasNoAssignee)}
              className="justify-between gap-4"
              onSelect={() => {
                setFilters({
                  ...filters,
                  assigneeIds: null,
                  hasNoAssignee: filters.hasNoAssignee ? null : true,
                });
              }}
              value="No assignee"
            >
              <Flex align="center" className="min-w-0 flex-1" gap={2}>
                <Avatar
                  className="text-foreground/80"
                  color="primary"
                  size="sm"
                />
                <Text className="max-w-48 truncate">No assignee</Text>
              </Flex>
              <Flex align="center" className="shrink-0" gap={1}>
                {filters.hasNoAssignee ? (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                ) : null}
                <Text color="muted">0</Text>
              </Flex>
            </Command.Item>
          ) : null}
          {members.map((member, idx) => {
            const name = member.fullName || member.username || member.email;
            return (
              <Command.Item
                active={Boolean(filters[field]?.includes(member.id))}
                className="justify-between gap-4"
                key={member.id}
                onSelect={() => {
                  toggleMember(member.id);
                }}
                value={name}
              >
                <Flex align="center" className="min-w-0 flex-1" gap={2}>
                  <Avatar
                    color="primary"
                    name={name}
                    size="sm"
                    src={member.avatarUrl}
                  />
                  <Text className="max-w-48 truncate">{name}</Text>
                </Flex>
                <Flex align="center" className="shrink-0" gap={1}>
                  {filters[field]?.includes(member.id) ? (
                    <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                  ) : null}
                  <Text color="muted">
                    {idx + (field === "assigneeIds" ? 1 : 0)}
                  </Text>
                </Flex>
              </Command.Item>
            );
          })}
          {isFetchingNextPage ? (
            <Command.Loading className="p-2">
              <MenuLoadingSkeleton avatar rows={2} />
            </Command.Loading>
          ) : null}
        </Command.Group>
      </Command.List>
    </Command>
  );
};

export const PriorityEditor = ({
  filters,
  setFilters,
}: StoriesFilterEditorProps) => {
  const togglePriority = (priority: StoryPriority) => {
    const selected = filters.priorities ?? [];
    const priorities = selected.includes(priority)
      ? selected.filter((value) => value !== priority)
      : [...selected, priority];
    setFilters({ ...filters, priorities: normalizeArrayFilter(priorities) });
  };

  return (
    <Command>
      <Command.Input autoFocus placeholder="Change priority..." />
      <Divider className="my-2" />
      <Command.List className="w-full border-0 bg-transparent p-0 shadow-none backdrop-blur-none dark:bg-transparent">
        <Command.Empty className="px-3 py-2 text-left text-base">
          <Text color="muted">No priority found.</Text>
        </Command.Empty>
        <Command.Group>
          {PRIORITIES.map((priority, idx) => (
            <Command.Item
              active={Boolean(filters.priorities?.includes(priority))}
              className="justify-between gap-4"
              key={priority}
              onSelect={() => {
                togglePriority(priority);
              }}
              value={priority}
            >
              <Box className="grid min-w-0 flex-1 grid-cols-[24px_minmax(0,1fr)] items-center">
                <PriorityIcon priority={priority} />
                <Text className="truncate">{priority}</Text>
              </Box>
              <Flex align="center" className="shrink-0" gap={2}>
                {filters.priorities?.includes(priority) ? (
                  <CheckIcon className="h-5 w-auto" strokeWidth={2.1} />
                ) : null}
                <Text color="muted">{idx}</Text>
              </Flex>
            </Command.Item>
          ))}
        </Command.Group>
      </Command.List>
    </Command>
  );
};

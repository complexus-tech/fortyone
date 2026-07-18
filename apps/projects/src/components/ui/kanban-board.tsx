"use client";

import { Box, Button, Flex, Popover, Text } from "ui";
import { cn } from "lib";
import { MoreHorizontalIcon } from "icons";
import { useParams } from "next/navigation";
import type {
  GroupedStoriesResponse,
  StoryPriority,
  StoryGroup,
} from "@/modules/stories/types";
import { useStatuses, useTeamStatuses } from "@/lib/hooks/statuses";
import { useMembers } from "@/lib/hooks/members";
import { useTeamMembers } from "@/lib/hooks/team-members";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import type { Member } from "@/types";
import type { State } from "@/types/states";
import { BodyContainer } from "../shared/body";
import { KanbanGroupTitle, StoriesKanbanHeader } from "./kanban-header";
import { KanbanGroup } from "./kanban-group";
import { useBoard } from "./board-context";
import {
  getHiddenKanbanGroupKeys,
  hideKanbanGroup,
  showKanbanGroup,
} from "./kanban-hidden-groups";

const getKanbanGroupIdentity = ({
  group,
  groupBy,
  members,
  statuses,
}: {
  group: StoryGroup;
  groupBy: StoriesViewOptions["groupBy"];
  members: Member[];
  statuses: State[];
}) => {
  const status =
    groupBy === "status"
      ? statuses.find((status) => status.id === group.key)
      : undefined;
  const member =
    groupBy === "assignee"
      ? members.find((member) => member.id === group.key)
      : undefined;
  const priority =
    groupBy === "priority" ? (group.key as StoryPriority) : undefined;

  return { member, priority, status };
};

const GroupedKanbanHeader = ({
  group,
  groupBy,
  members,
  statuses,
  onHide,
}: {
  group: StoryGroup;
  groupBy: StoriesViewOptions["groupBy"];
  isInSearch?: boolean;
  viewOptions: StoriesViewOptions;
  members: Member[];
  statuses: State[];
  onHide?: () => void;
}) => {
  const { member, priority, status } = getKanbanGroupIdentity({
    group,
    groupBy,
    members,
    statuses,
  });

  return (
    <StoriesKanbanHeader
      group={group}
      groupBy={groupBy}
      key={group.key}
      member={member}
      onHide={onHide}
      priority={priority}
      status={status}
    />
  );
};

const GroupedKanbanStories = ({
  meta,
  group,
  groupBy,
  members,
  statuses,
}: {
  group: StoryGroup;
  meta: GroupedStoriesResponse["meta"];
  groupBy: StoriesViewOptions["groupBy"];
  isInSearch?: boolean;
  viewOptions: StoriesViewOptions;
  members: Member[];
  statuses: State[];
}) => {
  const { member, priority, status } = getKanbanGroupIdentity({
    group,
    groupBy,
    members,
    statuses,
  });

  return (
    <KanbanGroup
      group={group}
      groupBy={groupBy}
      key={group.key}
      member={member}
      meta={meta}
      priority={priority}
      status={status}
    />
  );
};

const HiddenKanbanHeader = ({ count }: { count: number }) => {
  if (count === 0) return null;

  return (
    <Flex align="center" className="w-[340px] pl-1" gap={2}>
      <Text color="muted" fontWeight="medium">
        Hidden columns
      </Text>
      <Text color="muted">{count}</Text>
    </Flex>
  );
};

const HiddenKanbanGroups = ({
  groups,
  groupBy,
  members,
  statuses,
  onShow,
}: {
  groups: StoryGroup[];
  groupBy: StoriesViewOptions["groupBy"];
  members: Member[];
  statuses: State[];
  onShow: (groupKey: string) => void;
}) => {
  if (groups.length === 0) return null;

  return (
    <Box className="w-[340px] shrink-0">
      <Flex direction="column" gap={3}>
        {groups.map((group) => {
          const { member, priority, status } = getKanbanGroupIdentity({
            group,
            groupBy,
            members,
            statuses,
          });

          return (
            <Box
              className="border-border bg-surface hover:bg-surface-elevated group flex min-h-14 cursor-pointer items-center justify-between rounded-xl border-[0.5px] px-4 transition duration-200 ease-linear select-none"
              key={group.key}
              onClick={() => {
                onShow(group.key);
              }}
              onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                  event.preventDefault();
                  onShow(group.key);
                }
              }}
              role="button"
              tabIndex={0}
            >
              <Flex align="center" className="min-w-0" gap={2}>
                <KanbanGroupTitle
                  groupBy={groupBy}
                  member={member}
                  priority={priority}
                  status={status}
                />
              </Flex>
              <Flex align="center" className="shrink-0" gap={1}>
                <Popover>
                  <Popover.Trigger asChild>
                    <Button
                      aria-label="Hidden column options"
                      className="opacity-0 transition group-hover:opacity-100 focus-visible:opacity-100"
                      color="tertiary"
                      onClick={(event) => {
                        event.stopPropagation();
                      }}
                      size="sm"
                      variant="naked"
                    >
                      <MoreHorizontalIcon
                        className="h-[1.15rem] w-auto"
                        strokeWidth={4}
                      />
                    </Button>
                  </Popover.Trigger>
                  <Popover.Content align="end" className="w-40 p-1.5">
                    <Button
                      className="justify-start px-2"
                      color="tertiary"
                      fullWidth
                      onClick={() => {
                        onShow(group.key);
                      }}
                      size="sm"
                      variant="naked"
                    >
                      Show column
                    </Button>
                  </Popover.Content>
                </Popover>
                <Text color="muted">{group.totalCount}</Text>
              </Flex>
            </Box>
          );
        })}
      </Flex>
    </Box>
  );
};

export const KanbanBoard = ({
  className,
  groupedStories,
}: {
  groupedStories: GroupedStoriesResponse;
  className?: string;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { viewOptions, setViewOptions } = useBoard();
  const { groupBy } = viewOptions;
  const { data: teamStatuses = [] } = useTeamStatuses(teamId);
  const { data: allStatuses = [] } = useStatuses();
  const { data: allMembers = [] } = useMembers();
  const { data: teamMembers = [] } = useTeamMembers(teamId);
  const members = teamId ? teamMembers : allMembers;
  const statuses = teamId ? teamStatuses : allStatuses;
  const hiddenGroupKeys = getHiddenKanbanGroupKeys(viewOptions);
  const visibleGroups = groupedStories.groups.filter(
    (group) => !hiddenGroupKeys.includes(group.key),
  );
  const hiddenGroups = groupedStories.groups.filter((group) =>
    hiddenGroupKeys.includes(group.key),
  );
  const handleHide = (groupKey: string) => {
    setViewOptions?.(hideKanbanGroup(viewOptions, groupKey));
  };
  const handleShow = (groupKey: string) => {
    setViewOptions?.(showKanbanGroup(viewOptions, groupKey));
  };

  return (
    <BodyContainer
      className={cn("dark:bg-background overflow-x-auto bg-white", className)}
    >
      <Box className="sticky top-0 z-1 h-14 w-max px-6 backdrop-blur">
        <Flex
          align="center"
          className="h-full shrink-0 overflow-x-auto"
          gap={6}
        >
          {visibleGroups.map((group) => (
            <GroupedKanbanHeader
              group={group}
              groupBy={groupBy}
              key={group.key}
              members={members}
              onHide={
                setViewOptions
                  ? () => {
                      handleHide(group.key);
                    }
                  : undefined
              }
              statuses={statuses}
              viewOptions={viewOptions}
            />
          ))}
          <HiddenKanbanHeader count={hiddenGroups.length} />
        </Flex>
      </Box>
      <Box className="flex h-[calc(100%-3.5rem)] w-max gap-x-6 px-7">
        {visibleGroups.map((group) => (
          <GroupedKanbanStories
            group={group}
            groupBy={groupBy}
            key={group.key}
            members={members}
            meta={groupedStories.meta}
            statuses={statuses}
            viewOptions={viewOptions}
          />
        ))}
        <HiddenKanbanGroups
          groupBy={groupBy}
          groups={hiddenGroups}
          members={members}
          onShow={handleShow}
          statuses={statuses}
        />
      </Box>
    </BodyContainer>
  );
};

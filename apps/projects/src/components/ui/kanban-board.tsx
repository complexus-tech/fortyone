"use client";

import { Box, Button, Flex, Text } from "ui";
import { cn } from "lib";
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
    <Box className="w-[340px] shrink-0 pt-1">
      <Flex align="center" className="h-9 px-1" gap={2}>
        <Text color="muted" fontWeight="medium">
          Hidden columns
        </Text>
        <Text color="muted">{groups.length}</Text>
      </Flex>
      <Flex className="mt-3" direction="column" gap={3}>
        {groups.map((group) => {
          const { member, priority, status } = getKanbanGroupIdentity({
            group,
            groupBy,
            members,
            statuses,
          });

          return (
            <Button
              align="center"
              className="border-border bg-surface-muted/60 hover:bg-surface-muted min-h-14 justify-between rounded-md px-4"
              color="tertiary"
              fullWidth
              key={group.key}
              onClick={() => {
                onShow(group.key);
              }}
              size="md"
              variant="outline"
            >
              <Flex align="center" className="min-w-0" gap={2}>
                <KanbanGroupTitle
                  groupBy={groupBy}
                  member={member}
                  priority={priority}
                  status={status}
                />
              </Flex>
              <Text color="muted">{group.totalCount}</Text>
            </Button>
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
      className={cn("bg-surface-muted/50 overflow-x-auto", className)}
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
          {hiddenGroups.length > 0 ? <Box className="w-[340px]" /> : null}
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

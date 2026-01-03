"use client";

import { Box, Flex } from "ui";
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
import { StoriesKanbanHeader } from "./kanban-header";
import { KanbanGroup } from "./kanban-group";
import { useBoard } from "./board-context";

const GroupedKanbanHeader = ({
  group,
  groupBy,
  members,
  statuses,
}: {
  group: StoryGroup;
  groupBy: StoriesViewOptions["groupBy"];
  isInSearch?: boolean;
  viewOptions: StoriesViewOptions;
  members: Member[];
  statuses: State[];
}) => {
  const getGroupProps = () => {
    switch (groupBy) {
      case "priority":
        return { priority: group.key as StoryPriority };
      case "status":
        return { status: statuses.find((status) => status.id === group.key) };
      case "assignee":
        return { member: members.find((member) => member.id === group.key) };
      case "none":
        return {};
    }
  };

  return (
    <StoriesKanbanHeader
      groupBy={groupBy}
      {...getGroupProps()}
      group={group}
      key={group.key}
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
  const getGroupProps = () => {
    switch (groupBy) {
      case "priority":
        return { priority: group.key as StoryPriority };
      case "status":
        return { status: statuses.find((status) => status.id === group.key) };
      case "assignee":
        return { member: members.find((member) => member.id === group.key) };
      case "none":
        return {};
    }
  };

  return (
    <KanbanGroup
      groupBy={groupBy}
      {...getGroupProps()}
      group={group}
      key={group.key}
      meta={meta}
    />
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
  const { viewOptions } = useBoard();
  const { groupBy } = viewOptions;
  const { data: teamStatuses = [] } = useTeamStatuses(teamId);
  const { data: allStatuses = [] } = useStatuses();
  const { data: allMembers = [] } = useMembers();
  const { data: teamMembers = [] } = useTeamMembers(teamId);
  const members = (teamId ? teamMembers : allMembers).filter(
    ({ role }) => role !== "system",
  );
  const statuses = teamId ? teamStatuses : allStatuses;

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
          {groupedStories.groups.map((group) => (
            <GroupedKanbanHeader
              group={group}
              groupBy={groupBy}
              key={group.key}
              members={members}
              statuses={statuses}
              viewOptions={viewOptions}
            />
          ))}
        </Flex>
      </Box>
      <Box className="flex h-[calc(100%-3.5rem)] w-max gap-x-6 px-7">
        {groupedStories.groups.map((group) => (
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
      </Box>
    </BodyContainer>
  );
};

"use client";

import { Box, Flex } from "ui";
import { cn } from "lib";
import { useParams } from "next/navigation";
import type { Story, StoryPriority } from "@/modules/stories/types";
import { useStatuses, useTeamStatuses } from "@/lib/hooks/statuses";
import { useMembers } from "@/lib/hooks/members";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { BodyContainer } from "../shared/body";
import { StoriesKanbanHeader } from "./kanban-header";
import { KanbanGroup } from "./kanban-group";
import { useBoard } from "./board-context";

export const KanbanBoard = ({
  stories,
  className,
}: {
  stories: Story[];
  className?: string;
}) => {
  const { teamId } = useParams<{ teamId: string }>();
  const { viewOptions } = useBoard();
  const { groupBy } = viewOptions;
  const { data: teamStatuses = [] } = useTeamStatuses(teamId);
  const { data: allStatuses = [] } = useStatuses();
  const { data: allMembers = [] } = useMembers();
  const { data: teamMembers = [] } = useTeamMembers(teamId);
  const members = teamId ? teamMembers : allMembers;
  const statuses = teamId ? teamStatuses : allStatuses;
  const priorities: StoryPriority[] = [
    "Urgent",
    "High",
    "Medium",
    "Low",
    "No Priority",
  ];

  return (
    <BodyContainer
      className={cn(
        "overflow-x-auto bg-gray-50/60 dark:bg-transparent",
        className,
      )}
    >
      <Box className="sticky top-0 z-[1] h-[3.5rem] w-max px-6 backdrop-blur">
        <Flex
          align="center"
          className="h-full shrink-0 overflow-x-auto"
          gap={6}
        >
          {groupBy === "Status" &&
            statuses.map((status) => (
              <StoriesKanbanHeader
                groupBy={groupBy}
                key={status.id}
                status={status}
                stories={stories}
              />
            ))}
          {groupBy === "Priority" &&
            priorities.map((priority) => (
              <StoriesKanbanHeader
                groupBy={groupBy}
                key={priority}
                priority={priority}
                stories={stories}
              />
            ))}
          {groupBy === "Assignee" &&
            members.map((member) => (
              <StoriesKanbanHeader
                groupBy={groupBy}
                key={member.id}
                member={member}
                stories={stories}
              />
            ))}
        </Flex>
      </Box>
      <Box className="flex h-[calc(100%-3.5rem)] w-max gap-x-6 px-7">
        {groupBy === "Status" &&
          statuses.map((status) => (
            <KanbanGroup
              groupBy={groupBy}
              key={status.id}
              status={status}
              stories={stories}
            />
          ))}
        {groupBy === "Priority" &&
          priorities.map((priority) => (
            <KanbanGroup
              groupBy={groupBy}
              key={priority}
              priority={priority}
              stories={stories}
            />
          ))}
        {groupBy === "Assignee" &&
          members.map((member) => (
            <KanbanGroup
              groupBy={groupBy}
              key={member.id}
              member={member}
              stories={stories}
            />
          ))}
      </Box>
    </BodyContainer>
  );
};

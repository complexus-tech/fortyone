"use client";
import { useState } from "react";
import { Flex, Button, Text, Box, Tooltip, Avatar } from "ui";
import { PlusIcon, StoryIcon } from "icons";
import { cn } from "lib";
import type { StoryGroup, StoryPriority } from "@/modules/stories/types";
import type { State } from "@/types/states";
import { useUserRole, useTerminology } from "@/hooks";
import type { Member } from "@/types";
import { StoryStatusIcon } from "./story-status-icon";
import { NewStoryDialog } from "./new-story-dialog";
import type { ViewOptionsGroupBy } from "./stories-view-options-button";
import { PriorityIcon } from "./priority-icon";
import { useBoard } from "./board-context";

export const StoriesKanbanHeader = ({
  status,
  priority,
  group,
  groupBy,
  member,
}: {
  status?: State;
  priority?: StoryPriority;
  member?: Member;
  group: StoryGroup;
  groupBy: ViewOptionsGroupBy;
}) => {
  const { getTermDisplay } = useTerminology();
  const { viewOptions } = useBoard();
  const { showEmptyGroups } = viewOptions;
  const [isOpen, setIsOpen] = useState(false);
  const { userRole } = useUserRole();

  return (
    <Box
      className={cn({
        hidden: group.loadedCount === 0 && !showEmptyGroups,
      })}
    >
      <Flex
        align="center"
        className="w-[340px] pl-1"
        gap={2}
        justify="between"
        key={status?.id}
      >
        <Flex align="center" gap={2}>
          {groupBy === "status" && (
            <>
              <StoryStatusIcon statusId={status?.id} />
              {status?.name}
            </>
          )}
          {groupBy === "priority" && (
            <>
              <PriorityIcon priority={priority} />
              {priority}
            </>
          )}
          {groupBy === "assignee" && (
            <>
              <Avatar
                className={cn({
                  "text-black dark:text-white": !member?.fullName,
                })}
                name={member?.fullName}
                size="xs"
                src={member?.avatarUrl}
              />
              {member?.username || "Unassigned"}
            </>
          )}
          <Tooltip side="bottom" title="Total stories">
            <span>
              <StoryIcon className="ml-3 h-5 w-auto" strokeWidth={2} />
            </span>
          </Tooltip>
          <Text color="muted">
            {group.totalCount}{" "}
            {getTermDisplay("storyTerm", { variant: "plural" })}
          </Text>
        </Flex>
        <Button
          color="tertiary"
          disabled={userRole === "guest"}
          onClick={() => {
            if (userRole !== "guest") {
              setIsOpen(true);
            }
          }}
          size="sm"
          variant="naked"
        >
          <PlusIcon className="h-[1.2rem] w-auto" />
        </Button>
      </Flex>
      <NewStoryDialog
        assigneeId={member?.id}
        isOpen={isOpen}
        priority={priority}
        setIsOpen={setIsOpen}
        statusId={status?.id}
      />
    </Box>
  );
};

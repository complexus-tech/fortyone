"use client";
import { useState } from "react";
import { Flex, Button, Text, Box, Tooltip, Avatar } from "ui";
import { MinimizeIcon, PlusIcon, StoryIcon } from "icons";
import { cn } from "lib";
import type { Story, StoryPriority } from "@/modules/stories/types";
import type { State } from "@/types/states";
import { useUserRole, useTerminologyDisplay } from "@/hooks";
import type { Member } from "@/types";
import { StoryStatusIcon } from "./story-status-icon";
import { NewStoryDialog } from "./new-story-dialog";
import type { ViewOptionsGroupBy } from "./stories-view-options-button";
import { PriorityIcon } from "./priority-icon";
import { useBoard } from "./board-context";

export const StoriesKanbanHeader = ({
  status,
  priority,
  stories,
  groupBy,
  member,
}: {
  status?: State;
  priority?: StoryPriority;
  member?: Member;
  stories: Story[];
  groupBy: ViewOptionsGroupBy;
}) => {
  const { getTermDisplay } = useTerminologyDisplay();
  const { viewOptions } = useBoard();
  const { showEmptyGroups } = viewOptions;
  const [isOpen, setIsOpen] = useState(false);
  const { userRole } = useUserRole();

  let filteredStories: Story[] = [];
  if (groupBy === "Status") {
    filteredStories = stories.filter((story) => story.statusId === status?.id);
  } else if (groupBy === "Priority") {
    filteredStories = stories.filter((story) => story.priority === priority);
  } else if (groupBy === "Assignee") {
    filteredStories = stories.filter(
      (story) => story.assigneeId === member?.id,
    );
  }

  return (
    <Box
      className={cn({
        hidden: filteredStories.length === 0 && !showEmptyGroups,
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
          {groupBy === "Status" && (
            <>
              <StoryStatusIcon statusId={status?.id} />
              {status?.name}
            </>
          )}
          {groupBy === "Priority" && (
            <>
              <PriorityIcon priority={priority} />
              {priority}
            </>
          )}
          {groupBy === "Assignee" && (
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
            {filteredStories.length}{" "}
            {getTermDisplay("storyTerm", { variant: "plural" })}
          </Text>
        </Flex>
        <span className="flex items-center gap-1">
          <Button color="tertiary" size="sm" variant="naked">
            <MinimizeIcon className="h-[1.2rem] w-auto" />
          </Button>
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
        </span>
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

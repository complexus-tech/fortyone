"use client";
import { useState } from "react";
import { Flex, Button, Text, Box, Tooltip } from "ui";
import { MinimizeIcon, PlusIcon, StoryIcon } from "icons";
import type { Story, StoryPriority, StoryStatus } from "@/types/story";
import { StoryStatusIcon } from "./story-status-icon";
import { NewStoryDialog } from "./new-story-dialog";
import type { ViewOptionsGroupBy } from "./stories-view-options-button";
import { PriorityIcon } from "./priority-icon";
import { cn } from "lib";
import { useBoard } from "@/components/ui/stories-board";

export const StoriesKanbanHeader = ({
  status,
  priority,
  stories,
  groupBy,
}: {
  status?: StoryStatus;
  priority?: StoryPriority;
  stories: Story[];
  groupBy: ViewOptionsGroupBy;
}) => {
  const { viewOptions } = useBoard();
  const { showEmptyGroups } = viewOptions;
  const [isOpen, setIsOpen] = useState(false);

  const filteredStories =
    groupBy === "Status"
      ? stories.filter((story) => story.status === status)
      : stories.filter((story) => story.priority === priority);
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
        key={status}
      >
        <Flex align="center" gap={2}>
          {groupBy === "Status" && (
            <>
              <StoryStatusIcon status={status} />
              {status}
            </>
          )}
          {groupBy === "Priority" && (
            <>
              <PriorityIcon priority={priority} />
              {priority}
            </>
          )}
          <Tooltip title="Total stories" side="bottom">
            <span>
              <StoryIcon strokeWidth={2} className="ml-3 h-5 w-auto" />
            </span>
          </Tooltip>
          <Text color="muted">{filteredStories.length} stories</Text>
        </Flex>
        <span className="flex items-center gap-1">
          <Button color="tertiary" size="sm" variant="naked">
            <MinimizeIcon className="h-[1.2rem] w-auto" />
          </Button>
          <Button
            color="tertiary"
            onClick={() => {
              setIsOpen(true);
            }}
            size="sm"
            variant="naked"
          >
            <PlusIcon className="h-[1.2rem] w-auto" />
          </Button>
        </span>
      </Flex>
      <NewStoryDialog
        isOpen={isOpen}
        priority={priority}
        setIsOpen={setIsOpen}
        status={status}
      />
    </Box>
  );
};

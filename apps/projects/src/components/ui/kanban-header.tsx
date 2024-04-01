"use client";
import { useState } from "react";
import { Flex, Button, Text } from "ui";
import { MinimizeIcon, PlusIcon } from "icons";
import type { Story, StoryPriority, StoryStatus } from "@/types/story";
import { StoryStatusIcon } from "./story-status-icon";
import { NewStoryDialog } from "./new-story-dialog";
import type { ViewOptionsGroupBy } from "./stories-view-options-button";
import { PriorityIcon } from "./priority-icon";

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
  const [isOpen, setIsOpen] = useState(false);

  const filteredStories =
    groupBy === "Status"
      ? stories.filter((story) => story.status === status)
      : stories.filter((story) => story.priority === priority);
  return (
    <>
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
          <Text as="span" color="muted">
            {filteredStories.length}
          </Text>
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
    </>
  );
};

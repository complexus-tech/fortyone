"use client";
import { useState } from "react";
import { Flex, Button, Text } from "ui";
import { MinimizeIcon, PlusIcon } from "icons";
import type { Story, StoryStatus } from "@/types/story";
import { StoryStatusIcon } from "./story-status-icon";
import { NewStoryDialog } from "./new-story-dialog";

export const StoriesKanbanHeader = ({
  status,
  stories,
}: {
  status: StoryStatus;
  stories: Story[];
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const filteredStories = stories.filter((story) => story.status === status);
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
          <StoryStatusIcon status={status} />
          {status}
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
      <NewStoryDialog isOpen={isOpen} setIsOpen={setIsOpen} status={status} />
    </>
  );
};

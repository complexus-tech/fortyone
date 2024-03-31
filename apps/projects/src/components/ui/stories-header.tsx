"use client";
import { useState } from "react";
import { Button, Container, Flex, Text, Tooltip } from "ui";
import { cn } from "lib";
import { PlusIcon } from "icons";
import type { StoryStatus } from "../../types/story";
import { StoryStatusIcon } from "./story-status-icon";
import { NewStoryDialog } from "./new-story-dialog";

type StoryHeaderProps = {
  status?: StoryStatus;
  count: number;
  className?: string;
};
export const StoriesHeader = ({
  count,
  className,
  status = "Backlog",
}: StoryHeaderProps) => {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <Container
      className={cn(
        "sticky top-0 z-[1] select-none bg-gray-50/90 py-[0.4rem] backdrop-blur dark:bg-dark-300/90",
        className,
      )}
    >
      <Flex align="center" justify="between">
        <Flex align="center" gap={2}>
          <StoryStatusIcon status={status} />
          <Text fontWeight="medium">{status}</Text>
          <Text color="muted">{count}</Text>
        </Flex>
        <Tooltip side="left" title="Add story">
          <Button
            color="tertiary"
            leftIcon={
              <PlusIcon className="h-[1.1rem] w-auto dark:text-gray-200" />
            }
            onClick={() => {
              setIsOpen(true);
            }}
            size="sm"
            variant="naked"
          >
            <span className="sr-only">Add story</span>
          </Button>
        </Tooltip>
      </Flex>
      <NewStoryDialog isOpen={isOpen} setIsOpen={setIsOpen} status={status} />
    </Container>
  );
};

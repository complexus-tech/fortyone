"use client";
import { useState } from "react";
import { Button, Container, Flex, Text, Tooltip } from "ui";
import { cn } from "lib";
import { PlusIcon } from "icons";
import type { StoryPriority, StoryStatus } from "@/types/story";
import type { ViewOptionsGroupBy } from "@/components/ui/stories-view-options-button";
import { StoryStatusIcon } from "./story-status-icon";
import { NewStoryDialog } from "./new-story-dialog";
import { PriorityIcon } from "./priority-icon";

type StoryHeaderProps = {
  status?: StoryStatus;
  count: number;
  className?: string;
  priority?: StoryPriority;
  groupBy: ViewOptionsGroupBy;
};
export const StoriesHeader = ({
  count,
  className,
  status = "Backlog",
  priority,
  groupBy,
}: StoryHeaderProps) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Container
      className={cn(
        "sticky top-0 z-[1] select-none bg-gray-50/90 py-[0.4rem] backdrop-blur dark:bg-[#181818]/90",
        {
          "border-b-[0.5px] border-gray-100 dark:border-dark-50/60":
            count === 0,
        },
        className,
      )}
    >
      <Flex align="center" justify="between">
        <Flex align="center" gap={2}>
          {groupBy === "Status" && (
            <>
              <StoryStatusIcon status={status} />
              <Text fontWeight="medium">{status}</Text>
            </>
          )}
          {groupBy === "Priority" && (
            <>
              <PriorityIcon priority={priority} />
              <Text fontWeight="medium">{priority}</Text>
            </>
          )}
          <Text color="muted">{count}</Text>
        </Flex>
        <Tooltip side="left" title="New Story">
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
            <span className="sr-only">New Story</span>
          </Button>
        </Tooltip>
      </Flex>
      <NewStoryDialog
        isOpen={isOpen}
        priority={priority}
        setIsOpen={setIsOpen}
        status={status}
      />
    </Container>
  );
};

"use client";
import { useState } from "react";
import { Button, Container, Flex, Text, Tooltip } from "ui";
import { cn } from "lib";
import { ArrowDownIcon, PlusIcon, StoryIcon } from "icons";
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
  isCollapsed: boolean;
  setIsCollapsed: (value: boolean) => void;
};
export const StoriesHeader = ({
  count,
  className,
  status = "Backlog",
  priority,
  groupBy,
  isCollapsed,
  setIsCollapsed,
}: StoryHeaderProps) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Container
      className={cn(
        "sticky top-0 z-[1] select-none border-b-[0.5px] border-gray-100 bg-gray-50/90 py-[0.4rem] backdrop-blur dark:border-dark-50/50 dark:bg-[#191919]/90",
        {
          "border-b-[0.5px] border-gray-100 dark:border-dark-50/60":
            count === 0,
        },
        className,
      )}
    >
      <Flex align="center" justify="between">
        <Flex align="center" className="gap-1.5">
          <Button
            color="tertiary"
            onClick={() => {
              setIsCollapsed(!isCollapsed);
            }}
            rightIcon={
              <ArrowDownIcon
                className={cn("h-4 w-auto transition dark:text-gray-200", {
                  "-rotate-90": isCollapsed,
                })}
                strokeWidth={1}
              />
            }
            size="sm"
            variant="naked"
          >
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
          </Button>
          <Tooltip side="bottom" title="Total stories">
            <span>
              <StoryIcon className="ml-3 h-5 w-auto" strokeWidth={2} />
            </span>
          </Tooltip>
          <Text color="muted">{count} stories</Text>
        </Flex>
        <Flex gap={2}>
          <Tooltip side="top" title="New Story">
            <Button
              color="tertiary"
              leftIcon={
                <PlusIcon className="h-[1.1rem] w-auto dark:text-gray-200" />
              }
              onClick={() => {
                setIsOpen(true);
              }}
              size="sm"
            >
              <span className="sr-only">New Story</span>
            </Button>
          </Tooltip>
        </Flex>
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

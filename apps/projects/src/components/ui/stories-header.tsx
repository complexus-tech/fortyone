"use client";
import { useState } from "react";
import { Avatar, Button, Checkbox, Container, Flex, Text, Tooltip } from "ui";
import { cn } from "lib";
import { ArrowDownIcon, PlusIcon, StoryIcon } from "icons";
import type { Story, StoryPriority } from "@/modules/stories/types";
import type { ViewOptionsGroupBy } from "@/components/ui/stories-view-options-button";
import type { State } from "@/types/states";
import { useBoard } from "@/components/ui/board-context";
import type { Member } from "@/types";
import { useUserRole } from "@/hooks";
import { StoryStatusIcon } from "./story-status-icon";
import { NewStoryDialog } from "./new-story-dialog";
import { PriorityIcon } from "./priority-icon";

type StoryHeaderProps = {
  status?: State;
  className?: string;
  priority?: StoryPriority;
  groupBy: ViewOptionsGroupBy;
  stories: Story[];
  isCollapsed: boolean;
  setIsCollapsed: (value: boolean) => void;
  assignee?: Member;
};
export const StoriesHeader = ({
  className,
  stories,
  status,
  priority,
  groupBy,
  isCollapsed,
  setIsCollapsed,
  assignee,
}: StoryHeaderProps) => {
  const count = stories.length;
  const [isOpen, setIsOpen] = useState(false);
  const { selectedStories, setSelectedStories } = useBoard();
  const { userRole } = useUserRole();

  const groupedStories = stories.map((s) => s.id);

  return (
    <Container
      className={cn(
        "sticky top-0 z-[1] select-none border-b-[0.5px] border-gray-100 bg-gray-50/90 py-[0.4rem] backdrop-blur dark:border-dark-50/50 dark:bg-dark-200/75",
        {
          "border-b-[0.5px] border-gray-100 dark:border-dark-50/60":
            count === 0,
        },
        className,
      )}
    >
      <Flex align="center" justify="between">
        <Flex align="center" className="relative gap-1.5">
          <Checkbox
            checked={
              groupedStories.every((s) => selectedStories.includes(s)) &&
              groupedStories.length > 0
            }
            className="absolute -left-[1.6rem] rounded-[0.35rem]"
            disabled={userRole === "guest"}
            onCheckedChange={(checked) => {
              if (checked) {
                setSelectedStories(
                  Array.from(new Set([...selectedStories, ...groupedStories])),
                );
              } else {
                setSelectedStories([
                  ...selectedStories.filter((s) => !groupedStories.includes(s)),
                ]);
              }
            }}
          />
          <Button
            color="tertiary"
            onClick={() => {
              setIsCollapsed(!isCollapsed);
            }}
            rightIcon={
              <ArrowDownIcon
                className={cn(
                  "h-4 w-auto text-gray transition dark:text-gray-300",
                  {
                    "-rotate-90": isCollapsed,
                  },
                )}
                strokeWidth={1}
              />
            }
            size="sm"
            variant="naked"
          >
            {groupBy === "Assignee" && (
              <Flex align="center" className="gap-1.5">
                <Avatar
                  className={cn({
                    "text-black dark:text-white": !assignee?.fullName,
                  })}
                  name={assignee?.fullName}
                  size="xs"
                  src={assignee?.avatarUrl}
                />
                <Text
                  className={cn("relative -top-[1px]", {
                    "top-[0px]": !assignee?.fullName,
                  })}
                  fontWeight="medium"
                >
                  {assignee?.username || "Unassigned"}
                </Text>
              </Flex>
            )}
            {groupBy === "Status" && (
              <>
                <StoryStatusIcon statusId={status?.id} />
                <Text fontWeight="medium">{status?.name}</Text>
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
              <StoryIcon
                className="ml-1 h-5 w-auto text-gray dark:text-gray-300"
                strokeWidth={2}
              />
            </span>
          </Tooltip>
          <Text color="muted">{count} stories</Text>
        </Flex>
        <Flex gap={2}>
          <Tooltip side="top" title="New Story">
            <Button
              color="tertiary"
              disabled={userRole === "guest"}
              leftIcon={
                <PlusIcon className="h-[1.1rem] w-auto dark:text-gray-200" />
              }
              onClick={() => {
                if (userRole !== "guest") {
                  setIsOpen(true);
                }
              }}
              size="sm"
              variant="outline"
            >
              <span className="sr-only">New Story</span>
            </Button>
          </Tooltip>
        </Flex>
      </Flex>
      <NewStoryDialog
        assigneeId={assignee?.id || null}
        isOpen={isOpen}
        priority={priority}
        setIsOpen={setIsOpen}
        statusId={status?.id}
      />
    </Container>
  );
};

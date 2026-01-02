"use client";
import { useState } from "react";
import { Avatar, Button, Checkbox, Container, Flex, Text, Tooltip } from "ui";
import { cn } from "lib";
import { ArrowDownIcon, PlusIcon, StoryIcon } from "icons";
import type { StoryGroup, StoryPriority } from "@/modules/stories/types";
import type { ViewOptionsGroupBy } from "@/components/ui/stories-view-options-button";
import type { State } from "@/types/states";
import { useBoard } from "@/components/ui/board-context";
import type { Member } from "@/types";
import { useUserRole, useTerminology } from "@/hooks";
import { StoryStatusIcon } from "./story-status-icon";
import { NewStoryDialog } from "./new-story-dialog";
import { PriorityIcon } from "./priority-icon";

type StoryHeaderProps = {
  status?: State;
  className?: string;
  priority?: StoryPriority;
  groupBy: ViewOptionsGroupBy;
  group: StoryGroup;
  isCollapsed: boolean;
  setIsCollapsed: (value: boolean) => void;
  assignee?: Member;
};
export const StoriesHeader = ({
  className,
  group,
  status,
  priority,
  groupBy,
  isCollapsed,
  setIsCollapsed,
  assignee,
}: StoryHeaderProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const { selectedStories, setSelectedStories } = useBoard();
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();

  const groupedStories = group.stories.map((s) => s.id);

  return (
    <Container
      className={cn(
        "border-border bg-surface-muted/85 sticky top-0 z-1 border-b-[0.5px] py-[0.4rem] backdrop-blur select-none",
        {
          "border-border border-b-[0.5px]": group.loadedCount === 0,
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
            className="absolute -left-[1.6rem] hidden rounded-[0.35rem] md:inline"
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
                className={cn("text-text-muted h-4 w-auto transition", {
                  "-rotate-90": isCollapsed,
                })}
                strokeWidth={1}
              />
            }
            size="sm"
            variant="naked"
          >
            {groupBy === "assignee" && (
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
                  className={cn("relative -top-px", {
                    "top-0": !assignee?.fullName,
                  })}
                  fontWeight="medium"
                >
                  {assignee?.username || "Unassigned"}
                </Text>
              </Flex>
            )}
            {groupBy === "status" && (
              <>
                <StoryStatusIcon statusId={status?.id} />
                <Text fontWeight="medium">{status?.name}</Text>
              </>
            )}
            {groupBy === "priority" && (
              <>
                <PriorityIcon priority={priority} />
                <Text fontWeight="medium">{priority}</Text>
              </>
            )}
          </Button>
          <Tooltip side="bottom" title="Total stories">
            <span>
              <StoryIcon
                className="text-text-muted ml-1 h-5 w-auto"
                strokeWidth={2}
              />
            </span>
          </Tooltip>
          <Text color="muted">
            {group.totalCount}{" "}
            {getTermDisplay("storyTerm", {
              variant: group.totalCount === 1 ? "singular" : "plural",
            })}
          </Text>
        </Flex>
        <Flex gap={2}>
          <Tooltip
            side="top"
            title={`New ${getTermDisplay("storyTerm", { capitalize: true })}`}
          >
            <Button
              color="tertiary"
              disabled={userRole === "guest"}
              leftIcon={
                <PlusIcon className="text-foreground h-[1.1rem] w-auto" />
              }
              onClick={() => {
                if (userRole !== "guest") {
                  setIsOpen(true);
                }
              }}
              size="sm"
              variant="outline"
            >
              <span className="sr-only">
                New {getTermDisplay("storyTerm", { capitalize: true })}
              </span>
            </Button>
          </Tooltip>
        </Flex>
      </Flex>
      <NewStoryDialog
        assigneeId={assignee?.id}
        isOpen={isOpen}
        priority={priority}
        setIsOpen={setIsOpen}
        statusId={status?.id}
      />
    </Container>
  );
};

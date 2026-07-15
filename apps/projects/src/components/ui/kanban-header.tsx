"use client";
import { useState } from "react";
import { Flex, Button, Text, Box, Tooltip, Avatar, Popover } from "ui";
import { MoreHorizontalIcon, PlusIcon, StoryIcon } from "icons";
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

export const KanbanGroupTitle = ({
  status,
  priority,
  groupBy,
  member,
}: {
  status?: State;
  priority?: StoryPriority;
  member?: Member;
  groupBy: ViewOptionsGroupBy;
}) => (
  <>
    {groupBy === "status" && (
      <>
        <StoryStatusIcon statusId={status?.id} />
        <span
          className="inline-block max-w-[20ch] truncate"
          title={status?.name}
        >
          {status?.name}
        </span>
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
        <span
          className="inline-block max-w-[20ch] truncate"
          title={member?.username || "Unassigned"}
        >
          {member?.username || "Unassigned"}
        </span>
      </>
    )}
  </>
);

export const StoriesKanbanHeader = ({
  status,
  priority,
  group,
  groupBy,
  member,
  onHide,
}: {
  status?: State;
  priority?: StoryPriority;
  member?: Member;
  group: StoryGroup;
  groupBy: ViewOptionsGroupBy;
  onHide?: () => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const { newStoryDefaults, viewOptions } = useBoard();
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
          <KanbanGroupTitle
            groupBy={groupBy}
            member={member}
            priority={priority}
            status={status}
          />
          <Tooltip side="bottom" title="Total stories">
            <span>
              <StoryIcon className="ml-0 h-5 w-auto" strokeWidth={2} />
            </span>
          </Tooltip>
          <Text color="muted">
            {group.totalCount}{" "}
            {getTermDisplay("storyTerm", {
              variant: group.totalCount === 1 ? "singular" : "plural",
            })}
          </Text>
        </Flex>
        <Flex align="center" gap={1}>
          {onHide ? (
            <Popover>
              <Popover.Trigger asChild>
                <Button
                  aria-label="Column options"
                  color="tertiary"
                  size="sm"
                  variant="naked"
                >
                  <MoreHorizontalIcon
                    className="h-[1.15rem] w-auto"
                    strokeWidth={4}
                  />
                </Button>
              </Popover.Trigger>
              <Popover.Content align="end" className="w-44 p-1.5">
                <Button
                  className="justify-start px-2"
                  color="tertiary"
                  fullWidth
                  onClick={onHide}
                  size="sm"
                  variant="naked"
                >
                  Hide column
                </Button>
              </Popover.Content>
            </Popover>
          ) : null}
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
      </Flex>
      <NewStoryDialog
        assigneeId={groupBy === "assignee" ? member?.id ?? null : undefined}
        isOpen={isOpen}
        objectiveId={newStoryDefaults.objectiveId}
        priority={priority}
        setIsOpen={setIsOpen}
        sprintId={newStoryDefaults.sprintId}
        statusId={status?.id}
        teamId={newStoryDefaults.teamId}
      />
    </Box>
  );
};

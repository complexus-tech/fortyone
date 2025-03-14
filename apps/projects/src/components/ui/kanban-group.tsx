"use client";
import { useState, type ReactNode } from "react";
import { useDroppable } from "@dnd-kit/core";
import { cn } from "lib";
import { Box, Button } from "ui";
import { PlusIcon } from "icons";
import type { Story, StoryPriority } from "@/modules/stories/types";
import type { State } from "@/types/states";
import type { Member } from "@/types";
import { StoryCard } from "./story/card";
import type { ViewOptionsGroupBy } from "./stories-view-options-button";
import { NewStoryDialog } from "./new-story-dialog";
import { useBoard } from "./board-context";

const List = ({
  children,
  id,
  totalStories,
}: {
  children: ReactNode;
  id: string | number;
  totalStories: number;
}) => {
  const { viewOptions } = useBoard();
  const { showEmptyGroups } = viewOptions;
  const { isOver, setNodeRef } = useDroppable({
    id,
  });
  return (
    <Box
      className={cn({
        hidden: totalStories === 0 && !showEmptyGroups,
      })}
    >
      <div
        className={cn(
          "flex h-full w-[340px] flex-col gap-4 overflow-y-auto rounded-[0.45rem] pb-6 transition",
          {
            "bg-gray-100/30 dark:bg-dark-300/50": totalStories === 0,
            "bg-gray-100/40 dark:bg-dark-300/40": isOver,
          },
        )}
        ref={setNodeRef}
      >
        {children}
      </div>
    </Box>
  );
};

export const KanbanGroup = ({
  stories,
  status,
  priority,
  member,
  groupBy = "Status",
}: {
  stories: Story[];
  status?: State;
  priority?: StoryPriority;
  member?: Member;
  groupBy: ViewOptionsGroupBy;
}) => {
  const [isOpen, setIsOpen] = useState(false);

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

  const getId = () => {
    if (groupBy === "Status") return status?.id;
    if (groupBy === "Assignee") return member?.id;
    return priority;
  };

  const id = getId() || "";

  return (
    <List id={id} key={id} totalStories={filteredStories.length}>
      {filteredStories.map((story) => (
        <StoryCard key={story.id} story={story} />
      ))}
      <Button
        className="relative min-h-[2.35rem] w-[340px]"
        color="tertiary"
        onClick={() => {
          setIsOpen(true);
        }}
        size="sm"
        variant="naked"
      >
        <PlusIcon className="relative -top-[0.3px] h-[1.15rem] w-auto" /> New
        Story
      </Button>
      <NewStoryDialog
        assigneeId={member?.id}
        isOpen={isOpen}
        priority={priority}
        setIsOpen={setIsOpen}
        statusId={status?.id}
      />
    </List>
  );
};

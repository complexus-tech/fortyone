"use client";
import { useState, type ReactNode } from "react";
import { useDroppable } from "@dnd-kit/core";
import { cn } from "lib";
import { Box, Button } from "ui";
import { PlusIcon } from "icons";
import type { Story, StoryPriority, StoryStatus } from "@/types/story";
import { StoryCard } from "./story/card";
import type { ViewOptionsGroupBy } from "./stories-view-options-button";
import { NewStoryDialog } from "./new-story-dialog";

const List = ({
  children,
  id,
  totalStories,
}: {
  children: ReactNode;
  id: string | number;
  totalStories: number;
}) => {
  const { isOver, setNodeRef } = useDroppable({
    id,
  });
  return (
    <Box>
      <div
        className={cn(
          "flex h-full w-[340px] flex-col gap-3 overflow-y-auto rounded-[0.45rem] pb-6 transition",
          {
            "bg-gray-100/30 dark:bg-dark-300/20": totalStories === 0,
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
  groupBy = "Status",
}: {
  stories: Story[];
  status?: StoryStatus;
  priority?: StoryPriority;
  groupBy: ViewOptionsGroupBy;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const filteredStories =
    groupBy === "Status"
      ? stories.filter((story) => story.status === status)
      : stories.filter((story) => story.priority === priority);

  return (
    <List
      id={(groupBy === "Status" ? status : priority) as string}
      key={groupBy === "Status" ? status : priority}
      totalStories={filteredStories.length}
    >
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
        isOpen={isOpen}
        priority={priority}
        setIsOpen={setIsOpen}
        status={status}
      />
    </List>
  );
};

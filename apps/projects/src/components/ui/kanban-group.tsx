"use client";
import type { ReactNode } from "react";
import { useDroppable } from "@dnd-kit/core";
import { cn } from "lib";
import type { Story, StoryStatus } from "@/types/story";
import { StoryCard } from "./story/card";

const List = ({
  children,
  id,
}: {
  children: ReactNode;
  id: string | number;
}) => {
  const { isOver, setNodeRef } = useDroppable({
    id,
  });
  return (
    <div
      className={cn(
        "flex h-full flex-col gap-3 overflow-y-auto rounded-lg pb-6 transition",
        {
          "bg-white opacity-60 dark:bg-dark-200/60": isOver,
        },
      )}
      ref={setNodeRef}
    >
      {children}
    </div>
  );
};

export const KanbanGroup = ({
  stories,
  status,
}: {
  stories: Story[];
  status: StoryStatus;
}) => {
  const filteredStories = stories.filter((story) => story.status === status);
  return (
    <List id={status} key={status}>
      {filteredStories.map((story) => (
        <StoryCard story={story} key={story.id} />
      ))}
    </List>
  );
};

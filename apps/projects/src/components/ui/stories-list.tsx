"use client";
import { useDroppable } from "@dnd-kit/core";
import { cn } from "lib";
import type { Story as StoryType } from "@/types/story";
import { StoryRow } from "./story/row";

export const StoriesList = ({
  stories,
  id,
}: {
  stories: StoryType[];
  id: string | number;
}) => {
  const { isOver, setNodeRef } = useDroppable({
    id,
  });
  return (
    <div
      className={cn("border-0 border-transparent transition", {
        "border border-primary": isOver,
      })}
      ref={setNodeRef}
    >
      {stories.map((story) => (
        <StoryRow key={story.id} story={story} />
      ))}
    </div>
  );
};

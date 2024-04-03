"use client";
import { cn } from "lib";
import { useDroppable } from "@dnd-kit/core";
import type { StoryStatus, Story, StoryPriority } from "@/types/story";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { StoriesHeader } from "./stories-header";
import { StoriesList } from "./stories-list";

export const StoriesGroup = ({
  stories,
  status,
  priority,
  className,
  viewOptions,
}: {
  stories: Story[];
  status?: StoryStatus;
  priority?: StoryPriority;
  className?: string;
  viewOptions: StoriesViewOptions;
}) => {
  const { groupBy, showEmptyGroups } = viewOptions;
  const id = (groupBy === "Status" ? status : priority) as string;
  const { isOver, setNodeRef } = useDroppable({
    id,
  });
  const filteredStories =
    groupBy === "Status"
      ? stories.filter((story) => story.status === status)
      : stories.filter((story) => story.priority === priority);

  return (
    <div
      className={cn("border-0 border-transparent transition", {
        "border border-primary": isOver,
        hidden: !showEmptyGroups && filteredStories.length === 0,
      })}
      ref={setNodeRef}
    >
      <StoriesHeader
        className={className}
        count={filteredStories.length}
        groupBy={groupBy}
        priority={priority}
        status={status}
      />
      <StoriesList stories={filteredStories} />
    </div>
  );
};

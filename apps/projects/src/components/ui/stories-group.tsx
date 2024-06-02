"use client";
import { cn } from "lib";
import { useDroppable } from "@dnd-kit/core";
import { usePathname } from "next/navigation";
import { Text } from "ui";
import type { StoryStatus, Story, StoryPriority } from "@/types/story";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";
import { StoriesHeader } from "./stories-header";
import { StoriesList } from "./stories-list";
import { RowWrapper } from "./row-wrapper";

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
  const pathname = usePathname();
  const { groupBy, showEmptyGroups } = viewOptions;
  const id = (groupBy === "Status" ? status : priority) as string;
  const collapseKey = pathname + id;
  const [isCollapsed, setIsCollapsed] = useLocalStorage(collapseKey, false);
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
        isCollapsed={isCollapsed}
        priority={priority}
        setIsCollapsed={setIsCollapsed}
        status={status}
      />
      {!isCollapsed && <StoriesList stories={filteredStories} />}
      {!isCollapsed && (
        <RowWrapper>
          <Text color="muted">
            Showing <b>{filteredStories.length}</b> stor
            {filteredStories.length === 1 ? "y" : "ies"} with{" "}
            {groupBy.toLowerCase()} <b>{id}</b>
          </Text>
        </RowWrapper>
      )}
    </div>
  );
};

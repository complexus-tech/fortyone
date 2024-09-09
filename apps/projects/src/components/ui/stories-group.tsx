"use client";
import { cn } from "lib";
import { useDroppable } from "@dnd-kit/core";
import { usePathname } from "next/navigation";
import { Text } from "ui";
import type { Story, StoryPriority } from "@/modules/stories/types";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage } from "@/hooks";
import { StoriesHeader } from "./stories-header";
import { StoriesList } from "./stories-list";
import { RowWrapper } from "./row-wrapper";
import { State, StateCategory } from "@/types/states";
import { useStatuses } from "@/lib/hooks/statuses";

export const StoriesGroup = ({
  stories,
  status,
  priority,
  className,
  viewOptions,
}: {
  stories: Story[];
  status?: State;
  priority?: StoryPriority;
  className?: string;
  viewOptions: StoriesViewOptions;
}) => {
  const pathname = usePathname();
  const { data: statuses = [] } = useStatuses();
  const { id: defaultStatusId } = statuses.at(0)!!;
  const { groupBy, showEmptyGroups } = viewOptions;
  const id = (groupBy === "Status" ? status?.id : priority) as string;
  const collapseKey = pathname + id;
  const defaultClosedStatuses: StateCategory[] = [
    "cancelled",
    "completed",
    "paused",
  ];

  const getDefaultCollapsed = () => {
    if (
      groupBy === "Status" &&
      status &&
      defaultClosedStatuses.includes(status?.category)
    ) {
      return true;
    }
    return false;
  };

  const [isCollapsed, setIsCollapsed] = useLocalStorage(
    collapseKey,
    getDefaultCollapsed(),
  );
  const { isOver, setNodeRef } = useDroppable({
    id,
  });

  const mappedStories = stories.map(({ statusId, priority, ...rest }) => ({
    ...rest,
    priority: priority ?? "No Priority",
    statusId: statusId ?? defaultStatusId,
  }));
  const filteredStories =
    groupBy === "Status"
      ? mappedStories.filter((story) => story.statusId === status?.id)
      : mappedStories.filter((story) => story.priority === priority);

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
        stories={filteredStories}
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
            {groupBy.toLowerCase()}{" "}
            <b>{groupBy === "Status" ? status?.name : priority}</b>
          </Text>
        </RowWrapper>
      )}
    </div>
  );
};

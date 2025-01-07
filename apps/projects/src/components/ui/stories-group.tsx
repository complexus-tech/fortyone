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
import { Member } from "@/types";

export const StoriesGroup = ({
  stories,
  status,
  priority,
  className,
  viewOptions,
  assignee,
}: {
  stories: Story[];
  status?: State;
  priority?: StoryPriority;
  assignee?: Member;
  className?: string;
  viewOptions: StoriesViewOptions;
}) => {
  const pathname = usePathname();
  const { data: statuses = [] } = useStatuses();
  const { id: defaultStatusId } = statuses.at(0)!!;
  const { groupBy, showEmptyGroups } = viewOptions;
  const id = (
    groupBy === "Status"
      ? status?.id
      : groupBy === "Assignee"
        ? assignee?.id
        : priority
  ) as string;

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
  const filteredStories = mappedStories.filter((story) => {
    if (groupBy === "Status") return story.statusId === id;
    if (groupBy === "Assignee") return story.assigneeId === id;
    if (groupBy === "Priority") return story.priority === id;
    return false;
  });

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
        assignee={assignee}
      />
      {!isCollapsed && <StoriesList stories={filteredStories} />}
      {!isCollapsed && (
        <RowWrapper>
          <Text color="muted">
            Showing{" "}
            <span className="font-semibold">{filteredStories.length}</span> stor
            {filteredStories.length === 1 ? "y" : "ies"} with{" "}
            {groupBy.toLowerCase()}{" "}
            <span className="font-semibold">
              {groupBy === "Status"
                ? status?.name
                : groupBy === "Assignee"
                  ? assignee?.username || "Unassigned"
                  : priority}
            </span>
          </Text>
        </RowWrapper>
      )}
    </div>
  );
};

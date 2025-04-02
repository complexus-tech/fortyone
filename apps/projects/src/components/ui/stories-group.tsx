"use client";
import { cn } from "lib";
import { useDroppable } from "@dnd-kit/core";
import { usePathname } from "next/navigation";
import { Text } from "ui";
import type { Story, StoryPriority } from "@/modules/stories/types";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage, useTerminology } from "@/hooks";
import type { State, StateCategory } from "@/types/states";
import { useStatuses } from "@/lib/hooks/statuses";
import type { Member } from "@/types";
import { StoriesHeader } from "./stories-header";
import { StoriesList } from "./stories-list";
import { RowWrapper } from "./row-wrapper";

const getId = ({
  groupBy,
  status,
  assignee,
  priority,
}: {
  groupBy: StoriesViewOptions["groupBy"];
  status?: State;
  assignee?: Member;
  priority?: StoryPriority;
}) => {
  if (groupBy === "Status") return status?.id;
  if (groupBy === "Assignee") return assignee?.id;
  return priority;
};

const getGroupLabel = ({
  groupBy,
  status,
  assignee,
  priority,
}: {
  groupBy: StoriesViewOptions["groupBy"];
  status?: State;
  assignee?: Member;
  priority?: StoryPriority;
}) => {
  if (groupBy === "Status") return status?.name;
  if (groupBy === "Assignee") return assignee?.username || "Unassigned";
  return priority;
};

export const StoriesGroup = ({
  isInSearch,
  stories,
  status,
  priority,
  className,
  viewOptions,
  assignee,
}: {
  isInSearch?: boolean;
  stories: Story[];
  status?: State;
  priority?: StoryPriority;
  assignee?: Member;
  className?: string;
  viewOptions: StoriesViewOptions;
}) => {
  const { getTermDisplay } = useTerminology();
  const pathname = usePathname();
  const { data: statuses = [] } = useStatuses();
  const { id: defaultStatusId } = statuses.at(0)!;
  const { groupBy, showEmptyGroups } = viewOptions;
  const id = getId({ groupBy, status, assignee, priority })!;

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
      defaultClosedStatuses.includes(status.category)
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
    priority,
    statusId: statusId || defaultStatusId,
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
        assignee={assignee}
        className={className}
        groupBy={groupBy}
        isCollapsed={isCollapsed}
        priority={priority}
        setIsCollapsed={setIsCollapsed}
        status={status}
        stories={filteredStories}
      />
      {!isCollapsed && (
        <StoriesList isInSearch={isInSearch} stories={filteredStories} />
      )}
      {!isCollapsed && (
        <RowWrapper>
          <Text color="muted">
            Showing{" "}
            <span className="font-semibold">
              {filteredStories.length}{" "}
              {getTermDisplay("storyTerm", {
                variant: filteredStories.length === 1 ? "singular" : "plural",
              })}
            </span>{" "}
            with {groupBy.toLowerCase()}{" "}
            <span className="font-semibold">
              {getGroupLabel({ groupBy, status, assignee, priority })}
            </span>
          </Text>
        </RowWrapper>
      )}
    </div>
  );
};

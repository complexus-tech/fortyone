"use client";
import { cn } from "lib";
import { useDroppable } from "@dnd-kit/core";
import { usePathname } from "next/navigation";
import { Text } from "ui";
import type { Story, StoryPriority } from "@/modules/stories/types";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage, useTerminology } from "@/hooks";
import type { Member } from "@/types";
import type { State } from "@/types/states";
import { StoriesHeader } from "./stories-header";
import { StoriesList } from "./stories-list";
import { RowWrapper } from "./row-wrapper";

const getGroupLabel = ({
  groupBy,
  assignee,
  priority,
  status,
}: {
  groupBy: StoriesViewOptions["groupBy"];
  assignee?: Member;
  priority?: StoryPriority;
  status?: State;
}) => {
  if (groupBy === "status") return status?.name;
  if (groupBy === "assignee") return assignee?.username || "Unassigned";
  return priority;
};

export const StoriesGroup = ({
  isInSearch,
  id,
  stories,
  status,
  priority,
  className,
  viewOptions,
  assignee,
  rowClassName,
}: {
  id: string;
  isInSearch?: boolean;
  stories: Story[];
  status?: State;
  priority?: StoryPriority;
  assignee?: Member;
  className?: string;
  viewOptions: StoriesViewOptions;
  rowClassName?: string;
}) => {
  const { getTermDisplay } = useTerminology();
  const pathname = usePathname();
  const { groupBy, showEmptyGroups } = viewOptions;

  const collapseKey = pathname + id;

  const [isCollapsed, setIsCollapsed] = useLocalStorage(collapseKey, false);
  const { isOver, setNodeRef, active } = useDroppable({
    id,
  });

  return (
    <div
      className={cn("border-0 border-transparent transition", {
        "border border-primary": isOver && active?.id,
        hidden: !showEmptyGroups && stories.length === 0,
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
        stories={stories}
      />
      {!isCollapsed && (
        <StoriesList
          isInSearch={isInSearch}
          rowClassName={rowClassName}
          stories={stories}
        />
      )}
      {!isCollapsed && (
        <RowWrapper>
          <Text color="muted">
            Showing{" "}
            <span className="font-semibold">
              {stories.length}{" "}
              {getTermDisplay("storyTerm", {
                variant: stories.length === 1 ? "singular" : "plural",
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

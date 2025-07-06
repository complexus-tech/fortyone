"use client";
import { cn } from "lib";
import { useDroppable } from "@dnd-kit/core";
import { usePathname } from "next/navigation";
import { Text, Checkbox, Flex, Skeleton } from "ui";
import type {
  StoryGroup,
  StoryPriority,
  GroupStoryParams,
} from "@/modules/stories/types";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { useLocalStorage, useTerminology } from "@/hooks";
import type { Member } from "@/types";
import type { State } from "@/types/states";
import { useGroupStoriesInfinite } from "@/modules/stories/hooks/use-group-stories-infinite";
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
  group,
  status,
  priority,
  className,
  viewOptions,
  assignee,
  rowClassName,
}: {
  id: string;
  isInSearch?: boolean;
  group: StoryGroup;
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
  const { isOver, setNodeRef, active } = useDroppable({ id });

  const params: GroupStoryParams = {
    groupKey: group.key,
    groupBy,
    // Add other params as needed from viewOptions/filters
  };

  const {
    data: infiniteData,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
  } = useGroupStoriesInfinite(params, group);

  const allStories = infiniteData.pages.flatMap((page) => page.stories);

  const handleLoadMore = () => {
    fetchNextPage();
  };

  return (
    <div
      className={cn("border-0 border-transparent transition", {
        "border border-primary": isOver && active?.id,
        hidden: !showEmptyGroups && group.loadedCount === 0,
      })}
      ref={setNodeRef}
    >
      <StoriesHeader
        assignee={assignee}
        className={className}
        group={group}
        groupBy={groupBy}
        isCollapsed={isCollapsed}
        priority={priority}
        setIsCollapsed={setIsCollapsed}
        status={status}
      />
      {!isCollapsed && (
        <StoriesList
          isInSearch={isInSearch}
          rowClassName={rowClassName}
          stories={allStories}
        />
      )}
      {isFetchingNextPage
        ? Array.from({ length: 8 }).map((_, i) => (
            <RowWrapper className="pointer-events-none relative gap-4" key={i}>
              <Checkbox className="absolute left-6 hidden opacity-70 md:block" />
              <Flex align="center" className="relative shrink pl-1" gap={3}>
                <Skeleton className="h-5 w-10" />
                <Skeleton
                  className={cn("h-5 w-24 md:w-32", {
                    "w-40 md:w-56": i % 2 === 0,
                  })}
                />
              </Flex>
              <Flex align="center" className="shrink-0" gap={4}>
                <Skeleton className="h-5 w-20" />
                <Skeleton className="hidden h-5 w-16 md:block" />
                <Skeleton className="hidden h-5 w-20 md:block" />
                <Skeleton className="size-8 rounded-full" />
              </Flex>
            </RowWrapper>
          ))
        : null}
      {!isCollapsed && (
        <RowWrapper className="grid h-12 py-0 md:grid-cols-3">
          {hasNextPage ? (
            <button
              className="flex hover:underline"
              disabled={isFetchingNextPage}
              onClick={handleLoadMore}
              type="button"
            >
              {isFetchingNextPage
                ? "Loading..."
                : `Load more ${getTermDisplay("storyTerm", { variant: "plural" })}`}
            </button>
          ) : (
            <Text color="muted">
              Showing{" "}
              <span className="font-semibold">
                {allStories.length}{" "}
                {getTermDisplay("storyTerm", {
                  variant: allStories.length === 1 ? "singular" : "plural",
                })}{" "}
              </span>{" "}
              with {groupBy.toLowerCase()}{" "}
              <span className="font-semibold">
                {getGroupLabel({ groupBy, status, assignee, priority })}
              </span>
            </Text>
          )}
        </RowWrapper>
      )}
    </div>
  );
};

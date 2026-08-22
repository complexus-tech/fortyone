"use client";
import { useEffect, useRef, useState, type ReactNode } from "react";
import { useDroppable } from "@dnd-kit/core";
import { useIntersectionObserver } from "react-intersection-observer-hook";
import type { IntersectionObserverHookRootRefCallback } from "react-intersection-observer-hook";
import { cn } from "lib";
import { Box, Button, Skeleton } from "ui";
import { PlusIcon } from "icons";
import type {
  Story,
  StoryGroup,
  StoryPriority,
  GroupStoryParams,
  GroupedStoriesResponse,
} from "@/modules/stories/types";
import type { State } from "@/types/states";
import type { Member } from "@/types";
import { useTerminology } from "@/hooks";
import { useGroupStoriesInfinite } from "@/modules/stories/hooks/use-group-stories-infinite";
import { StoryCard } from "./story/card";
import type { ViewOptionsGroupBy } from "./stories-view-options-button";
import { NewStoryDialog } from "./new-story-dialog";
import { useBoard } from "./board-context";
import { StoryDialog } from "./story-dialog";
import { groupFilters } from "./group-filters";

const List = ({
  children,
  id,
  scrollRootRef,
  totalStories,
}: {
  children: ReactNode;
  id: string | number;
  scrollRootRef: IntersectionObserverHookRootRefCallback;
  totalStories: number;
}) => {
  const { viewOptions } = useBoard();
  const { showEmptyGroups } = viewOptions;
  const { isOver, setNodeRef } = useDroppable({
    id,
  });

  return (
    <Box
      className={cn("h-full min-h-0 w-[340px] shrink-0", {
        hidden: totalStories === 0 && !showEmptyGroups,
      })}
    >
      <div
        className={cn("h-full min-h-0 w-[340px] rounded-md transition", {
          "bg-surface-muted/50": totalStories === 0,
          "bg-surface-muted": isOver,
        })}
        ref={setNodeRef}
      >
        <div
          className="flex h-full min-h-0 w-full flex-col gap-3 overflow-y-auto overscroll-y-contain rounded-md pb-6"
          ref={scrollRootRef}
        >
          {children}
        </div>
      </div>
    </Box>
  );
};

export const KanbanGroup = ({
  group,
  meta,
  status,
  priority,
  member,
  groupBy = "status",
}: {
  group: StoryGroup;
  meta: GroupedStoriesResponse["meta"];
  status?: State;
  priority?: StoryPriority;
  member?: Member;
  groupBy: ViewOptionsGroupBy;
}) => {
  const { getTermDisplay } = useTerminology();
  const { newStoryDefaults } = useBoard();
  const [isOpen, setIsOpen] = useState(false);

  const [storyId, setStoryId] = useState<string | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const params: GroupStoryParams = {
    groupKey: group.key,
    ...groupFilters(meta),
    groupBy,
  };
  const paginationScope = JSON.stringify(params);

  const {
    data: infiniteData,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
  } = useGroupStoriesInfinite(params, group);

  const uniqueStories = new Map<string, Story>();
  for (const page of infiniteData.pages) {
    for (const story of page.stories) {
      uniqueStories.set(story.id, story);
    }
  }
  const allStories = Array.from(uniqueStories.values());

  const [triggerRef, { entry, rootRef }] = useIntersectionObserver({
    threshold: 0,
    rootMargin: "0px 0px 300px",
  });
  const lastRequestRef = useRef<{
    entry: IntersectionObserverEntry;
    paginationScope: string;
  } | null>(null);

  useEffect(() => {
    if (!entry?.isIntersecting) {
      lastRequestRef.current = null;
      return;
    }

    if (
      (lastRequestRef.current?.entry === entry &&
        lastRequestRef.current.paginationScope === paginationScope) ||
      !hasNextPage ||
      isFetchingNextPage
    ) {
      return;
    }

    lastRequestRef.current = { entry, paginationScope };
    void fetchNextPage();
  }, [entry, fetchNextPage, hasNextPage, isFetchingNextPage, paginationScope]);

  const handleNavigate = (newStoryId: string) => {
    setStoryId(newStoryId);
  };

  return (
    <List
      id={group.key}
      key={group.key}
      scrollRootRef={rootRef}
      totalStories={allStories.length}
    >
      {allStories.map((story) => (
        <StoryCard
          handleStoryClick={(storyId) => {
            setStoryId(storyId);
            setIsDialogOpen(true);
          }}
          key={story.id}
          story={story}
        />
      ))}

      {hasNextPage ? <div className="h-6 w-full" ref={triggerRef} /> : null}

      {isFetchingNextPage ? (
        <Box>
          {Array.from({ length: 6 }).map((_, index) => (
            <Skeleton
              className="mb-2 h-28 shadow-sm dark:shadow-none"
              key={index}
            />
          ))}
        </Box>
      ) : null}

      <Button
        align="center"
        className="border-border bg-surface-muted relative min-h-[2.35rem] w-[340px]"
        color="tertiary"
        fullWidth
        onClick={() => {
          setIsOpen(true);
        }}
        size="sm"
      >
        <PlusIcon className="relative -top-[0.3px] h-[1.15rem] w-auto" /> New{" "}
        {getTermDisplay("storyTerm", { capitalize: true })}
      </Button>

      <NewStoryDialog
        assigneeId={groupBy === "assignee" ? member?.id ?? null : undefined}
        isOpen={isOpen}
        objectiveId={newStoryDefaults.objectiveId}
        priority={priority}
        setIsOpen={setIsOpen}
        sprintId={newStoryDefaults.sprintId}
        statusId={status?.id}
        teamId={newStoryDefaults.teamId}
      />
      {storyId ? (
        <StoryDialog
          isOpen={isDialogOpen}
          onNavigate={handleNavigate}
          setIsOpen={setIsDialogOpen}
          stories={allStories}
          storyId={storyId}
        />
      ) : null}
    </List>
  );
};

"use client";
import { useState, type ReactNode } from "react";
import { useDroppable } from "@dnd-kit/core";
import { cn } from "lib";
import { Box, Button } from "ui";
import { PlusIcon } from "icons";
import type {
  StoryGroup,
  StoryPriority,
  GroupStoryParams,
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

const List = ({
  children,
  id,
  totalStories,
}: {
  children: ReactNode;
  id: string | number;
  totalStories: number;
}) => {
  const { viewOptions } = useBoard();
  const { showEmptyGroups } = viewOptions;
  const { isOver, setNodeRef } = useDroppable({
    id,
  });
  return (
    <Box
      className={cn({
        hidden: totalStories === 0 && !showEmptyGroups,
      })}
    >
      <div
        className={cn(
          "flex h-full w-[340px] flex-col gap-4 overflow-y-auto rounded-[0.45rem] pb-6 transition",
          {
            "bg-gray-100/20 dark:bg-dark-200/10": totalStories === 0,
            "bg-gray-100/40 dark:bg-dark-200/50": isOver,
          },
        )}
        ref={setNodeRef}
      >
        {children}
      </div>
    </Box>
  );
};

export const KanbanGroup = ({
  group,
  status,
  priority,
  member,
  groupBy = "status",
}: {
  group: StoryGroup;
  status?: State;
  priority?: StoryPriority;
  member?: Member;
  groupBy: ViewOptionsGroupBy;
}) => {
  const { getTermDisplay } = useTerminology();
  const [isOpen, setIsOpen] = useState(false);

  const getId = () => {
    if (groupBy === "status") return status?.id;
    if (groupBy === "assignee") return member?.id;
    return priority;
  };

  const id = getId() || "";
  const [storyId, setStoryId] = useState<string | null>(null);
  const [isDialogOpen, setIsDialogOpen] = useState(false);

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

  const handleNavigate = (newStoryId: string) => {
    setStoryId(newStoryId);
  };

  const handleLoadMore = () => {
    fetchNextPage();
  };

  return (
    <List id={id} key={id} totalStories={allStories.length}>
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

      {hasNextPage ? (
        <Button
          align="center"
          className="relative min-h-[2.35rem] w-[340px] border-gray-100/80 dark:border-dark-200 dark:bg-dark-200/60"
          color="tertiary"
          disabled={isFetchingNextPage}
          fullWidth
          onClick={handleLoadMore}
          size="sm"
        >
          {isFetchingNextPage
            ? "Loading..."
            : `Load more ${getTermDisplay("storyTerm", { variant: "plural" })}`}
        </Button>
      ) : null}

      <Button
        align="center"
        className="relative min-h-[2.35rem] w-[340px] border-gray-100/80 dark:border-dark-200 dark:bg-dark-200/60"
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
        assigneeId={member?.id}
        isOpen={isOpen}
        priority={priority}
        setIsOpen={setIsOpen}
        statusId={status?.id}
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

"use client";
import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core";
import { DndContext, DragOverlay } from "@dnd-kit/core";
import { Flex, Text } from "ui";
import { useState } from "react";
import { createPortal } from "react-dom";
import type {
  Story,
  StoryPriority,
  StoryStatus,
} from "@/modules/stories/types";
import type {
  DisplayColumn,
  StoriesViewOptions,
} from "@/components/ui/stories-view-options-button";
import { KanbanBoard } from "./kanban-board";
import { StoryStatusIcon } from "./story-status-icon";
import { StoryCard } from "./story/card";
import { ListBoard } from "./list-board";
import { StoriesToolbar } from "./stories-toolbar";
import { BoardContext } from "./board-context";

export type StoriesLayout = "list" | "kanban" | null;

const StoryOverlay = ({
  story,
  layout,
  selectedStories = 0,
}: {
  story: Story | null;
  layout: StoriesLayout;
  selectedStories: number;
}) => {
  return (
    <DragOverlay
      className="pointer-events-none"
      dropAnimation={{
        duration: 300,
        easing: "cubic-bezier(0.18, 0.67, 0.6, 1.22)",
      }}
    >
      {layout === "kanban" ? (
        <StoryCard
          className="border-gray-100 shadow-lg dark:border-dark-50/60 dark:shadow-dark"
          story={story!}
        />
      ) : (
        <Flex
          align="center"
          className="w-max rounded-lg border border-gray-100 bg-gray-50/70 px-3 py-3.5 shadow backdrop-blur dark:border-dark-100 dark:bg-dark-200/70"
          gap={2}
        >
          {selectedStories > 1 ? (
            <Text className="w-60 truncate pl-2" fontWeight="medium">
              {selectedStories} stories selected
            </Text>
          ) : (
            <>
              <StoryStatusIcon status={story?.status} />
              <Text color="muted">COM-{story?.id}</Text>
              <Text className="max-w-xs truncate" fontWeight="medium">
                {story?.title}
              </Text>
            </>
          )}
        </Flex>
      )}
    </DragOverlay>
  );
};

export const StoriesBoard = ({
  layout,
  stories,
  className,
  viewOptions,
}: {
  layout: StoriesLayout;
  stories: Story[];
  className?: string;
  viewOptions: StoriesViewOptions;
}) => {
  const [storiesBoard, setStoriesBoard] = useState<Story[]>(stories);
  const [activeStory, setActiveStory] = useState<Story | null>(null);
  const [selectedStories, setSelectedStories] = useState<number[]>([]);

  const isColumnVisible = (column: DisplayColumn) =>
    viewOptions.displayColumns.includes(column);

  const handleDragStart = (e: DragStartEvent) => {
    const story = stories.find(({ id }) => id === Number(e.active.id))!;
    setActiveStory(story);
  };

  const handleDragEnd = (e: DragEndEvent) => {
    const { groupBy } = viewOptions;
    if (groupBy === "Status") {
      const newStatus = e.over?.id as StoryStatus | null;
      if (newStatus) {
        const index = storiesBoard.findIndex(
          ({ id }) => id === Number(e.active.id),
        );
        storiesBoard[index].status = newStatus;
        setStoriesBoard([...storiesBoard]);
      }
    }

    if (groupBy === "Priority") {
      const newPriority = e.over?.id as StoryPriority | null;
      if (newPriority) {
        const index = storiesBoard.findIndex(
          ({ id }) => id === Number(e.active.id),
        );
        storiesBoard[index].priority = newPriority;
        setStoriesBoard([...storiesBoard]);
      }
    }

    setActiveStory(null);
  };

  return (
    <BoardContext.Provider
      value={{
        selectedStories,
        setSelectedStories,
        viewOptions,
        isColumnVisible,
      }}
    >
      <DndContext onDragEnd={handleDragEnd} onDragStart={handleDragStart}>
        {layout === "kanban" ? (
          <KanbanBoard className={className} stories={storiesBoard} />
        ) : (
          <ListBoard
            className={className}
            stories={storiesBoard}
            viewOptions={viewOptions}
          />
        )}

        {typeof window !== "undefined" &&
          createPortal(
            <StoryOverlay
              layout={layout}
              selectedStories={selectedStories.length}
              story={activeStory}
            />,
            document.body,
          )}

        {/* This toolbar pops up when the user selects stories */}
        {selectedStories.length > 0 && <StoriesToolbar />}
      </DndContext>
    </BoardContext.Provider>
  );
};

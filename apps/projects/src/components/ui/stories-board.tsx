"use client";
import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core";
import { DndContext, DragOverlay } from "@dnd-kit/core";
import { Flex, Text } from "ui";
import { useEffect, useState } from "react";
import { createPortal } from "react-dom";
import type { Story, StoryPriority } from "@/modules/stories/types";
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
import { DetailedStory } from "@/modules/story/types";
import { updateStoryAction } from "@/modules/story/actions/update-story";
import { toast } from "sonner";
import { useStore } from "@/hooks/store";

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
  const { teams } = useStore();
  const team = teams.find(({ id }) => id === story?.teamId);
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
              <StoryStatusIcon statusId={story?.statusId} />
              <Text color="muted">
                {team?.code}-{story?.sequenceId}
              </Text>
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
  const [selectedStories, setSelectedStories] = useState<string[]>([]);

  const isColumnVisible = (column: DisplayColumn) =>
    viewOptions.displayColumns.includes(column);

  const handleDragStart = (e: DragStartEvent) => {
    const story = stories.find(({ id }) => id === e.active.id)!;
    setActiveStory(story);
  };

  const updateStory = async (storyId: string, data: Partial<DetailedStory>) => {
    try {
      const _ = await updateStoryAction(storyId, data);
    } catch (error) {
      toast.error("Error updating story", {
        description: "Failed to update story, reverted changes.",
      });
    }
  };

  const handleDragEnd = (e: DragEndEvent) => {
    const { groupBy } = viewOptions;
    if (groupBy === "Status") {
      const newStatus = e.over?.id.toString() ?? null;
      if (newStatus) {
        const index = storiesBoard.findIndex(({ id }) => id === e.active.id);
        storiesBoard[index].statusId = newStatus;
        setStoriesBoard([...storiesBoard]);
        updateStory(e.active.id.toString(), {
          statusId: newStatus,
        });
      }
    }

    if (groupBy === "Priority") {
      const newPriority = e.over?.id as StoryPriority | null;
      if (newPriority) {
        const index = storiesBoard.findIndex(({ id }) => id === e.active.id);
        storiesBoard[index].priority = newPriority;
        setStoriesBoard([...storiesBoard]);
        updateStory(e.active.id.toString(), {
          priority: newPriority,
        });
      }
    }
    setActiveStory(null);
  };

  const orderStories = (stories: Story[] = []) => {
    const getSortValue = (story: Story): number => {
      switch (viewOptions.orderBy) {
        case "Created":
          return new Date(story.createdAt).getTime();
        case "Updated":
          return new Date(story.updatedAt).getTime();
        case "Due date":
          const date = new Date(story.endDate);
          return isNaN(date.getTime()) ? Infinity : date.getTime();
        case "Priority":
          const prioritiesMap: Record<StoryPriority, number> = {
            "No Priority": 0,
            Low: 1,
            Medium: 2,
            High: 3,
            Urgent: 4,
          };
          return (
            prioritiesMap[story.priority] * 1e15 -
            new Date(story.createdAt).getTime()
          );
        default:
          return 0;
      }
    };

    return stories.sort((a, b) => getSortValue(b) - getSortValue(a));
  };

  useEffect(() => {
    setStoriesBoard(stories);
  }, [stories]);

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
          <KanbanBoard
            className={className}
            stories={orderStories(storiesBoard)}
          />
        ) : (
          <ListBoard
            className={className}
            stories={orderStories(storiesBoard)}
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

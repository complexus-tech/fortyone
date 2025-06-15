"use client";
import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core";
import {
  DndContext,
  DragOverlay,
  useSensor,
  useSensors,
  MouseSensor,
  PointerSensor,
  TouchSensor,
} from "@dnd-kit/core";
import { Box, Flex, Text } from "ui";
import { useEffect, useState, useMemo, useCallback } from "react";
import { createPortal } from "react-dom";
import { PlusIcon, StoryMissingIcon } from "icons";
import { useParams } from "next/navigation";
import type {
  GroupedStoriesResponse,
  Story,
  StoryPriority,
} from "@/modules/stories/types";
import type {
  DisplayColumn,
  StoriesViewOptions,
} from "@/components/ui/stories-view-options-button";
import type { DetailedStory } from "@/modules/story/types";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useFeatures, useTerminology } from "@/hooks";
import { KanbanBoard } from "./kanban-board";
import { StoryStatusIcon } from "./story-status-icon";
import { StoryCard } from "./story/card";
import { ListBoard } from "./list-board";
import { GanttBoard } from "./gantt-board";
import { StoriesToolbar } from "./stories-toolbar";
import { BoardContext } from "./board-context";
import { NewStoryButton } from "./new-story-button";

export type StoriesLayout = "list" | "kanban" | "gantt" | null;

const StoryOverlay = ({
  story,
  layout,
  selectedStories = 0,
}: {
  story: Story | null;
  layout: StoriesLayout;
  selectedStories: number;
}) => {
  const { data: teams = [] } = useTeams();
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
          handleStoryClick={() => {}}
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

// Move orderStories out of the component for better performance
const orderStories = (stories: Story[] = [], orderBy = ""): Story[] => {
  const getSortValue = (story: Story): number => {
    switch (orderBy) {
      case "Created":
        return new Date(story.createdAt).getTime();
      case "Updated":
        return new Date(story.updatedAt).getTime();
      case "Deadline": {
        const date = new Date(story.endDate!);
        return isNaN(date.getTime()) ? Infinity : date.getTime();
      }
      case "Priority": {
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
      }
      default:
        return 0;
    }
  };

  return [...stories].sort((a, b) => getSortValue(b) - getSortValue(a));
};

// Move EmptyState component outside StoriesBoard to prevent recreation on each render
const EmptyState = ({
  objectiveId,
  sprintId,
  teamId,
  getTermDisplay,
}: {
  objectiveId?: string;
  sprintId?: string;
  teamId?: string;
  getTermDisplay: ReturnType<typeof useTerminology>["getTermDisplay"];
}) => (
  <Box className="flex h-[70vh] items-center justify-center">
    <Box className="flex flex-col items-center">
      <StoryMissingIcon className="h-20 w-auto rotate-12" strokeWidth={1.3} />
      <Text className="mb-6 mt-8" fontSize="3xl">
        No {getTermDisplay("storyTerm", { variant: "plural" })} found
      </Text>
      <Text className="mb-6 max-w-md text-center" color="muted">
        Oops! This board is empty. Why not create a{" "}
        {getTermDisplay("storyTerm")}?
      </Text>
      <Flex gap={2}>
        <NewStoryButton
          color="tertiary"
          leftIcon={<PlusIcon />}
          objectiveId={objectiveId}
          size="md"
          sprintId={sprintId}
          teamId={teamId}
        >
          Create new {getTermDisplay("storyTerm")}
        </NewStoryButton>
      </Flex>
    </Box>
  </Box>
);

export const StoriesBoard = ({
  isInSearch,
  layout,
  stories,
  groupedStories,
  className,
  viewOptions,
}: {
  isInSearch?: boolean;
  layout: StoriesLayout;
  stories: Story[];
  groupedStories?: GroupedStoriesResponse;
  className?: string;
  viewOptions: StoriesViewOptions;
}) => {
  const { getTermDisplay } = useTerminology();
  const { objectiveId, sprintId, teamId } = useParams<{
    objectiveId: string;
    sprintId: string;
    teamId: string;
  }>();
  const features = useFeatures();
  const [storiesBoard, setStoriesBoard] = useState<Story[]>(stories);
  const [activeStory, setActiveStory] = useState<Story | null>(null);
  const [selectedStories, setSelectedStories] = useState<string[]>([]);

  const { mutate } = useUpdateStoryMutation();

  // Memoize the isColumnVisible function
  const isColumnVisible = useCallback(
    (column: DisplayColumn) => {
      if (column === "Sprint" && !features.sprintEnabled) return false;
      if (column === "Objective" && !features.objectiveEnabled) return false;
      return viewOptions.displayColumns.includes(column);
    },
    [
      features.objectiveEnabled,
      features.sprintEnabled,
      viewOptions.displayColumns,
    ],
  );

  const handleDragStart = (e: DragStartEvent) => {
    const story = stories.find(({ id }) => id === e.active.id)!;
    setActiveStory(story);
  };

  const handleDragEnd = useCallback(
    (e: DragEndEvent) => {
      const updateStory = (
        storyId: string,
        payload: Partial<DetailedStory>,
      ) => {
        mutate({ storyId, payload });
      };

      const { groupBy } = viewOptions;

      if (e.over) {
        const storyId = e.active.id.toString();
        const updatePayload: Partial<DetailedStory> = {};

        if (groupBy === "status") {
          const newStatus = e.over.id.toString();
          updatePayload.statusId = newStatus;

          setStoriesBoard((prev) =>
            prev.map((story) =>
              story.id === storyId ? { ...story, statusId: newStatus } : story,
            ),
          );
        }

        if (groupBy === "priority") {
          const newPriority = e.over.id as StoryPriority;
          updatePayload.priority = newPriority;

          setStoriesBoard((prev) =>
            prev.map((story) =>
              story.id === storyId
                ? { ...story, priority: newPriority }
                : story,
            ),
          );
        }

        if (groupBy === "assignee") {
          const newAssignee = e.over.id as string;
          updatePayload.assigneeId = newAssignee;

          setStoriesBoard((prev) =>
            prev.map((story) =>
              story.id === storyId
                ? { ...story, assigneeId: newAssignee }
                : story,
            ),
          );
        }

        // Only call the API if we have updates
        if (Object.keys(updatePayload).length > 0) {
          updateStory(storyId, updatePayload);
        }
      }

      setActiveStory(null);
    },
    [viewOptions, mutate],
  );

  // Memoize the ordered stories
  const orderedStories = useMemo(
    () => orderStories(storiesBoard, viewOptions.orderBy),
    [storiesBoard, viewOptions.orderBy],
  );

  const mouseSensor = useSensor(MouseSensor, {
    activationConstraint: {
      distance: 8,
    },
  });
  const pointerSensor = useSensor(PointerSensor, {
    activationConstraint: {
      distance: 8,
    },
  });

  const touchSensor = useSensor(TouchSensor, {
    activationConstraint: {
      delay: 200,
      tolerance: 8,
    },
  });

  const sensors = useSensors(mouseSensor, pointerSensor, touchSensor);

  useEffect(() => {
    setStoriesBoard(stories);
  }, [stories]);

  // Memoize the context value to prevent unnecessary re-renders
  const boardContextValue = useMemo(
    () => ({
      selectedStories,
      setSelectedStories,
      viewOptions,
      isColumnVisible,
    }),
    [selectedStories, viewOptions, isColumnVisible],
  );

  return (
    <BoardContext.Provider value={boardContextValue}>
      <Box>
        {storiesBoard.length === 0 && !isInSearch && (
          <EmptyState
            getTermDisplay={getTermDisplay}
            objectiveId={objectiveId}
            sprintId={sprintId}
            teamId={teamId}
          />
        )}

        {storiesBoard.length > 0 && (
          <DndContext
            onDragEnd={handleDragEnd}
            onDragStart={handleDragStart}
            sensors={sensors}
          >
            {layout === "kanban" && (
              <KanbanBoard className={className} stories={orderedStories} />
            )}
            {layout === "gantt" && (
              <GanttBoard className={className} stories={orderedStories} />
            )}
            {(layout === "list" || !layout) && (
              <ListBoard
                className={className}
                groupedStories={groupedStories!}
                isInSearch={isInSearch}
                viewOptions={viewOptions}
              />
            )}

            {activeStory && typeof window !== "undefined"
              ? createPortal(
                  <StoryOverlay
                    layout={layout}
                    selectedStories={selectedStories.length}
                    story={activeStory}
                  />,
                  document.body,
                )
              : null}
          </DndContext>
        )}

        {/* This toolbar pops up when the user selects stories */}
        {selectedStories.length > 0 && <StoriesToolbar />}
      </Box>
    </BoardContext.Provider>
  );
};

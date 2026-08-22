"use client";
import type {
  CollisionDetection,
  DragEndEvent,
  DragStartEvent,
} from "@dnd-kit/core";
import {
  DndContext,
  DragOverlay,
  pointerWithin,
  rectIntersection,
  useSensor,
  useSensors,
  PointerSensor,
  TouchSensor,
} from "@dnd-kit/core";
import { Box, Flex, Text } from "ui";
import { useState, useMemo, useCallback } from "react";
import type { ReactNode } from "react";
import { createPortal } from "react-dom";
import { PlusIcon } from "icons";
import { useParams } from "next/navigation";
import { cn } from "lib";
import type { GroupedStoriesResponse, Story } from "@/modules/stories/types";
import type {
  DisplayColumn,
  StoriesViewOptions,
} from "@/components/ui/stories-view-options-button";
import type { DetailedStory } from "@/modules/story/types";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useFeatures, useTerminology, useSprintsEnabled } from "@/hooks";
import { StoriesList } from "@/components/ui/stories-list";
import { BodyContainer } from "@/components/shared/body";
import { KanbanBoard } from "./kanban-board";
import { StoryStatusIcon } from "./story-status-icon";
import { StoryCardPreview } from "./story/card";
import { ListBoard } from "./list-board";
import { GanttBoard } from "./gantt-board";
import { StoriesToolbar } from "./stories-toolbar";
import { BoardContext } from "./board-context";
import { NewStoryButton } from "./new-story-button";
import { getStoryDropUpdate } from "./story-drag";
import { StoriesEmptyIllustration } from "./illustrations/stories-empty-illustration";

export type StoriesLayout = "list" | "kanban" | "gantt" | null;

const storyCollisionDetection: CollisionDetection = (args) => {
  const pointerCollisions = pointerWithin(args);
  return pointerCollisions.length > 0
    ? pointerCollisions
    : rectIntersection(args);
};

const getDraggedStory = (
  event: DragStartEvent | DragEndEvent,
  groupedStories?: GroupedStoriesResponse,
) => {
  const draggedStory = event.active.data.current?.story as Story | undefined;
  if (draggedStory) return draggedStory;

  const storyId = event.active.id.toString();
  return groupedStories?.groups
    .flatMap((group) => group.stories)
    .find((story) => story.id === storyId);
};

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
  let overlayContent: ReactNode = null;

  if (story && layout === "kanban") {
    overlayContent = <StoryCardPreview className="rotate-2" story={story} />;
  } else if (story) {
    overlayContent = (
      <Flex
        align="center"
        className="border-border bg-surface-muted shadow-shadow w-max rounded-xl border px-3 py-3.5 shadow backdrop-blur"
        gap={2}
      >
        {selectedStories > 1 ? (
          <Text className="w-60 truncate pl-2" fontWeight="medium">
            {selectedStories} stories selected
          </Text>
        ) : (
          <>
            <StoryStatusIcon statusId={story.statusId} />
            <Text color="muted">
              {team?.code}-{story.sequenceId}
            </Text>
            <Text className="max-w-xs truncate" fontWeight="medium">
              {story.title}
            </Text>
          </>
        )}
      </Flex>
    );
  }

  return (
    <DragOverlay
      className="pointer-events-none"
      dropAnimation={{
        duration: 300,
        easing: "cubic-bezier(0.18, 0.67, 0.6, 1.22)",
      }}
    >
      {overlayContent}
    </DragOverlay>
  );
};

// Move EmptyState component outside StoriesBoard to prevent recreation on each render
const EmptyState = ({
  objectiveId,
  sprintId,
  teamId,
  getTermDisplay,
  illustration,
}: {
  objectiveId?: string;
  sprintId?: string;
  teamId?: string;
  getTermDisplay: ReturnType<typeof useTerminology>["getTermDisplay"];
  illustration?: ReactNode;
}) => (
  <Box className="flex h-[70dvh] items-center justify-center">
    <Box className="flex flex-col items-center">
      {illustration ?? <StoriesEmptyIllustration />}
      <Text className="mt-8 mb-6" fontSize="3xl">
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
  groupedStories: allStories,
  className,
  viewOptions,
  setViewOptions,
  rowClassName,
  emptyStateIllustration,
}: {
  isInSearch?: boolean;
  layout: StoriesLayout;
  groupedStories?: GroupedStoriesResponse;
  className?: string;
  viewOptions: StoriesViewOptions;
  setViewOptions?: (value: StoriesViewOptions) => void;
  rowClassName?: string;
  emptyStateIllustration?: ReactNode;
}) => {
  const { getTermDisplay } = useTerminology();
  const { objectiveId, sprintId, teamId } = useParams<{
    objectiveId: string;
    sprintId: string;
    teamId: string;
  }>();

  const features = useFeatures();
  const sprintsEnabled = useSprintsEnabled(teamId);
  const [activeStory, setActiveStory] = useState<Story | null>(null);
  const [selectedStories, setSelectedStories] = useState<string[]>([]);
  const { mutate } = useUpdateStoryMutation();
  const updateStory = useCallback(
    (storyId: string, payload: Partial<DetailedStory>) => {
      mutate({ storyId, payload });
    },
    [mutate],
  );

  // Memoize the isColumnVisible function
  const isColumnVisible = useCallback(
    (column: DisplayColumn) => {
      if (column === "Sprint" && !sprintsEnabled) return false;
      if (
        (column === "Objective" || column === "Key Result") &&
        !features.objectiveEnabled
      )
        return false;
      return viewOptions.displayColumns.includes(column);
    },
    [features.objectiveEnabled, sprintsEnabled, viewOptions.displayColumns],
  );

  const handleDragStart = useCallback(
    (event: DragStartEvent) => {
      setActiveStory(getDraggedStory(event, allStories) ?? null);
    },
    [allStories],
  );

  const handleDragCancel = useCallback(() => {
    setActiveStory(null);
  }, []);

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      const { groupBy } = viewOptions;
      const story = getDraggedStory(event, allStories);
      const targetKey = event.over?.id.toString();

      if (!story || !targetKey) {
        setActiveStory(null);
        return;
      }

      const updatePayload = getStoryDropUpdate(story, groupBy, targetKey);
      if (updatePayload) updateStory(story.id, updatePayload);
      setActiveStory(null);
    },
    [allStories, updateStory, viewOptions],
  );

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

  const sensors = useSensors(pointerSensor, touchSensor);

  // Memoize the context value to prevent unnecessary re-renders
  const boardContextValue = useMemo(
    () => ({
      selectedStories,
      setSelectedStories,
      viewOptions,
      setViewOptions,
      isColumnVisible,
      updateStory,
      newStoryDefaults: {
        teamId,
        objectiveId,
        sprintId,
      },
    }),
    [
      selectedStories,
      setViewOptions,
      viewOptions,
      isColumnVisible,
      updateStory,
      teamId,
      objectiveId,
      sprintId,
    ],
  );

  const hasStories = allStories?.groups.some(
    (group) => group.stories.length > 0,
  );

  return (
    <BoardContext.Provider value={boardContextValue}>
      <Box
        className={cn("min-h-0 w-full min-w-0", {
          "h-full overflow-hidden": !isInSearch,
        })}
      >
        {!isInSearch && !hasStories && (
          <EmptyState
            getTermDisplay={getTermDisplay}
            illustration={emptyStateIllustration}
            objectiveId={objectiveId}
            sprintId={sprintId}
            teamId={teamId}
          />
        )}

        {hasStories ? (
          <DndContext
            collisionDetection={storyCollisionDetection}
            onDragCancel={handleDragCancel}
            onDragEnd={handleDragEnd}
            onDragStart={handleDragStart}
            sensors={sensors}
          >
            {layout === "gantt" && (
              <GanttBoard className={className} stories={[]} />
            )}
            {allStories?.meta.groupBy === "none" ? (
              <BodyContainer
                className={cn(
                  "overflow-x-auto pb-6",
                  {
                    "h-auto pb-0": isInSearch,
                  },
                  className,
                )}
              >
                <StoriesList
                  isInSearch={isInSearch}
                  rowClassName={rowClassName}
                  stories={allStories.groups[0].stories}
                />
              </BodyContainer>
            ) : (
              <>
                {layout === "kanban" && (
                  <KanbanBoard
                    className={className}
                    groupedStories={allStories!}
                  />
                )}
                {(layout === "list" || !layout) && (
                  <ListBoard
                    className={className}
                    groupedStories={allStories!}
                    isInSearch={isInSearch}
                    rowClassName={rowClassName}
                    viewOptions={viewOptions}
                  />
                )}
              </>
            )}
            {typeof window !== "undefined"
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
        ) : null}

        {/* This toolbar pops up when the user selects stories */}
        {selectedStories.length > 0 && <StoriesToolbar />}
      </Box>
    </BoardContext.Provider>
  );
};

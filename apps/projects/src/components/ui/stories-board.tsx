"use client";
import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core";
import { DndContext, DragOverlay } from "@dnd-kit/core";
import { Flex, Text } from "ui";
import { cn } from "lib";
import { useState } from "react";
import { createPortal } from "react-dom";
import type { Story, StoryStatus } from "@/types/story";
import { BodyContainer } from "../shared/body";
import { KanbanBoard } from "./kanban-board";
import { StoriesGroup } from "./stories-group";
import { StoryStatusIcon } from "./story-status-icon";
import { StoryCard } from "./story/card";

export type StoriesLayout = "list" | "kanban" | null;

const StoryOverlay = ({
  story,
  layout,
}: {
  story: Story | null;
  layout: StoriesLayout;
}) => {
  return (
    <DragOverlay
      dropAnimation={{
        duration: 300,
        easing: "cubic-bezier(0.18, 0.67, 0.6, 1.22)",
      }}
    >
      {layout === "kanban" ? (
        <StoryCard
          className="border-gray-200 shadow-lg dark:border-dark-50/60 dark:shadow-dark"
          story={story!}
        />
      ) : (
        <Flex
          align="center"
          className="w-max rounded-lg border border-gray-100 bg-gray-50/70 px-3 py-3.5 shadow backdrop-blur dark:border-dark-100 dark:bg-dark-200/70"
          gap={2}
        >
          <StoryStatusIcon status={story?.status} />
          <Text color="muted">COM-{story?.id}</Text>
          <Text className="max-w-xs truncate" fontWeight="medium">
            {story?.title}
          </Text>
        </Flex>
      )}
    </DragOverlay>
  );
};

export const StoriesBoard = ({
  layout,
  stories,
  statuses,
  className,
}: {
  layout: StoriesLayout;
  stories: Story[];
  statuses: StoryStatus[];
  className?: string;
}) => {
  const [storiesBoard, setStoriesBoard] = useState<Story[]>(stories);
  const [activeStory, setActiveStory] = useState<Story | null>(null);

  const handleDragStart = (e: DragStartEvent) => {
    const story = stories.find(({ id }) => id === Number(e.active.id))!;
    setActiveStory(story);
  };

  const handleDragEnd = (e: DragEndEvent) => {
    const newStatus = e.over?.id as StoryStatus | null;
    if (newStatus) {
      const index = storiesBoard.findIndex(
        ({ id }) => id === Number(e.active.id),
      )!;
      storiesBoard[index].status = newStatus;
      setStoriesBoard([...storiesBoard]);
    }
    setActiveStory(null);
  };

  return (
    <DndContext onDragEnd={handleDragEnd} onDragStart={handleDragStart}>
      {layout === "kanban" ? (
        <KanbanBoard
          className={className}
          statuses={statuses}
          stories={storiesBoard}
        />
      ) : (
        <BodyContainer className={cn("overflow-x-auto pb-6", className)}>
          {statuses.map((status) => (
            <StoriesGroup
              className="-top-[0.5px]"
              key={status}
              status={status}
              stories={storiesBoard}
            />
          ))}
        </BodyContainer>
      )}

      {typeof window !== "undefined" &&
        createPortal(
          <StoryOverlay layout={layout} story={activeStory} />,
          document.body,
        )}
    </DndContext>
  );
};

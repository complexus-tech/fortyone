"use client";

import type { ComponentProps, ReactNode } from "react";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import { Box, Text } from "ui";
import type { RoadmapKeyResultSummary } from "./roadmap-key-results";

type Objective = ComponentProps<typeof RoadmapKeyResultSummary>["objective"];

export type ObjectiveBoardCardProps = {
  canDrag?: boolean;
  children?: ReactNode;
  objective: Objective;
  onSelect: (objective: Objective) => void;
  teamCode?: string;
  isOverlay?: boolean;
};

export const ObjectiveBoardCard = ({
  canDrag = true,
  children,
  objective,
  teamCode,
  onSelect,
  isOverlay = false,
}: ObjectiveBoardCardProps) => {
  const objectiveReference = teamCode
    ? `${teamCode}-${objective.sequenceId}`
    : String(objective.sequenceId);
  const { attributes, isDragging, listeners, setNodeRef } = useDraggable({
    id: objective.id,
    disabled: !canDrag || isOverlay,
  });

  return (
    <div
      className={cn(
        "border-border shadow-shadow hover:bg-surface-elevated dark:border-border/70 dark:bg-surface w-[340px] rounded-xl border-[0.5px] bg-white px-4 pb-4 shadow-lg backdrop-blur transition duration-200 ease-linear select-none",
        {
          "rotate-2 shadow-xl": isOverlay,
          "bg-surface-muted opacity-60": isDragging,
        },
      )}
      ref={setNodeRef}
    >
      <Box
        className={cn("cursor-grab pt-3 pb-1.5", {
          "cursor-grabbing": isDragging,
        })}
        {...attributes}
        {...listeners}
      >
        <button
          className="focus-visible:ring-primary flex w-full justify-between gap-2 rounded-sm text-left outline-none focus-visible:ring-1"
          onClick={() => {
            if (!isDragging) onSelect(objective);
          }}
          type="button"
        >
          <Text className="line-clamp-3 text-[1.1rem] leading-[1.4rem]">
            {objective.name}
          </Text>
          {objective.sequenceId > 0 ? (
            <Text
              className="shrink-0 text-[0.95rem] leading-[1.4rem] uppercase"
              color="muted"
            >
              {objectiveReference}
            </Text>
          ) : null}
        </button>
      </Box>
      {children}
    </div>
  );
};

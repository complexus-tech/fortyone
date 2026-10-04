"use client";

import type { ComponentProps, ReactNode } from "react";
import Link from "next/link";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import { Box, Text } from "ui";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import type { RoadmapKeyResultSummary } from "./roadmap-key-results";

type Objective = ComponentProps<typeof RoadmapKeyResultSummary>["objective"];

export type ObjectiveBoardCardProps = {
  canDrag?: boolean;
  children?: ReactNode;
  objective: Objective;
  teamCode?: string;
  isOverlay?: boolean;
};

export const ObjectiveBoardCard = ({
  canDrag = true,
  children,
  objective,
  teamCode,
  isOverlay = false,
}: ObjectiveBoardCardProps) => {
  const objectiveReference = teamCode
    ? `${teamCode}-${objective.sequenceId}`
    : String(objective.sequenceId);
  const { withWorkspace } = useWorkspacePath();
  const { isDragging, listeners, setNodeRef } = useDraggable({
    id: objective.id,
    disabled: !canDrag || isOverlay,
  });
  const heading = (
    <>
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
    </>
  );

  return (
    <div
      aria-hidden={isOverlay || undefined}
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
        className={cn("pt-3 pb-1.5", {
          "cursor-grab": canDrag && !isOverlay,
          "cursor-grabbing": isDragging,
        })}
        {...listeners}
      >
        {isOverlay ? (
          <Box className="flex w-full justify-between gap-2">{heading}</Box>
        ) : (
          <Link
            className="focus-visible:ring-primary flex w-full justify-between gap-2 rounded-sm text-left outline-none focus-visible:ring-1"
            data-objective-link={objective.id}
            href={withWorkspace(
              `/teams/${objective.teamId}/objectives/${objective.id}`,
            )}
            onClick={(event) => {
              if (isDragging) event.preventDefault();
            }}
          >
            {heading}
          </Link>
        )}
      </Box>
      {children}
    </div>
  );
};

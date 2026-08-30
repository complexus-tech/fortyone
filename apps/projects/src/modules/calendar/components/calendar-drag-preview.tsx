import { useRef, useState } from "react";
import type { DragMoveEvent } from "@dnd-kit/core";
import { useDndMonitor } from "@dnd-kit/core";
import { cn } from "lib";
import { Box, Text } from "ui";
import { getCalendarStoryBlockStyle } from "./calendar-block";
import { getCalendarManualChange } from "./calendar-drag";
import { getCalendarDragData } from "./calendar-dnd";
import {
  HOUR_HEIGHT,
  SCHEDULED_STORY_STATUS_CLASS,
  toMoveTimeLabel,
} from "./calendar-presentation";
import type { CalendarDragData } from "./calendar-types";

export const CalendarDragPreview = ({
  drag,
}: {
  drag: CalendarDragData | null;
}) => {
  const [movePreview, setMovePreview] = useState<{
    endAt: Date;
    startAt: Date;
  } | null>(null);
  const moveTargetDayRef = useRef<Date | null>(null);

  const resetMovePreview = () => {
    moveTargetDayRef.current = null;
    setMovePreview(null);
  };

  const updateMovePreview = (event: DragMoveEvent) => {
    const moveDrag = getCalendarDragData(event.active);
    if (moveDrag?.kind !== "move") return;

    const targetDay = event.over?.data.current?.calendarDay as
      | string
      | undefined;
    if (targetDay) {
      moveTargetDayRef.current = new Date(targetDay);
    }

    const nextPreview = getCalendarManualChange({
      block: moveDrag.block,
      deltaY: event.delta.y,
      hourHeight: HOUR_HEIGHT,
      kind: "move",
      targetDay: moveTargetDayRef.current ?? new Date(moveDrag.block.startAt),
    });
    setMovePreview((currentPreview) =>
      currentPreview?.startAt.getTime() === nextPreview.startAt.getTime() &&
      currentPreview.endAt.getTime() === nextPreview.endAt.getTime()
        ? currentPreview
        : nextPreview,
    );
  };

  useDndMonitor({
    onDragCancel: resetMovePreview,
    onDragEnd: resetMovePreview,
    onDragMove: updateMovePreview,
    onDragOver: updateMovePreview,
    onDragStart: (event) => {
      const moveDrag = getCalendarDragData(event.active);
      moveTargetDayRef.current =
        moveDrag?.kind === "move" ? new Date(moveDrag.block.startAt) : null;
      setMovePreview(null);
    },
  });

  if (!drag) return null;
  if (drag.kind === "resize") {
    return <Box className="bg-border-strong h-full w-full rounded-sm" />;
  }

  const storyStyle = getCalendarStoryBlockStyle(drag.block);
  const previewStartAt = movePreview?.startAt ?? new Date(drag.block.startAt);
  const previewEndAt = movePreview?.endAt ?? new Date(drag.block.endAt);
  const moveFeedbackLabel = toMoveTimeLabel(previewStartAt, previewEndAt);

  return (
    <Box className="relative h-full w-full overflow-visible">
      <Box
        className={cn(
          "shadow-shadow relative h-full w-full overflow-hidden rounded-md border px-3 py-1 backdrop-blur-sm",
          storyStyle
            ? SCHEDULED_STORY_STATUS_CLASS
            : "border-border-strong/60 bg-surface-muted/95",
        )}
        style={storyStyle}
      >
        <span
          aria-hidden="true"
          className={cn(
            "absolute top-1/2 left-1 h-[calc(100%-0.5rem)] max-h-14 w-[0.1875rem] -translate-y-1/2 rounded-full",
            storyStyle
              ? "bg-[var(--calendar-story-accent)]"
              : "bg-border-strong",
          )}
        />
        <Text className="truncate leading-tight" fontWeight="medium">
          {drag.block.title}
        </Text>
      </Box>
      <span
        aria-hidden="true"
        className="border-border-strong/70 bg-surface-elevated/95 text-foreground shadow-shadow pointer-events-none absolute -top-7 right-1 z-10 rounded-md border px-2 py-0.5 text-xs font-medium whitespace-nowrap tabular-nums"
      >
        {moveFeedbackLabel}
      </span>
    </Box>
  );
};

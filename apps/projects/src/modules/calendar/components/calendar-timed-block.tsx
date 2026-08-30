import type {
  PointerEvent as ReactPointerEvent,
  TouchEvent as ReactTouchEvent,
} from "react";
import { useDraggable } from "@dnd-kit/core";
import { cn } from "lib";
import { CalendarIcon, CheckIcon, TimeScheduleIcon, Video02Icon } from "icons";
import { Box, Text } from "ui";
import { formatTimeNeeded } from "@/lib/time-needed";
import type {
  CalendarEventSummary,
  CalendarScheduleBlock,
} from "@/lib/queries/calendar/types";
import {
  CROSS_WORKSPACE_CALENDAR_BLOCK_TITLE,
  CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP,
  RESERVED_TIME_BLOCK_CLASS,
  getCalendarScheduleBlockSecondaryLabel,
  getCalendarStoryBlockStyle,
  getMayaCalendarBlockReason,
  isCalendarScheduleBlockCompleted,
  isCalendarScheduleBlockEditable,
} from "./calendar-block";
import {
  CALENDAR_RESIZE_STEP_MINUTES,
  activateCalendarResize,
  isCalendarBlockResizeTerminalDay,
  resizeCalendarBlockByMinutes,
  snapCalendarDeltaMinutes,
} from "./calendar-drag";
import {
  COMPLETED_CALENDAR_BLOCK_PATTERN,
  HOUR_HEIGHT,
  SCHEDULED_STORY_STATUS_CLASS,
  SCHEDULED_STORY_STATUS_HOVER_CLASS,
  SCHEDULED_TASK_BACKGROUND_CLASS,
  SCHEDULED_TASK_HOVER_BACKGROUND_CLASS,
  TIMED_BLOCK_VERTICAL_GAP,
  TIMED_BLOCK_VERTICAL_INSET,
  TWO_LINE_TITLE_MINIMUM_HEIGHT,
  getBusyWindowTitle,
  toResizeEndLabel,
  toTimeLabel,
} from "./calendar-presentation";
import type { CalendarItem, CalendarManualChange } from "./calendar-types";
import {
  getCalendarEventTitle,
  getCalendarEventTimeLabel,
} from "./calendar-event-details-dialog";

export const CalendarTimedBlock = ({
  day,
  item,
  isDragDisabled,
  layout,
  onEdit,
  onManualChange,
  onSelectEvent,
  today,
}: {
  day: Date;
  item: CalendarItem;
  isDragDisabled: boolean;
  layout: { top: number; height: number; lane: number; laneCount: number };
  onEdit: (block: CalendarScheduleBlock) => void;
  onManualChange: (change: CalendarManualChange) => void;
  onSelectEvent: (event: CalendarEventSummary) => void;
  today: Date;
}) => {
  const draggableBlock =
    item.kind === "block" &&
    !isCalendarScheduleBlockCompleted(item.block) &&
    !item.block.isCrossWorkspace &&
    item.block.blockType === "work" &&
    item.block.storyId
      ? item.block
      : null;
  const resizableBlock =
    draggableBlock && isCalendarBlockResizeTerminalDay(draggableBlock, day)
      ? draggableBlock
      : null;
  const dragSegmentID = `${item.id}-${day.getTime()}`;
  const moveDrag = useDraggable({
    id: `calendar-move-${dragSegmentID}`,
    disabled: !draggableBlock || isDragDisabled,
    data: draggableBlock
      ? { calendarDrag: { kind: "move", block: draggableBlock } }
      : undefined,
  });
  const {
    isDragging: isResizeDragging,
    listeners: resizeListeners,
    setNodeRef: setResizeNodeRef,
    transform: resizeTransform,
  } = useDraggable({
    id: `calendar-resize-${dragSegmentID}`,
    disabled: !resizableBlock || isDragDisabled,
    data: resizableBlock
      ? { calendarDrag: { kind: "resize", block: resizableBlock } }
      : undefined,
  });
  const resizePointerDown = resizeListeners?.onPointerDown as
    | ((event: ReactPointerEvent<HTMLButtonElement>) => void)
    | undefined;
  const resizeTouchStart = resizeListeners?.onTouchStart as
    | ((event: ReactTouchEvent<HTMLButtonElement>) => void)
    | undefined;
  const laneWidth = 100 / layout.laneCount;
  const isPast = new Date(item.endAt).getTime() <= today.getTime();
  const resizeDeltaMinutes = isResizeDragging
    ? snapCalendarDeltaMinutes(resizeTransform?.y ?? 0, HOUR_HEIGHT, "resize")
    : 0;
  const resizeDelta = resizeDeltaMinutes * (HOUR_HEIGHT / 60);
  const resizePreview =
    resizableBlock && isResizeDragging
      ? resizeCalendarBlockByMinutes(resizableBlock, resizeDeltaMinutes)
      : null;
  const renderedHeight = Math.max(
    18,
    layout.height - TIMED_BLOCK_VERTICAL_GAP + resizeDelta,
  );
  const style = {
    height: `${renderedHeight}px`,
    left: `calc(${layout.lane * laneWidth}% + 0.25rem)`,
    top: `${layout.top + TIMED_BLOCK_VERTICAL_INSET}px`,
    width: `calc(${laneWidth}% - 0.5rem)`,
  };
  const isCompactEvent =
    item.kind === "event" &&
    renderedHeight < HOUR_HEIGHT - TIMED_BLOCK_VERTICAL_GAP;
  const showSecondaryLine =
    item.kind === "event" ? renderedHeight >= 31 : renderedHeight >= 40;
  const canShowTwoLineTitle = layout.height >= TWO_LINE_TITLE_MINIMUM_HEIGHT;
  const titleLineClass = canShowTwoLineTitle
    ? "line-clamp-2 leading-5"
    : "truncate leading-[0.9375rem]";
  const secondaryLineClass = canShowTwoLineTitle
    ? "mt-1 leading-[1.1rem]"
    : "mt-0.5 leading-[0.9375rem]";
  const blockPaddingClass =
    layout.height >= HOUR_HEIGHT ? "py-1 pr-2.5 pl-3" : "py-px pr-2.5 pl-3";

  if (item.kind === "event") {
    const EventIcon = item.event.meetingUrl ? Video02Icon : CalendarIcon;
    const eventTextClass = isPast ? "text-text-muted" : "text-[#3c90ff]";
    const eventAccentClass = isPast ? "bg-border-strong" : "bg-[#3c90ff]";

    return (
      <button
        aria-label={`Open ${getCalendarEventTitle(item.event)} details, ${getCalendarEventTimeLabel(item.event)}`}
        className={cn(
          "absolute overflow-hidden rounded-md border text-left backdrop-blur-sm transition-colors focus-visible:ring-2 focus-visible:outline-none",
          isPast
            ? "border-border-strong/40 bg-surface-muted/55 dark:bg-surface-elevated/55 hover:bg-surface-muted/65 dark:hover:bg-surface-elevated/65 bg-[repeating-linear-gradient(135deg,transparent_0,transparent_5px,rgba(100,116,139,0.12)_5px,rgba(100,116,139,0.12)_8px)]"
            : "border-[#3c90ff]/20 bg-[#3c90ff]/20 hover:bg-[#3c90ff]/25 focus-visible:ring-[#3c90ff]/50",
          blockPaddingClass,
        )}
        onClick={() => {
          onSelectEvent(item.event);
        }}
        style={style}
        type="button"
      >
        <span
          aria-hidden="true"
          className={cn(
            "absolute top-1/2 left-1 h-[calc(100%-0.5rem)] max-h-14 w-[0.1875rem] -translate-y-1/2 rounded-full",
            eventAccentClass,
          )}
        />
        <Box
          className={cn(
            "flex min-w-0 items-start gap-1.5",
            isCompactEvent && "-mt-0.5",
          )}
        >
          <EventIcon
            aria-hidden="true"
            className={cn(
              "h-4 w-4 shrink-0",
              isCompactEvent ? "mt-1" : "mt-0.5",
              eventTextClass,
            )}
          />
          <Box className="-mt-px min-w-0 flex-1">
            <Text
              as="span"
              className={cn("min-w-0", titleLineClass, eventTextClass)}
              fontSize="md"
              fontWeight="medium"
            >
              {getCalendarEventTitle(item.event)}
            </Text>
            {showSecondaryLine ? (
              <Text
                as="span"
                className={cn(
                  "block truncate text-[0.9375rem]",
                  secondaryLineClass,
                  isCompactEvent && "mt-0 leading-[0.9375rem]",
                  eventTextClass,
                )}
              >
                {toTimeLabel(item.startAt, item.endAt)}
              </Text>
            ) : null}
          </Box>
        </Box>
      </button>
    );
  }

  if (item.kind === "busy") {
    return (
      <Box
        className={cn(
          "absolute overflow-hidden rounded-md border border-dashed backdrop-blur-sm",
          isPast
            ? "border-border-strong/40 bg-surface-muted/55 dark:bg-surface-elevated/55 bg-[repeating-linear-gradient(135deg,transparent_0,transparent_5px,rgba(100,116,139,0.12)_5px,rgba(100,116,139,0.12)_8px)]"
            : "border-border-strong/40 bg-surface-muted/35 bg-[repeating-linear-gradient(135deg,transparent_0,transparent_5px,rgba(148,163,184,0.08)_5px,rgba(148,163,184,0.08)_8px)]",
          blockPaddingClass,
        )}
        style={style}
      >
        <span
          aria-hidden="true"
          className="bg-border-strong absolute top-1/2 left-1 h-[calc(100%-0.5rem)] max-h-14 w-[0.1875rem] -translate-y-1/2 rounded-full"
        />
        <Text
          className={cn(
            isPast ? "text-text-muted" : "text-foreground",
            titleLineClass,
          )}
          fontSize="md"
          fontWeight="medium"
        >
          {getBusyWindowTitle(item.window)}
        </Text>
        {showSecondaryLine ? (
          <Text
            className={cn(
              "truncate text-[0.9375rem]",
              secondaryLineClass,
              isPast ? "text-text-muted" : "text-foreground",
            )}
          >
            {toTimeLabel(item.startAt, item.endAt)}
          </Text>
        ) : null}
      </Box>
    );
  }

  const { block } = item;
  const isCompleted = isCalendarScheduleBlockCompleted(block);
  const isCrossWorkspace = Boolean(block.isCrossWorkspace);
  const isMayaManaged = block.source === "maya";
  const isEditable = isCalendarScheduleBlockEditable(block);
  const mayaReason = getMayaCalendarBlockReason(block);
  const blockTitle = isCrossWorkspace
    ? CROSS_WORKSPACE_CALENDAR_BLOCK_TITLE
    : block.title;
  const storyStyle = getCalendarStoryBlockStyle(block);
  let blockColorClass = RESERVED_TIME_BLOCK_CLASS;
  if (block.blockType === "work") {
    blockColorClass = storyStyle
      ? SCHEDULED_STORY_STATUS_CLASS
      : cn(
          "border-border-strong/70 dark:border-border-strong",
          SCHEDULED_TASK_BACKGROUND_CLASS,
        );
    if (!isMayaManaged) {
      blockColorClass = cn(
        blockColorClass,
        storyStyle
          ? SCHEDULED_STORY_STATUS_HOVER_CLASS
          : SCHEDULED_TASK_HOVER_BACKGROUND_CLASS,
      );
    }
  }
  if (block.hasConflict) {
    blockColorClass = "border-danger/60 bg-danger/20";
    if (!isMayaManaged) {
      blockColorClass += " hover:bg-danger/25";
    }
  }
  if (isCompleted && block.blockType === "work") {
    blockColorClass = cn(
      "border-border-strong/40",
      storyStyle
        ? SCHEDULED_STORY_STATUS_CLASS
        : SCHEDULED_TASK_BACKGROUND_CLASS,
    );
  }
  if (isCrossWorkspace) {
    blockColorClass = RESERVED_TIME_BLOCK_CLASS;
  }
  const isScheduledStory = block.blockType === "work";
  const isStandardHeightBlock =
    layout.height >= HOUR_HEIGHT &&
    layout.height < TWO_LINE_TITLE_MINIMUM_HEIGHT;
  const hasLeadingIcon = isCompleted || isCrossWorkspace || !isScheduledStory;
  const timeLabel = toTimeLabel(block.startAt, block.endAt);
  const resizeStartAt = resizePreview?.startAt ?? new Date(block.startAt);
  const resizeEndAt = resizePreview?.endAt ?? new Date(block.endAt);
  const resizeDurationMinutes = Math.round(
    (resizeEndAt.getTime() - resizeStartAt.getTime()) / 60_000,
  );
  const resizeFeedbackLabel = `Ends ${toResizeEndLabel(resizeStartAt, resizeEndAt)} · ${formatTimeNeeded(resizeDurationMinutes)}`;
  const secondaryLabel = getCalendarScheduleBlockSecondaryLabel(
    block,
    timeLabel,
  );
  const canOpenBlock = !isCrossWorkspace && (isScheduledStory || isEditable);
  let blockActionLabel = isScheduledStory
    ? "Open scheduled story details"
    : "Edit calendar block";
  if (block.hasConflict) {
    blockActionLabel = "Resolve calendar block conflict";
  }
  if (isMayaManaged) {
    if (isScheduledStory) {
      blockActionLabel = "Open Maya-managed scheduled story details";
    } else if (block.isLocked) {
      blockActionLabel = "Locked Maya-managed calendar block";
    } else {
      blockActionLabel = "Maya-managed calendar block";
    }
    if (mayaReason) {
      blockActionLabel += `. ${mayaReason}`;
    }
  }
  if (isCrossWorkspace) {
    blockActionLabel = CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP;
  } else if (isCompleted) {
    blockActionLabel = `Open completed ${blockTitle} details`;
  }
  let blockTooltip = block.storyId ? undefined : mayaReason ?? undefined;
  if (isCrossWorkspace) {
    blockTooltip = CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP;
  }
  const resizeByKeyboard = (deltaMinutes: number) => {
    if (!resizableBlock || isDragDisabled) return;
    const { endAt, startAt } = resizeCalendarBlockByMinutes(
      resizableBlock,
      deltaMinutes,
    );
    if (endAt.getTime() === new Date(resizableBlock.endAt).getTime()) return;

    onManualChange({
      block: resizableBlock,
      change: "resize",
      startAt,
      endAt,
    });
  };
  let blockTitleColorClass = isCompleted
    ? "text-text-muted"
    : "text-foreground";
  if (block.hasConflict && !isCompleted) {
    blockTitleColorClass = "text-danger";
  }
  let blockSecondaryColorClass = "text-text-muted";
  let blockIconClass = "text-text-muted";
  let blockAccentClass = storyStyle
    ? "bg-[var(--calendar-story-accent)]"
    : "bg-border-strong";
  if (block.hasConflict && !isCompleted) {
    blockSecondaryColorClass = "text-danger";
    blockIconClass = "text-danger";
    blockAccentClass = "bg-danger";
  }
  if (isCrossWorkspace) {
    blockTitleColorClass = "text-text-muted";
    blockSecondaryColorClass = "text-text-muted";
    blockIconClass = "text-text-muted";
    blockAccentClass = "bg-border-strong";
  }
  const blockContent = (
    <div
      {...moveDrag.listeners}
      className={cn(
        "absolute flex items-center rounded-md border backdrop-blur-sm transition-colors",
        resizableBlock ? "overflow-visible" : "overflow-hidden",
        isResizeDragging && "z-30",
        draggableBlock && !isDragDisabled
          ? "cursor-grab touch-none active:cursor-grabbing"
          : null,
        moveDrag.isDragging ? "opacity-0" : null,
        blockPaddingClass,
        blockColorClass,
      )}
      ref={moveDrag.setNodeRef}
      style={{
        ...style,
        ...storyStyle,
        ...(isCompleted && block.blockType === "work"
          ? { backgroundImage: COMPLETED_CALENDAR_BLOCK_PATTERN }
          : null),
      }}
      title={blockTooltip}
    >
      <span
        aria-hidden="true"
        className={cn(
          "absolute top-1/2 left-1 h-[calc(100%-0.5rem)] max-h-14 w-[0.1875rem] -translate-y-1/2 rounded-full",
          blockAccentClass,
        )}
      />
      <button
        aria-label={blockActionLabel}
        className="focus-visible:ring-primary/40 absolute inset-0 z-0 rounded-md focus-visible:ring-2 focus-visible:outline-none focus-visible:ring-inset"
        disabled={!canOpenBlock}
        onClick={() => {
          if (canOpenBlock) {
            onEdit(block);
          }
        }}
        type="button"
      />
      <Box
        className={cn(
          "pointer-events-none relative z-10 max-h-full min-w-0 overflow-hidden",
          hasLeadingIcon
            ? cn(
                "flex gap-1.5",
                showSecondaryLine ? "items-start" : "items-center",
              )
            : null,
        )}
      >
        {hasLeadingIcon && isCompleted ? (
          <CheckIcon
            aria-hidden="true"
            className={cn(
              "h-4 w-4 shrink-0",
              showSecondaryLine && "mt-0.5",
              blockIconClass,
            )}
            strokeWidth={2.2}
          />
        ) : null}
        {hasLeadingIcon && !isCompleted ? (
          <TimeScheduleIcon
            aria-hidden="true"
            className={cn(
              "h-4 w-4 shrink-0",
              showSecondaryLine && "mt-0.5",
              blockIconClass,
            )}
          />
        ) : null}
        <Box
          className={cn(
            "min-w-0",
            hasLeadingIcon && "flex-1",
            hasLeadingIcon && showSecondaryLine && "-mt-px",
          )}
        >
          <Text
            className={cn(
              "min-w-0",
              titleLineClass,
              isStandardHeightBlock && "leading-4",
              isCompleted && "line-through decoration-current/70",
              blockTitleColorClass,
            )}
            fontSize="md"
            fontWeight="medium"
          >
            {blockTitle}
          </Text>
          {showSecondaryLine ? (
            <Text
              className={cn(
                "truncate text-[0.9375rem]",
                secondaryLineClass,
                isStandardHeightBlock && "mt-1 leading-4",
                blockSecondaryColorClass,
              )}
            >
              {secondaryLabel}
            </Text>
          ) : null}
        </Box>
      </Box>
      {resizePreview ? (
        <span
          aria-hidden="true"
          className="border-border-strong/70 bg-surface-elevated/95 text-foreground shadow-shadow pointer-events-none absolute right-1 bottom-6 z-30 rounded-md border px-2 py-0.5 text-xs font-medium whitespace-nowrap tabular-nums"
        >
          {resizeFeedbackLabel}
        </span>
      ) : null}
      {resizableBlock ? (
        <span aria-atomic="true" aria-live="polite" className="sr-only">
          {resizeFeedbackLabel}
        </span>
      ) : null}
      {resizableBlock ? (
        <button
          {...resizeListeners}
          aria-keyshortcuts="ArrowUp ArrowDown"
          aria-label={`Resize ${blockTitle}. ${resizeFeedbackLabel}. Use Arrow Up to shorten or Arrow Down to extend by five minutes.`}
          aria-roledescription="resize handle"
          className={cn(
            "group/resize focus-visible:ring-primary/60 absolute inset-x-0 -bottom-[3px] z-20 h-6 max-h-full touch-none rounded-md focus-visible:ring-2 focus-visible:outline-none focus-visible:ring-inset",
            isDragDisabled ? "cursor-wait" : "cursor-ns-resize",
          )}
          disabled={isDragDisabled}
          onKeyDown={(event) => {
            if (event.key !== "ArrowUp" && event.key !== "ArrowDown") return;
            event.preventDefault();
            event.stopPropagation();
            resizeByKeyboard(
              event.key === "ArrowUp"
                ? -CALENDAR_RESIZE_STEP_MINUTES
                : CALENDAR_RESIZE_STEP_MINUTES,
            );
          }}
          onPointerDown={(event) => {
            activateCalendarResize(event, resizePointerDown);
          }}
          onTouchStart={(event) => {
            activateCalendarResize(event, resizeTouchStart);
          }}
          ref={setResizeNodeRef}
          type="button"
        >
          <span
            aria-hidden="true"
            className={cn(
              "bg-border-strong/80 group-hover/resize:bg-foreground/70 group-focus-visible/resize:bg-primary group-active/resize:bg-primary absolute bottom-0.5 left-1/2 h-0.5 w-7 -translate-x-1/2 rounded-full",
              isResizeDragging && "bg-primary",
            )}
          />
        </button>
      ) : null}
    </div>
  );

  return blockContent;
};

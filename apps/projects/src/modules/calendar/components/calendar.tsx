"use client";

import { useState } from "react";
import {
  addDays,
  addHours,
  format,
  isSameDay,
  isSameMonth,
  startOfDay,
} from "date-fns";
import type {
  CollisionDetection,
  DragEndEvent,
  DragStartEvent,
  Modifier,
} from "@dnd-kit/core";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  TouchSensor,
  pointerWithin,
  rectIntersection,
  useDraggable,
  useDroppable,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { cn } from "lib";
import {
  ArrowDown2Icon,
  CalendarIcon,
  CheckIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  DeleteIcon,
  GoogleCalendarIcon,
  StoryIcon,
  TimeScheduleIcon,
  Video02Icon,
} from "icons";
import {
  Box,
  Button,
  Command,
  Dialog,
  Divider,
  Flex,
  Input,
  Menu,
  Popover,
  Text,
} from "ui";
import { useLocalStorage, useTerminology, useWorkspacePath } from "@/hooks";
import {
  useCalendarIntegration,
  useCreateCalendarConnectSession,
  useCalendarSchedule,
  useCreateCalendarScheduleBlock,
  useDeleteCalendarScheduleBlock,
  useManualRescheduleCalendarScheduleBlock,
  useSyncCalendarConnection,
  useUpdateCalendarScheduleBlock,
} from "@/lib/hooks/calendar";
import type {
  CalendarBusyWindow,
  CalendarEventSummary,
  CalendarScheduleBlock,
  CalendarScheduleBlockInput,
} from "@/lib/queries/calendar/types";
import type { CalendarConnection } from "@/modules/settings/workspace/integrations/calendar/types";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import type { Story } from "@/modules/stories/types";
import { useUpdateStoryMutation } from "@/modules/story/hooks/update-mutation";
import {
  CROSS_WORKSPACE_CALENDAR_BLOCK_TITLE,
  CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP,
  RESERVED_TIME_BLOCK_CLASS,
  getCalendarScheduleBlockSecondaryLabel,
  getCalendarStoryBlockStyle,
  getMayaCalendarBlockLabel,
  getMayaCalendarBlockReason,
  isCalendarScheduleBlockEditable,
} from "./calendar-block";
import type { CalendarDragKind } from "./calendar-drag";
import {
  getCalendarManualChange,
  snapCalendarDeltaPixels,
} from "./calendar-drag";
import {
  buildCalendarEventLayouts,
  deriveCalendarVisibleHours,
  getDisplayBusyWindows,
  parseCalendarDate,
} from "./calendar-layout";
import {
  CalendarEventDetailsDialog,
  getCalendarEventTitle,
  getCalendarEventTimeLabel,
} from "./calendar-event-details-dialog";
import { CalendarScheduleBlockDetailsDialog } from "./calendar-schedule-block-details-dialog";
import {
  getCalendarViewDays,
  getCalendarViewRange,
  getCalendarViewTitle,
  moveCalendarCursor,
  normalizeCalendarView,
} from "./calendar-view";
import type { CalendarView } from "./calendar-view";
import { CalendarGridSkeleton } from "./calendar-skeleton";

const defaultVisibleStartHour = 8;
const defaultVisibleEndHour = 24;
const hourHeight = 52;
const timedBlockVerticalGap = 6;
const timedBlockVerticalInset = timedBlockVerticalGap / 2;
const twoLineTitleMinimumHeight = hourHeight * 1.5;
const timeRailWidth = 8;
const calendarHistoryDays = 7;
const calendarLookaheadDays = 90;
const calendarViews = ["day", "week", "month"] as const;
const scheduledTaskBackgroundClass =
  "bg-surface-muted dark:bg-surface-prominent/65";
const scheduledTaskHoverBackgroundClass =
  "hover:bg-accent dark:hover:bg-surface-prominent/70";
const scheduledStoryStatusClass =
  "border-[var(--calendar-story-border)] bg-[var(--calendar-story-background)]";
const scheduledStoryStatusHoverClass = "hover:bg-[var(--calendar-story-hover)]";
const completedCalendarBlockPattern =
  "repeating-linear-gradient(135deg, transparent 0, transparent 5px, rgba(100, 116, 139, 0.12) 5px, rgba(100, 116, 139, 0.12) 8px)";

type CalendarItem =
  | {
      kind: "event";
      id: string;
      startAt: string;
      endAt: string;
      event: CalendarEventSummary;
    }
  | {
      kind: "busy";
      id: string;
      startAt: string;
      endAt: string;
      window: CalendarBusyWindow;
    }
  | {
      kind: "block";
      id: string;
      startAt: string;
      endAt: string;
      block: CalendarScheduleBlock;
    };

type CalendarDragData = {
  kind: CalendarDragKind;
  block: CalendarScheduleBlock;
};

type CalendarManualChange = {
  block: CalendarScheduleBlock;
  change: CalendarDragKind;
  startAt: Date;
  endAt: Date;
};

const toDateTimeInputValue = (value: Date | string) =>
  format(new Date(value), "yyyy-MM-dd'T'HH:mm");

const toClockLabel = (value: Date, includePeriod: boolean) => {
  const timePattern = value.getMinutes() === 0 ? "h" : "h:mm";
  return format(
    value,
    includePeriod ? `${timePattern}a` : timePattern,
  ).toLowerCase();
};

const toTimeLabel = (startAt: string, endAt: string) => {
  const start = new Date(startAt);
  const end = new Date(endAt);
  const isSamePeriod = format(start, "a") === format(end, "a");
  return `${toClockLabel(start, !isSamePeriod)} – ${toClockLabel(end, true)}`;
};

const getUtcOffsetLabel = (date: Date) => {
  const offsetMinutes = -date.getTimezoneOffset();
  const sign = offsetMinutes >= 0 ? "+" : "-";
  const absoluteOffset = Math.abs(offsetMinutes);
  const hours = Math.floor(absoluteOffset / 60)
    .toString()
    .padStart(2, "0");
  const minutes = absoluteOffset % 60;
  return `GMT${sign}${hours}${minutes ? `:${minutes.toString().padStart(2, "0")}` : ""}`;
};

const getLocalTimeZoneName = () => {
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
  const location = timeZone.split("/").at(-1);
  return location?.replaceAll("_", " ") || "Local time";
};

const roundToNextHalfHour = (date: Date) => {
  const next = new Date(date);
  const minutes = next.getMinutes();
  next.setMinutes(minutes < 30 ? 30 : 60, 0, 0);
  return next;
};

const getCalendarDragData = (active: DragEndEvent["active"]) =>
  (active.data.current?.calendarDrag as CalendarDragData | undefined) ?? null;

const calendarCollisionDetection: CollisionDetection = (args) => {
  const pointerCollisions = pointerWithin(args);
  return pointerCollisions.length > 0
    ? pointerCollisions
    : rectIntersection(args);
};

const snapCalendarDragModifier: Modifier = ({ active, transform }) => {
  if (!active?.data.current?.calendarDrag) return transform;
  return {
    ...transform,
    y: snapCalendarDeltaPixels(transform.y, hourHeight),
  };
};

const calendarDragModifiers = [snapCalendarDragModifier];

const CalendarDragPreview = ({ drag }: { drag: CalendarDragData | null }) => {
  if (!drag) return null;
  if (drag.kind === "resize") {
    return <Box className="bg-border-strong h-full w-full rounded-sm" />;
  }

  const storyStyle = getCalendarStoryBlockStyle(drag.block);

  return (
    <Box
      className={cn(
        "shadow-shadow relative h-full w-full overflow-hidden rounded-md border px-3 py-1 backdrop-blur-sm",
        storyStyle
          ? scheduledStoryStatusClass
          : "border-border-strong/60 bg-surface-muted/95",
      )}
      style={storyStyle}
    >
      <span
        aria-hidden="true"
        className={cn(
          "absolute top-1/2 left-1 h-[calc(100%-0.5rem)] w-[0.1875rem] -translate-y-1/2 rounded-md",
          storyStyle ? "bg-[var(--calendar-story-accent)]" : "bg-border-strong",
        )}
      />
      <Text className="truncate leading-tight" fontWeight="medium">
        {drag.block.title}
      </Text>
    </Box>
  );
};

const overlapsDay = (
  item: Pick<CalendarItem, "startAt" | "endAt">,
  day: Date,
) => {
  const dayStart = startOfDay(day);
  const dayEnd = addDays(dayStart, 1);
  return new Date(item.startAt) < dayEnd && new Date(item.endAt) > dayStart;
};

const calendarEventOverlapsDay = (event: CalendarEventSummary, day: Date) => {
  if (event.isAllDay) {
    const startDate = parseCalendarDate(event.startDate);
    const endDate = parseCalendarDate(event.endDate);
    if (startDate && endDate) {
      const dayStart = startOfDay(day);
      const dayEnd = addDays(dayStart, 1);
      return startDate < dayEnd && endDate > dayStart;
    }
  }
  return overlapsDay(event, day);
};

const getStoryCode = (story: Story) =>
  story.team?.code
    ? `${story.team.code}-${story.sequenceId}`
    : `#${story.sequenceId}`;

const getBusyWindowTitle = (window: CalendarBusyWindow) => {
  if (window.isPrivate) {
    return "Busy";
  }
  return window.title?.trim() || "Busy";
};

const CalendarTimedBlock = ({
  item,
  isDragDisabled,
  layout,
  onEdit,
  onSelectEvent,
  today,
}: {
  item: CalendarItem;
  isDragDisabled: boolean;
  layout: { top: number; height: number; lane: number; laneCount: number };
  onEdit: (block: CalendarScheduleBlock) => void;
  onSelectEvent: (event: CalendarEventSummary) => void;
  today: Date;
}) => {
  const draggableBlock =
    item.kind === "block" &&
    !item.block.isCrossWorkspace &&
    item.block.blockType === "work" &&
    item.block.storyId
      ? item.block
      : null;
  const moveDrag = useDraggable({
    id: `calendar-move-${item.id}`,
    disabled: !draggableBlock || isDragDisabled,
    data: draggableBlock
      ? { calendarDrag: { kind: "move", block: draggableBlock } }
      : undefined,
  });
  const resizeDrag = useDraggable({
    id: `calendar-resize-${item.id}`,
    disabled: !draggableBlock || isDragDisabled,
    data: draggableBlock
      ? { calendarDrag: { kind: "resize", block: draggableBlock } }
      : undefined,
  });
  const laneWidth = 100 / layout.laneCount;
  const isCompleted = new Date(item.endAt).getTime() <= today.getTime();
  const resizeDelta = resizeDrag.isDragging
    ? snapCalendarDeltaPixels(resizeDrag.transform?.y ?? 0, hourHeight)
    : 0;
  const renderedHeight = Math.max(
    18,
    layout.height - timedBlockVerticalGap + resizeDelta,
  );
  const style = {
    height: `${renderedHeight}px`,
    left: `calc(${layout.lane * laneWidth}% + 0.25rem)`,
    top: `${layout.top + timedBlockVerticalInset}px`,
    width: `calc(${laneWidth}% - 0.5rem)`,
  };
  const isCompactEvent =
    item.kind === "event" &&
    renderedHeight < hourHeight - timedBlockVerticalGap;
  const showSecondaryLine =
    item.kind === "event" ? renderedHeight >= 31 : renderedHeight >= 40;
  const canShowTwoLineTitle = layout.height >= twoLineTitleMinimumHeight;
  const titleLineClass = canShowTwoLineTitle
    ? "line-clamp-2 leading-5"
    : "truncate leading-[0.9375rem]";
  const secondaryLineClass = canShowTwoLineTitle
    ? "mt-1 leading-[1.1rem]"
    : "mt-0.5 leading-[0.9375rem]";
  const blockPaddingClass =
    layout.height >= hourHeight ? "py-1 pr-2.5 pl-3" : "py-px pr-2.5 pl-3";

  if (item.kind === "event") {
    const EventIcon = item.event.meetingUrl ? Video02Icon : CalendarIcon;
    const eventTextClass = isCompleted ? "text-text-muted" : "text-[#3c90ff]";
    const eventAccentClass = isCompleted ? "bg-border-strong" : "bg-[#3c90ff]";

    return (
      <button
        aria-label={`Open ${getCalendarEventTitle(item.event)} details, ${getCalendarEventTimeLabel(item.event)}`}
        className={cn(
          "absolute overflow-hidden rounded-md border text-left backdrop-blur-sm transition-colors focus-visible:ring-2 focus-visible:outline-none",
          isCompleted
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
            "absolute top-1/2 left-1 h-[calc(100%-0.5rem)] w-[0.1875rem] -translate-y-1/2 rounded-md",
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
          isCompleted
            ? "border-border-strong/40 bg-surface-muted/55 dark:bg-surface-elevated/55 bg-[repeating-linear-gradient(135deg,transparent_0,transparent_5px,rgba(100,116,139,0.12)_5px,rgba(100,116,139,0.12)_8px)]"
            : "border-border-strong/40 bg-surface-muted/35 bg-[repeating-linear-gradient(135deg,transparent_0,transparent_5px,rgba(148,163,184,0.08)_5px,rgba(148,163,184,0.08)_8px)]",
          blockPaddingClass,
        )}
        style={style}
      >
        <span
          aria-hidden="true"
          className="bg-border-strong absolute top-1/2 left-1 h-[calc(100%-0.5rem)] w-[0.1875rem] -translate-y-1/2 rounded-md"
        />
        <Text
          className={cn(
            isCompleted ? "text-text-muted" : "text-foreground",
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
              isCompleted ? "text-text-muted" : "text-foreground",
            )}
          >
            {toTimeLabel(item.startAt, item.endAt)}
          </Text>
        ) : null}
      </Box>
    );
  }

  const { block } = item;
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
      ? scheduledStoryStatusClass
      : cn(
          "border-border-strong/70 dark:border-border-strong",
          scheduledTaskBackgroundClass,
        );
    if (!isMayaManaged) {
      blockColorClass = cn(
        blockColorClass,
        storyStyle
          ? scheduledStoryStatusHoverClass
          : scheduledTaskHoverBackgroundClass,
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
      storyStyle ? scheduledStoryStatusClass : scheduledTaskBackgroundClass,
    );
  }
  if (isCrossWorkspace) {
    blockColorClass = RESERVED_TIME_BLOCK_CLASS;
  }
  const isScheduledStory = block.blockType === "work";
  const isStandardHeightBlock =
    layout.height >= hourHeight && layout.height < twoLineTitleMinimumHeight;
  const hasLeadingIcon = isCrossWorkspace || !isScheduledStory;
  const timeLabel = toTimeLabel(block.startAt, block.endAt);
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
  }
  let blockTooltip = block.storyId ? undefined : mayaReason ?? undefined;
  if (isCrossWorkspace) {
    blockTooltip = CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP;
  }
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
      {...moveDrag.attributes}
      {...moveDrag.listeners}
      className={cn(
        "absolute flex items-center overflow-hidden rounded-md border backdrop-blur-sm transition-colors",
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
          ? { backgroundImage: completedCalendarBlockPattern }
          : null),
      }}
      title={blockTooltip}
    >
      <span
        aria-hidden="true"
        className={cn(
          "absolute top-1/2 left-1 h-[calc(100%-0.5rem)] w-[0.1875rem] -translate-y-1/2 rounded-md",
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
          "pointer-events-none relative z-10 min-w-0",
          hasLeadingIcon
            ? cn(
                "flex gap-1.5",
                showSecondaryLine ? "items-start" : "items-center",
              )
            : null,
        )}
      >
        {hasLeadingIcon ? (
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
      {draggableBlock ? (
        <button
          {...resizeDrag.attributes}
          {...resizeDrag.listeners}
          aria-label={`Resize ${blockTitle}`}
          className={cn(
            "absolute inset-x-1 bottom-0 z-20 h-2 rounded-sm focus-visible:ring-2 focus-visible:outline-none focus-visible:ring-inset",
            isDragDisabled ? "cursor-wait" : "cursor-ns-resize",
          )}
          disabled={isDragDisabled}
          onPointerDown={(event) => {
            event.stopPropagation();
          }}
          ref={resizeDrag.setNodeRef}
          type="button"
        />
      ) : null}
    </div>
  );

  return blockContent;
};

const CalendarAllDayEvent = ({
  event,
  onSelect,
  today,
}: {
  event: CalendarEventSummary;
  onSelect: (event: CalendarEventSummary) => void;
  today: Date;
}) => {
  const isCompleted = new Date(event.endAt).getTime() <= today.getTime();

  return (
    <button
      aria-label={`Open ${getCalendarEventTitle(event)} details, ${getCalendarEventTimeLabel(event)}`}
      className={cn(
        "w-full truncate rounded-md border px-3 py-1.5 text-left text-base font-medium backdrop-blur-sm transition-colors focus-visible:outline-none",
        isCompleted
          ? "border-border-strong/40 bg-surface-muted/55 text-text-muted dark:bg-surface-elevated/55 hover:bg-surface-muted/65 dark:hover:bg-surface-elevated/65 bg-[repeating-linear-gradient(135deg,transparent_0,transparent_5px,rgba(100,116,139,0.12)_5px,rgba(100,116,139,0.12)_8px)]"
          : "border-[#3c90ff]/80 bg-[#3c90ff]/20 text-[#3c90ff] hover:bg-[#3c90ff]/25 focus-visible:ring-2 focus-visible:ring-[#3c90ff]/50",
      )}
      onClick={() => {
        onSelect(event);
      }}
      type="button"
    >
      {getCalendarEventTitle(event)}
    </button>
  );
};

const CalendarStoryPicker = ({
  onSelect,
  selectedStoryId,
  stories,
  storyTerm,
  storyTermPlural,
}: {
  onSelect: (storyId: string) => void;
  selectedStoryId: string;
  stories: Story[];
  storyTerm: string;
  storyTermPlural: string;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState("");
  const selectedStory = stories.find((story) => story.id === selectedStoryId);

  const changeOpen = (open: boolean) => {
    setIsOpen(open);
    if (!open) {
      setQuery("");
    }
  };

  return (
    <Popover onOpenChange={changeOpen} open={isOpen}>
      <Popover.Trigger asChild>
        <Button
          align="between"
          aria-expanded={isOpen}
          className="min-w-0 text-base"
          color="tertiary"
          fullWidth
          rightIcon={<ArrowDown2Icon className="h-4 shrink-0" />}
          size="md"
          variant="outline"
        >
          {selectedStory ? (
            <span className="min-w-0 flex-1 truncate text-left">
              <span className="text-text-muted mr-2">
                {getStoryCode(selectedStory)}
              </span>
              {selectedStory.title}
            </span>
          ) : (
            <span className="text-text-muted">Select {storyTerm}</span>
          )}
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="start"
        className="border-border-strong w-[var(--radix-popover-trigger-width)] max-w-[34rem] min-w-[22rem] border"
      >
        <Command>
          <Command.Input
            aria-label={`Search ${storyTermPlural}`}
            autoFocus
            className="text-base"
            onValueChange={setQuery}
            placeholder={`Search ${storyTermPlural}...`}
            value={query}
          />
          <Divider className="my-2" />
          <Command.List className="max-h-72 overflow-y-auto">
            <Command.Empty className="py-4 text-base">
              <Text color="muted" fontSize="md">
                No matching {storyTermPlural}.
              </Text>
            </Command.Empty>
            <Command.Group>
              {stories.map((story) => {
                const code = getStoryCode(story);
                const isSelected = story.id === selectedStoryId;
                return (
                  <Command.Item
                    active={isSelected}
                    className="justify-between gap-4 py-2 text-base"
                    key={story.id}
                    onSelect={() => {
                      onSelect(story.id);
                      changeOpen(false);
                    }}
                    value={`${code} ${story.title}`}
                  >
                    <Flex align="center" className="min-w-0" gap={2}>
                      <StoryIcon className="h-[1.1rem] shrink-0" />
                      <Text className="min-w-0 truncate" fontSize="md">
                        <span className="text-text-muted mr-2">{code}</span>
                        {story.title}
                      </Text>
                    </Flex>
                    {isSelected ? (
                      <CheckIcon className="h-5 shrink-0" strokeWidth={2.1} />
                    ) : null}
                  </Command.Item>
                );
              })}
            </Command.Group>
          </Command.List>
        </Command>
      </Popover.Content>
    </Popover>
  );
};

const CalendarDialog = ({
  candidateStories,
  editingBlock,
  isOpen,
  mode,
  onOpenChange,
}: {
  candidateStories: Story[];
  editingBlock: CalendarScheduleBlock | null;
  isOpen: boolean;
  mode: "work" | "focus";
  onOpenChange: (value: boolean) => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const createBlock = useCreateCalendarScheduleBlock();
  const updateBlock = useUpdateCalendarScheduleBlock();
  const deleteBlock = useDeleteCalendarScheduleBlock();
  const defaultStart = roundToNextHalfHour(addHours(new Date(), 1));
  const defaultStoryId = candidateStories.at(0)?.id ?? "";
  const [selectedStoryId, setSelectedStoryId] = useState(
    editingBlock?.storyId ?? defaultStoryId,
  );
  const [title, setTitle] = useState(editingBlock?.title ?? "Focus time");
  const [startAt, setStartAt] = useState(() =>
    toDateTimeInputValue(editingBlock?.startAt ?? defaultStart),
  );
  const [endAt, setEndAt] = useState(() =>
    toDateTimeInputValue(editingBlock?.endAt ?? addHours(defaultStart, 1)),
  );
  const storyTerm = getTermDisplay("storyTerm");
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const selectedStory = candidateStories.find(
    (story) => story.id === selectedStoryId,
  );
  const isWork = mode === "work";
  const parsedStartAt = new Date(startAt);
  const parsedEndAt = new Date(endAt);
  const earliestScheduleAt = addDays(new Date(), -calendarHistoryDays);
  const latestScheduleAt = addDays(new Date(), calendarLookaheadDays);
  const hasChronologicalRange =
    Number.isFinite(parsedStartAt.getTime()) &&
    Number.isFinite(parsedEndAt.getTime()) &&
    parsedEndAt.getTime() > parsedStartAt.getTime();
  const isWithinScheduleHorizon =
    hasChronologicalRange &&
    parsedStartAt >= earliestScheduleAt &&
    parsedEndAt <= latestScheduleAt;
  const hasRequiredContent = isWork
    ? Boolean(selectedStoryId)
    : title.trim().length > 0;
  const canSubmit =
    hasRequiredContent && hasChronologicalRange && isWithinScheduleHorizon;
  const isSaving = createBlock.isPending || updateBlock.isPending;
  let dialogTitle = "Add focus time";
  if (editingBlock) {
    dialogTitle = "Edit calendar block";
  } else if (isWork) {
    dialogTitle = `Schedule ${storyTerm}`;
  }

  const close = () => {
    onOpenChange(false);
  };

  const submit = () => {
    if (!canSubmit) {
      return;
    }
    const input: CalendarScheduleBlockInput = {
      blockType: mode,
      title: isWork
        ? selectedStory?.title ?? editingBlock?.title ?? storyTerm
        : title,
      storyId: isWork ? selectedStoryId : null,
      startAt: new Date(startAt).toISOString(),
      endAt: new Date(endAt).toISOString(),
      isLocked: true,
    };

    if (editingBlock) {
      updateBlock.mutate(
        { blockId: editingBlock.id, input },
        { onSuccess: close },
      );
      return;
    }
    createBlock.mutate(input, { onSuccess: close });
  };

  const handleDelete = () => {
    if (!editingBlock) {
      return;
    }
    deleteBlock.mutate(editingBlock.id, { onSuccess: close });
  };

  return (
    <Dialog onOpenChange={onOpenChange} open={isOpen}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-0.5 text-lg">
            {dialogTitle}
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-4">
          {isWork ? (
            <Box>
              <Text className="mb-2" fontSize="md" fontWeight="medium">
                {storyTerm}
              </Text>
              <CalendarStoryPicker
                onSelect={setSelectedStoryId}
                selectedStoryId={selectedStoryId}
                stories={candidateStories}
                storyTerm={storyTerm}
                storyTermPlural={storyTermPlural}
              />
              {candidateStories.length === 0 ? (
                <Text className="mt-2" color="muted" fontSize="md">
                  No assigned{" "}
                  {getTermDisplay("storyTerm", { variant: "plural" })} found.
                </Text>
              ) : null}
            </Box>
          ) : (
            <Input
              className="text-base"
              label="Title"
              labelClassName="text-base"
              onChange={(event) => {
                setTitle(event.target.value);
              }}
              value={title}
            />
          )}
          <Box className="grid grid-cols-1 gap-3 md:grid-cols-2">
            <Input
              className="text-base"
              label="Start"
              labelClassName="text-base"
              max={toDateTimeInputValue(latestScheduleAt)}
              min={toDateTimeInputValue(earliestScheduleAt)}
              onChange={(event) => {
                setStartAt(event.target.value);
              }}
              type="datetime-local"
              value={startAt}
            />
            <Input
              className="text-base"
              label="End"
              labelClassName="text-base"
              max={toDateTimeInputValue(latestScheduleAt)}
              min={toDateTimeInputValue(earliestScheduleAt)}
              onChange={(event) => {
                setEndAt(event.target.value);
              }}
              type="datetime-local"
              value={endAt}
            />
          </Box>
          {!hasChronologicalRange ? (
            <Text color="danger" fontSize="md">
              End time must be after start time.
            </Text>
          ) : null}
          {hasChronologicalRange && !isWithinScheduleHorizon ? (
            <Text color="danger" fontSize="md">
              Choose a time from the last {calendarHistoryDays} days through the
              next {calendarLookaheadDays} days.
            </Text>
          ) : null}
        </Dialog.Body>
        <Dialog.Footer className="justify-between gap-3 border-0 pt-2">
          {editingBlock ? (
            <Button
              className="text-base"
              color="danger"
              leftIcon={<DeleteIcon className="h-4" />}
              loading={deleteBlock.isPending}
              onClick={handleDelete}
              variant="naked"
            >
              Delete
            </Button>
          ) : (
            <span />
          )}
          <Flex align="center" gap={2}>
            <Button
              className="text-base"
              color="tertiary"
              onClick={close}
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              className="text-base"
              color="invert"
              disabled={!canSubmit}
              loading={isSaving}
              onClick={submit}
            >
              {editingBlock ? "Save" : "Add"}
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

const CalendarToolbar = ({
  canNavigateNext,
  canNavigatePrevious,
  currentView,
  onFocus,
  onNext,
  onPrevious,
  onToday,
  onViewChange,
  title,
}: {
  canNavigateNext: boolean;
  canNavigatePrevious: boolean;
  currentView: CalendarView;
  onFocus: () => void;
  onNext: () => void;
  onPrevious: () => void;
  onToday: () => void;
  onViewChange: (view: CalendarView) => void;
  title: string;
}) => (
  <Flex
    align="center"
    className="border-border/70 h-16 shrink-0 gap-5 overflow-x-auto border-b px-5 py-3"
    justify="between"
  >
    <Flex align="center" className="shrink-0" gap={3}>
      <Flex align="center" gap={1}>
        <Button
          aria-label={`Previous ${currentView}`}
          asIcon
          className="focus-visible:ring-primary/40 focus-visible:ring-2"
          color="tertiary"
          disabled={!canNavigatePrevious}
          onClick={onPrevious}
          size="sm"
          variant="naked"
        >
          <ChevronLeftIcon className="h-5" />
        </Button>
        <Button
          aria-label={`Next ${currentView}`}
          asIcon
          className="focus-visible:ring-primary/40 focus-visible:ring-2"
          color="tertiary"
          disabled={!canNavigateNext}
          onClick={onNext}
          size="sm"
          variant="naked"
        >
          <ChevronRightIcon className="h-5" />
        </Button>
      </Flex>
      <Text
        as="h2"
        className="whitespace-nowrap"
        fontSize="xl"
        fontWeight="medium"
      >
        {title}
      </Text>
    </Flex>
    <Flex align="center" className="shrink-0" gap={2}>
      <Button color="tertiary" onClick={onToday} size="sm">
        Today
      </Button>
      <Menu>
        <Menu.Button>
          <Button
            className="justify-between capitalize"
            color="tertiary"
            rightIcon={<ArrowDown2Icon className="h-4" />}
            size="sm"
            variant="outline"
          >
            {currentView}
          </Button>
        </Menu.Button>
        <Menu.Items align="end" className="w-36">
          <Menu.Group>
            {calendarViews.map((view) => (
              <Menu.Item
                active={currentView === view}
                className="py-2.5 text-base capitalize"
                key={view}
                onSelect={() => {
                  onViewChange(view);
                }}
              >
                {view}
              </Menu.Item>
            ))}
          </Menu.Group>
        </Menu.Items>
      </Menu>
      <span className="text-text-secondary mx-1 hidden opacity-40 md:inline">
        |
      </span>
      <Button
        color="tertiary"
        leftIcon={<TimeScheduleIcon className="h-[1.1rem]" strokeWidth={2} />}
        onClick={onFocus}
        size="sm"
        variant="outline"
      >
        Block focus time
      </Button>
    </Flex>
  </Flex>
);

const CalendarNotices = ({
  canReadEventDetails,
  conflictCount,
  connectHref,
  connection,
  hasIntegrationError,
  isIntegrationPending,
  isReconnectPending,
  isSyncing,
  onReconnect,
  onSync,
}: {
  canReadEventDetails: boolean;
  conflictCount: number;
  connectHref: string;
  connection?: CalendarConnection;
  hasIntegrationError: boolean;
  isIntegrationPending: boolean;
  isReconnectPending: boolean;
  isSyncing: boolean;
  onReconnect: () => void;
  onSync: (connectionId: string) => void;
}) => (
  <>
    {!isIntegrationPending && !hasIntegrationError && !connection ? (
      <Box className="border-border border-b px-5 py-3">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" className="min-w-0" gap={3}>
            <Box className="flex h-10 w-10 shrink-0 items-center justify-center">
              <GoogleCalendarIcon aria-hidden="true" className="h-6 w-6" />
            </Box>
            <Box className="min-w-0">
              <Text fontSize="md" fontWeight="medium">
                Google Calendar is not connected
              </Text>
              <Text className="line-clamp-1" color="muted" fontSize="md">
                FortyOne can still schedule work blocks, but availability will
                be incomplete until you connect your primary calendar.
              </Text>
            </Box>
          </Flex>
          <Button
            className="text-base"
            color="tertiary"
            href={connectHref}
            variant="outline"
          >
            Connect
          </Button>
        </Flex>
      </Box>
    ) : null}

    {connection?.requiresReauthorization ? (
      <Box className="border-border bg-surface-prominent/30 border-b px-5 py-3">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" className="min-w-0" gap={3}>
            <Box className="flex h-10 w-10 shrink-0 items-center justify-center">
              <GoogleCalendarIcon aria-hidden="true" className="h-6 w-6" />
            </Box>
            <Box className="min-w-0">
              <Text fontSize="md" fontWeight="medium">
                Reconnect to update Google Calendar work blocks
              </Text>
              <Text className="line-clamp-1" color="muted" fontSize="md">
                Reconnect once to let FortyOne add and update scheduled work in
                your primary calendar.
              </Text>
            </Box>
          </Flex>
          <Button
            className="text-base"
            color="tertiary"
            loading={isReconnectPending}
            onClick={onReconnect}
            variant="outline"
          >
            Reconnect
          </Button>
        </Flex>
      </Box>
    ) : null}

    {connection &&
    !connection.requiresReauthorization &&
    !canReadEventDetails ? (
      <Box className="border-border border-b px-5 py-3">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" className="min-w-0" gap={3}>
            <Box className="flex h-10 w-10 shrink-0 items-center justify-center">
              <GoogleCalendarIcon aria-hidden="true" className="h-6 w-6" />
            </Box>
            <Box className="min-w-0">
              <Text fontSize="md" fontWeight="medium">
                Event details are not enabled
              </Text>
              <Text className="line-clamp-1" color="muted" fontSize="md">
                Reconnect your primary Google Calendar to show event titles
                instead of availability-only busy blocks.
              </Text>
            </Box>
          </Flex>
          <Button
            className="text-base"
            color="tertiary"
            loading={isReconnectPending}
            onClick={onReconnect}
            variant="outline"
          >
            Reconnect
          </Button>
        </Flex>
      </Box>
    ) : null}

    {connection?.syncStatus === "failed" ? (
      <Box className="border-warning/30 border-b px-5 py-3">
        <Flex align="center" gap={3} justify="between">
          <Box className="min-w-0">
            <Text fontSize="md" fontWeight="medium">
              Calendar sync failed
            </Text>
            <Text color="muted" fontSize="md">
              {connection.syncError?.trim() ||
                "Google Calendar could not be refreshed."}{" "}
              {connection.lastSyncedAt
                ? `Showing the last successful sync from ${format(new Date(connection.lastSyncedAt), "MMM d 'at' h:mm a")}.`
                : "No successful calendar sync is available yet."}
            </Text>
          </Box>
          <Button
            className="text-base"
            color="tertiary"
            loading={isSyncing}
            onClick={() => {
              onSync(connection.id);
            }}
            variant="outline"
          >
            Retry
          </Button>
        </Flex>
      </Box>
    ) : null}

    {conflictCount > 0 ? (
      <Box className="border-danger/30 border-b px-5 py-3">
        <Text fontSize="md" fontWeight="medium">
          {conflictCount === 1
            ? "A scheduled block now overlaps a meeting"
            : `${conflictCount} scheduled blocks now overlap meetings`}
        </Text>
        <Text color="muted" fontSize="md">
          Open a red block to choose another time. FortyOne will not move locked
          work without your approval.
        </Text>
      </Box>
    ) : null}
  </>
);

const CalendarTimedDayColumn = ({
  day,
  dayItems,
  hours,
  isDragDisabled,
  onEdit,
  onSelectEvent,
  today,
  visibleEndHour,
  visibleStartHour,
}: {
  day: Date;
  dayItems: CalendarItem[];
  hours: number[];
  isDragDisabled: boolean;
  onEdit: (block: CalendarScheduleBlock) => void;
  onSelectEvent: (event: CalendarEventSummary) => void;
  today: Date;
  visibleEndHour: number;
  visibleStartHour: number;
}) => {
  const { isOver, setNodeRef } = useDroppable({
    id: `calendar-day-${day.toISOString()}`,
    data: { calendarDay: day.toISOString() },
  });
  const layouts = buildCalendarEventLayouts({
    day,
    events: dayItems.map((item) => ({
      id: `${item.kind}-${item.id}`,
      startAt: item.startAt,
      endAt: item.endAt,
    })),
    hourHeight,
    visibleEndHour,
    visibleStartHour,
  });
  const layoutById = new Map(layouts.map((layout) => [layout.id, layout]));

  return (
    <div
      className={cn(
        "border-border/60 relative border-l transition-colors duration-100",
        isOver ? "bg-state-hover/30" : null,
      )}
      ref={setNodeRef}
      style={{
        height: `${(visibleEndHour - visibleStartHour) * hourHeight}px`,
      }}
    >
      {hours.slice(1, -1).map((hour) => (
        <Box
          className="border-border/60 absolute inset-x-0 border-t"
          key={hour}
          style={{
            top: `${(hour - visibleStartHour) * hourHeight}px`,
          }}
        />
      ))}
      {dayItems.map((item) => {
        const key = `${item.kind}-${item.id}`;
        const layout = layoutById.get(key);
        if (!layout) return null;
        return (
          <CalendarTimedBlock
            isDragDisabled={isDragDisabled}
            item={item}
            key={key}
            layout={layout}
            onEdit={onEdit}
            onSelectEvent={onSelectEvent}
            today={today}
          />
        );
      })}
    </div>
  );
};

const CalendarTimeGrid = ({
  allDayEvents,
  days,
  hours,
  isDaySelectable,
  isManualChangePending,
  onEdit,
  onManualChange,
  onSelectDay,
  onSelectEvent,
  timedCalendarItems,
  timeZoneLabel,
  timeZoneName,
  today,
  visibleEndHour,
  visibleStartHour,
}: {
  allDayEvents: CalendarEventSummary[];
  days: Date[];
  hours: number[];
  isDaySelectable: (day: Date) => boolean;
  isManualChangePending: boolean;
  onEdit: (block: CalendarScheduleBlock) => void;
  onManualChange: (change: CalendarManualChange) => void;
  onSelectDay: (day: Date) => void;
  onSelectEvent: (event: CalendarEventSummary) => void;
  timedCalendarItems: CalendarItem[];
  timeZoneLabel: string;
  timeZoneName: string;
  today: Date;
  visibleEndHour: number;
  visibleStartHour: number;
}) => {
  const isDayView = days.length === 1;
  const dayColumn = isDayView ? "minmax(0, 1fr)" : "minmax(9.5rem, 1fr)";
  const gridTemplateColumns = `${timeRailWidth}rem repeat(${days.length}, ${dayColumn})`;
  const minimumWidthClass = isDayView ? "min-w-full" : "min-w-[72rem]";
  const [activeDrag, setActiveDrag] = useState<CalendarDragData | null>(null);
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } }),
    useSensor(TouchSensor, {
      activationConstraint: { delay: 150, tolerance: 8 },
    }),
  );
  const handleDragStart = (event: DragStartEvent) => {
    setActiveDrag(getCalendarDragData(event.active));
  };
  const handleDragEnd = (event: DragEndEvent) => {
    const dragData = getCalendarDragData(event.active);
    const targetDay = event.over?.data.current?.calendarDay as
      | string
      | undefined;
    setActiveDrag(null);
    if (!dragData || !targetDay) return;

    const originalStart = new Date(dragData.block.startAt);
    const originalEnd = new Date(dragData.block.endAt);
    const { endAt, startAt } = getCalendarManualChange({
      block: dragData.block,
      deltaY: event.delta.y,
      hourHeight,
      kind: dragData.kind,
      targetDay: new Date(targetDay),
    });

    if (
      startAt.getTime() === originalStart.getTime() &&
      endAt.getTime() === originalEnd.getTime()
    ) {
      return;
    }
    onManualChange({
      block: dragData.block,
      change: dragData.kind,
      startAt,
      endAt,
    });
  };

  return (
    <DndContext
      collisionDetection={calendarCollisionDetection}
      modifiers={calendarDragModifiers}
      onDragCancel={() => {
        setActiveDrag(null);
      }}
      onDragEnd={handleDragEnd}
      onDragStart={handleDragStart}
      sensors={sensors}
    >
      <Box className="min-h-0 flex-1 overflow-auto overscroll-contain">
        <Box
          className={cn(
            "border-border/70 bg-background sticky top-0 z-30 grid h-18 border-b",
            minimumWidthClass,
          )}
          style={{ gridTemplateColumns }}
        >
          <Box className="flex flex-col items-center justify-between px-3 pt-2.5 pb-2 text-center">
            <Text
              className="max-w-full truncate"
              fontSize="md"
              fontWeight="medium"
            >
              {timeZoneName}
            </Text>
            <Text className="tabular-nums" color="muted" fontSize="md">
              {timeZoneLabel}
            </Text>
          </Box>
          {days.map((day) => {
            const isToday = isSameDay(day, today);
            const canSelectDay = isDaySelectable(day);
            return (
              <Box
                className="border-border/60 flex flex-col items-center justify-between border-l px-3 pt-2.5 pb-1"
                key={day.toISOString()}
              >
                <Text
                  className="text-[0.875rem] leading-none tracking-[0.08em]"
                  color={isToday ? "primary" : "muted"}
                  fontWeight="medium"
                  transform="uppercase"
                >
                  {format(day, "EEE")}
                </Text>
                <button
                  aria-current={isToday ? "date" : undefined}
                  aria-label={`Open ${format(day, "MMMM d, yyyy")} in day view`}
                  className={cn(
                    "group focus-visible:ring-primary/40 grid size-10 place-items-center rounded-full focus-visible:ring-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40",
                    isToday
                      ? "text-primary-foreground"
                      : "text-foreground hover:bg-state-hover",
                  )}
                  disabled={!canSelectDay}
                  onClick={() => {
                    onSelectDay(day);
                  }}
                  type="button"
                >
                  <Text
                    as="span"
                    className={cn(
                      "grid place-items-center leading-none text-current tabular-nums",
                      isToday
                        ? "bg-primary group-hover:bg-primary/90 size-8 rounded-full text-base transition-colors"
                        : "text-lg",
                    )}
                    fontWeight="medium"
                  >
                    {format(day, "d")}
                  </Text>
                </button>
              </Box>
            );
          })}
        </Box>
        {allDayEvents.length > 0 ? (
          <Box
            className={cn(
              "border-border/70 bg-background sticky top-18 z-20 grid min-h-14 border-b",
              minimumWidthClass,
            )}
            style={{ gridTemplateColumns }}
          >
            <Box className="flex items-center justify-center px-3 py-2">
              <Text color="muted" fontSize="md">
                All day
              </Text>
            </Box>
            {days.map((day) => {
              const dayEvents = allDayEvents.filter((event) =>
                calendarEventOverlapsDay(event, day),
              );
              return (
                <Box
                  className="border-border/60 min-h-14 space-y-1.5 border-l p-2"
                  key={day.toISOString()}
                >
                  {dayEvents.map((event) => (
                    <CalendarAllDayEvent
                      event={event}
                      key={event.id}
                      onSelect={onSelectEvent}
                      today={today}
                    />
                  ))}
                </Box>
              );
            })}
          </Box>
        ) : null}
        <Box
          className={cn("grid", minimumWidthClass)}
          style={{ gridTemplateColumns }}
        >
          <Box className="relative">
            {hours.slice(1, -1).map((hour) => (
              <Box
                className="absolute right-4 -translate-y-1/2"
                key={hour}
                style={{ top: `${(hour - visibleStartHour) * hourHeight}px` }}
              >
                <Text
                  className="text-[0.9375rem] tabular-nums"
                  color="muted"
                  fontWeight="medium"
                >
                  {format(new Date(2026, 0, 1, hour), "h a")}
                </Text>
              </Box>
            ))}
          </Box>
          {days.map((day) => (
            <CalendarTimedDayColumn
              day={day}
              dayItems={timedCalendarItems.filter((item) =>
                overlapsDay(item, day),
              )}
              hours={hours}
              isDragDisabled={isManualChangePending}
              key={day.toISOString()}
              onEdit={onEdit}
              onSelectEvent={onSelectEvent}
              today={today}
              visibleEndHour={visibleEndHour}
              visibleStartHour={visibleStartHour}
            />
          ))}
        </Box>
      </Box>
      <DragOverlay
        className="pointer-events-none"
        dropAnimation={null}
        zIndex={50}
      >
        <CalendarDragPreview drag={activeDrag} />
      </DragOverlay>
    </DndContext>
  );
};

const getMonthItemTitle = (item: CalendarItem) => {
  if (item.kind === "event") {
    return getCalendarEventTitle(item.event);
  }
  if (item.kind === "busy") {
    return getBusyWindowTitle(item.window);
  }
  return item.block.title;
};

const CalendarMonthItem = ({
  item,
  onEdit,
  onSelectEvent,
}: {
  item: CalendarItem;
  onEdit: (block: CalendarScheduleBlock) => void;
  onSelectEvent: (event: CalendarEventSummary) => void;
}) => {
  const title = getMonthItemTitle(item);
  const start = new Date(item.startAt);
  const time = toClockLabel(start, true);

  if (item.kind === "event") {
    if (item.event.isAllDay) {
      return (
        <button
          aria-label={`Open ${title} details, ${getCalendarEventTimeLabel(item.event)}`}
          className="w-full truncate rounded-md border border-[#3c90ff]/80 bg-[#3c90ff]/20 px-2 py-1 text-left text-base font-medium text-[#3c90ff] backdrop-blur-sm transition-colors hover:bg-[#3c90ff]/25 focus-visible:ring-2 focus-visible:ring-[#3c90ff]/50 focus-visible:outline-none"
          onClick={() => {
            onSelectEvent(item.event);
          }}
          type="button"
        >
          {title}
        </button>
      );
    }

    return (
      <button
        aria-label={`Open ${title} details, ${getCalendarEventTimeLabel(item.event)}`}
        className="hover:bg-state-hover focus-visible:ring-primary/40 flex w-full min-w-0 items-center gap-2 rounded-md px-2 py-1 text-left text-base focus-visible:ring-2 focus-visible:outline-none"
        onClick={() => {
          onSelectEvent(item.event);
        }}
        type="button"
      >
        <span
          aria-hidden="true"
          className="size-2 shrink-0 rounded-full border border-[#3c90ff]"
        />
        <span className="shrink-0 text-[#3c90ff] tabular-nums">{time}</span>
        <span className="text-foreground truncate">{title}</span>
      </button>
    );
  }

  if (item.kind === "busy") {
    return (
      <Flex
        align="center"
        className="min-w-0 gap-2 rounded-md px-2 py-1 text-base"
      >
        <span
          aria-hidden="true"
          className="border-border-strong size-2 shrink-0 rounded-full border"
        />
        <span className="text-text-muted shrink-0 tabular-nums">{time}</span>
        <span className="text-foreground truncate">{title}</span>
      </Flex>
    );
  }

  const { block } = item;
  const isCrossWorkspace = Boolean(block.isCrossWorkspace);
  const isMayaManaged = block.source === "maya";
  const isEditable = isCalendarScheduleBlockEditable(block);
  const mayaLabel = getMayaCalendarBlockLabel(block);
  const mayaReason = getMayaCalendarBlockReason(block);
  let displayTitle = mayaLabel ? `${mayaLabel} · ${title}` : title;
  if (isCrossWorkspace) {
    displayTitle = CROSS_WORKSPACE_CALENDAR_BLOCK_TITLE;
  }
  const isScheduledStory = block.blockType === "work";
  const canOpenBlock = !isCrossWorkspace && (isScheduledStory || isEditable);
  const storyStyle = getCalendarStoryBlockStyle(block);
  let toneClass = isScheduledStory
    ? cn(
        storyStyle
          ? scheduledStoryStatusClass
          : cn("border-border-strong/70", scheduledTaskBackgroundClass),
        "text-text-muted",
      )
    : RESERVED_TIME_BLOCK_CLASS;
  if (block.hasConflict) {
    toneClass = "border-danger/60 bg-danger/20 text-danger";
  }
  if (isCrossWorkspace) {
    toneClass = RESERVED_TIME_BLOCK_CLASS;
  }
  let blockActionLabel = isScheduledStory
    ? `Open ${title} details`
    : `Edit ${title}`;
  if (block.hasConflict) {
    blockActionLabel = `Resolve conflict for ${title}`;
  }
  if (isMayaManaged) {
    if (isScheduledStory) {
      blockActionLabel = `Open Maya-managed ${title} details`;
    } else if (block.isLocked) {
      blockActionLabel = `Locked Maya-managed block for ${title}`;
    } else {
      blockActionLabel = `Maya-managed block for ${title}`;
    }
    if (mayaReason) {
      blockActionLabel += `. ${mayaReason}`;
    }
  }
  if (isCrossWorkspace) {
    blockActionLabel = CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP;
  }
  let blockTooltip = block.storyId ? undefined : mayaReason ?? undefined;
  if (isCrossWorkspace) {
    blockTooltip = CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP;
  }

  const blockContent = (
    <Box
      className={cn(
        "relative flex min-w-0 items-center gap-2 overflow-hidden rounded-md border px-2 py-1 text-base backdrop-blur-sm transition-colors",
        toneClass,
        storyStyle && !isMayaManaged ? scheduledStoryStatusHoverClass : null,
      )}
      style={storyStyle}
      title={blockTooltip}
    >
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
      {isCrossWorkspace ? (
        <TimeScheduleIcon
          aria-hidden="true"
          className="pointer-events-none size-4 shrink-0"
        />
      ) : (
        <span
          aria-hidden="true"
          className="pointer-events-none size-2 shrink-0 rounded-full bg-current"
          style={
            storyStyle ? { backgroundColor: block.storyStatusColor } : undefined
          }
        />
      )}
      <span className="pointer-events-none relative z-10 shrink-0 tabular-nums">
        {time}
      </span>
      <span className="pointer-events-none relative z-10 min-w-0 flex-1 truncate">
        {displayTitle}
      </span>
    </Box>
  );

  return blockContent;
};

const CalendarMonthGrid = ({
  calendarItems,
  cursor,
  days,
  isDaySelectable,
  onEdit,
  onSelectDay,
  onSelectEvent,
  today,
}: {
  calendarItems: CalendarItem[];
  cursor: Date;
  days: Date[];
  isDaySelectable: (day: Date) => boolean;
  onEdit: (block: CalendarScheduleBlock) => void;
  onSelectDay: (day: Date) => void;
  onSelectEvent: (event: CalendarEventSummary) => void;
  today: Date;
}) => {
  const weeks = Math.ceil(days.length / 7);
  const weekDays = days.slice(0, 7);
  const gridTemplateRows = `3.5rem repeat(${weeks}, minmax(9rem, 1fr))`;

  return (
    <Box className="min-h-0 flex-1 overflow-auto overscroll-contain">
      <Box
        className="grid min-h-full min-w-[72rem] grid-cols-7"
        style={{ gridTemplateRows }}
      >
        {weekDays.map((day, index) => (
          <Box
            className={cn(
              "border-border/60 bg-background sticky top-0 z-20 flex items-center justify-center border-b px-3",
              index > 0 && "border-l",
            )}
            key={`weekday-${day.toISOString()}`}
          >
            <Text
              className="text-[0.875rem] tracking-[0.08em]"
              color="muted"
              fontWeight="medium"
              transform="uppercase"
            >
              {format(day, "EEE")}
            </Text>
          </Box>
        ))}
        {days.map((day, index) => {
          const dayItems = calendarItems
            .filter((item) =>
              item.kind === "event"
                ? calendarEventOverlapsDay(item.event, day)
                : overlapsDay(item, day),
            )
            .sort((first, second) => {
              const firstIsAllDay =
                first.kind === "event" && first.event.isAllDay;
              const secondIsAllDay =
                second.kind === "event" && second.event.isAllDay;
              if (firstIsAllDay !== secondIsAllDay) {
                return firstIsAllDay ? -1 : 1;
              }
              return (
                new Date(first.startAt).getTime() -
                new Date(second.startAt).getTime()
              );
            });
          const visibleItems = dayItems.slice(0, 4);
          const hiddenItemCount = dayItems.length - visibleItems.length;
          const isToday = isSameDay(day, today);
          const isCurrentMonth = isSameMonth(day, cursor);
          const canSelectDay = isDaySelectable(day);
          const isLastColumn = index % 7 === 6;
          const isLastRow = index >= days.length - 7;

          return (
            <Box
              className={cn(
                "border-border/60 min-h-36 min-w-0 overflow-hidden p-3",
                !isLastColumn && "border-r",
                !isLastRow && "border-b",
              )}
              key={day.toISOString()}
            >
              <button
                aria-current={isToday ? "date" : undefined}
                aria-label={`Open ${format(day, "MMMM d, yyyy")} in day view`}
                className={cn(
                  "focus-visible:ring-primary/40 mb-2 grid size-9 place-items-center rounded-full text-base font-medium tabular-nums focus-visible:ring-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40",
                  isToday
                    ? "bg-primary text-primary-foreground"
                    : "hover:bg-state-hover",
                  !isToday && !isCurrentMonth && "text-text-muted",
                )}
                disabled={!canSelectDay}
                onClick={() => {
                  onSelectDay(day);
                }}
                type="button"
              >
                {format(day, "d")}
              </button>
              <Box className="space-y-1">
                {visibleItems.map((item) => (
                  <CalendarMonthItem
                    item={item}
                    key={`${item.kind}-${item.id}`}
                    onEdit={onEdit}
                    onSelectEvent={onSelectEvent}
                  />
                ))}
                {hiddenItemCount > 0 ? (
                  <button
                    className="text-text-muted hover:bg-state-hover focus-visible:ring-primary/40 w-full rounded-md px-2 py-1 text-left text-base font-medium focus-visible:ring-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-40"
                    disabled={!canSelectDay}
                    onClick={() => {
                      onSelectDay(day);
                    }}
                    type="button"
                  >
                    +{hiddenItemCount} more
                  </button>
                ) : null}
              </Box>
            </Box>
          );
        })}
      </Box>
    </Box>
  );
};

export const PersonalCalendar = ({
  isScheduleDialogOpen,
  onScheduleDialogOpenChange,
}: {
  isScheduleDialogOpen: boolean;
  onScheduleDialogOpenChange: (open: boolean) => void;
}) => {
  const [storedCalendarView, setCalendarView] = useLocalStorage<CalendarView>(
    "calendarView",
    "week",
  );
  const calendarView = normalizeCalendarView(storedCalendarView);
  const [cursor, setCursor] = useState(() => new Date());
  const [dialogMode, setDialogMode] = useState<"work" | "focus" | null>(null);
  const [editingBlock, setEditingBlock] =
    useState<CalendarScheduleBlock | null>(null);
  const [selectedEvent, setSelectedEvent] =
    useState<CalendarEventSummary | null>(null);
  const [selectedBlock, setSelectedBlock] =
    useState<CalendarScheduleBlock | null>(null);
  const { withWorkspace } = useWorkspacePath();
  const viewRange = getCalendarViewRange(cursor, calendarView);
  const scheduleStartAt = viewRange.start.toISOString();
  const scheduleEndAt = viewRange.end.toISOString();
  const days = getCalendarViewDays(cursor, calendarView);
  const today = new Date();
  const earliestCalendarDate = startOfDay(addDays(today, -calendarHistoryDays));
  const latestCalendarDateExclusive = addDays(
    startOfDay(today),
    calendarLookaheadDays + 1,
  );
  const previousCursor = moveCalendarCursor(cursor, calendarView, -1);
  const nextCursor = moveCalendarCursor(cursor, calendarView, 1);
  const previousRange = getCalendarViewRange(previousCursor, calendarView);
  const nextRange = getCalendarViewRange(nextCursor, calendarView);
  const canNavigatePrevious = previousRange.end > earliestCalendarDate;
  const canNavigateNext = nextRange.start < latestCalendarDateExclusive;
  const scheduleQuery = useCalendarSchedule({
    startAt: scheduleStartAt,
    endAt: scheduleEndAt,
  });
  const integrationQuery = useCalendarIntegration();
  const schedule = scheduleQuery.data;
  const integration = integrationQuery.data;
  const connection = integration?.connections[0];
  const timeZoneLabel = getUtcOffsetLabel(viewRange.start);
  const timeZoneName = getLocalTimeZoneName();
  const canReadEventDetails = Boolean(connection?.canReadEventDetails);
  const createConnectSession = useCreateCalendarConnectSession();
  const syncCalendar = useSyncCalendarConnection();
  const manualReschedule = useManualRescheduleCalendarScheduleBlock();
  const updateStory = useUpdateStoryMutation();
  const { data: assignedStories } = useMyStoriesGrouped("none", {
    assignedToMe: true,
    categories: ["backlog", "unstarted", "started", "paused"],
    orderBy: "deadline",
    showSubStories: false,
    storiesPerGroup: 50,
  });
  const candidateStories =
    assignedStories?.groups.flatMap((group) => group.stories) ?? [];
  const displayBusyWindows = getDisplayBusyWindows({
    busyWindows: schedule?.busyWindows ?? [],
    events: schedule?.events ?? [],
  });
  const calendarItems: CalendarItem[] = [
    ...(schedule?.events ?? []).map((event) => ({
      kind: "event" as const,
      id: event.id,
      startAt: event.startAt,
      endAt: event.endAt,
      event,
    })),
    ...displayBusyWindows.map((window) => ({
      kind: "busy" as const,
      id: window.id,
      startAt: window.startAt,
      endAt: window.endAt,
      window,
    })),
    ...(schedule?.blocks ?? []).map((block) => ({
      kind: "block" as const,
      id: block.id,
      startAt: block.startAt,
      endAt: block.endAt,
      block,
    })),
  ].sort(
    (first, second) =>
      new Date(first.startAt).getTime() - new Date(second.startAt).getTime(),
  );
  const timedCalendarItems = calendarItems.filter(
    (item) => item.kind !== "event" || !item.event.isAllDay,
  );
  const allDayEvents = (schedule?.events ?? []).filter(
    (event) => event.isAllDay,
  );
  const conflictingBlocks = (schedule?.blocks ?? []).filter(
    (block) => block.hasConflict,
  );
  const { visibleEndHour, visibleStartHour } = deriveCalendarVisibleHours({
    defaultEndHour: defaultVisibleEndHour,
    defaultStartHour: defaultVisibleStartHour,
    events: timedCalendarItems,
  });
  const hours = Array.from(
    { length: visibleEndHour - visibleStartHour + 1 },
    (_, index) => visibleStartHour + index,
  );
  const hasCalendarLoadError =
    scheduleQuery.isError || integrationQuery.isError;
  const isCalendarInitialLoading =
    scheduleQuery.isPending || integrationQuery.isPending;
  const activeDialogMode = isScheduleDialogOpen ? "work" : dialogMode;

  const openDialog = (mode: "work" | "focus") => {
    onScheduleDialogOpenChange(false);
    setSelectedBlock(null);
    setEditingBlock(null);
    setDialogMode(mode);
  };
  const openEditDialog = (block: CalendarScheduleBlock) => {
    onScheduleDialogOpenChange(false);
    setSelectedBlock(null);
    setEditingBlock(block);
    setDialogMode(block.blockType);
  };
  const openBlock = (block: CalendarScheduleBlock) => {
    if (block.blockType === "work") {
      setSelectedEvent(null);
      setSelectedBlock(block);
      return;
    }
    openEditDialog(block);
  };
  const openEventDetails = (event: CalendarEventSummary) => {
    setSelectedBlock(null);
    setSelectedEvent(event);
  };
  const closeDialog = (value: boolean) => {
    if (!value) {
      onScheduleDialogOpenChange(false);
      setEditingBlock(null);
      setDialogMode(null);
      return;
    }
    setDialogMode(dialogMode ?? "work");
  };
  const syncConnection = (connectionID: string) => {
    syncCalendar.mutate({ connectionId: connectionID });
  };
  const isDaySelectable = (day: Date) => {
    const dayStart = startOfDay(day);
    return (
      dayStart >= earliestCalendarDate && dayStart < latestCalendarDateExclusive
    );
  };
  const selectDay = (day: Date) => {
    if (!isDaySelectable(day)) {
      return;
    }
    setCursor(day);
    setCalendarView("day");
  };
  const handleManualCalendarChange = ({
    block,
    change,
    endAt,
    startAt,
  }: CalendarManualChange) => {
    manualReschedule.mutate(
      {
        blockId: block.id,
        input: {
          startAt: startAt.toISOString(),
          endAt: endAt.toISOString(),
          expectedUpdatedAt: block.updatedAt,
          timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
          change,
          clientMutationId: crypto.randomUUID(),
        },
      },
      {
        onSuccess: () => {
          if (change !== "resize" || !block.storyId) return;
          updateStory.mutate({
            storyId: block.storyId,
            payload: {
              estimatedDurationMinutes: Math.round(
                (endAt.getTime() - startAt.getTime()) / 60_000,
              ),
            },
          });
        },
      },
    );
  };

  return (
    <Box className="flex h-[calc(100dvh-4rem)] min-h-0 flex-col overflow-hidden">
      <CalendarNotices
        canReadEventDetails={canReadEventDetails}
        conflictCount={conflictingBlocks.length}
        connectHref={withWorkspace("/settings/account/calendar")}
        connection={connection}
        hasIntegrationError={integrationQuery.isError}
        isIntegrationPending={integrationQuery.isPending}
        isReconnectPending={createConnectSession.isPending}
        isSyncing={syncCalendar.isPending}
        onReconnect={() => {
          createConnectSession.mutate();
        }}
        onSync={syncConnection}
      />

      <Box className="flex min-h-0 flex-1 flex-col overflow-hidden">
        <CalendarToolbar
          canNavigateNext={canNavigateNext}
          canNavigatePrevious={canNavigatePrevious}
          currentView={calendarView}
          onFocus={() => {
            openDialog("focus");
          }}
          onNext={() => {
            if (canNavigateNext) setCursor(nextCursor);
          }}
          onPrevious={() => {
            if (canNavigatePrevious) setCursor(previousCursor);
          }}
          onToday={() => {
            setCursor(new Date());
          }}
          onViewChange={setCalendarView}
          title={getCalendarViewTitle(cursor, calendarView)}
        />

        {hasCalendarLoadError ? (
          <Box
            className="flex min-h-0 flex-1 items-center justify-center px-6 py-12 text-center"
            role="alert"
          >
            <Box className="flex flex-col items-center">
              <Text fontSize="md" fontWeight="semibold">
                Couldn&apos;t load your calendar
              </Text>
              <Text className="mt-1" color="muted" fontSize="md">
                Your calendar data is still safe. Try loading this view again.
              </Text>
              <Button
                className="mt-4 text-base"
                color="tertiary"
                loading={
                  scheduleQuery.isFetching || integrationQuery.isFetching
                }
                onClick={() => {
                  void Promise.all([
                    scheduleQuery.refetch(),
                    integrationQuery.refetch(),
                  ]);
                }}
                variant="outline"
              >
                Try again
              </Button>
            </Box>
          </Box>
        ) : null}
        {!hasCalendarLoadError && isCalendarInitialLoading ? (
          <CalendarGridSkeleton view={calendarView} />
        ) : null}
        {!hasCalendarLoadError &&
        !isCalendarInitialLoading &&
        calendarView === "month" ? (
          <CalendarMonthGrid
            calendarItems={calendarItems}
            cursor={cursor}
            days={days}
            isDaySelectable={isDaySelectable}
            onEdit={openBlock}
            onSelectDay={selectDay}
            onSelectEvent={openEventDetails}
            today={today}
          />
        ) : null}
        {!hasCalendarLoadError &&
        !isCalendarInitialLoading &&
        calendarView !== "month" ? (
          <CalendarTimeGrid
            allDayEvents={allDayEvents}
            days={days}
            hours={hours}
            isDaySelectable={isDaySelectable}
            isManualChangePending={manualReschedule.isPending}
            onEdit={openBlock}
            onManualChange={handleManualCalendarChange}
            onSelectDay={selectDay}
            onSelectEvent={openEventDetails}
            timeZoneLabel={timeZoneLabel}
            timeZoneName={timeZoneName}
            timedCalendarItems={timedCalendarItems}
            today={today}
            visibleEndHour={visibleEndHour}
            visibleStartHour={visibleStartHour}
          />
        ) : null}
      </Box>

      {activeDialogMode ? (
        <CalendarDialog
          candidateStories={candidateStories}
          editingBlock={editingBlock}
          isOpen
          mode={activeDialogMode}
          onOpenChange={closeDialog}
        />
      ) : null}
      <CalendarEventDetailsDialog
        event={selectedEvent}
        onOpenChange={(open) => {
          if (!open) setSelectedEvent(null);
        }}
      />
      <CalendarScheduleBlockDetailsDialog
        block={selectedBlock}
        onEdit={openEditDialog}
        onOpenChange={(open) => {
          if (!open) setSelectedBlock(null);
        }}
      />
    </Box>
  );
};

import { format, isSameDay, isSameMonth } from "date-fns";
import { cn } from "lib";
import { CheckIcon, TimeScheduleIcon } from "icons";
import { Box, Flex, Text } from "ui";
import type {
  CalendarEventSummary,
  CalendarScheduleBlock,
} from "@/lib/queries/calendar/types";
import {
  CROSS_WORKSPACE_CALENDAR_BLOCK_TITLE,
  CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP,
  RESERVED_TIME_BLOCK_CLASS,
  getCalendarStoryBlockStyle,
  getMayaCalendarBlockLabel,
  getMayaCalendarBlockReason,
  isCalendarScheduleBlockCompleted,
  isCalendarScheduleBlockEditable,
} from "./calendar-block";
import {
  COMPLETED_CALENDAR_BLOCK_PATTERN,
  SCHEDULED_STORY_STATUS_CLASS,
  SCHEDULED_STORY_STATUS_HOVER_CLASS,
  SCHEDULED_TASK_BACKGROUND_CLASS,
  calendarEventOverlapsDay,
  getBusyWindowTitle,
  overlapsDay,
  toClockLabel,
} from "./calendar-presentation";
import type { CalendarItem } from "./calendar-types";
import {
  getCalendarEventTitle,
  getCalendarEventTimeLabel,
} from "./calendar-event-details-dialog";

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
  const isCompleted = isCalendarScheduleBlockCompleted(block);
  const isCrossWorkspace = Boolean(block.isCrossWorkspace);
  const isMayaManaged = block.source === "maya";
  const isEditable = isCalendarScheduleBlockEditable(block);
  const mayaLabel = getMayaCalendarBlockLabel(block);
  const mayaReason = getMayaCalendarBlockReason(block);
  let displayTitle =
    mayaLabel && !isCompleted ? `${mayaLabel} · ${title}` : title;
  if (isCrossWorkspace) {
    displayTitle = CROSS_WORKSPACE_CALENDAR_BLOCK_TITLE;
  }
  const isScheduledStory = block.blockType === "work";
  const canOpenBlock = !isCrossWorkspace && (isScheduledStory || isEditable);
  const storyStyle = getCalendarStoryBlockStyle(block);
  let toneClass = isScheduledStory
    ? cn(
        storyStyle
          ? SCHEDULED_STORY_STATUS_CLASS
          : cn("border-border-strong/70", SCHEDULED_TASK_BACKGROUND_CLASS),
        "text-text-muted",
      )
    : RESERVED_TIME_BLOCK_CLASS;
  if (block.hasConflict && !isCompleted) {
    toneClass = "border-danger/60 bg-danger/20 text-danger";
  }
  if (isCompleted && isScheduledStory) {
    toneClass = cn(
      "border-border-strong/40 text-text-muted",
      storyStyle
        ? SCHEDULED_STORY_STATUS_CLASS
        : SCHEDULED_TASK_BACKGROUND_CLASS,
    );
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
  } else if (isCompleted) {
    blockActionLabel = `Open completed ${title} details`;
  }
  let blockTooltip = block.storyId ? undefined : mayaReason ?? undefined;
  if (isCrossWorkspace) {
    blockTooltip = CROSS_WORKSPACE_CALENDAR_BLOCK_TOOLTIP;
  }

  return (
    <Box
      className={cn(
        "relative flex min-w-0 items-center gap-2 overflow-hidden rounded-md border px-2 py-1 text-base backdrop-blur-sm transition-colors",
        toneClass,
        storyStyle && !isMayaManaged
          ? SCHEDULED_STORY_STATUS_HOVER_CLASS
          : null,
      )}
      style={{
        ...storyStyle,
        ...(isCompleted
          ? { backgroundImage: COMPLETED_CALENDAR_BLOCK_PATTERN }
          : null),
      }}
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
      {isCompleted ? (
        <CheckIcon
          aria-hidden="true"
          className="pointer-events-none size-4 shrink-0"
          strokeWidth={2.2}
        />
      ) : null}
      {!isCompleted && isCrossWorkspace ? (
        <TimeScheduleIcon
          aria-hidden="true"
          className="pointer-events-none size-4 shrink-0"
        />
      ) : null}
      {!isCompleted && !isCrossWorkspace ? (
        <span
          aria-hidden="true"
          className="pointer-events-none size-2 shrink-0 rounded-full bg-current"
          style={
            storyStyle ? { backgroundColor: block.storyStatusColor } : undefined
          }
        />
      ) : null}
      <span className="pointer-events-none relative z-10 shrink-0 tabular-nums">
        {time}
      </span>
      <span
        className={cn(
          "pointer-events-none relative z-10 min-w-0 flex-1 truncate",
          isCompleted && "line-through decoration-current/70",
        )}
      >
        {displayTitle}
      </span>
    </Box>
  );
};

export const CalendarMonthGrid = ({
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

import { useState } from "react";
import { format, isSameDay } from "date-fns";
import type { DragEndEvent, DragStartEvent } from "@dnd-kit/core";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  TouchSensor,
  useDroppable,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import { cn } from "lib";
import { Box, Text } from "ui";
import type {
  CalendarEventSummary,
  CalendarScheduleBlock,
} from "@/lib/queries/calendar/types";
import { CalendarDragPreview } from "./calendar-drag-preview";
import {
  calendarCollisionDetection,
  calendarDragModifiers,
  getCalendarDragData,
} from "./calendar-dnd";
import { getCalendarManualChange } from "./calendar-drag";
import { buildCalendarEventLayouts } from "./calendar-layout";
import {
  HOUR_HEIGHT,
  TIME_RAIL_WIDTH,
  calendarEventOverlapsDay,
  overlapsDay,
} from "./calendar-presentation";
import { CalendarTimedBlock } from "./calendar-timed-block";
import type {
  CalendarDragData,
  CalendarItem,
  CalendarManualChange,
} from "./calendar-types";
import {
  getCalendarEventTitle,
  getCalendarEventTimeLabel,
} from "./calendar-event-details-dialog";

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

const CalendarTimedDayColumn = ({
  day,
  dayItems,
  hours,
  isDragDisabled,
  onEdit,
  onManualChange,
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
  onManualChange: (change: CalendarManualChange) => void;
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
    hourHeight: HOUR_HEIGHT,
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
        height: `${(visibleEndHour - visibleStartHour) * HOUR_HEIGHT}px`,
      }}
    >
      {hours.slice(1, -1).map((hour) => (
        <Box
          className="border-border/60 absolute inset-x-0 border-t"
          key={hour}
          style={{
            top: `${(hour - visibleStartHour) * HOUR_HEIGHT}px`,
          }}
        />
      ))}
      {dayItems.map((item) => {
        const key = `${item.kind}-${item.id}`;
        const layout = layoutById.get(key);
        if (!layout) return null;
        return (
          <CalendarTimedBlock
            day={day}
            isDragDisabled={isDragDisabled}
            item={item}
            key={key}
            layout={layout}
            onEdit={onEdit}
            onManualChange={onManualChange}
            onSelectEvent={onSelectEvent}
            today={today}
          />
        );
      })}
    </div>
  );
};

export const CalendarTimeGrid = ({
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
  const gridTemplateColumns = `${TIME_RAIL_WIDTH}rem repeat(${days.length}, ${dayColumn})`;
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
    if (!dragData || (dragData.kind === "move" && !targetDay)) return;

    const originalStart = new Date(dragData.block.startAt);
    const originalEnd = new Date(dragData.block.endAt);
    const { endAt, startAt } = getCalendarManualChange({
      block: dragData.block,
      deltaY: event.delta.y,
      hourHeight: HOUR_HEIGHT,
      kind: dragData.kind,
      targetDay: targetDay ? new Date(targetDay) : originalStart,
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
            "border-border/70 bg-background/70 sticky top-0 z-30 grid h-18 border-b backdrop-blur-xl",
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
                style={{ top: `${(hour - visibleStartHour) * HOUR_HEIGHT}px` }}
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
              onManualChange={onManualChange}
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

"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import {
  addDays,
  addHours,
  addWeeks,
  format,
  isSameDay,
  isSameMonth,
  isSameYear,
  startOfDay,
  startOfWeek,
  subWeeks,
} from "date-fns";
import { cn } from "lib";
import {
  ArrowLeftIcon,
  ArrowRightIcon,
  CalendarIcon,
  ClockIcon,
  DeleteIcon,
  PlusIcon,
  ReloadIcon,
} from "icons";
import { Box, Button, Dialog, Flex, Input, Select, Skeleton, Text } from "ui";
import { useTerminology, useWorkspacePath } from "@/hooks";
import {
  useCalendarIntegration,
  useCreateCalendarConnectSession,
  useCalendarSchedule,
  useCreateCalendarScheduleBlock,
  useDeleteCalendarScheduleBlock,
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
import { getStoryPath } from "@/modules/story/utils/story-url";
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

const weekStartsOn = 1 as const;
const defaultVisibleStartHour = 8;
const defaultVisibleEndHour = 18;
const hourHeight = 84;
const timeRailWidth = 4.25;
const automaticSyncStaleAfter = 5 * 60 * 1000;
const calendarHistoryDays = 7;
const calendarLookaheadDays = 90;

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

const toDateTimeInputValue = (value: Date | string) =>
  format(new Date(value), "yyyy-MM-dd'T'HH:mm");

const toTimeLabel = (startAt: string, endAt: string) =>
  `${format(new Date(startAt), "h:mm a")} - ${format(new Date(endAt), "h:mm a")}`;

const getWeekTitle = (weekStart: Date) => {
  const weekEnd = addDays(weekStart, 6);
  if (isSameMonth(weekStart, weekEnd)) {
    return format(weekStart, "MMMM yyyy");
  }
  if (isSameYear(weekStart, weekEnd)) {
    return `${format(weekStart, "MMMM")} – ${format(weekEnd, "MMMM yyyy")}`;
  }
  return `${format(weekStart, "MMMM yyyy")} – ${format(weekEnd, "MMMM yyyy")}`;
};

const roundToNextHalfHour = (date: Date) => {
  const next = new Date(date);
  const minutes = next.getMinutes();
  next.setMinutes(minutes < 30 ? 30 : 60, 0, 0);
  return next;
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

const getStoryHref = (
  withWorkspace: (path: string) => string,
  storyId: string,
  storyCode?: string,
) => withWorkspace(getStoryPath({ id: storyCode || storyId }));

const getBusyWindowTitle = (window: CalendarBusyWindow) => {
  if (window.isPrivate) {
    return "Busy";
  }
  return window.title?.trim() || "Busy";
};

const CalendarTimedBlock = ({
  item,
  layout,
  onEdit,
  onSelectEvent,
}: {
  item: CalendarItem;
  layout: { top: number; height: number; lane: number; laneCount: number };
  onEdit: (block: CalendarScheduleBlock) => void;
  onSelectEvent: (event: CalendarEventSummary) => void;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const laneWidth = 100 / layout.laneCount;
  const style = {
    height: `${layout.height}px`,
    left: `calc(${layout.lane * laneWidth}% + 0.25rem)`,
    top: `${layout.top}px`,
    width: `calc(${laneWidth}% - 0.5rem)`,
  };

  if (item.kind === "event") {
    return (
      <button
        aria-label={`Open ${getCalendarEventTitle(item.event)} details, ${getCalendarEventTimeLabel(item.event)}`}
        className="absolute overflow-hidden rounded-md border border-blue-500/25 bg-blue-500/10 px-2.5 py-1.5 text-left shadow-[0_1px_2px_rgba(15,23,42,0.05)] transition hover:bg-blue-500/15 focus-visible:ring-2 focus-visible:ring-blue-500/60 focus-visible:outline-none"
        onClick={() => {
          onSelectEvent(item.event);
        }}
        style={style}
        type="button"
      >
        <Text
          className="truncate text-blue-600 dark:text-blue-300"
          fontSize="sm"
          fontWeight="semibold"
        >
          {getCalendarEventTitle(item.event)}
        </Text>
        <Text className="truncate text-[0.78rem] text-blue-500/80 dark:text-blue-200/75">
          {toTimeLabel(item.startAt, item.endAt)}
        </Text>
      </button>
    );
  }

  if (item.kind === "busy") {
    return (
      <Box
        className="absolute overflow-hidden rounded-md border border-blue-500/15 bg-blue-500/[0.07] px-2.5 py-1.5"
        style={style}
      >
        <Text
          className="truncate text-blue-600 dark:text-blue-300"
          fontSize="sm"
          fontWeight="semibold"
        >
          {getBusyWindowTitle(item.window)}
        </Text>
        <Text className="truncate text-[0.78rem] text-blue-500/80 dark:text-blue-200/75">
          {toTimeLabel(item.startAt, item.endAt)}
        </Text>
      </Box>
    );
  }

  const { block } = item;
  const label =
    block.blockType === "work"
      ? block.storyCode || block.teamCode || "Work"
      : "Focus";
  const statusLabel = block.hasConflict ? "Conflict" : label;
  const href =
    block.blockType === "work" && block.storyId
      ? getStoryHref(withWorkspace, block.storyId, block.storyCode)
      : null;
  let blockColorClass =
    "border-emerald-500/20 bg-emerald-500/10 hover:bg-emerald-500/15";
  if (block.blockType === "work") {
    blockColorClass = "border-primary/25 bg-primary/10 hover:bg-primary/15";
  }
  if (block.hasConflict) {
    blockColorClass = "border-danger/40 bg-danger/10 hover:bg-danger/15";
  }

  return (
    <Box
      className={cn(
        "absolute overflow-hidden rounded-md border px-2.5 py-1.5 shadow-[0_1px_2px_rgba(15,23,42,0.06)] transition",
        blockColorClass,
      )}
      style={style}
    >
      <Flex align="start" className="h-full" gap={2} justify="between">
        <Box className="min-w-0">
          {href ? (
            <Link className="block" href={href}>
              <Text
                className="text-primary hover:text-primary line-clamp-2"
                fontSize="sm"
                fontWeight="semibold"
              >
                {block.title}
              </Text>
            </Link>
          ) : (
            <Text
              className={cn(
                "line-clamp-2",
                block.hasConflict && "text-danger",
                !block.hasConflict &&
                  block.blockType === "focus" &&
                  "text-emerald-700 dark:text-emerald-300",
              )}
              fontSize="sm"
              fontWeight="semibold"
            >
              {block.title}
            </Text>
          )}
          <Flex align="center" className="mt-0.5 min-w-0" gap={1}>
            <Text className="truncate text-[0.78rem]" color="muted">
              {statusLabel}
            </Text>
            <span className="bg-border h-1 w-1 rounded-full" />
            <Text className="truncate text-[0.78rem]" color="muted">
              {toTimeLabel(block.startAt, block.endAt)}
            </Text>
          </Flex>
        </Box>
        <Button
          aria-label={
            block.hasConflict
              ? "Resolve calendar block conflict"
              : "Edit calendar block"
          }
          asIcon
          className="shrink-0"
          color="tertiary"
          onClick={(event) => {
            event.preventDefault();
            onEdit(block);
          }}
          size="xs"
          variant="naked"
        >
          <CalendarIcon className="h-3.5" />
        </Button>
      </Flex>
    </Box>
  );
};

const CalendarAllDayEvent = ({
  event,
  onSelect,
}: {
  event: CalendarEventSummary;
  onSelect: (event: CalendarEventSummary) => void;
}) => (
  <button
    aria-label={`Open ${getCalendarEventTitle(event)} details, ${getCalendarEventTimeLabel(event)}`}
    className="w-full truncate rounded-md border border-blue-500/25 bg-blue-500/10 px-2 py-1 text-left text-sm font-medium text-blue-600 transition hover:bg-blue-500/15 focus-visible:ring-2 focus-visible:ring-blue-500/60 focus-visible:outline-none dark:text-blue-300"
    onClick={() => {
      onSelect(event);
    }}
    type="button"
  >
    {getCalendarEventTitle(event)}
  </button>
);

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
              <Text className="mb-1.5" fontSize="sm" fontWeight="medium">
                {storyTerm}
              </Text>
              <Select
                onValueChange={(value) => {
                  setSelectedStoryId(value);
                }}
                value={selectedStoryId}
              >
                <Select.Trigger className="bg-surface h-[2.8rem] rounded-lg">
                  <Select.Input placeholder={`Select ${storyTerm}`} />
                </Select.Trigger>
                <Select.Content className="max-w-[34rem]">
                  {candidateStories.map((story) => (
                    <Select.Option key={story.id} value={story.id}>
                      <span className="text-text-muted mr-2">
                        {getStoryCode(story)}
                      </span>
                      {story.title}
                    </Select.Option>
                  ))}
                </Select.Content>
              </Select>
              {candidateStories.length === 0 ? (
                <Text className="mt-2" color="muted" fontSize="sm">
                  No assigned{" "}
                  {getTermDisplay("storyTerm", { variant: "plural" })} found.
                </Text>
              ) : null}
            </Box>
          ) : (
            <Input
              label="Title"
              onChange={(event) => {
                setTitle(event.target.value);
              }}
              value={title}
            />
          )}
          <Box className="grid grid-cols-1 gap-3 md:grid-cols-2">
            <Input
              label="Start"
              max={toDateTimeInputValue(latestScheduleAt)}
              min={toDateTimeInputValue(earliestScheduleAt)}
              onChange={(event) => {
                setStartAt(event.target.value);
              }}
              type="datetime-local"
              value={startAt}
            />
            <Input
              label="End"
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
            <Text color="danger" fontSize="sm">
              End time must be after start time.
            </Text>
          ) : null}
          {hasChronologicalRange && !isWithinScheduleHorizon ? (
            <Text color="danger" fontSize="sm">
              Choose a time from the last {calendarHistoryDays} days through the
              next {calendarLookaheadDays} days.
            </Text>
          ) : null}
        </Dialog.Body>
        <Dialog.Footer className="justify-between gap-3 border-0 pt-2">
          {editingBlock ? (
            <Button
              color="danger"
              leftIcon={<DeleteIcon className="h-4" />}
              loading={deleteBlock.isPending}
              onClick={handleDelete}
              size="sm"
              variant="naked"
            >
              Delete
            </Button>
          ) : (
            <span />
          )}
          <Flex align="center" gap={2}>
            <Button
              color="tertiary"
              onClick={close}
              size="sm"
              variant="outline"
            >
              Cancel
            </Button>
            <Button
              color="invert"
              disabled={!canSubmit}
              loading={isSaving}
              onClick={submit}
              size="sm"
            >
              {editingBlock ? "Save" : "Add"}
            </Button>
          </Flex>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

const CalendarSkeleton = () => (
  <Box className="border-border bg-surface overflow-hidden rounded-lg border">
    <Box
      className="border-border bg-surface-muted/40 grid border-b"
      style={{
        gridTemplateColumns: `${timeRailWidth}rem repeat(7, minmax(11rem, 1fr))`,
      }}
    >
      <Box />
      {Array.from({ length: 7 }).map((_, index) => (
        <Box className="border-border border-l px-3 py-3" key={index}>
          <Skeleton className="h-5 w-24" />
        </Box>
      ))}
    </Box>
    <Box
      className="grid"
      style={{
        gridTemplateColumns: `${timeRailWidth}rem repeat(7, minmax(11rem, 1fr))`,
      }}
    >
      <Box className="border-border border-r" />
      {Array.from({ length: 7 }).map((_, dayIndex) => (
        <Box
          className="border-border relative border-l"
          key={dayIndex}
          style={{
            height: `${(defaultVisibleEndHour - defaultVisibleStartHour) * hourHeight}px`,
          }}
        >
          {Array.from({ length: 4 }).map((__, index) => (
            <Skeleton
              className="absolute right-3 left-3 h-14"
              key={index}
              style={{ top: `${(index * 2 + 1) * hourHeight}px` }}
            />
          ))}
        </Box>
      ))}
    </Box>
  </Box>
);

const CalendarToolbar = ({
  canNavigateNext,
  canNavigatePrevious,
  canReadEventDetails,
  connectHref,
  connectionId,
  hasIntegrationError,
  isIntegrationPending,
  isSyncing,
  onFocus,
  onNext,
  onPrevious,
  onSchedule,
  onSync,
  onToday,
  storyTerm,
  weekStart,
}: {
  canNavigateNext: boolean;
  canNavigatePrevious: boolean;
  canReadEventDetails: boolean;
  connectHref: string;
  connectionId?: string;
  hasIntegrationError: boolean;
  isIntegrationPending: boolean;
  isSyncing: boolean;
  onFocus: () => void;
  onNext: () => void;
  onPrevious: () => void;
  onSchedule: () => void;
  onSync: (connectionId: string) => void;
  onToday: () => void;
  storyTerm: string;
  weekStart: Date;
}) => (
  <Flex
    align="center"
    className="mb-4 flex-col gap-3 md:flex-row"
    justify="between"
  >
    <Flex align="center" gap={3}>
      <Flex align="center" gap={1}>
        <Button
          aria-label="Previous week"
          asIcon
          color="tertiary"
          disabled={!canNavigatePrevious}
          onClick={onPrevious}
          size="sm"
          variant="naked"
        >
          <ArrowLeftIcon className="h-4" />
        </Button>
        <Button
          aria-label="Next week"
          asIcon
          color="tertiary"
          disabled={!canNavigateNext}
          onClick={onNext}
          size="sm"
          variant="naked"
        >
          <ArrowRightIcon className="h-4" />
        </Button>
      </Flex>
      <Text as="h2" fontSize="xl" fontWeight="semibold">
        {getWeekTitle(weekStart)}
      </Text>
    </Flex>
    <Flex align="center" className="flex-wrap" gap={2}>
      <Button color="tertiary" onClick={onToday} size="sm" variant="naked">
        Today
      </Button>
      {isIntegrationPending ? <Skeleton className="h-8 w-32" /> : null}
      {!isIntegrationPending && connectionId ? (
        <Button
          color="tertiary"
          leftIcon={<ReloadIcon className="h-4" />}
          loading={isSyncing}
          onClick={() => {
            onSync(connectionId);
          }}
          size="sm"
          variant="outline"
        >
          {canReadEventDetails ? "Sync primary" : "Sync availability"}
        </Button>
      ) : null}
      {!isIntegrationPending && !hasIntegrationError && !connectionId ? (
        <Button
          color="tertiary"
          href={connectHref}
          leftIcon={<CalendarIcon className="h-4" />}
          size="sm"
          variant="outline"
        >
          Connect calendar
        </Button>
      ) : null}
      <Button
        color="tertiary"
        leftIcon={<ClockIcon className="h-4" />}
        onClick={onFocus}
        size="sm"
        variant="outline"
      >
        Focus time
      </Button>
      <Button
        color="invert"
        leftIcon={<PlusIcon className="h-4 text-current dark:text-current" />}
        onClick={onSchedule}
        size="sm"
      >
        Schedule {storyTerm}
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
      <Box className="border-border bg-surface-muted/40 mb-4 rounded-lg border px-4 py-3">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" className="min-w-0" gap={3}>
            <Box className="bg-info/10 text-info flex h-9 w-9 shrink-0 items-center justify-center rounded-lg">
              <CalendarIcon className="h-4 text-current dark:text-current" />
            </Box>
            <Box className="min-w-0">
              <Text fontSize="sm" fontWeight="medium">
                Google Calendar is not connected
              </Text>
              <Text className="line-clamp-1" color="muted" fontSize="sm">
                FortyOne can still schedule work blocks, but availability will
                be incomplete until you connect your primary calendar.
              </Text>
            </Box>
          </Flex>
          <Button
            color="tertiary"
            href={connectHref}
            size="sm"
            variant="outline"
          >
            Connect
          </Button>
        </Flex>
      </Box>
    ) : null}

    {connection && !canReadEventDetails ? (
      <Box className="border-border bg-surface-muted/40 mb-4 rounded-lg border px-4 py-3">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" className="min-w-0" gap={3}>
            <Box className="bg-info/10 text-info flex h-9 w-9 shrink-0 items-center justify-center rounded-lg">
              <CalendarIcon className="h-4 text-current dark:text-current" />
            </Box>
            <Box className="min-w-0">
              <Text fontSize="sm" fontWeight="medium">
                Event details are not enabled
              </Text>
              <Text className="line-clamp-1" color="muted" fontSize="sm">
                Reconnect your primary Google Calendar to show event titles
                instead of availability-only busy blocks.
              </Text>
            </Box>
          </Flex>
          <Button
            color="tertiary"
            loading={isReconnectPending}
            onClick={onReconnect}
            size="sm"
            variant="outline"
          >
            Reconnect
          </Button>
        </Flex>
      </Box>
    ) : null}

    {connection?.syncStatus === "failed" ? (
      <Box className="border-warning/30 bg-warning/5 mb-4 rounded-lg border px-4 py-3">
        <Flex align="center" gap={3} justify="between">
          <Box className="min-w-0">
            <Text fontSize="sm" fontWeight="medium">
              Calendar sync failed
            </Text>
            <Text color="muted" fontSize="sm">
              {connection.syncError?.trim() ||
                "Google Calendar could not be refreshed."}{" "}
              {connection.lastSyncedAt
                ? `Showing the last successful sync from ${format(new Date(connection.lastSyncedAt), "MMM d 'at' h:mm a")}.`
                : "No successful calendar sync is available yet."}
            </Text>
          </Box>
          <Button
            color="tertiary"
            loading={isSyncing}
            onClick={() => {
              onSync(connection.id);
            }}
            size="sm"
            variant="outline"
          >
            Retry
          </Button>
        </Flex>
      </Box>
    ) : null}

    {conflictCount > 0 ? (
      <Box className="border-danger/30 bg-danger/5 mb-4 rounded-lg border px-4 py-3">
        <Text fontSize="sm" fontWeight="medium">
          {conflictCount === 1
            ? "A scheduled block now overlaps a meeting"
            : `${conflictCount} scheduled blocks now overlap meetings`}
        </Text>
        <Text color="muted" fontSize="sm">
          Open a red block to choose another time. FortyOne will not move locked
          work without your approval.
        </Text>
      </Box>
    ) : null}

    {connection ? (
      <Text className="mb-2 block" color="muted" fontSize="sm">
        Showing Google primary calendar · 7 days back, 90 days ahead
      </Text>
    ) : null}
  </>
);

const CalendarWeekGrid = ({
  allDayEvents,
  days,
  hours,
  onEdit,
  onSelectEvent,
  timedCalendarItems,
  today,
  visibleEndHour,
  visibleStartHour,
}: {
  allDayEvents: CalendarEventSummary[];
  days: Date[];
  hours: number[];
  onEdit: (block: CalendarScheduleBlock) => void;
  onSelectEvent: (event: CalendarEventSummary) => void;
  timedCalendarItems: CalendarItem[];
  today: Date;
  visibleEndHour: number;
  visibleStartHour: number;
}) => (
  <Box className="border-border bg-surface overflow-x-auto rounded-lg border shadow-[0_1px_8px_rgba(15,23,42,0.04)]">
    <Box
      className="border-border bg-surface-muted/45 grid min-w-[72rem] border-b"
      style={{
        gridTemplateColumns: `${timeRailWidth}rem repeat(7, minmax(9.5rem, 1fr))`,
      }}
    >
      <Box className="flex items-end px-3 py-3">
        <Text className="text-[0.78rem]" color="muted" fontWeight="medium">
          {Intl.DateTimeFormat()
            .resolvedOptions()
            .timeZone.split("/")
            .pop()
            ?.replace("_", " ") ?? "Local"}
        </Text>
      </Box>
      {days.map((day) => {
        const isToday = isSameDay(day, today);
        return (
          <Box
            className="border-border border-l px-3 py-3"
            key={day.toISOString()}
          >
            <Flex align="center" gap={2} justify="between">
              <Text color={isToday ? "primary" : "muted"} fontSize="sm">
                {format(day, "EEE")}
              </Text>
              <Box
                className={cn(
                  "flex h-7 min-w-7 items-center justify-center rounded-md px-2",
                  isToday
                    ? "bg-primary text-primary-foreground"
                    : "text-foreground",
                )}
              >
                <Text
                  as="span"
                  className="text-current"
                  fontSize="sm"
                  fontWeight="semibold"
                >
                  {format(day, "d")}
                </Text>
              </Box>
            </Flex>
          </Box>
        );
      })}
    </Box>
    <Box
      className="border-border bg-surface-muted/20 grid min-w-[72rem] border-b"
      style={{
        gridTemplateColumns: `${timeRailWidth}rem repeat(7, minmax(9.5rem, 1fr))`,
      }}
    >
      <Box className="flex items-start justify-end px-3 py-2">
        <Text className="text-[0.78rem]" color="muted">
          All day
        </Text>
      </Box>
      {days.map((day) => {
        const dayEvents = allDayEvents.filter((event) =>
          calendarEventOverlapsDay(event, day),
        );
        return (
          <Box
            className="border-border min-h-11 space-y-1 border-l p-1.5"
            key={day.toISOString()}
          >
            {dayEvents.map((event) => (
              <CalendarAllDayEvent
                event={event}
                key={event.id}
                onSelect={onSelectEvent}
              />
            ))}
          </Box>
        );
      })}
    </Box>
    <Box
      className="grid min-w-[72rem]"
      style={{
        gridTemplateColumns: `${timeRailWidth}rem repeat(7, minmax(9.5rem, 1fr))`,
      }}
    >
      <Box className="border-border relative border-r">
        {hours.slice(0, -1).map((hour) => (
          <Box
            className="absolute right-3 -translate-y-2"
            key={hour}
            style={{ top: `${(hour - visibleStartHour) * hourHeight}px` }}
          >
            <Text className="text-[0.78rem]" color="muted">
              {format(new Date(2026, 0, 1, hour), "ha")}
            </Text>
          </Box>
        ))}
      </Box>
      {days.map((day) => {
        const dayItems = timedCalendarItems.filter((item) =>
          overlapsDay(item, day),
        );
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
        const layoutById = new Map(
          layouts.map((layout) => [layout.id, layout]),
        );

        return (
          <Box
            className={cn(
              "border-border relative border-l",
              isSameDay(day, today) && "bg-primary/[0.025]",
            )}
            key={day.toISOString()}
            style={{
              height: `${(visibleEndHour - visibleStartHour) * hourHeight}px`,
            }}
          >
            {hours.slice(0, -1).map((hour) => (
              <Box
                className="border-border/80 absolute inset-x-0 border-t"
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
                  item={item}
                  key={key}
                  layout={layout}
                  onEdit={onEdit}
                  onSelectEvent={onSelectEvent}
                />
              );
            })}
          </Box>
        );
      })}
    </Box>
  </Box>
);

export const MyWorkCalendar = () => {
  const [weekCursor, setWeekCursor] = useState(() =>
    startOfWeek(new Date(), { weekStartsOn }),
  );
  const [dialogMode, setDialogMode] = useState<"work" | "focus" | null>(null);
  const [editingBlock, setEditingBlock] =
    useState<CalendarScheduleBlock | null>(null);
  const [selectedEvent, setSelectedEvent] =
    useState<CalendarEventSummary | null>(null);
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();
  const weekStart = startOfWeek(weekCursor, { weekStartsOn });
  const scheduleStartAt = weekStart.toISOString();
  const scheduleEndAt = addDays(weekStart, 7).toISOString();
  const days = Array.from({ length: 7 }, (_, index) =>
    addDays(weekStart, index),
  );
  const today = new Date();
  const earliestWeekStart = startOfWeek(addDays(today, -calendarHistoryDays), {
    weekStartsOn,
  });
  const latestWeekStart = startOfWeek(addDays(today, calendarLookaheadDays), {
    weekStartsOn,
  });
  const canNavigatePrevious = weekStart.getTime() > earliestWeekStart.getTime();
  const canNavigateNext = weekStart.getTime() < latestWeekStart.getTime();
  const scheduleQuery = useCalendarSchedule({
    startAt: scheduleStartAt,
    endAt: scheduleEndAt,
  });
  const integrationQuery = useCalendarIntegration();
  const schedule = scheduleQuery.data;
  const integration = integrationQuery.data;
  const connection = integration?.connections[0];
  const canReadEventDetails = Boolean(connection?.canReadEventDetails);
  const createConnectSession = useCreateCalendarConnectSession();
  const syncCalendar = useSyncCalendarConnection();
  const automaticSyncAttempt = useRef<{
    connectionId: string;
    attemptedAt: number;
  } | null>(null);
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
  const connectionId = connection?.id;
  const connectionLastSyncedAt = connection?.lastSyncedAt;
  const connectionSyncStatus = connection?.syncStatus;
  const syncCalendarMutate = syncCalendar.mutate;
  const syncCalendarPending = syncCalendar.isPending;
  const hasCalendarLoadError =
    scheduleQuery.isError || integrationQuery.isError;
  const isCalendarInitialLoading =
    scheduleQuery.isPending || integrationQuery.isPending;

  useEffect(() => {
    if (!connectionId) {
      return;
    }

    const syncIfStale = () => {
      const now = Date.now();
      const lastSyncedAt = connectionLastSyncedAt
        ? Date.parse(connectionLastSyncedAt)
        : Number.NaN;
      const isFresh =
        connectionSyncStatus !== "failed" &&
        Number.isFinite(lastSyncedAt) &&
        now - lastSyncedAt < automaticSyncStaleAfter;
      const previousAttempt = automaticSyncAttempt.current;
      const attemptedRecently =
        previousAttempt?.connectionId === connectionId &&
        now - previousAttempt.attemptedAt < automaticSyncStaleAfter;

      if (syncCalendarPending || isFresh || attemptedRecently) {
        return;
      }

      automaticSyncAttempt.current = { connectionId, attemptedAt: now };
      syncCalendarMutate({ connectionId, silent: true });
    };

    syncIfStale();
    const interval = window.setInterval(syncIfStale, automaticSyncStaleAfter);
    return () => {
      window.clearInterval(interval);
    };
  }, [
    connectionId,
    connectionLastSyncedAt,
    connectionSyncStatus,
    syncCalendarMutate,
    syncCalendarPending,
  ]);

  const openDialog = (mode: "work" | "focus") => {
    setEditingBlock(null);
    setDialogMode(mode);
  };
  const openEditDialog = (block: CalendarScheduleBlock) => {
    setEditingBlock(block);
    setDialogMode(block.blockType);
  };
  const closeDialog = (value: boolean) => {
    if (!value) {
      setEditingBlock(null);
      setDialogMode(null);
      return;
    }
    setDialogMode(dialogMode ?? "work");
  };
  const syncConnection = (connectionID: string) => {
    automaticSyncAttempt.current = {
      connectionId: connectionID,
      attemptedAt: Date.now(),
    };
    syncCalendar.mutate({ connectionId: connectionID });
  };

  return (
    <Box className="bg-surface h-[calc(100dvh-4rem)] overflow-auto px-5 py-4 md:px-6">
      <CalendarToolbar
        canNavigateNext={canNavigateNext}
        canNavigatePrevious={canNavigatePrevious}
        canReadEventDetails={canReadEventDetails}
        connectHref={withWorkspace("/settings/account/calendar")}
        connectionId={connection?.id}
        hasIntegrationError={integrationQuery.isError}
        isIntegrationPending={integrationQuery.isPending}
        isSyncing={syncCalendar.isPending}
        onFocus={() => {
          openDialog("focus");
        }}
        onNext={() => {
          if (canNavigateNext) setWeekCursor((value) => addWeeks(value, 1));
        }}
        onPrevious={() => {
          if (canNavigatePrevious) {
            setWeekCursor((value) => subWeeks(value, 1));
          }
        }}
        onSchedule={() => {
          openDialog("work");
        }}
        onSync={syncConnection}
        onToday={() => {
          setWeekCursor(startOfWeek(new Date(), { weekStartsOn }));
        }}
        storyTerm={getTermDisplay("storyTerm")}
        weekStart={weekStart}
      />

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

      {hasCalendarLoadError ? (
        <Box
          className="border-border bg-surface-muted/30 flex min-h-80 items-center justify-center rounded-lg border px-6 py-12 text-center"
          role="alert"
        >
          <Box>
            <Text fontWeight="semibold">Couldn&apos;t load your calendar</Text>
            <Text className="mt-1" color="muted" fontSize="sm">
              Your calendar data is still safe. Try loading this week again.
            </Text>
            <Button
              className="mt-4"
              color="tertiary"
              loading={scheduleQuery.isFetching || integrationQuery.isFetching}
              onClick={() => {
                void Promise.all([
                  scheduleQuery.refetch(),
                  integrationQuery.refetch(),
                ]);
              }}
              size="sm"
              variant="outline"
            >
              Try again
            </Button>
          </Box>
        </Box>
      ) : null}
      {!hasCalendarLoadError && isCalendarInitialLoading ? (
        <CalendarSkeleton />
      ) : null}
      {!hasCalendarLoadError && !isCalendarInitialLoading ? (
        <CalendarWeekGrid
          allDayEvents={allDayEvents}
          days={days}
          hours={hours}
          onEdit={openEditDialog}
          onSelectEvent={setSelectedEvent}
          timedCalendarItems={timedCalendarItems}
          today={today}
          visibleEndHour={visibleEndHour}
          visibleStartHour={visibleStartHour}
        />
      ) : null}

      {dialogMode ? (
        <CalendarDialog
          candidateStories={candidateStories}
          editingBlock={editingBlock}
          isOpen={Boolean(dialogMode)}
          mode={dialogMode}
          onOpenChange={closeDialog}
        />
      ) : null}
      <CalendarEventDetailsDialog
        event={selectedEvent}
        onOpenChange={(open) => {
          if (!open) setSelectedEvent(null);
        }}
      />
    </Box>
  );
};

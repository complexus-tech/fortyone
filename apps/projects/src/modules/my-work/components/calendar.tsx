"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import {
  addDays,
  addHours,
  addWeeks,
  format,
  isSameDay,
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
  CalendarScheduleBlock,
  CalendarScheduleBlockInput,
} from "@/lib/queries/calendar/types";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import type { Story } from "@/modules/stories/types";
import { slugify } from "@/utils";
import { buildCalendarEventLayouts } from "./calendar-layout";

const weekStartsOn = 1 as const;
const visibleStartHour = 8;
const visibleEndHour = 18;
const hourHeight = 84;
const timeRailWidth = 4.25;

type CalendarItem =
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

const getStoryCode = (story: Story) =>
  story.team?.code
    ? `${story.team.code}-${story.sequenceId}`
    : `#${story.sequenceId}`;

const getStoryHref = (
  withWorkspace: (path: string) => string,
  storyId: string,
  title: string,
) => withWorkspace(`/story/${storyId}/${slugify(title)}`);

const hours = Array.from(
  { length: visibleEndHour - visibleStartHour + 1 },
  (_, index) => visibleStartHour + index,
);

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
}: {
  item: CalendarItem;
  layout: { top: number; height: number; lane: number; laneCount: number };
  onEdit: (block: CalendarScheduleBlock) => void;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const laneWidth = 100 / layout.laneCount;
  const style = {
    height: `${layout.height}px`,
    left: `calc(${layout.lane * laneWidth}% + 0.25rem)`,
    top: `${layout.top}px`,
    width: `calc(${laneWidth}% - 0.5rem)`,
  };

  if (item.kind === "busy") {
    return (
      <Box
        className="absolute overflow-hidden rounded-md border border-blue-500/20 bg-blue-500/10 px-2.5 py-1.5 shadow-[0_1px_2px_rgba(15,23,42,0.05)]"
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
  const href =
    block.blockType === "work" && block.storyId
      ? getStoryHref(
          withWorkspace,
          block.storyId,
          block.storyTitle ?? block.title,
        )
      : null;

  return (
    <Box
      className={cn(
        "absolute overflow-hidden rounded-md border px-2.5 py-1.5 shadow-[0_1px_2px_rgba(15,23,42,0.06)] transition",
        block.blockType === "work"
          ? "border-primary/25 bg-primary/10 hover:bg-primary/15"
          : "border-emerald-500/20 bg-emerald-500/10 hover:bg-emerald-500/15",
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
              {label}
            </Text>
            <span className="bg-border h-1 w-1 rounded-full" />
            <Text className="truncate text-[0.78rem]" color="muted">
              {toTimeLabel(block.startAt, block.endAt)}
            </Text>
          </Flex>
        </Box>
        <Button
          aria-label="Edit calendar block"
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
  const [startAt, setStartAt] = useState(
    toDateTimeInputValue(editingBlock?.startAt ?? defaultStart),
  );
  const [endAt, setEndAt] = useState(
    toDateTimeInputValue(editingBlock?.endAt ?? addHours(defaultStart, 1)),
  );
  const storyTerm = getTermDisplay("storyTerm");
  const selectedStory = candidateStories.find(
    (story) => story.id === selectedStoryId,
  );
  const isWork = mode === "work";
  const canSubmit = isWork ? Boolean(selectedStoryId) : title.trim().length > 0;
  const isSaving = createBlock.isPending || updateBlock.isPending;
  let dialogTitle = "Add focus time";
  if (editingBlock) {
    dialogTitle = "Edit calendar block";
  } else if (isWork) {
    dialogTitle = `Schedule ${storyTerm}`;
  }

  useEffect(() => {
    if (!isOpen) {
      return;
    }
    const nextDefaultStart = roundToNextHalfHour(addHours(new Date(), 1));
    setSelectedStoryId(editingBlock?.storyId ?? defaultStoryId);
    setTitle(editingBlock?.title ?? "Focus time");
    setStartAt(toDateTimeInputValue(editingBlock?.startAt ?? nextDefaultStart));
    setEndAt(
      toDateTimeInputValue(
        editingBlock?.endAt ?? addHours(nextDefaultStart, 1),
      ),
    );
  }, [
    defaultStoryId,
    editingBlock?.endAt,
    editingBlock?.id,
    editingBlock?.startAt,
    editingBlock?.storyId,
    editingBlock?.title,
    isOpen,
    mode,
  ]);

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
              onChange={(event) => {
                setStartAt(event.target.value);
              }}
              type="datetime-local"
              value={startAt}
            />
            <Input
              label="End"
              onChange={(event) => {
                setEndAt(event.target.value);
              }}
              type="datetime-local"
              value={endAt}
            />
          </Box>
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
            height: `${(visibleEndHour - visibleStartHour) * hourHeight}px`,
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

export const MyWorkCalendar = () => {
  const [weekCursor, setWeekCursor] = useState(
    startOfWeek(new Date(), { weekStartsOn }),
  );
  const [dialogMode, setDialogMode] = useState<"work" | "focus" | null>(null);
  const [editingBlock, setEditingBlock] =
    useState<CalendarScheduleBlock | null>(null);
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();
  const weekStart = startOfWeek(weekCursor, { weekStartsOn });
  const scheduleStartAt = weekStart.toISOString();
  const scheduleEndAt = addDays(weekStart, 7).toISOString();
  const days = useMemo(
    () => Array.from({ length: 7 }, (_, index) => addDays(weekStart, index)),
    [weekStart],
  );
  const { data: schedule, isPending } = useCalendarSchedule({
    startAt: scheduleStartAt,
    endAt: scheduleEndAt,
  });
  const { data: integration } = useCalendarIntegration();
  const connection = integration?.connections[0];
  const canReadEventDetails = Boolean(connection?.canReadEventDetails);
  const createConnectSession = useCreateCalendarConnectSession();
  const syncCalendar = useSyncCalendarConnection();
  const { data: assignedStories } = useMyStoriesGrouped("none", {
    assignedToMe: true,
    categories: ["backlog", "unstarted", "started", "paused"],
    orderBy: "deadline",
    showSubStories: false,
    storiesPerGroup: 50,
  });
  const candidateStories =
    assignedStories?.groups.flatMap((group) => group.stories) ?? [];
  const calendarItems: CalendarItem[] = [
    ...(schedule?.busyWindows ?? []).map((window) => ({
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

  return (
    <Box className="bg-surface h-[calc(100dvh-7.7rem)] overflow-auto px-5 py-4 md:px-6">
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
              onClick={() => {
                setWeekCursor((value) => subWeeks(value, 1));
              }}
              size="sm"
              variant="naked"
            >
              <ArrowLeftIcon className="h-4" />
            </Button>
            <Button
              aria-label="Next week"
              asIcon
              color="tertiary"
              onClick={() => {
                setWeekCursor((value) => addWeeks(value, 1));
              }}
              size="sm"
              variant="naked"
            >
              <ArrowRightIcon className="h-4" />
            </Button>
          </Flex>
          <Text as="h2" fontSize="xl" fontWeight="semibold">
            {format(weekStart, "MMMM yyyy")}
          </Text>
        </Flex>
        <Flex align="center" className="flex-wrap" gap={2}>
          <Button
            color="tertiary"
            onClick={() => {
              setWeekCursor(startOfWeek(new Date(), { weekStartsOn }));
            }}
            size="sm"
            variant="naked"
          >
            Today
          </Button>
          {connection ? (
            <Button
              color="tertiary"
              leftIcon={<ReloadIcon className="h-4" />}
              loading={syncCalendar.isPending}
              onClick={() => {
                syncCalendar.mutate(connection.id);
              }}
              size="sm"
              variant="outline"
            >
              {canReadEventDetails ? "Sync events" : "Sync availability"}
            </Button>
          ) : (
            <Button
              color="tertiary"
              href={withWorkspace("/settings/workspace/integrations/calendar")}
              leftIcon={<CalendarIcon className="h-4" />}
              size="sm"
              variant="outline"
            >
              Connect calendar
            </Button>
          )}
          <Button
            color="tertiary"
            leftIcon={<ClockIcon className="h-4" />}
            onClick={() => {
              openDialog("focus");
            }}
            size="sm"
            variant="outline"
          >
            Focus time
          </Button>
          <Button
            color="invert"
            leftIcon={
              <PlusIcon className="h-4 text-current dark:text-current" />
            }
            onClick={() => {
              openDialog("work");
            }}
            size="sm"
          >
            Schedule {getTermDisplay("storyTerm")}
          </Button>
        </Flex>
      </Flex>

      {!connection ? (
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
                  be incomplete.
                </Text>
              </Box>
            </Flex>
            <Button
              color="tertiary"
              href={withWorkspace("/settings/workspace/integrations/calendar")}
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
                  Reconnect Google Calendar to show event titles instead of
                  availability-only busy blocks.
                </Text>
              </Box>
            </Flex>
            <Button
              color="tertiary"
              loading={createConnectSession.isPending}
              onClick={() => {
                createConnectSession.mutate();
              }}
              size="sm"
              variant="outline"
            >
              Reconnect
            </Button>
          </Flex>
        </Box>
      ) : null}

      {isPending ? (
        <CalendarSkeleton />
      ) : (
        <Box className="border-border bg-surface overflow-x-auto rounded-lg border shadow-[0_1px_8px_rgba(15,23,42,0.04)]">
          <Box
            className="border-border bg-surface-muted/45 grid min-w-[72rem] border-b"
            style={{
              gridTemplateColumns: `${timeRailWidth}rem repeat(7, minmax(9.5rem, 1fr))`,
            }}
          >
            <Box className="flex items-end px-3 py-3">
              <Text
                className="text-[0.78rem]"
                color="muted"
                fontWeight="medium"
              >
                {Intl.DateTimeFormat()
                  .resolvedOptions()
                  .timeZone.split("/")
                  .pop()
                  ?.replace("_", " ") ?? "Local"}
              </Text>
            </Box>
            {days.map((day) => {
              const today = isSameDay(day, new Date());
              return (
                <Box
                  className="border-border border-l px-3 py-3"
                  key={day.toISOString()}
                >
                  <Flex align="center" gap={2} justify="between">
                    <Text color={today ? "primary" : "muted"} fontSize="sm">
                      {format(day, "EEE")}
                    </Text>
                    <Box
                      className={cn(
                        "flex h-7 min-w-7 items-center justify-center rounded-md px-2",
                        today ? "bg-primary text-primary-foreground" : "text-foreground",
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
                  style={{
                    top: `${(hour - visibleStartHour) * hourHeight}px`,
                  }}
                >
                  <Text className="text-[0.78rem]" color="muted">
                    {format(new Date(2026, 0, 1, hour), "ha")}
                  </Text>
                </Box>
              ))}
            </Box>
            {days.map((day) => {
              const dayItems = calendarItems.filter((item) =>
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
                    isSameDay(day, new Date()) && "bg-primary/[0.025]",
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
                    if (!layout) {
                      return null;
                    }
                    return (
                      <CalendarTimedBlock
                        item={item}
                        key={key}
                        layout={layout}
                        onEdit={openEditDialog}
                      />
                    );
                  })}
                </Box>
              );
            })}
          </Box>
        </Box>
      )}

      {dialogMode ? (
        <CalendarDialog
          candidateStories={candidateStories}
          editingBlock={editingBlock}
          isOpen={Boolean(dialogMode)}
          mode={dialogMode}
          onOpenChange={closeDialog}
        />
      ) : null}
    </Box>
  );
};

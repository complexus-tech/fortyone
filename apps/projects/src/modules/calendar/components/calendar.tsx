"use client";

import { useState } from "react";
import { addDays, startOfDay } from "date-fns";
import { Box, Button, Text } from "ui";
import { useLocalStorage, useTerminology, useWorkspacePath } from "@/hooks";
import {
  useCalendarIntegration,
  useCreateCalendarConnectSession,
  useCalendarSchedule,
  useManualRescheduleCalendarScheduleBlock,
  useSyncCalendarConnection,
} from "@/lib/hooks/calendar";
import type {
  CalendarEventSummary,
  CalendarScheduleBlock,
} from "@/lib/queries/calendar/types";
import { useMyStoriesGrouped } from "@/modules/stories/hooks/use-my-stories-grouped";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import { getVisibleCalendarScheduleBlocks } from "./calendar-block";
import {
  deriveCalendarVisibleHours,
  getDisplayBusyWindows,
} from "./calendar-layout";
import { CalendarMonthGrid } from "./calendar-month-grid";
import { CalendarNotices } from "./calendar-notices";
import {
  CALENDAR_HISTORY_DAYS,
  CALENDAR_LOOKAHEAD_DAYS,
  DEFAULT_VISIBLE_END_HOUR,
  DEFAULT_VISIBLE_START_HOUR,
  getLocalTimeZoneName,
  getUtcOffsetLabel,
} from "./calendar-presentation";
import { CalendarScheduleDialog } from "./calendar-schedule-dialog";
import { CalendarGridSkeleton } from "./calendar-skeleton";
import { CalendarTimeGrid } from "./calendar-time-grid";
import type { CalendarManualChange } from "./calendar-types";
import {
  getCalendarViewDays,
  getCalendarViewRange,
  getCalendarViewTitle,
  moveCalendarCursor,
  normalizeCalendarView,
} from "./calendar-view";
import type { CalendarView } from "./calendar-view";
import { CalendarEventDetailsDialog } from "./calendar-event-details-dialog";
import { CalendarScheduleBlockDetailsDialog } from "./calendar-schedule-block-details-dialog";
import { CalendarHeader } from "./header";

type PersonalCalendarProps = {
  isScheduleDialogOpen: boolean;
  onScheduleDialogOpenChange: (open: boolean) => void;
};

const usePersonalCalendar = ({
  isScheduleDialogOpen,
  onScheduleDialogOpenChange,
}: PersonalCalendarProps) => {
  const [storedCalendarView, setCalendarView] = useLocalStorage<CalendarView>(
    "calendarView",
    "week",
  );
  const [showCompletedWork, setShowCompletedWork] = useLocalStorage<boolean>(
    "calendarShowCompletedWork",
    false,
  );
  const [cursor, setCursor] = useState(() => new Date());
  const [dialogMode, setDialogMode] = useState<"work" | "focus" | null>(null);
  const [editingBlock, setEditingBlock] =
    useState<CalendarScheduleBlock | null>(null);
  const [selectedEvent, setSelectedEvent] =
    useState<CalendarEventSummary | null>(null);
  const [selectedBlock, setSelectedBlock] =
    useState<CalendarScheduleBlock | null>(null);
  const { getTermDisplay } = useTerminology();
  const { withWorkspace } = useWorkspacePath();
  const calendarView = normalizeCalendarView(storedCalendarView);
  const viewRange = getCalendarViewRange(cursor, calendarView);
  const scheduleStartAt = viewRange.start.toISOString();
  const scheduleEndAt = viewRange.end.toISOString();
  const days = getCalendarViewDays(cursor, calendarView);
  const today = new Date();
  const earliestCalendarDate = startOfDay(
    addDays(today, -CALENDAR_HISTORY_DAYS),
  );
  const latestCalendarDateExclusive = addDays(
    startOfDay(today),
    CALENDAR_LOOKAHEAD_DAYS + 1,
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
  const connections = integration?.connections ?? [];
  const connection =
    connections.find((item) => item.requiresReauthorization) ??
    connections.find((item) => item.syncStatus === "failed") ??
    connections.at(0);
  const timeZoneLabel = getUtcOffsetLabel(viewRange.start);
  const timeZoneName = getLocalTimeZoneName();
  const canReadEventDetails = connections.some(
    (item) => item.canReadEventDetails,
  );
  const createConnectSession = useCreateCalendarConnectSession();
  const syncCalendar = useSyncCalendarConnection();
  const manualReschedule = useManualRescheduleCalendarScheduleBlock();
  const { data: assignedStories } = useMyStoriesGrouped("none", {
    assignedToMe: true,
    categories: ["backlog", "unstarted", "started", "paused"],
    orderBy: "deadline",
    showSubStories: false,
    storiesPerGroup: 50,
  });
  const candidateStories =
    assignedStories?.groups.flatMap((group) => group.stories) ?? [];
  const storyTerm = getTermDisplay("storyTerm");
  const storyTermPlural = getTermDisplay("storyTerm", { variant: "plural" });
  const displayBusyWindows = getDisplayBusyWindows({
    busyWindows: schedule?.busyWindows ?? [],
    events: schedule?.events ?? [],
  });
  const visibleScheduleBlocks = getVisibleCalendarScheduleBlocks(
    schedule?.blocks ?? [],
    showCompletedWork,
  );
  const calendarItems = [
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
    ...visibleScheduleBlocks.map((block) => ({
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
  const conflictingBlocks = visibleScheduleBlocks.filter(
    (block) => block.hasConflict,
  );
  const { visibleEndHour, visibleStartHour } = deriveCalendarVisibleHours({
    defaultEndHour: DEFAULT_VISIBLE_END_HOUR,
    defaultStartHour: DEFAULT_VISIBLE_START_HOUR,
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
    manualReschedule.mutate({
      blockId: block.id,
      input: {
        startAt: startAt.toISOString(),
        endAt: endAt.toISOString(),
        expectedUpdatedAt: block.updatedAt,
        timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
        change,
        clientMutationId: crypto.randomUUID(),
      },
    });
  };

  return {
    activeDialogMode,
    allDayEvents,
    calendarItems,
    calendarView,
    canNavigateNext,
    canNavigatePrevious,
    canReadEventDetails,
    candidateStories,
    closeDialog,
    conflictCount: conflictingBlocks.length,
    connectHref: withWorkspace("/settings/account/calendar"),
    connection,
    createConnectSession,
    cursor,
    days,
    editingBlock,
    handleManualCalendarChange,
    hasCalendarLoadError,
    hours,
    integrationQuery,
    isCalendarInitialLoading,
    isDaySelectable,
    manualReschedule,
    nextCursor,
    onScheduleDialogOpenChange,
    openBlock,
    openDialog,
    openEditDialog,
    openEventDetails,
    previousCursor,
    scheduleQuery,
    selectedBlock,
    selectedEvent,
    selectDay,
    setCalendarView,
    setCursor,
    setSelectedBlock,
    setSelectedEvent,
    setShowCompletedWork,
    showCompletedWork,
    storyTerm,
    storyTermPlural,
    syncCalendar,
    syncConnection,
    timeZoneLabel,
    timeZoneName,
    timedCalendarItems,
    today,
    visibleEndHour,
    visibleStartHour,
  };
};

export const PersonalCalendar = (props: PersonalCalendarProps) => {
  const {
    activeDialogMode,
    allDayEvents,
    calendarItems,
    calendarView,
    canNavigateNext,
    canNavigatePrevious,
    canReadEventDetails,
    candidateStories,
    closeDialog,
    conflictCount,
    connectHref,
    connection,
    createConnectSession,
    cursor,
    days,
    editingBlock,
    handleManualCalendarChange,
    hasCalendarLoadError,
    hours,
    integrationQuery,
    isCalendarInitialLoading,
    isDaySelectable,
    manualReschedule,
    nextCursor,
    onScheduleDialogOpenChange,
    openBlock,
    openDialog,
    openEditDialog,
    openEventDetails,
    previousCursor,
    scheduleQuery,
    selectedBlock,
    selectedEvent,
    selectDay,
    setCalendarView,
    setCursor,
    setSelectedBlock,
    setSelectedEvent,
    setShowCompletedWork,
    showCompletedWork,
    storyTerm,
    storyTermPlural,
    syncCalendar,
    syncConnection,
    timeZoneLabel,
    timeZoneName,
    timedCalendarItems,
    today,
    visibleEndHour,
    visibleStartHour,
  } = usePersonalCalendar(props);

  return (
    <Box className="flex h-full min-h-0 flex-col overflow-hidden">
      <CalendarHeader
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
        onSchedule={() => {
          onScheduleDialogOpenChange(true);
        }}
        onShowCompletedWorkChange={setShowCompletedWork}
        onToday={() => {
          setCursor(new Date());
        }}
        onViewChange={setCalendarView}
        showCompletedWork={showCompletedWork}
        title={getCalendarViewTitle(cursor, calendarView)}
      />
      <Box className="flex min-h-0 flex-1 flex-col overflow-hidden">
        <CalendarNotices
          canReadEventDetails={canReadEventDetails}
          conflictCount={conflictCount}
          connectHref={connectHref}
          connection={connection}
          hasIntegrationError={integrationQuery.isError}
          isIntegrationPending={integrationQuery.isPending}
          isReconnectPending={createConnectSession.isPending}
          isSyncing={syncCalendar.isPending}
          onReconnect={() => {
            if (!connection) {
              return;
            }

            createConnectSession.mutate(
              connection.provider === "microsoft" ? "microsoft" : "google",
            );
          }}
          onSync={syncConnection}
        />
        <Box
          className="flex min-h-0 flex-1 flex-col overflow-hidden"
          data-walkthrough-target={walkthroughTargets.calendarGrid}
        >
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
          <CalendarScheduleDialog
            candidateStories={candidateStories}
            editingBlock={editingBlock}
            isOpen
            mode={activeDialogMode}
            onOpenChange={closeDialog}
            storyTerm={storyTerm}
            storyTermPlural={storyTermPlural}
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
    </Box>
  );
};

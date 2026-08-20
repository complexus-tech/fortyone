"use client";

import type { ReactNode } from "react";
import { useState, useSyncExternalStore } from "react";
import { addDays, addMinutes, format, startOfDay } from "date-fns";
import {
  CalendarIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  CloseIcon,
  ExternalLinkIcon,
  Video02Icon,
} from "icons";
import { Box, Button, Dialog, Flex, Input, Text } from "ui";
import { useWorkspacePath } from "@/hooks";
import {
  useCalendarIntegration,
  useCalendarSchedule,
  useOverrideCalendarScheduleIssue,
  useRetryCalendarScheduleIssue,
} from "@/lib/hooks/calendar";
import { formatTimeNeeded } from "@/lib/time-needed";
import type { CalendarScheduleIssue } from "@/lib/queries/calendar/types";
import { getStoryPath } from "@/modules/story/utils/story-url";
import {
  dismissMeetingUntilEnd,
  getUpcomingMeetings,
  isMeetingDismissed,
} from "./upcoming-meeting";
import type { UpcomingMeeting } from "./upcoming-meeting";

const CLOCK_REFRESH_INTERVAL_MS = 30 * 1000;
const subscribeToHydration = () => () => undefined;
let clockSnapshot = typeof window === "undefined" ? 0 : Date.now();
let clockInterval: number | undefined;
const clockSubscribers = new Set<() => void>();

const refreshClock = () => {
  clockSnapshot = Date.now();
  clockSubscribers.forEach((subscriber) => {
    subscriber();
  });
};

const subscribeToClock = (subscriber: () => void) => {
  clockSubscribers.add(subscriber);
  if (clockSubscribers.size === 1) {
    refreshClock();
    clockInterval = window.setInterval(refreshClock, CLOCK_REFRESH_INTERVAL_MS);
    window.addEventListener("focus", refreshClock);
  }

  return () => {
    clockSubscribers.delete(subscriber);
    if (clockSubscribers.size === 0) {
      if (clockInterval) window.clearInterval(clockInterval);
      clockInterval = undefined;
      window.removeEventListener("focus", refreshClock);
    }
  };
};

const getClockSnapshot = () => clockSnapshot;
const getServerClockSnapshot = () => 0;

type SidebarAssistantItem =
  | { id: string; kind: "meeting"; meeting: UpcomingMeeting }
  | { id: string; issue: CalendarScheduleIssue; kind: "schedule-issue" };

const getMeetingTitle = (title?: string) => title?.trim() || "Upcoming meeting";

const roundToNextHalfHour = (date: Date) => {
  const next = new Date(date);
  next.setMinutes(next.getMinutes() < 30 ? 30 : 60, 0, 0);
  return next;
};

const toDateTimeInputValue = (value: Date) =>
  format(value, "yyyy-MM-dd'T'HH:mm");

const ScheduleIssueDialog = ({
  issue,
  isSaving,
  now,
  onClose,
  onSubmit,
}: {
  issue: CalendarScheduleIssue;
  isSaving: boolean;
  now: number;
  onClose: () => void;
  onSubmit: (startAt: string) => void;
}) => {
  const defaultStart = roundToNextHalfHour(addMinutes(new Date(now), 30));
  const [startAt, setStartAt] = useState(() =>
    toDateTimeInputValue(defaultStart),
  );
  const parsedStartAt = new Date(startAt);
  const durationMinutes = issue.estimatedDurationMinutes ?? 0;
  const endAt = addMinutes(parsedStartAt, durationMinutes);
  const canSubmit =
    durationMinutes > 0 &&
    Number.isFinite(parsedStartAt.getTime()) &&
    parsedStartAt.getTime() >= now - 5 * 60 * 1000;

  return (
    <Dialog
      onOpenChange={(isOpen) => {
        if (!isOpen) onClose();
      }}
      open
    >
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title className="px-6 pt-0.5 text-lg">
            Choose a time for {issue.storyCode}
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="space-y-4">
          <Box>
            <Text fontWeight="medium">{issue.storyTitle}</Text>
            <Text className="mt-1" color="muted" fontSize="md">
              Maya will reserve {formatTimeNeeded(durationMinutes, "full")} and
              keep this exact time locked.
            </Text>
          </Box>
          <Input
            className="text-base"
            label="Start"
            labelClassName="text-base"
            min={toDateTimeInputValue(new Date(now))}
            onChange={(event) => {
              setStartAt(event.target.value);
            }}
            type="datetime-local"
            value={startAt}
          />
          {canSubmit ? (
            <Text color="muted" fontSize="md">
              Ends {format(endAt, "EEEE, MMM d 'at' h:mm a")}. If this overlaps
              a meeting or another task, FortyOne will show the conflict but
              keep your choice.
            </Text>
          ) : (
            <Text color="danger" fontSize="md">
              Choose a future start time.
            </Text>
          )}
        </Dialog.Body>
        <Dialog.Footer className="gap-3 border-0 pt-2">
          <Button color="tertiary" onClick={onClose} variant="outline">
            Cancel
          </Button>
          <Button
            color="invert"
            disabled={!canSubmit}
            loading={isSaving}
            onClick={() => {
              onSubmit(parsedStartAt.toISOString());
            }}
          >
            Lock this time
          </Button>
        </Dialog.Footer>
      </Dialog.Content>
    </Dialog>
  );
};

const MeetingContent = ({ meeting }: { meeting: UpcomingMeeting }) => {
  const startAt = new Date(meeting.event.startAt);
  const endAt = new Date(meeting.event.endAt);
  const statusLabel =
    meeting.status === "in-progress"
      ? "Happening now"
      : `Starts in ${meeting.minutesUntilStart} min`;

  return (
    <>
      <Flex align="center" className="gap-1.5 pr-6">
        <span className="bg-primary relative flex size-2 rounded-full">
          <span className="bg-primary absolute inset-0 animate-ping rounded-full opacity-40 motion-reduce:animate-none" />
        </span>
        <Text
          className="text-[0.72rem] tracking-[0.08em] uppercase"
          color="muted"
          fontWeight="semibold"
        >
          Meeting
        </Text>
        <Text
          className="bg-state-selected ml-auto rounded-md px-1.5 py-0.5 text-[0.78rem] tabular-nums"
          fontWeight="medium"
        >
          {statusLabel}
        </Text>
      </Flex>

      <Text className="mt-2 line-clamp-2 leading-snug" fontWeight="medium">
        {getMeetingTitle(meeting.event.title)}
      </Text>
      <Text className="mt-1 tabular-nums" color="muted" fontSize="sm">
        {format(startAt, "h:mm a")}–{format(endAt, "h:mm a")}
      </Text>

      <a
        className="bg-background-inverse text-foreground-inverse focus-visible:ring-ring mt-3 flex h-8.5 w-full items-center justify-center gap-1.5 rounded-lg px-2.5 text-[0.94rem] font-medium transition-opacity hover:opacity-90 focus-visible:ring-2 focus-visible:outline-none"
        href={meeting.meetingUrl}
        rel="noreferrer noopener"
        target="_blank"
      >
        <Video02Icon
          aria-hidden="true"
          className="text-foreground-inverse dark:text-foreground-inverse h-4 w-auto"
          strokeWidth={2}
        />
        Join meeting
        <ExternalLinkIcon
          aria-hidden="true"
          className="text-foreground-inverse dark:text-foreground-inverse h-3.5 w-auto"
          strokeWidth={2}
        />
      </a>
    </>
  );
};

const ScheduleIssueContent = ({
  issue,
  isRetrying,
  onChooseTime,
  onRetry,
}: {
  issue: CalendarScheduleIssue;
  isRetrying: boolean;
  onChooseTime: () => void;
  onRetry: () => void;
}) => {
  const { withWorkspace } = useWorkspacePath();
  const reason =
    issue.autoSchedulingReason?.trim() ||
    "Maya could not find enough available time in the planning window.";

  return (
    <>
      <Flex align="center" className="gap-1.5 pr-6">
        <span className="bg-warning flex size-2 rounded-full" />
        <Text
          className="text-[0.72rem] tracking-[0.08em] uppercase"
          color="muted"
          fontWeight="semibold"
        >
          Maya needs your help
        </Text>
      </Flex>

      <a
        className="focus-visible:ring-ring mt-2 block rounded-sm leading-snug font-medium outline-none hover:underline focus-visible:ring-2"
        href={withWorkspace(getStoryPath({ id: issue.storyId }))}
      >
        {issue.storyTitle}
      </a>
      <Text
        className="mt-1.5 line-clamp-3 leading-snug"
        color="muted"
        fontSize="sm"
      >
        {reason}
      </Text>
      {issue.estimatedDurationMinutes ? (
        <Flex align="center" className="mt-2 gap-1.5">
          <CalendarIcon aria-hidden="true" className="h-3.5" />
          <Text color="muted" fontSize="sm">
            {formatTimeNeeded(issue.estimatedDurationMinutes, "full")} needed
          </Text>
        </Flex>
      ) : null}

      <Flex className="mt-3 gap-2">
        <Button
          className="min-w-0 flex-1 px-2"
          color="tertiary"
          loading={isRetrying}
          onClick={onRetry}
          size="sm"
          variant="outline"
        >
          Retry
        </Button>
        <Button
          className="min-w-0 flex-1 px-2"
          color="invert"
          onClick={onChooseTime}
          size="sm"
        >
          Choose time
        </Button>
      </Flex>
    </>
  );
};

export const SidebarAssistantCards = ({
  fallback = null,
}: {
  fallback?: ReactNode;
}) => {
  const now = useSyncExternalStore(
    subscribeToClock,
    getClockSnapshot,
    getServerClockSnapshot,
  );
  const [activeIndex, setActiveIndex] = useState(0);
  const [dismissedMeetingIds, setDismissedMeetingIds] = useState<Set<string>>(
    () => new Set(),
  );
  const [dismissedIssueIds, setDismissedIssueIds] = useState<Set<string>>(
    () => new Set(),
  );
  const [selectedIssue, setSelectedIssue] =
    useState<CalendarScheduleIssue | null>(null);
  const integrationQuery = useCalendarIntegration();
  const connection = integrationQuery.data?.connections[0];
  const rangeStart = startOfDay(now);
  const scheduleQuery = useCalendarSchedule(
    {
      startAt: rangeStart.toISOString(),
      endAt: addDays(rangeStart, 2).toISOString(),
    },
    { enabled: Boolean(connection) },
  );
  const retryIssue = useRetryCalendarScheduleIssue();
  const overrideIssue = useOverrideCalendarScheduleIssue();
  const isHydrated = useSyncExternalStore(
    subscribeToHydration,
    () => true,
    () => false,
  );

  const meetings = isHydrated
    ? getUpcomingMeetings(scheduleQuery.data?.events ?? [], now).filter(
        (meeting) =>
          !dismissedMeetingIds.has(meeting.event.id) &&
          !isMeetingDismissed(meeting.event.id),
      )
    : [];
  const issues = (scheduleQuery.data?.scheduleIssues ?? []).filter(
    (issue) => !dismissedIssueIds.has(issue.storyId),
  );
  const items: SidebarAssistantItem[] = [
    ...meetings.map((meeting) => ({
      id: `meeting:${meeting.event.id}`,
      kind: "meeting" as const,
      meeting,
    })),
    ...issues.map((issue) => ({
      id: `schedule-issue:${issue.storyId}`,
      issue,
      kind: "schedule-issue" as const,
    })),
  ];
  if (items.length === 0) return fallback;

  const safeActiveIndex = activeIndex % items.length;
  const activeItem = items[safeActiveIndex];

  const dismissActiveItem = () => {
    if (activeItem.kind === "meeting") {
      dismissMeetingUntilEnd(activeItem.meeting);
      setDismissedMeetingIds((current) => {
        const next = new Set(current);
        next.add(activeItem.meeting.event.id);
        return next;
      });
    } else {
      setDismissedIssueIds((current) => {
        const next = new Set(current);
        next.add(activeItem.issue.storyId);
        return next;
      });
    }
    setActiveIndex(0);
  };

  return (
    <>
      <Box aria-live="polite" className="relative pb-1.5">
        {items.length > 1 ? (
          <Box
            aria-hidden="true"
            className="border-border bg-surface-elevated absolute inset-x-2 bottom-0 h-4 rounded-b-xl border-[0.5px] opacity-70"
          />
        ) : null}
        <Box className="border-border bg-surface-elevated shadow-shadow relative rounded-xl border-[0.5px] px-3.5 pt-3 pb-3.5 shadow-lg">
          <button
            aria-label={
              activeItem.kind === "meeting"
                ? "Dismiss meeting"
                : "Dismiss scheduling message"
            }
            className="text-foreground hover:bg-state-hover focus-visible:ring-ring absolute top-2 right-1.5 z-1 grid size-6 place-items-center rounded-md transition-colors outline-none focus-visible:ring-2"
            onClick={dismissActiveItem}
            type="button"
          >
            <CloseIcon
              aria-hidden="true"
              className="h-3.5 w-auto"
              strokeWidth={2.5}
            />
          </button>

          <Box>
            {activeItem.kind === "meeting" ? (
              <MeetingContent meeting={activeItem.meeting} />
            ) : (
              <ScheduleIssueContent
                isRetrying={Boolean(
                  retryIssue.isPending &&
                    retryIssue.variables === activeItem.issue.storyId,
                )}
                issue={activeItem.issue}
                onChooseTime={() => {
                  setSelectedIssue(activeItem.issue);
                }}
                onRetry={() => {
                  retryIssue.mutate(activeItem.issue.storyId);
                }}
              />
            )}
          </Box>

          {items.length > 1 ? (
            <Flex
              align="center"
              className="mt-3 border-t pt-2"
              justify="between"
            >
              <Text className="tabular-nums" color="muted" fontSize="xs">
                {safeActiveIndex + 1} of {items.length}
              </Text>
              <Flex className="gap-1">
                <button
                  aria-label="Previous message"
                  className="hover:bg-state-hover focus-visible:ring-ring grid size-6 place-items-center rounded-md outline-none focus-visible:ring-2"
                  onClick={() => {
                    setActiveIndex(
                      (safeActiveIndex - 1 + items.length) % items.length,
                    );
                  }}
                  type="button"
                >
                  <ChevronLeftIcon aria-hidden="true" className="h-3.5" />
                </button>
                <button
                  aria-label="Next message"
                  className="hover:bg-state-hover focus-visible:ring-ring grid size-6 place-items-center rounded-md outline-none focus-visible:ring-2"
                  onClick={() => {
                    setActiveIndex((safeActiveIndex + 1) % items.length);
                  }}
                  type="button"
                >
                  <ChevronRightIcon aria-hidden="true" className="h-3.5" />
                </button>
              </Flex>
            </Flex>
          ) : null}
        </Box>
      </Box>

      {selectedIssue ? (
        <ScheduleIssueDialog
          isSaving={overrideIssue.isPending}
          issue={selectedIssue}
          now={now}
          onClose={() => {
            setSelectedIssue(null);
          }}
          onSubmit={(startAt) => {
            overrideIssue.mutate(
              {
                storyId: selectedIssue.storyId,
                startAt,
                timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
              },
              {
                onSuccess: () => {
                  setSelectedIssue(null);
                },
              },
            );
          }}
        />
      ) : null}
    </>
  );
};

// Retain the existing export for nearby consumers while the component now owns
// the full sidebar assistant stack, not just meeting reminders.
export const UpcomingMeetingCard = SidebarAssistantCards;

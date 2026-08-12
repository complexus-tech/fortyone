"use client";

import type { ReactNode } from "react";
import { useEffect, useState } from "react";
import { addDays, format, startOfDay } from "date-fns";
import { CloseIcon, ExternalLinkIcon, Video02Icon } from "icons";
import { Box, Flex, Text } from "ui";
import {
  useCalendarIntegration,
  useCalendarSchedule,
} from "@/lib/hooks/calendar";
import { getUpcomingMeeting } from "./upcoming-meeting";

const CLOCK_REFRESH_INTERVAL_MS = 30 * 1000;

const getMeetingTitle = (title?: string) => title?.trim() || "Upcoming meeting";

export const UpcomingMeetingCard = ({
  fallback = null,
}: {
  fallback?: ReactNode;
}) => {
  const [now, setNow] = useState(() => Date.now());
  const [dismissedMeetingId, setDismissedMeetingId] = useState<string | null>(
    null,
  );
  const integrationQuery = useCalendarIntegration();
  const connection = integrationQuery.data?.connections[0];
  const rangeStart = startOfDay(now);
  const scheduleQuery = useCalendarSchedule(
    {
      startAt: rangeStart.toISOString(),
      endAt: addDays(rangeStart, 2).toISOString(),
    },
    { enabled: Boolean(connection?.canReadEventDetails) },
  );

  useEffect(() => {
    const refreshClock = () => {
      setNow(Date.now());
    };
    const interval = window.setInterval(
      refreshClock,
      CLOCK_REFRESH_INTERVAL_MS,
    );
    window.addEventListener("focus", refreshClock);

    return () => {
      window.clearInterval(interval);
      window.removeEventListener("focus", refreshClock);
    };
  }, []);

  const meeting = getUpcomingMeeting(scheduleQuery.data?.events ?? [], now);

  if (!meeting || meeting.event.id === dismissedMeetingId) return fallback;

  const startAt = new Date(meeting.event.startAt);
  const endAt = new Date(meeting.event.endAt);
  const statusLabel =
    meeting.status === "in-progress"
      ? "Happening now"
      : `Starts in ${meeting.minutesUntilStart} min`;

  return (
    <Box className="border-border bg-surface-elevated shadow-shadow rounded-xl border-[0.5px] px-3.5 pt-3 pb-3.5 shadow-lg">
      <Flex align="center" justify="between">
        <Flex align="center" className="gap-1.5">
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
        </Flex>
        <Flex align="center" className="-mr-2 ml-auto gap-1">
          <Text
            className="bg-state-selected rounded-md px-1.5 py-0.5 text-[0.78rem] tabular-nums"
            fontWeight="medium"
          >
            {statusLabel}
          </Text>
          <button
            aria-label="Dismiss meeting"
            className="text-foreground hover:bg-state-hover focus-visible:ring-ring dark:text-foreground grid size-6 shrink-0 place-items-center rounded-md transition-colors outline-none focus-visible:ring-2"
            onClick={() => {
              setDismissedMeetingId(meeting.event.id);
            }}
            type="button"
          >
            <CloseIcon
              aria-hidden="true"
              className="text-foreground dark:text-foreground h-3.5 w-auto"
              strokeWidth={2.5}
            />
          </button>
        </Flex>
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
    </Box>
  );
};

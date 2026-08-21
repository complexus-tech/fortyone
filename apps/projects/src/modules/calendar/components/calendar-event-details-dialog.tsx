"use client";

import { format, isSameDay } from "date-fns";
import { useState } from "react";
import {
  CalendarIcon,
  ClockIcon,
  ExternalLinkIcon,
  UserMultiple02Icon,
  Video02Icon,
} from "icons";
import { Badge, Box, Button, Dialog, Flex, Skeleton, Text } from "ui";
import { useCalendarEvent } from "@/lib/hooks/calendar";
import type {
  CalendarEventAttendee,
  CalendarEventSummary,
} from "@/lib/queries/calendar/types";
import { parseCalendarDate } from "./calendar-layout";

const getSafeExternalUrl = (value?: string) => {
  if (!value) return null;
  try {
    const url = new URL(value);
    return url.protocol === "https:" ? url.toString() : null;
  } catch {
    return null;
  }
};

export const getCalendarEventTitle = (
  event: Pick<CalendarEventSummary, "isPrivate" | "title">,
) => {
  if (event.isPrivate) return "Busy";
  return event.title?.trim() || "Untitled event";
};

export const getCalendarEventTimeLabel = (
  event: Pick<
    CalendarEventSummary,
    "endAt" | "endDate" | "isAllDay" | "startAt" | "startDate"
  >,
) => {
  const start = new Date(event.startAt);
  const end = new Date(event.endAt);
  if (event.isAllDay) {
    const dateOnlyStart = parseCalendarDate(event.startDate);
    const dateOnlyEnd = parseCalendarDate(event.endDate);
    const visibleStart = dateOnlyStart ?? start;
    const exclusiveEnd = dateOnlyEnd ?? end;
    const inclusiveEnd = new Date(exclusiveEnd);
    inclusiveEnd.setDate(inclusiveEnd.getDate() - 1);
    return isSameDay(visibleStart, inclusiveEnd)
      ? format(visibleStart, "EEEE, MMMM d")
      : `${format(visibleStart, "MMM d")} – ${format(inclusiveEnd, "MMM d")}`;
  }
  if (isSameDay(start, end)) {
    return `${format(start, "EEEE, MMMM d")} · ${format(start, "h:mm a")} – ${format(end, "h:mm a")}`;
  }
  return `${format(start, "MMM d, h:mm a")} – ${format(end, "MMM d, h:mm a")}`;
};

const getCalendarLabel = (event?: CalendarEventSummary | null) => {
  const calendarId = event?.calendarId?.trim();
  if (calendarId && calendarId !== "primary") return calendarId;
  if (event?.provider === "google") return "Google Calendar";
  if (event?.provider === "microsoft") return "Outlook Calendar";
  return "External calendar";
};

const getPersonLabel = ({
  displayName,
  email,
}: {
  displayName?: string;
  email?: string;
}) => displayName?.trim() || email?.trim() || "Unknown attendee";

const getResponseLabel = (attendee: CalendarEventAttendee) => {
  const labels: Record<string, string> = {
    accepted: "Accepted",
    declined: "Declined",
    needsAction: "Awaiting response",
    tentative: "Tentative",
  };
  return attendee.responseStatus
    ? labels[attendee.responseStatus] ?? attendee.responseStatus
    : null;
};

const DetailsLoading = () => (
  <Box aria-label="Loading event details" className="space-y-5" role="status">
    <Skeleton className="h-5 w-32" />
    <Skeleton className="h-12 w-full" />
    <Skeleton className="h-20 w-full" />
  </Box>
);

const ATTENDEES_PAGE_SIZE = 20;

const CalendarEventAttendees = ({
  attendees,
  attendeesOmitted,
}: {
  attendees: CalendarEventAttendee[];
  attendeesOmitted: boolean;
}) => {
  const [visibleAttendeeCount, setVisibleAttendeeCount] =
    useState(ATTENDEES_PAGE_SIZE);
  const visibleAttendees = attendees.slice(0, visibleAttendeeCount);
  const hasMoreAttendees = visibleAttendeeCount < attendees.length;

  return (
    <Box>
      <Flex align="center" className="mb-2" gap={2}>
        <UserMultiple02Icon className="text-text-muted h-5 w-auto" />
        <Text fontSize="md" fontWeight="medium">
          Attendees · {attendees.length}
          {attendeesOmitted ? "+" : ""}
        </Text>
      </Flex>
      <Box className="divide-border bg-surface-muted/60 divide-y overflow-hidden rounded-lg dark:divide-white/[0.08] dark:bg-white/[0.04]">
        {visibleAttendees.map((attendee) => (
          <Flex
            align="center"
            className="px-3 py-2.5"
            gap={3}
            justify="between"
            key={`${attendee.email ?? attendee.displayName ?? "attendee"}-${attendee.responseStatus ?? "unknown"}`}
          >
            <Box className="min-w-0">
              <Text className="truncate" fontSize="md">
                {getPersonLabel(attendee)}
              </Text>
              {attendee.optional ? (
                <Text color="muted" fontSize="md">
                  Optional
                </Text>
              ) : null}
            </Box>
            {getResponseLabel(attendee) ? (
              <Text className="shrink-0" color="muted" fontSize="md">
                {getResponseLabel(attendee)}
              </Text>
            ) : null}
          </Flex>
        ))}
      </Box>
      {hasMoreAttendees ? (
        <Button
          className="mt-3 w-full"
          color="tertiary"
          onClick={() => {
            setVisibleAttendeeCount((current) => current + ATTENDEES_PAGE_SIZE);
          }}
          size="sm"
          variant="naked"
        >
          Load more attendees
        </Button>
      ) : null}
    </Box>
  );
};

export const CalendarEventDetailsDialog = ({
  event,
  onOpenChange,
}: {
  event: CalendarEventSummary | null;
  onOpenChange: (open: boolean) => void;
}) => {
  const eventQuery = useCalendarEvent(
    event && !event.isPrivate ? event.id : null,
  );

  const details = event?.isPrivate ? undefined : eventQuery.data;
  const visibleEvent = details ?? event;
  const isPrivate = Boolean(visibleEvent?.isPrivate);
  const meetingUrl = getSafeExternalUrl(visibleEvent?.meetingUrl);
  const htmlLink = getSafeExternalUrl(visibleEvent?.htmlLink);
  const attendees = details?.attendees ?? [];

  return (
    <Dialog onOpenChange={onOpenChange} open={Boolean(event)}>
      <Dialog.Content
        className="border-border/70 shadow-shadow bg-surface-elevated mr-6 mb-8 max-w-148 rounded-3xl border-[0.5px] shadow-2xl outline-none md:mt-auto"
        overlayClassName="justify-end bg-black/10"
      >
        <Dialog.Header className="border-border flex min-h-16 items-center border-b-[0.5px] px-6 pr-16">
          <Dialog.Title className="min-w-0 truncate text-lg font-semibold">
            {visibleEvent
              ? getCalendarEventTitle(visibleEvent)
              : "Calendar event"}
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Description className="sr-only">
          Details for the selected calendar event.
        </Dialog.Description>

        <Dialog.Body className="h-[calc(100dvh-8rem)] max-h-[calc(100dvh-8rem)] px-6 py-6">
          {event ? (
            <Box className="space-y-6">
              <Box>
                <Text as="h2" className="text-2xl" fontWeight="semibold">
                  {visibleEvent ? getCalendarEventTitle(visibleEvent) : null}
                </Text>
                {!isPrivate && visibleEvent ? (
                  <Badge
                    className="bg-surface-muted/70 mt-3 h-7.5 border-0 px-2 py-0 text-base dark:bg-white/[0.06]"
                    color="tertiary"
                  >
                    {getCalendarLabel(visibleEvent)}
                  </Badge>
                ) : null}
              </Box>

              <Flex align="start" gap={3}>
                <ClockIcon className="text-text-muted mt-0.5 h-5 w-auto shrink-0" />
                <Text fontSize="md">
                  {visibleEvent
                    ? getCalendarEventTimeLabel(visibleEvent)
                    : null}
                </Text>
              </Flex>

              {!isPrivate ? (
                <>
                  {meetingUrl || htmlLink ? (
                    <Flex align="center" className="flex-wrap" gap={2}>
                      {meetingUrl ? (
                        <a
                          className="bg-background-inverse text-foreground-inverse flex h-10 w-max items-center gap-2 rounded-xl border px-3 text-base font-medium transition hover:opacity-90"
                          href={meetingUrl}
                          rel="noreferrer noopener"
                          target="_blank"
                        >
                          <Video02Icon
                            aria-hidden="true"
                            className="text-foreground-inverse dark:text-foreground-inverse h-4 w-auto"
                            strokeWidth={2}
                          />
                          Join meeting
                        </a>
                      ) : null}
                      {htmlLink ? (
                        <a
                          className="border-border hover:bg-state-hover flex h-10 w-max items-center gap-2 rounded-xl border px-3 text-base font-medium transition"
                          href={htmlLink}
                          rel="noreferrer noopener"
                          target="_blank"
                        >
                          <ExternalLinkIcon className="h-4 w-auto" />
                          Open in {getCalendarLabel(visibleEvent)}
                        </a>
                      ) : null}
                    </Flex>
                  ) : null}

                  {visibleEvent?.location ? (
                    <Flex align="start" gap={3}>
                      <CalendarIcon className="text-text-muted mt-0.5 h-5 w-auto shrink-0" />
                      <Box>
                        <Text color="muted" fontSize="md">
                          Location
                        </Text>
                        <Text
                          className="mt-0.5 whitespace-pre-wrap"
                          fontSize="md"
                        >
                          {visibleEvent.location}
                        </Text>
                      </Box>
                    </Flex>
                  ) : null}

                  {eventQuery.isPending ? <DetailsLoading /> : null}

                  {eventQuery.isError ? (
                    <Box className="border-border bg-surface-muted/40 rounded-xl border px-4 py-3">
                      <Text fontSize="md" fontWeight="medium">
                        Couldn&apos;t load every detail
                      </Text>
                      <Text className="mt-1" color="muted" fontSize="md">
                        The event remains visible from the latest calendar sync.
                      </Text>
                    </Box>
                  ) : null}

                  {details?.organizer ? (
                    <Flex align="start" gap={3}>
                      <UserMultiple02Icon className="text-text-muted mt-0.5 h-5 w-auto shrink-0" />
                      <Box>
                        <Text color="muted" fontSize="md">
                          Organizer
                        </Text>
                        <Text className="mt-0.5" fontSize="md">
                          {getPersonLabel(details.organizer)}
                        </Text>
                      </Box>
                    </Flex>
                  ) : null}

                  {attendees.length > 0 ? (
                    <CalendarEventAttendees
                      attendees={attendees}
                      attendeesOmitted={Boolean(details?.attendeesOmitted)}
                      key={event.id}
                    />
                  ) : null}

                  {details?.description ? (
                    <Box>
                      <Text className="mb-2" fontSize="md" fontWeight="medium">
                        Description
                      </Text>
                      <Text
                        className="whitespace-pre-wrap"
                        color="muted"
                        fontSize="md"
                      >
                        {details.description}
                      </Text>
                    </Box>
                  ) : null}
                </>
              ) : null}
            </Box>
          ) : null}
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};

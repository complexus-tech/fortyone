/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ElementType, ReactNode } from "react";
import type { CalendarEventSummary } from "@/lib/queries/calendar/types";
import {
  CalendarEventDetailsDialog,
  getCalendarEventTimeLabel,
} from "./calendar-event-details-dialog";

const useCalendarEvent = jest.fn();

jest.mock("@/lib/hooks/calendar", () => ({
  useCalendarEvent: (eventId: string | null) => useCalendarEvent(eventId),
}));

jest.mock("icons", () => ({
  CalendarIcon: () => <span aria-hidden>Calendar icon</span>,
  ClockIcon: () => <span aria-hidden>Clock icon</span>,
  ExternalLinkIcon: () => <span aria-hidden>External link icon</span>,
  UserMultiple02Icon: () => <span aria-hidden>People icon</span>,
  Video02Icon: () => <span aria-hidden>Video icon</span>,
}));

jest.mock("ui", () => {
  const Container = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  const Dialog = ({ children }: { children: ReactNode }) => (
    <div>{children}</div>
  );
  Dialog.Content = Container;
  Dialog.Header = Container;
  Dialog.Description = Container;
  Dialog.Body = Container;
  Dialog.Title = Container;

  return {
    Badge: Container,
    Box: Container,
    Button: ({
      children,
      onClick,
    }: {
      children: ReactNode;
      onClick: () => void;
    }) => (
      <button onClick={onClick} type="button">
        {children}
      </button>
    ),
    Dialog,
    Flex: Container,
    Skeleton: Container,
    Text: ({
      as: Component = "span",
      children,
    }: {
      as?: ElementType;
      children: ReactNode;
    }) => <Component>{children}</Component>,
  };
});

const event: CalendarEventSummary = {
  id: "event-1",
  provider: "google",
  calendarId: "primary",
  title: "Customer review",
  location: "Studio 2",
  meetingUrl: "https://meet.google.com/example",
  htmlLink: "https://calendar.google.com/event?eid=example",
  startAt: new Date(2026, 7, 7, 14).toISOString(),
  endAt: new Date(2026, 7, 7, 15).toISOString(),
  isAllDay: false,
  isPrivate: false,
};

describe("CalendarEventDetailsDialog", () => {
  beforeEach(() => {
    useCalendarEvent.mockReset();
    useCalendarEvent.mockReturnValue({
      data: {
        ...event,
        description: "Review the launch plan.",
        organizer: { displayName: "Tariro" },
        attendees: [
          {
            displayName: "Joseph",
            responseStatus: "accepted",
            optional: false,
            organizer: false,
            self: true,
          },
        ],
        attendeesOmitted: false,
      },
      isError: false,
      isPending: false,
    });
  });

  it("loads owner-only details when a visible event is opened", () => {
    render(
      <CalendarEventDetailsDialog event={event} onOpenChange={jest.fn()} />,
    );

    expect(useCalendarEvent).toHaveBeenCalledWith("event-1");
    expect(screen.getByText("Studio 2")).toBeInTheDocument();
    expect(screen.getByText("Tariro")).toBeInTheDocument();
    expect(screen.getByText("Joseph")).toBeInTheDocument();
    expect(screen.getByText("Review the launch plan.")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Join meeting" })).toHaveAttribute(
      "href",
      "https://meet.google.com/example",
    );
  });

  it("never fetches or reveals metadata for a private event", () => {
    render(
      <CalendarEventDetailsDialog
        event={{
          ...event,
          title: "Secret board discussion",
          location: "Private room",
          isPrivate: true,
        }}
        onOpenChange={jest.fn()}
      />,
    );

    expect(useCalendarEvent).toHaveBeenCalledWith(null);
    expect(screen.getAllByText("Busy").length).toBeGreaterThan(0);
    expect(
      screen.queryByText("Secret board discussion"),
    ).not.toBeInTheDocument();
    expect(screen.queryByText("Private room")).not.toBeInTheDocument();
    expect(screen.queryByText("Google Calendar")).not.toBeInTheDocument();
  });

  it("replaces stale summary metadata when refreshed details become private", () => {
    useCalendarEvent.mockReturnValue({
      data: {
        ...event,
        title: "Busy",
        location: undefined,
        meetingUrl: undefined,
        htmlLink: undefined,
        isPrivate: true,
        attendees: [],
        attendeesOmitted: false,
      },
      isError: false,
      isPending: false,
    });

    render(
      <CalendarEventDetailsDialog event={event} onOpenChange={jest.fn()} />,
    );

    expect(screen.getAllByText("Busy").length).toBeGreaterThan(0);
    expect(screen.queryByText("Customer review")).not.toBeInTheDocument();
    expect(screen.queryByText("Studio 2")).not.toBeInTheDocument();
    expect(screen.queryByText("Google Calendar")).not.toBeInTheDocument();
  });

  it("does not render non-HTTPS provider links", () => {
    useCalendarEvent.mockReturnValue({
      data: {
        ...event,
        meetingUrl: "http://example.com/meeting",
        htmlLink: ["java", "script:alert(1)"].join(""),
        attendees: [],
        attendeesOmitted: false,
      },
      isError: false,
      isPending: false,
    });

    render(
      <CalendarEventDetailsDialog
        event={{
          ...event,
          meetingUrl: "http://example.com/meeting",
          htmlLink: ["java", "script:alert(1)"].join(""),
        }}
        onOpenChange={jest.fn()}
      />,
    );

    expect(screen.queryByRole("link", { name: "Join meeting" })).toBeNull();
    expect(
      screen.queryByRole("link", { name: "Open in Google Calendar" }),
    ).toBeNull();
  });

  it("uses provider date-only values for all-day labels", () => {
    expect(
      getCalendarEventTimeLabel({
        startAt: "2026-08-06T22:00:00Z",
        endAt: "2026-08-07T22:00:00Z",
        startDate: "2026-08-07",
        endDate: "2026-08-08",
        isAllDay: true,
      }),
    ).toBe("Friday, August 7");
  });

  it("marks attendee counts as truncated when Google omits attendees", () => {
    useCalendarEvent.mockReturnValue({
      data: {
        ...event,
        attendees: [
          {
            displayName: "Joseph",
            optional: false,
            organizer: false,
            self: true,
          },
        ],
        attendeesOmitted: true,
      },
      isError: false,
      isPending: false,
    });

    render(
      <CalendarEventDetailsDialog event={event} onOpenChange={jest.fn()} />,
    );

    expect(screen.getByText("Attendees · 1+")).toBeInTheDocument();
  });

  it("reveals attendees in batches of twenty", () => {
    const attendees = Array.from({ length: 21 }, (_, index) => ({
      displayName: `Attendee ${index + 1}`,
      optional: false,
      organizer: false,
      self: false,
    }));
    useCalendarEvent.mockReturnValue({
      data: { ...event, attendees, attendeesOmitted: false },
      isError: false,
      isPending: false,
    });

    render(
      <CalendarEventDetailsDialog event={event} onOpenChange={jest.fn()} />,
    );

    expect(screen.getByText("Attendee 1")).toBeInTheDocument();
    expect(screen.getByText("Attendee 20")).toBeInTheDocument();
    expect(screen.queryByText("Attendee 21")).not.toBeInTheDocument();

    fireEvent.click(
      screen.getByRole("button", { name: "Load more attendees" }),
    );

    expect(screen.getByText("Attendee 21")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Load more attendees" }),
    ).not.toBeInTheDocument();
  });
});

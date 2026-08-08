/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import type { CalendarSchedule } from "@/lib/queries/calendar/types";
import { UpcomingMeetingCard } from "./upcoming-meeting-card";

const useCalendarSchedule = jest.fn();
const syncCalendar = jest.fn();

jest.mock("ui", () => ({
  Box: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Text: ({ children }: { children: ReactNode }) => <span>{children}</span>,
}));

jest.mock("icons", () => ({
  CloseIcon: () => null,
  ExternalLinkIcon: () => null,
  Video02Icon: () => null,
}));

jest.mock("@/lib/hooks/calendar", () => ({
  useCalendarAutoSync: () => undefined,
  useCalendarIntegration: () => ({
    data: {
      connections: [
        {
          id: "connection-1",
          canReadEventDetails: true,
          lastSyncedAt: "2026-08-08T09:59:00.000Z",
          syncStatus: "synced",
        },
      ],
    },
  }),
  useCalendarSchedule: (...args: unknown[]) => useCalendarSchedule(...args),
  useSyncCalendarConnection: () => ({
    isPending: false,
    mutate: syncCalendar,
  }),
}));

const createSchedule = (): CalendarSchedule => ({
  startAt: "2026-08-08T00:00:00.000Z",
  endAt: "2026-08-10T00:00:00.000Z",
  events: [
    {
      id: "meeting-1",
      provider: "google",
      title: "Weekly planning",
      meetingUrl: "https://meet.google.com/abc-defg-hij",
      startAt: "2026-08-08T10:10:00.000Z",
      endAt: "2026-08-08T10:40:00.000Z",
      isAllDay: false,
      isPrivate: false,
    },
  ],
  busyWindows: [],
  blocks: [],
});

describe("UpcomingMeetingCard", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date("2026-08-08T10:00:00.000Z"));
    useCalendarSchedule.mockReset();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it("replaces the fallback with the imminent meeting", () => {
    useCalendarSchedule.mockReturnValue({ data: createSchedule() });

    const { unmount } = render(
      <UpcomingMeetingCard fallback={<div>Upgrade plan</div>} />,
    );

    expect(screen.getByText("Weekly planning")).toBeInTheDocument();
    expect(screen.getByText("Starts in 10 min")).toBeInTheDocument();
    expect(screen.queryByText("Upgrade plan")).toBeNull();
    expect(screen.getByRole("link", { name: /join meeting/i })).toHaveAttribute(
      "href",
      "https://meet.google.com/abc-defg-hij",
    );

    unmount();
  });

  it("returns the fallback when there is no relevant meeting", () => {
    useCalendarSchedule.mockReturnValue({
      data: { ...createSchedule(), events: [] },
    });

    const { unmount } = render(
      <UpcomingMeetingCard fallback={<div>Upgrade plan</div>} />,
    );

    expect(screen.getByText("Upgrade plan")).toBeInTheDocument();
    expect(screen.queryByText("Join meeting")).toBeNull();

    unmount();
  });

  it("restores the fallback when the meeting is dismissed", () => {
    useCalendarSchedule.mockReturnValue({ data: createSchedule() });

    const { unmount } = render(
      <UpcomingMeetingCard fallback={<div>Normal footer</div>}>
        <div>Meeting actions</div>
      </UpcomingMeetingCard>,
    );

    expect(screen.getByText("Weekly planning")).toBeInTheDocument();
    expect(screen.getByText("Starts in 10 min")).toBeInTheDocument();
    expect(screen.getByText("Meeting actions")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /join meeting/i })).toHaveAttribute(
      "href",
      "https://meet.google.com/abc-defg-hij",
    );

    fireEvent.click(screen.getByRole("button", { name: "Dismiss meeting" }));

    expect(screen.getByText("Normal footer")).toBeInTheDocument();
    expect(screen.queryByText("Meeting actions")).toBeNull();
    expect(screen.queryByRole("link", { name: /join meeting/i })).toBeNull();

    unmount();
  });
});

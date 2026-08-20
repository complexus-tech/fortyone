/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import type { CalendarSchedule } from "@/lib/queries/calendar/types";
import { UpcomingMeetingCard } from "./upcoming-meeting-card";

const useCalendarSchedule = jest.fn();
const syncCalendar = jest.fn();
const retryScheduleIssue = jest.fn();
const overrideScheduleIssue = jest.fn();

jest.mock("ui", () => {
  function Dialog({ children }: { children: ReactNode }) {
    return <div>{children}</div>;
  }
  function DialogContent({ children }: { children: ReactNode }) {
    return <div>{children}</div>;
  }
  function DialogHeader({ children }: { children: ReactNode }) {
    return <div>{children}</div>;
  }
  function DialogTitle({ children }: { children: ReactNode }) {
    return <h2>{children}</h2>;
  }
  function DialogBody({ children }: { children: ReactNode }) {
    return <div>{children}</div>;
  }
  function DialogFooter({ children }: { children: ReactNode }) {
    return <div>{children}</div>;
  }
  Dialog.Content = DialogContent;
  Dialog.Header = DialogHeader;
  Dialog.Title = DialogTitle;
  Dialog.Body = DialogBody;
  Dialog.Footer = DialogFooter;

  return {
    Box: ({ children }: { children: ReactNode }) => <div>{children}</div>,
    Button: ({
      children,
      leftIcon,
      onClick,
    }: {
      children: ReactNode;
      leftIcon?: ReactNode;
      onClick?: () => void;
    }) => (
      <button onClick={onClick} type="button">
        {leftIcon}
        {children}
      </button>
    ),
    Dialog,
    Flex: ({ children }: { children: ReactNode }) => <div>{children}</div>,
    Input: ({
      label,
      labelClassName: _labelClassName,
      ...props
    }: React.InputHTMLAttributes<HTMLInputElement> & {
      label?: string;
      labelClassName?: string;
    }) => <input aria-label={label} {...props} />,
    Text: ({ children }: { children: ReactNode }) => <span>{children}</span>,
  };
});

jest.mock("icons", () => ({
  CalendarIcon: () => null,
  ChevronLeftIcon: () => null,
  ChevronRightIcon: () => null,
  CloseIcon: () => null,
  ExternalLinkIcon: () => null,
  RefreshIcon: () => null,
  Video02Icon: () => null,
}));

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
    workspaceSlug: "acme",
  }),
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
  useOverrideCalendarScheduleIssue: () => ({
    isPending: false,
    mutate: overrideScheduleIssue,
  }),
  useRetryCalendarScheduleIssue: () => ({
    isPending: false,
    mutate: retryScheduleIssue,
    variables: undefined,
  }),
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
  scheduleIssues: [],
});

describe("UpcomingMeetingCard", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date("2026-08-08T10:00:00.000Z"));
    useCalendarSchedule.mockReset();
    retryScheduleIssue.mockReset();
    overrideScheduleIssue.mockReset();
    document.cookie =
      "fortyone_meeting_dismissed_meeting-1=; Path=/; Max-Age=0";
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
      <UpcomingMeetingCard fallback={<div>Normal footer</div>} />,
    );

    expect(screen.getByText("Weekly planning")).toBeInTheDocument();
    expect(screen.getByText("Starts in 10 min")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /join meeting/i })).toHaveAttribute(
      "href",
      "https://meet.google.com/abc-defg-hij",
    );

    fireEvent.click(screen.getByRole("button", { name: "Dismiss meeting" }));

    expect(screen.getByText("Normal footer")).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: /join meeting/i })).toBeNull();
    expect(document.cookie).toContain("fortyone_meeting_dismissed_meeting-1=1");

    unmount();

    render(<UpcomingMeetingCard fallback={<div>Normal footer</div>} />);
    expect(screen.getByText("Normal footer")).toBeInTheDocument();
  });

  it("shows an actionable Maya card for a story that could not fit", () => {
    useCalendarSchedule.mockReturnValue({
      data: {
        ...createSchedule(),
        events: [],
        scheduleIssues: [
          {
            storyId: "story-1",
            storyTitle: "Prepare the launch brief",
            storyCode: "ENG-42",
            teamId: "team-1",
            teamName: "Engineering",
            teamCode: "ENG",
            estimatedDurationMinutes: 90,
            autoSchedulingStatus: "cannot_fit",
            autoSchedulingReason: "No safe focus window remains before Friday.",
            updatedAt: "2026-08-08T09:55:00.000Z",
          },
        ],
      },
    });

    render(<UpcomingMeetingCard fallback={<div>Normal footer</div>} />);

    expect(screen.getByText("Maya needs your help")).toBeInTheDocument();
    expect(screen.getByText("Prepare the launch brief")).toBeInTheDocument();
    expect(screen.queryByText("ENG-42")).not.toBeInTheDocument();
    expect(screen.getByText("1 hour 30 minutes needed")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(retryScheduleIssue).toHaveBeenCalledWith("story-1");

    fireEvent.click(screen.getByRole("button", { name: "Choose time" }));
    expect(
      screen.getByRole("heading", { name: "Choose a time for ENG-42" }),
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Lock this time" }));
    expect(overrideScheduleIssue).toHaveBeenCalledWith(
      expect.objectContaining({
        storyId: "story-1",
        timezone: expect.any(String),
      }),
      expect.objectContaining({ onSuccess: expect.any(Function) }),
    );
  });
});

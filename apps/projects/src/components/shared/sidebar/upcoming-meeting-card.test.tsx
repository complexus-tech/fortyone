/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import type {
  CalendarSchedule,
  CalendarScheduleIssue,
} from "@/lib/queries/calendar/types";
import type { Objective } from "@/modules/objectives/types";
import { UpcomingMeetingCard } from "./upcoming-meeting-card";

const useCalendarSchedule = jest.fn();
const syncCalendar = jest.fn();
const retryScheduleIssue = jest.fn();
const overrideScheduleIssue = jest.fn();
const mockObjectives: Objective[] = [];

jest.mock("next/link", () => ({
  __esModule: true,
  default: ({
    children,
    ...props
  }: React.AnchorHTMLAttributes<HTMLAnchorElement>) => (
    <a data-next-link="true" {...props}>
      {children}
    </a>
  ),
}));

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
  function Popover({ children }: { children: ReactNode }) {
    return <div>{children}</div>;
  }
  function PopoverTrigger({ children }: { children: ReactNode }) {
    return <>{children}</>;
  }
  function PopoverContent({ children }: { children: ReactNode }) {
    return <div>{children}</div>;
  }
  Popover.Trigger = PopoverTrigger;
  Popover.Content = PopoverContent;

  return {
    Box: ({
      children,
      className,
    }: {
      children: ReactNode;
      className?: string;
    }) => <div className={className}>{children}</div>,
    Button: ({
      children,
      href,
      leftIcon,
      onClick,
    }: {
      children: ReactNode;
      href?: string;
      leftIcon?: ReactNode;
      onClick?: () => void;
    }) =>
      href ? (
        <a href={href}>
          {leftIcon}
          {children}
        </a>
      ) : (
        <button onClick={onClick} type="button">
          {leftIcon}
          {children}
        </button>
      ),
    Dialog,
    Flex: ({
      children,
      className,
    }: {
      children: ReactNode;
      className?: string;
    }) => <div className={className}>{children}</div>,
    Input: ({
      hasError: _hasError,
      helpText,
      label,
      labelClassName: _labelClassName,
      rightIcon,
      ...props
    }: React.InputHTMLAttributes<HTMLInputElement> & {
      hasError?: boolean;
      helpText?: string;
      label?: string;
      labelClassName?: string;
      rightIcon?: ReactNode;
    }) => (
      <label>
        {label}
        <input aria-label={label} {...props} />
        {rightIcon}
        {helpText}
      </label>
    ),
    Popover,
    Text: ({
      children,
      className,
      title,
    }: {
      children: ReactNode;
      className?: string;
      title?: string;
    }) => (
      <span className={className} title={title}>
        {children}
      </span>
    ),
    Wrapper: ({
      children,
      className,
    }: {
      children: ReactNode;
      className?: string;
    }) => <div className={className}>{children}</div>,
  };
});

jest.mock("icons", () => ({
  CalendarIcon: () => null,
  ChevronLeftIcon: () => null,
  ChevronRightIcon: () => null,
  CloseIcon: () => null,
  ExternalLinkIcon: () => null,
  InfoIcon: () => null,
  RefreshIcon: () => null,
  Video02Icon: () => null,
  WarningIcon: ({ className }: { className?: string }) => (
    <svg className={className} data-testid="collapsed-warning-icon" />
  ),
}));

jest.mock("@/hooks", () => ({
  useTerminology: () => ({
    getTermDisplay: () => "story",
  }),
  useWorkspacePath: () => ({
    withWorkspace: (path: string) => `/acme${path}`,
    workspaceSlug: "acme",
  }),
}));

jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "user-1" } } }),
}));

jest.mock("@/modules/objectives/hooks/use-objectives", () => ({
  useObjectives: () => ({ data: mockObjectives }),
}));

jest.mock("@/lib/hooks/objective-statuses", () => ({
  useObjectiveStatuses: () => ({
    data: [{ category: "started", id: "status-1" }],
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

const createScheduleIssue = (
  overrides: Partial<CalendarScheduleIssue> = {},
): CalendarScheduleIssue => ({
  storyId: "story-1",
  storyTitle: "Prepare the launch brief",
  storyCode: "ENG-42",
  teamId: "team-1",
  teamName: "Engineering",
  teamCode: "ENG",
  estimatedDurationMinutes: 90,
  scheduledDurationMinutes: 60,
  remainingDurationMinutes: 30,
  autoSchedulingStatus: "cannot_fit",
  autoSchedulingReason: "No safe focus window remains before Friday.",
  updatedAt: "2026-08-08T09:55:00.000Z",
  ...overrides,
});

const createObjectiveRisk = (overrides: Partial<Objective> = {}): Objective =>
  ({
    endDate: "2026-08-21",
    forecastCauseStory: {
      id: "story-42",
      sequenceId: 42,
      source: "planning",
      title: "Prepare the launch brief",
    },
    forecastDaysDelta: 8,
    forecastEndDate: "2026-08-29",
    id: "objective-1",
    leadUser: "user-1",
    name: "Launch the new workspace experience",
    scheduleStatus: "at_risk",
    statusId: "status-1",
    teamId: "team-1",
    ...overrides,
  }) as Objective;

describe("UpcomingMeetingCard", () => {
  beforeEach(() => {
    jest.useFakeTimers();
    jest.setSystemTime(new Date("2026-08-08T10:00:00.000Z"));
    useCalendarSchedule.mockReset();
    retryScheduleIssue.mockReset();
    overrideScheduleIssue.mockReset();
    mockObjectives.length = 0;
    window.localStorage.clear();
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

    expect(screen.getByText("Weekly planning")).toHaveClass("line-clamp-1");
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
    const issue = createScheduleIssue();
    useCalendarSchedule.mockReturnValue({
      data: {
        ...createSchedule(),
        events: [],
        scheduleIssues: [
          issue,
          {
            ...issue,
            storyCode: "ENG-43",
            storyId: "story-2",
            storyTitle: "Prepare the launch checklist",
          },
        ],
      },
    });

    render(
      <UpcomingMeetingCard fallback={<div>Normal footer</div>} isCollapsed />,
    );

    const collapsedIndicator = screen.getByRole("button", {
      name: "Open Maya scheduling message",
    });
    expect(collapsedIndicator).toHaveClass("text-primary", "dark:text-primary");
    expect(screen.getByTestId("collapsed-warning-icon")).toHaveClass(
      "text-primary",
      "dark:text-primary",
    );
    expect(screen.getByText("Maya needs your help")).toBeInTheDocument();
    const storyLink = screen.getByText("Prepare the launch brief").closest("a");
    expect(storyLink).toHaveAttribute("data-next-link", "true");
    expect(storyLink).toHaveClass("line-clamp-1");
    expect(storyLink).toHaveAttribute("title", "Prepare the launch brief");
    expect(screen.queryByText("ENG-42")).not.toBeInTheDocument();
    const description = screen.getByText(
      "No safe focus window remains before Friday.",
    );
    expect(description).not.toHaveClass("line-clamp-2");
    expect(description).toHaveAttribute(
      "title",
      "30m left to schedule. No safe focus window remains before Friday.",
    );
    expect(screen.getByText("1 of 2").parentElement).toHaveClass(
      "border-border",
    );

    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(retryScheduleIssue).toHaveBeenCalledWith("story-1");

    fireEvent.click(screen.getByRole("button", { name: "Choose time" }));
    expect(
      screen.getByRole("heading", { name: "Choose a time for ENG-42" }),
    ).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Start"), {
      target: { value: "today at 8:07pm" },
    });
    expect(
      screen.getByText(/Starts Saturday, Aug 8 at 8:07 PM/),
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

  it("adds brief guidance when Maya only reports the remaining duration", () => {
    useCalendarSchedule.mockReturnValue({
      data: {
        ...createSchedule(),
        events: [],
        scheduleIssues: [
          createScheduleIssue({
            estimatedDurationMinutes: 60,
            scheduledDurationMinutes: 0,
            remainingDurationMinutes: 60,
            autoSchedulingReason: "1h left to schedule.",
          }),
        ],
      },
    });

    render(<UpcomingMeetingCard fallback={<div>Normal footer</div>} />);

    const description = screen.getByText(
      "1h remains. Choose a time or let Maya try again.",
    );
    expect(description).toHaveAttribute(
      "title",
      "1h remains. Choose a time or let Maya try again.",
    );
    expect(screen.queryByText("1h left to schedule.")).not.toBeInTheDocument();
  });

  it("shows forecast risk to the objective lead until that forecast is dismissed", () => {
    const objective = createObjectiveRisk();
    mockObjectives.push(objective);
    useCalendarSchedule.mockReturnValue({
      data: { ...createSchedule(), events: [] },
    });

    const { unmount } = render(
      <UpcomingMeetingCard fallback={<div>Normal footer</div>} />,
    );

    expect(screen.getByText("Forecast risk")).toBeInTheDocument();
    expect(screen.getByText("+8d")).toBeInTheDocument();
    const objectiveTitle = screen
      .getByText("Launch the new workspace experience")
      .closest("a");
    expect(objectiveTitle).toHaveClass("line-clamp-1");
    expect(objectiveTitle).toHaveAttribute(
      "href",
      "/acme/teams/team-1/objectives/objective-1?tab=overview",
    );
    expect(objectiveTitle).toHaveAttribute(
      "title",
      "Launch the new workspace experience",
    );
    const description = screen.getByText(
      /Linked work is forecast for Aug 29, 2026/,
    );
    expect(description).toHaveClass("line-clamp-3");
    expect(description).toHaveAttribute("title", description.textContent);
    const reviewLink = screen.getByRole("link", {
      name: "Review objective: Launch the new workspace experience",
    });
    expect(reviewLink).toHaveAttribute("data-next-link", "true");
    expect(reviewLink).toHaveClass("justify-center", "text-center");

    fireEvent.click(
      screen.getByRole("button", {
        name: "Dismiss objective forecast risk",
      }),
    );
    expect(screen.getByText("Normal footer")).toBeInTheDocument();

    unmount();
    const dismissedForecast = render(
      <UpcomingMeetingCard fallback={<div>Normal footer</div>} />,
    );
    expect(screen.queryByText("Forecast risk")).not.toBeInTheDocument();

    dismissedForecast.unmount();
    mockObjectives[0] = createObjectiveRisk({
      forecastDaysDelta: 10,
      forecastEndDate: "2026-08-31",
    });
    render(<UpcomingMeetingCard fallback={<div>Normal footer</div>} />);
    expect(screen.getByText("Forecast risk")).toBeInTheDocument();
    expect(screen.getByText("+10d")).toBeInTheDocument();
  });

  it("uses a pulsing pin instead of a two-digit collapsed count", () => {
    useCalendarSchedule.mockReturnValue({
      data: {
        ...createSchedule(),
        events: [],
        scheduleIssues: Array.from({ length: 10 }, (_, index) =>
          createScheduleIssue({
            storyCode: `ENG-${index + 1}`,
            storyId: `story-${index + 1}`,
            storyTitle: `Story ${index + 1}`,
          }),
        ),
      },
    });

    render(
      <UpcomingMeetingCard fallback={<div>Normal footer</div>} isCollapsed />,
    );

    expect(screen.queryByText("10")).not.toBeInTheDocument();
    expect(screen.getByTestId("collapsed-overflow-pin")).toHaveClass(
      "bg-primary",
    );
    expect(
      screen.getByTestId("collapsed-overflow-pin").firstElementChild,
    ).toHaveClass("animate-ping", "motion-reduce:animate-none");
  });
});

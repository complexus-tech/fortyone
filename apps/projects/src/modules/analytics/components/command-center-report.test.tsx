/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { CommandCenterReport } from "./command-center-report";

const mockUseCommandCenterReport = jest.fn();
const mockTrackEvent = jest.fn();

jest.mock("ui", () => {
  const Container = ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  );
  const Button = ({
    children,
    onClick,
  }: {
    children?: ReactNode;
    onClick?: () => void;
  }) => (
    <button onClick={onClick} type="button">
      {children}
    </button>
  );

  return {
    Badge: Container,
    Box: Container,
    Button,
    Flex: Container,
    Skeleton: () => <div data-testid="analytics-skeleton" />,
    Tabs: Object.assign(Container, {
      List: Container,
      Panel: Container,
      Tab: Button,
    }),
    Text: Container,
    Wrapper: Container,
  };
});

jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: () => ({
    getTermDisplay: (term: string, options?: { variant?: string }) =>
      options?.variant === "plural" ? `${term}s` : term,
  }),
}));

jest.mock("../hooks/filters", () => ({
  useAppliedFilters: () => ({ teamIds: ["team-1"] }),
}));

jest.mock("../hooks/command-center-report", () => ({
  useCommandCenterReport: (...args: unknown[]) =>
    mockUseCommandCenterReport(...args),
}));

jest.mock("../hooks/workspace-analytics-event", () => ({
  useWorkspaceAnalyticsEvent: () => ({ trackEvent: mockTrackEvent }),
}));

jest.mock("./filters", () => ({ Filters: () => <div>Filters</div> }));
jest.mock("./command-center-report/overview-tab", () => ({
  OverviewTab: () => <div data-testid="overview-tab" />,
}));
jest.mock("./command-center-report/workload-tab", () => ({
  WorkloadTab: () => <div data-testid="workload-tab" />,
}));
jest.mock("./command-center-report/flow-and-planning-tabs", () => ({
  FlowTab: () => <div data-testid="flow-tab" />,
  PlanningTab: () => <div data-testid="planning-tab" />,
}));
jest.mock("./command-center-report/engagement-tab", () => ({
  EngagementTab: () => <div data-testid="engagement-tab" />,
}));

const report = {
  engagement: { totalEvents: 19, uniqueUsers: 4 },
  overview: { metrics: { completedStories: 8, totalStories: 12 } },
  pulse: {
    risks: [{ kind: "overdue_stories" }],
    summary: { blockedStories: 1, overdueStories: 2 },
  },
  requests: {
    acceptedRequests: 6,
    declinedRequests: 1,
    pendingRequests: 3,
    totalRequests: 10,
  },
  sectionErrors: [{ section: "requests" }],
  workload: {
    risks: { overloadedMembers: [{ userId: "member-1" }] },
    summary: { totalEstimate: 21, totalOpenStories: 9 },
  },
};

describe("CommandCenterReport", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("renders the loading state while the report is pending", () => {
    mockUseCommandCenterReport.mockReturnValue({
      isError: false,
      isFetching: false,
      isPending: true,
      refetch: jest.fn(),
    });

    render(<CommandCenterReport />);

    expect(screen.getAllByTestId("analytics-skeleton")).toHaveLength(13);
  });

  it("renders an actionable error state and retries the report", () => {
    const refetch = jest.fn();
    mockUseCommandCenterReport.mockReturnValue({
      isError: true,
      isFetching: false,
      isPending: false,
      refetch,
    });

    render(<CommandCenterReport />);
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));

    expect(screen.getByText("Analytics are unavailable")).toBeInTheDocument();
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it("keeps the report shell, metrics, delayed-section warning, and tabs intact", () => {
    mockUseCommandCenterReport.mockReturnValue({
      data: report,
      isError: false,
      isFetching: true,
      isPending: false,
      refetch: jest.fn(),
    });

    render(<CommandCenterReport />);

    expect(screen.getByText("Some analytics sections are delayed"));
    expect(screen.getByText("Requests"));
    expect(screen.getByText("Refreshing"));
    expect(screen.getByText("9"));
    expect(screen.getByText("60%"));
    expect(screen.getByTestId("overview-tab")).toBeInTheDocument();
    expect(screen.getByTestId("workload-tab")).toBeInTheDocument();
    expect(screen.getByTestId("flow-tab")).toBeInTheDocument();
    expect(screen.getByTestId("planning-tab")).toBeInTheDocument();
    expect(screen.getByTestId("engagement-tab")).toBeInTheDocument();
    expect(mockTrackEvent).toHaveBeenCalledWith({
      eventName: "analytics_command_center_viewed",
      properties: { hasFilters: true },
      surface: "analytics_command_center",
    });
  });
});

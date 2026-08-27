/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ElementType, ReactNode } from "react";
import { render, screen } from "@testing-library/react";
import { AnalyticsReport } from "./analytics-report";

type PrimitiveProps = {
  children?: ReactNode;
  className?: string;
};

jest.mock("lib", () => ({
  cn: (...values: unknown[]) => values.filter(Boolean).join(" "),
}));

jest.mock("ui", () => ({
  Box: ({ children, className }: PrimitiveProps) => (
    <div className={className}>{children}</div>
  ),
  Button: ({ children, className }: PrimitiveProps) => (
    <button className={className} type="button">
      {children}
    </button>
  ),
  Flex: ({ children, className }: PrimitiveProps) => (
    <div className={className}>{children}</div>
  ),
  Text: ({
    as: Component = "span",
    children,
    className,
  }: PrimitiveProps & { as?: ElementType }) => (
    <Component className={className}>{children}</Component>
  ),
}));

jest.mock("@/hooks", () => ({
  useTerminology: () => ({
    getTermDisplay: (term: string) => (term === "storyTerm" ? "stories" : term),
  }),
}));

jest.mock("next-themes", () => ({
  useTheme: () => ({ resolvedTheme: "light" }),
}));

jest.mock("recharts", () => ({
  Bar: () => null,
  BarChart: ({
    children,
    data,
  }: {
    children?: ReactNode;
    data?: Record<string, unknown>[];
  }) => (
    <div>
      {data?.map((row) => (
        <span key={String(row.memberName)}>{String(row.memberName ?? "")}</span>
      ))}
      {children}
    </div>
  ),
  Line: () => null,
  LineChart: ({ children }: { children?: ReactNode }) => <div>{children}</div>,
  ResponsiveContainer: ({ children }: { children?: ReactNode }) => (
    <div>{children}</div>
  ),
  Tooltip: () => null,
  XAxis: () => null,
  YAxis: () => null,
}));

jest.mock("@/modules/sprints/stories/burndown", () => ({
  BurndownChart: () => null,
}));

const detailedCommandCenterOutput = {
  kind: "workspace-command-center-report",
  title: "Workspace analytics command center",
  report: {
    sectionErrors: [],
    overview: {
      metrics: {
        totalStories: 100,
        completedStories: 64,
      },
    },
    pulse: {
      summary: {
        overdueStories: 7,
        blockedStories: 3,
        overloadedMembers: 2,
        atRiskSprints: 1,
        atRiskObjectives: 1,
      },
      stories: {
        startedStories: 12,
        pausedStories: 4,
        unassignedStories: 5,
      },
      sprints: {
        activeSprints: 2,
        atRiskSprints: 1,
        overdueSprints: 0,
      },
      objectives: {
        activeObjectives: 3,
        atRiskObjectives: 1,
        offTrackObjectives: 1,
        overdueObjectives: 0,
      },
      risks: [
        {
          kind: "overdue_stories",
          severity: "high",
          title: "Overdue delivery",
          count: 7,
        },
      ],
    },
    workload: {
      summary: {
        totalOpenStories: 36,
        totalEstimate: 144,
        unassignedStories: 5,
        unestimatedStories: 8,
      },
      members: [
        {
          userId: "raw-member-identifier",
          fullName: "Ada Lovelace",
          username: "ada",
          openStories: 14,
          overdueStories: 2,
        },
      ],
    },
    sprints: {
      sprintProgress: [
        {
          sprintId: "raw-sprint-identifier",
          sprintName: "August launch",
          completed: 6,
          total: 10,
          status: "active",
        },
      ],
    },
    objectives: {
      keyResultsProgress: [
        {
          objectiveId: "raw-objective-identifier",
          objectiveName: "Improve activation",
          avgProgress: 72,
          completed: 2,
          total: 4,
        },
      ],
    },
    requests: {
      totalRequests: 18,
      pendingRequests: 4,
      providers: [
        {
          provider: "slack",
          totalRequests: 12,
          acceptanceRate: 0.75,
        },
      ],
    },
    engagement: {
      totalEvents: 240,
      uniqueUsers: 16,
      eventsByName: [{ name: "story_completed", count: 42 }],
    },
  },
};

const emptyCommandCenterOutput = {
  kind: "workspace-command-center-report",
  title: "Workspace analytics command center",
  report: {
    sectionErrors: [],
    overview: { metrics: {} },
    pulse: {
      summary: {},
      stories: {},
      sprints: {},
      objectives: {},
      risks: [],
    },
    workload: { summary: {}, members: [] },
    sprints: { sprintProgress: [] },
    objectives: { keyResultsProgress: [] },
    requests: { providers: [] },
    engagement: { eventsByName: [] },
  },
};

describe("AnalyticsReport workspace command center", () => {
  it("shows the complete workspace operating picture without exposing IDs", () => {
    const { container } = render(
      <AnalyticsReport output={detailedCommandCenterOutput} />,
    );

    expect(screen.getByText("Delivery health")).toBeInTheDocument();
    expect(screen.getByText("Workload")).toBeInTheDocument();
    expect(screen.getByText("Sprint health and progress")).toBeInTheDocument();
    expect(
      screen.getByText("Objective health and progress"),
    ).toBeInTheDocument();
    expect(screen.getAllByText("Active risks")).not.toHaveLength(0);
    expect(screen.getByText("Request sources")).toBeInTheDocument();
    expect(screen.getByText("Engagement")).toBeInTheDocument();
    expect(screen.getByText("August launch")).toBeInTheDocument();
    expect(screen.getByText("Improve activation")).toBeInTheDocument();
    expect(screen.getByText("Ada Lovelace")).toBeInTheDocument();
    expect(screen.getByText("Slack")).toBeInTheDocument();
    expect(screen.getByText("Story Completed")).toBeInTheDocument();
    expect(container).not.toHaveTextContent("raw-member-identifier");
    expect(container).not.toHaveTextContent("raw-sprint-identifier");
    expect(container).not.toHaveTextContent("raw-objective-identifier");
  });

  it("suppresses detailed sections that have no data", () => {
    render(<AnalyticsReport output={emptyCommandCenterOutput} />);

    expect(screen.getByText("Delivery health")).toBeInTheDocument();
    expect(screen.queryByText("Workload")).not.toBeInTheDocument();
    expect(
      screen.queryByText("Sprint health and progress"),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByText("Objective health and progress"),
    ).not.toBeInTheDocument();
    expect(screen.queryByText("Request sources")).not.toBeInTheDocument();
    expect(screen.queryByText("Engagement")).not.toBeInTheDocument();
  });

  it("names partial sections without exposing backend error details", () => {
    const output = {
      ...detailedCommandCenterOutput,
      report: {
        ...detailedCommandCenterOutput.report,
        sectionErrors: [
          {
            section: "objective_progress",
            message: "private backend failure detail",
          },
        ],
      },
    };

    const { container } = render(<AnalyticsReport output={output} />);

    expect(container).toHaveTextContent(
      "Some analytics are temporarily unavailable: Objective Progress.",
    );
    expect(container).not.toHaveTextContent("private backend failure detail");
  });
});

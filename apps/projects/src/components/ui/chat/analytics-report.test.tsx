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

jest.mock("@/hooks/use-terminology-display", () => ({
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
      {data?.map((row) => {
        const label = row.memberName ?? row.username;
        return <span key={String(label)}>{String(label ?? "")}</span>;
      })}
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

jest.mock("@/components/ui/burndown-chart", () => ({
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

  it("renders a single-sprint report with its owned report sections", () => {
    render(
      <AnalyticsReport
        output={{
          analytics: {
            burndown: [
              { date: "2026-08-01", ideal: 8, remaining: 10 },
              { date: "2026-08-02", ideal: 6, remaining: 7 },
            ],
            overview: { completionPercentage: 25, daysRemaining: 4 },
            storyBreakdown: {
              blocked: 1,
              completed: 2,
              inProgress: 3,
              total: 8,
            },
            teamAllocation: [{ assigned: 5, completed: 2, username: "Ada" }],
            workingDays: [1, 2, 3, 4, 5],
          },
          kind: "single-sprint-analytics-report",
          title: "August launch sprint",
        }}
      />,
    );

    expect(screen.getByText("August launch sprint")).toBeInTheDocument();
    expect(screen.getByText("Burndown")).toBeInTheDocument();
    expect(screen.getByText("Team allocation")).toBeInTheDocument();
    expect(screen.getByText("Ada")).toBeInTheDocument();
  });
});

const reportDispatchCases: {
  expected: string;
  name: string;
  output: Record<string, unknown>;
}[] = [
  {
    expected: "GitHub is not connected to this workspace.",
    name: "GitHub integration",
    output: {
      kind: "github-integration-report",
      summary: {},
    },
  },
  {
    expected: "Automation rules for Platform.",
    name: "GitHub team automation",
    output: {
      kind: "github-team-automation-report",
      rules: [],
      team: { name: "Platform" },
    },
  },
  {
    expected: "No GitHub links are attached to this stories.",
    name: "GitHub story links",
    output: {
      kind: "github-story-report",
      links: [],
      story: {},
    },
  },
  {
    expected: "Completion trend",
    name: "workspace performance",
    output: {
      kind: "workspace-performance-report",
      overview: { metrics: {} },
    },
  },
  {
    expected: "Highest workload",
    name: "pulse",
    output: {
      kind: "pulse-report",
      report: { risks: [], summary: {}, workload: { members: [] } },
    },
  },
  {
    expected: "Workload by member",
    name: "workload analysis",
    output: {
      analysis: { members: [], risks: {}, summary: {} },
      kind: "workload-analysis-report",
    },
  },
  {
    expected: "Status breakdown",
    name: "story performance",
    output: { analytics: {}, kind: "story-performance-report" },
  },
  {
    expected: "Key-result progress",
    name: "objective progress",
    output: { kind: "objective-progress-report", progress: {} },
  },
  {
    expected: "Team workload",
    name: "team performance",
    output: {
      focusMember: {},
      kind: "team-performance-report",
      performance: {},
    },
  },
  {
    expected: "Sprint progress",
    name: "sprint performance",
    output: { analytics: {}, kind: "sprint-performance-report" },
  },
  {
    expected: "Key metrics",
    name: "timeline trends",
    output: { kind: "timeline-trends-report", trends: {} },
  },
];

describe("AnalyticsReport report dispatch", () => {
  it.each(reportDispatchCases)(
    "renders the $name report",
    ({ expected, output }) => {
      render(<AnalyticsReport output={output} />);

      expect(screen.getByText(expected)).toBeInTheDocument();
    },
  );
});

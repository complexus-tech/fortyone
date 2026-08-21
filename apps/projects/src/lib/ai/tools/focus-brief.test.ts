/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { getPulseReport } from "@/modules/analytics/queries/get-pulse-report";
import type { PulseReport } from "@/modules/analytics/types";
import { getGroupedStories } from "@/modules/stories/queries/get-grouped-stories";
import type { GroupedStoriesResponse, Story } from "@/modules/stories/types";
import { focusBriefTool } from "./focus-brief";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/modules/analytics/queries/get-pulse-report", () => ({
  getPulseReport: jest.fn(),
}));

jest.mock("@/modules/stories/queries/get-grouped-stories", () => ({
  getGroupedStories: jest.fn(),
}));

const authMock = jest.mocked(auth);
const getPulseReportMock = jest.mocked(getPulseReport);
const getGroupedStoriesMock = jest.mocked(getGroupedStories);

const session = {
  user: {
    id: "user-1",
    name: "Joseph Mukorivo",
    email: "joseph@example.com",
    image: null,
    username: "joseph",
    fullName: "Joseph Mukorivo",
    isInternal: false,
    lastUsedWorkspaceId: "workspace-1",
  },
};

const toolOptions: ToolExecutionOptions = {
  toolCallId: "tool-call-1",
  messages: [],
  experimental_context: { workspaceSlug: "complexus" },
};

const executeTool = async <Input, Output>(
  execute: ToolExecuteFunction<Input, Output> | undefined,
  input: Input,
  options: ToolExecutionOptions = toolOptions,
): Promise<Output> => {
  if (!execute) throw new Error("Tool does not have an execute function");

  const result = execute(input, options);
  if (
    typeof result === "object" &&
    result !== null &&
    Symbol.asyncIterator in result
  ) {
    throw new Error("Streaming tool results are not supported by this test");
  }

  return (await result) as Output;
};

const story = (overrides: Partial<Story>): Story => ({
  id: "story-1",
  title: "Prepare the launch",
  estimateLabel: "M",
  estimateValue: 3,
  estimateScheme: "points",
  estimatedDurationMinutes: 120,
  minimumFocusBlockMinutes: null,
  autoSchedulingEnabled: false,
  autoSchedulingLocked: false,
  autoSchedulingStatus: "off",
  autoSchedulingReason: null,
  autoSchedulingUpdatedAt: null,
  statusId: "status-started",
  sprintId: null,
  sprint: null,
  objectiveId: null,
  objective: null,
  keyResultId: null,
  teamId: "team-1",
  team: { id: "team-1", name: "Product", code: "PRO" },
  workspaceId: "workspace-1",
  assigneeId: "user-1",
  assignee: {
    id: "user-1",
    username: "joseph",
    fullName: "Joseph Mukorivo",
    avatarUrl: null,
    isActive: true,
    isSystem: false,
  },
  collaboratorCount: 0,
  reporterId: "user-1",
  reporter: null,
  epicId: null,
  sequenceId: 41,
  priority: "Medium",
  startDate: null,
  endDate: null,
  createdAt: "2026-08-01T08:00:00.000Z",
  updatedAt: "2026-08-20T08:00:00.000Z",
  completedAt: null,
  deletedAt: null,
  archivedAt: null,
  labels: null,
  subStories: [],
  ...overrides,
});

const grouped = (stories: Story[]): GroupedStoriesResponse => ({
  groups: [
    {
      key: "none",
      loadedCount: stories.length,
      totalCount: stories.length,
      hasMore: false,
      nextPage: 0,
      stories,
    },
  ],
  meta: {
    totalGroups: 1,
    filters: {},
    groupBy: "none",
    orderBy: "priority",
    orderDirection: "asc",
  },
});

const pulse = {
  reportDate: "2026-08-21T08:00:00.000Z",
  risks: [
    {
      kind: "blocked_stories",
      severity: "high",
      count: 1,
      title: "Blocked work",
      description: "One active story is blocked.",
    },
  ],
  stories: {
    openStories: 7,
    startedStories: 2,
    pausedStories: 1,
    blockedStories: 1,
    overdueStories: 1,
    urgentStories: 1,
    highPriorityStories: 1,
    unestimatedStories: 2,
  },
  sprints: {
    atRiskSprints: 1,
    overdueSprints: 0,
  },
  objectives: {
    atRiskObjectives: 1,
    offTrackObjectives: 0,
    overdueObjectives: 0,
    objectivesDueSoon: 1,
  },
  requests: {
    pendingRequests: 2,
    urgentRequests: 1,
    highRequests: 0,
    staleRequests: 1,
  },
} as PulseReport;

describe("Maya focus brief", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    getPulseReportMock.mockResolvedValue(pulse);
  });

  it("builds a bounded current-user brief and ranks actionable stories", async () => {
    const blocked = story({ id: "blocked", priority: "Urgent" });
    const overdue = story({
      id: "overdue",
      sequenceId: 42,
      endDate: "2026-08-20T17:00:00.000Z",
    });
    const highPriority = story({
      id: "high",
      sequenceId: 43,
      priority: "High",
    });
    getGroupedStoriesMock
      .mockResolvedValueOnce(grouped([blocked]))
      .mockResolvedValueOnce(grouped([overdue]))
      .mockResolvedValueOnce(grouped([blocked, highPriority]));

    const result = (await executeTool(focusBriefTool.execute, {
      subject: { type: "current-user" as const },
    })) as Record<string, unknown> & {
      candidates: { id: string; signals: string[] }[];
    };

    expect(getPulseReportMock).toHaveBeenCalledWith(
      { session, workspaceSlug: "complexus" },
      {
        teamIds: undefined,
        assigneeIds: ["user-1"],
        sprintIds: undefined,
        objectiveIds: undefined,
      },
    );
    expect(getGroupedStoriesMock).toHaveBeenCalledTimes(3);
    expect(getGroupedStoriesMock.mock.calls[0]?.[1]).toMatchObject({
      assignedToMe: true,
      categories: ["backlog", "unstarted", "started", "paused"],
      hasBlockedBy: true,
      orderBy: "priority",
      orderDirection: "asc",
      storiesPerGroup: 5,
    });
    expect(getGroupedStoriesMock.mock.calls[1]?.[1]).toMatchObject({
      assignedToMe: true,
      orderBy: "deadline",
      orderDirection: "asc",
      storiesPerGroup: 10,
    });
    expect(getGroupedStoriesMock.mock.calls[2]?.[1]).toMatchObject({
      assignedToMe: true,
      orderBy: "priority",
      orderDirection: "asc",
      storiesPerGroup: 10,
    });
    expect(result.candidates.map((candidate) => candidate.id)).toEqual([
      "blocked",
      "overdue",
      "high",
    ]);
    expect(result.candidates[0]?.signals).toEqual(["blocked", "urgent"]);
    expect(result).toMatchObject({
      success: true,
      kind: "focus-brief-data",
      scope: { type: "current-user" },
      signals: {
        risks: [
          {
            kind: "blocked_stories",
            reason: "One active story is blocked.",
          },
        ],
        workload: { open: 7, blocked: 1, overdue: 1 },
      },
    });
    expect(result).not.toHaveProperty("report");
    expect(result).not.toHaveProperty("message");
  });

  it("applies resolved-member and entity filters without using assignedToMe", async () => {
    getGroupedStoriesMock.mockResolvedValue(grouped([]));

    await executeTool(focusBriefTool.execute, {
      subject: { type: "member" as const, memberId: "member-2" },
      teamIds: ["team-1", "team-1"],
      sprintIds: ["sprint-1"],
      objectiveId: "objective-1",
    });

    expect(getPulseReportMock).toHaveBeenCalledWith(
      { session, workspaceSlug: "complexus" },
      {
        teamIds: ["team-1"],
        assigneeIds: ["member-2"],
        sprintIds: ["sprint-1"],
        objectiveIds: ["objective-1"],
      },
    );
    for (const [, params] of getGroupedStoriesMock.mock.calls) {
      expect(params).toMatchObject({
        assigneeIds: ["member-2"],
        teamIds: ["team-1"],
        sprintIds: ["sprint-1"],
        objectiveId: "objective-1",
      });
      expect(params.assignedToMe).toBeUndefined();
    }
  });

  it("caps the focus candidates after deterministic ranking", async () => {
    const urgentStories = Array.from({ length: 7 }, (_, index) =>
      story({
        id: `urgent-${index + 1}`,
        sequenceId: 50 + index,
        priority: "Urgent",
      }),
    );
    getGroupedStoriesMock
      .mockResolvedValueOnce(grouped([]))
      .mockResolvedValueOnce(grouped([]))
      .mockResolvedValueOnce(grouped(urgentStories));

    const result = (await executeTool(focusBriefTool.execute, {
      subject: { type: "current-user" as const },
    })) as { candidates: { id: string }[] };

    expect(result.candidates.map((candidate) => candidate.id)).toEqual([
      "urgent-1",
      "urgent-2",
      "urgent-3",
      "urgent-4",
      "urgent-5",
    ]);
  });

  it("keeps workspace scope unassigned to any one member", async () => {
    getGroupedStoriesMock.mockResolvedValue(grouped([]));

    await executeTool(focusBriefTool.execute, {
      subject: { type: "workspace" as const },
    });

    expect(getPulseReportMock.mock.calls[0]?.[1]?.assigneeIds).toBeUndefined();
    for (const [, params] of getGroupedStoriesMock.mock.calls) {
      expect(params.assigneeIds).toBeUndefined();
      expect(params.assignedToMe).toBeUndefined();
    }
  });

  it("does not query project data without authentication", async () => {
    authMock.mockResolvedValue(null);

    const result = await executeTool(focusBriefTool.execute, {
      subject: { type: "current-user" as const },
    });

    expect(result).toEqual({
      success: false,
      error: "Authentication required to build a focus brief",
    });
    expect(getPulseReportMock).not.toHaveBeenCalled();
    expect(getGroupedStoriesMock).not.toHaveBeenCalled();
  });
});

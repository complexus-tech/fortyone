/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { getMembers } from "@/lib/queries/members/get-members";
import { getStatuses } from "@/lib/queries/states/get-states";
import { searchQuery } from "@/modules/search/queries/search";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { searchTool } from "./search";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/modules/search/queries/search", () => ({
  searchQuery: jest.fn(),
}));

jest.mock("@/lib/queries/members/get-members", () => ({
  getMembers: jest.fn(),
}));

jest.mock("@/lib/queries/states/get-states", () => ({
  getStatuses: jest.fn(),
}));

jest.mock("@/modules/teams/queries/get-teams", () => ({
  getTeams: jest.fn(),
}));

const authMock = jest.mocked(auth);
const getMembersMock = jest.mocked(getMembers);
const getStatusesMock = jest.mocked(getStatuses);
const getTeamsMock = jest.mocked(getTeams);
const searchQueryMock = jest.mocked(searchQuery);

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
  toolCallId: "search-call",
  messages: [],
  experimental_context: { workspaceSlug: "complexus" },
};

const executeTool = async <Input, Output>(
  execute: ToolExecuteFunction<Input, Output> | undefined,
  input: Input,
): Promise<Output> => {
  if (!execute) throw new Error("Tool does not have an execute function");

  const result = execute(input, toolOptions);
  if (
    typeof result === "object" &&
    result !== null &&
    Symbol.asyncIterator in result
  ) {
    throw new Error("Streaming tool results are not supported by this test");
  }

  return (await result) as Output;
};

const pageStory = (index: number) => ({
  id: `story-${index}`,
  title: `Story ${index}`,
  priority: "Medium" as const,
  statusId: "status-1",
  statusName: "In Progress",
  statusColor: "#eab308",
  statusCategory: "started" as const,
  assigneeId: "user-1",
  assigneeFullName: "Joseph Mukorivo",
  assigneeUsername: "joseph",
  teamId: "team-1",
  teamName: "Product",
  teamCode: "PROD",
  createdAt: "2026-08-27T08:00:00.000Z",
  updatedAt: "2026-08-27T08:00:00.000Z",
});

const pageObjective = (index: number) => ({
  id: `objective-${index}`,
  sequenceId: index,
  name: `Objective ${index}`,
  teamId: "team-1",
  teamName: "Product",
  teamCode: "PROD",
  leadUser: "user-1",
  leadFullName: "Joseph Mukorivo",
  leadUsername: "joseph",
  startDate: "2026-08-01",
  endDate: "2026-09-30",
  priority: "High" as const,
  health: "On Track" as const,
  createdAt: "2026-08-27T08:00:00.000Z",
  updatedAt: "2026-08-27T08:00:00.000Z",
});

describe("searchTool", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    searchQueryMock.mockResolvedValue({
      stories: Array.from({ length: 100 }, (_, index) => pageStory(index)),
      objectives: Array.from({ length: 100 }, (_, index) =>
        pageObjective(index),
      ),
      totalStories: 100,
      totalObjectives: 100,
      totalPages: 2,
      page: 1,
      pageSize: 100,
    } as never);
  });

  it("enriches a maximum-size current page with one bounded search request", async () => {
    const result = (await executeTool(searchTool.execute, {
      action: "search-all" as const,
      query: "launch",
      page: 1,
      pageSize: 100,
    })) as {
      stories: Record<string, unknown>[];
      objectives: Record<string, unknown>[];
    };

    expect(searchQueryMock).toHaveBeenCalledTimes(1);
    expect(getStatusesMock).not.toHaveBeenCalled();
    expect(getTeamsMock).not.toHaveBeenCalled();
    expect(getMembersMock).not.toHaveBeenCalled();
    expect(result.stories).toHaveLength(100);
    expect(result.objectives).toHaveLength(100);
    expect(result.stories[0]).toMatchObject({
      status: { name: "In Progress", category: "started" },
      team: { id: "team-1", name: "Product", code: "PROD" },
      assignee: { id: "user-1", username: "joseph" },
    });
    expect(result.objectives[0]).toMatchObject({
      team: { id: "team-1", name: "Product", code: "PROD" },
      leadUser: { id: "user-1", username: "joseph" },
    });
  });
});

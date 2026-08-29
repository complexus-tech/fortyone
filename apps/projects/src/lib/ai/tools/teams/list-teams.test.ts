/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { getPublicTeamsPage } from "@/modules/teams/queries/get-public-teams";
import { getJoinedTeamsPage } from "@/modules/teams/queries/get-teams";
import type { Team, TeamsPage } from "@/modules/teams/types";
import { listPublicTeams } from "./list-public-teams";
import { listTeams } from "./list-teams";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/modules/teams/queries/get-public-teams", () => ({
  getPublicTeamsPage: jest.fn(),
}));

jest.mock("@/modules/teams/queries/get-teams", () => ({
  getJoinedTeamsPage: jest.fn(),
}));

const authMock = jest.mocked(auth);
const getJoinedTeamsPageMock = jest.mocked(getJoinedTeamsPage);
const getPublicTeamsPageMock = jest.mocked(getPublicTeamsPage);

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

const team = (overrides: Partial<Team>): Team => ({
  id: "team-1",
  name: "Product",
  code: "PROD",
  color: "#6366F1",
  isPrivate: false,
  workspaceId: "workspace-1",
  createdAt: "2026-08-01T08:00:00.000Z",
  updatedAt: "2026-08-01T08:00:00.000Z",
  memberCount: 2,
  sprintsEnabled: true,
  ...overrides,
});

const page = (teams: Team[]): TeamsPage => ({
  teams,
  pagination: {
    page: 1,
    pageSize: 20,
    hasMore: false,
    nextPage: 0,
  },
});

describe("Maya team membership tools", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
  });

  it("lists only teams the current user has joined", async () => {
    const product = team({});
    getJoinedTeamsPageMock.mockResolvedValue(page([product]));

    const result = await executeTool(listTeams.execute, {
      searchQuery: "",
      page: 1,
      pageSize: 20,
    });

    expect(getJoinedTeamsPageMock).toHaveBeenCalledWith(
      { session, workspaceSlug: "complexus" },
      "",
      1,
      20,
    );
    expect(result).toMatchObject({
      success: true,
      teams: [{ id: "team-1", name: "Product" }],
    });
  });

  it("keeps an unjoined public team available for explicit discovery", async () => {
    const growth = team({
      id: "team-2",
      name: "Growth",
      code: "GROW",
    });
    getPublicTeamsPageMock.mockResolvedValue(page([growth]));

    const result = await executeTool(listPublicTeams.execute, {
      searchQuery: "Growth",
      page: 1,
      pageSize: 20,
    });

    expect(getPublicTeamsPageMock).toHaveBeenCalledWith(
      { session, workspaceSlug: "complexus" },
      "Growth",
      1,
      20,
    );
    expect(getJoinedTeamsPageMock).not.toHaveBeenCalled();
    expect(result).toMatchObject({
      success: true,
      teams: [{ id: "team-2", name: "Growth" }],
    });
  });
});

/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import {
  getMembersPage,
  getTeamMembersPage,
} from "@/lib/queries/members/get-members";
import type { Member, MembersPage } from "@/types";
import { resolveMemberTool } from "./members";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/lib/queries/members/get-members", () => ({
  getMembers: jest.fn(),
  getMembersPage: jest.fn(),
  getTeamMembersPage: jest.fn(),
}));

jest.mock("@/modules/teams/queries/get-teams", () => ({
  getTeams: jest.fn(),
}));

const authMock = jest.mocked(auth);
const getMembersPageMock = jest.mocked(getMembersPage);
const getTeamMembersPageMock = jest.mocked(getTeamMembersPage);

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

const member: Member = {
  id: "member-1",
  fullName: "Thomas Moyo",
  username: "thomas",
  role: "member",
  avatarUrl: "",
  email: "thomas@example.com",
  isActive: true,
  isSystem: false,
  isInternal: false,
  createdAt: "2026-08-01T08:00:00.000Z",
  updatedAt: "2026-08-01T08:00:00.000Z",
};

const page: MembersPage = {
  members: [member],
  pagination: {
    page: 1,
    pageSize: 10,
    hasMore: false,
    nextPage: 0,
  },
};

describe("Maya member resolution", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
  });

  it("returns a non-presentational exact member match", async () => {
    getMembersPageMock.mockResolvedValue(page);

    const result = await executeTool(resolveMemberTool.execute, {
      query: "@thomas",
    });

    expect(getMembersPageMock).toHaveBeenCalledWith(
      { session, workspaceSlug: "complexus" },
      "@thomas",
      1,
      10,
    );
    expect(result).toEqual({
      success: true,
      query: "@thomas",
      resolution: "resolved",
      match: {
        id: "member-1",
        name: "Thomas Moyo",
        username: "thomas",
        role: "member",
      },
      matches: [
        {
          id: "member-1",
          name: "Thomas Moyo",
          username: "thomas",
          role: "member",
        },
      ],
      hasMore: false,
    });
    expect(result).not.toHaveProperty("members");
  });

  it("can constrain resolution to a team", async () => {
    getTeamMembersPageMock.mockResolvedValue(page);

    await executeTool(resolveMemberTool.execute, {
      query: "Thomas",
      teamId: "team-1",
    });

    expect(getTeamMembersPageMock).toHaveBeenCalledWith(
      "team-1",
      { session, workspaceSlug: "complexus" },
      "Thomas",
      1,
      10,
    );
    expect(getMembersPageMock).not.toHaveBeenCalled();
  });

  it("does not query members without authentication", async () => {
    authMock.mockResolvedValue(null);

    const result = await executeTool(resolveMemberTool.execute, {
      query: "Thomas",
    });

    expect(result).toEqual({
      success: false,
      error: "Authentication required to resolve members",
    });
    expect(getMembersPageMock).not.toHaveBeenCalled();
  });
});

/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { joinPublicTeamAction } from "@/modules/teams/actions/join-public-team";
import { joinTeam } from "./join-team";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/modules/teams/actions/join-public-team", () => ({
  joinPublicTeamAction: jest.fn(),
}));

const authMock = jest.mocked(auth);
const joinPublicTeamActionMock = jest.mocked(joinPublicTeamAction);

const teamID = "00000000-0000-4000-8000-000000000001";
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
  toolCallId: "join-team-call",
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

describe("joinTeam", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    joinPublicTeamActionMock.mockResolvedValue({ data: { teamId: teamID } });
  });

  it("uses the dedicated self-join endpoint without forwarding a member UUID", async () => {
    await expect(
      executeTool(joinTeam.execute, { teamId: teamID }),
    ).resolves.toMatchObject({ success: true });

    expect(joinPublicTeamActionMock).toHaveBeenCalledWith(teamID, "complexus");
  });

  it("does not attempt to join a team without an authenticated actor", async () => {
    authMock.mockResolvedValue(null);

    await expect(
      executeTool(joinTeam.execute, { teamId: teamID }),
    ).resolves.toEqual({
      success: false,
      error: "Authentication required to join teams",
    });
    expect(joinPublicTeamActionMock).not.toHaveBeenCalled();
  });
});

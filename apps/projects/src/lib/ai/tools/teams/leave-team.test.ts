/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { leaveTeamAction } from "@/modules/teams/actions/leave-team";
import { leaveTeam } from "./leave-team";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/modules/teams/actions/leave-team", () => ({
  leaveTeamAction: jest.fn(),
}));

const authMock = jest.mocked(auth);
const leaveTeamActionMock = jest.mocked(leaveTeamAction);

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
  toolCallId: "leave-team-call",
  messages: [],
  experimental_context: { workspaceSlug: "complexus" },
};

const executeTool = async <Input, Output>(
  execute: ToolExecuteFunction<Input, Output> | undefined,
  input: Input,
): Promise<Output> => {
  if (!execute) throw new Error("Tool does not have an execute function");
  return (await execute(input, toolOptions)) as Output;
};

describe("leaveTeam", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    leaveTeamActionMock.mockResolvedValue({ data: null });
  });

  it("delegates to the actor-bound self-leave action", async () => {
    const result = await executeTool(leaveTeam.execute, { teamId: "team-1" });

    expect(leaveTeamActionMock).toHaveBeenCalledWith("team-1", "complexus");
    expect(result).toMatchObject({
      success: true,
      message: "Successfully left the team.",
    });
  });
});

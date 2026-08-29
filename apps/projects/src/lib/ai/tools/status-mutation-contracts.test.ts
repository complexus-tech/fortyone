/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { createStateAction } from "@/lib/actions/states/create";
import { updateStateAction } from "@/lib/actions/states/update";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { getTeams } from "@/modules/teams/queries/get-teams";
import { createObjectiveStatusAction } from "@/modules/objectives/actions/statuses/create";
import { objectiveStatusesTool } from "./objective-statuses";
import { statusesTool } from "./statuses";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/actions/states/create", () => ({
  createStateAction: jest.fn(),
}));
jest.mock("@/lib/actions/states/update", () => ({
  updateStateAction: jest.fn(),
}));
jest.mock("@/lib/actions/states/delete", () => ({
  deleteStateAction: jest.fn(),
}));
jest.mock("@/lib/queries/states/get-states", () => ({
  getStatuses: jest.fn(),
}));
jest.mock("@/lib/queries/states/get-team-states", () => ({
  getTeamStatuses: jest.fn(),
}));
jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
}));
jest.mock("@/modules/teams/queries/get-teams", () => ({
  getTeams: jest.fn(),
}));
jest.mock("@/modules/objectives/actions/statuses/create", () => ({
  createObjectiveStatusAction: jest.fn(),
}));
jest.mock("@/modules/objectives/actions/statuses/update", () => ({
  updateObjectiveStatusAction: jest.fn(),
}));
jest.mock("@/modules/objectives/actions/statuses/delete", () => ({
  deleteObjectiveStatusAction: jest.fn(),
}));
jest.mock("@/modules/objectives/queries/statuses", () => ({
  getObjectiveStatuses: jest.fn(),
}));

const authMock = jest.mocked(auth);
const createStateActionMock = jest.mocked(createStateAction);
const updateStateActionMock = jest.mocked(updateStateAction);
const getWorkspaceMock = jest.mocked(getWorkspace);
const getTeamsMock = jest.mocked(getTeams);
const createObjectiveStatusActionMock = jest.mocked(
  createObjectiveStatusAction,
);

const execute = async (
  definition: { execute?: (input: never, options: never) => unknown },
  input: Record<string, unknown>,
) => {
  if (!definition.execute) throw new Error("Tool has no execute function.");

  return definition.execute(
    input as never,
    {
      experimental_context: { workspaceSlug: "acme" },
      messages: [],
      toolCallId: "status-call",
    } as never,
  ) as Promise<Record<string, unknown>>;
};

describe("status mutation tool contracts", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue({ user: { id: "user-1" } } as never);
    getWorkspaceMock.mockResolvedValue({ userRole: "admin" } as never);
    getTeamsMock.mockResolvedValue([
      { id: "team-1", name: "Product" },
    ] as never);
    createStateActionMock.mockResolvedValue({
      data: { id: "status-1" },
    } as never);
    updateStateActionMock.mockResolvedValue({
      data: { id: "status-1" },
    } as never);
    createObjectiveStatusActionMock.mockResolvedValue({
      data: { id: "objective-status-1" },
    } as never);
  });

  it("preserves the approved default flag when creating statuses", async () => {
    await execute(statusesTool, {
      action: "create-status",
      category: "started",
      color: "#3366ff",
      isDefault: true,
      name: "In review",
      teamId: "team-1",
    });
    await execute(objectiveStatusesTool, {
      action: "create-objective-status",
      category: "started",
      color: "#3366ff",
      isDefault: true,
      name: "In review",
    });

    expect(createStateActionMock).toHaveBeenCalledWith(
      expect.objectContaining({ isDefault: true }),
      "acme",
    );
    expect(createObjectiveStatusActionMock).toHaveBeenCalledWith(
      expect.objectContaining({ isDefault: true }),
      "acme",
    );
  });

  it("updates color and rejects unsupported category changes explicitly", async () => {
    await execute(statusesTool, {
      action: "update-status",
      color: "#ff6600",
      statusId: "status-1",
    });
    const categoryResult = await execute(statusesTool, {
      action: "update-status",
      category: "completed",
      statusId: "status-1",
    });

    expect(updateStateActionMock).toHaveBeenCalledWith(
      "status-1",
      { color: "#ff6600" },
      "acme",
    );
    expect(categoryResult).toMatchObject({
      success: false,
      error: expect.stringContaining("cannot be changed"),
    });
    expect(updateStateActionMock).toHaveBeenCalledTimes(1);
  });

  it("rejects an update that contains no changed field", async () => {
    const result = await execute(statusesTool, {
      action: "update-status",
      statusId: "status-1",
    });

    expect(result).toMatchObject({
      success: false,
      error: expect.stringContaining("at least one"),
    });
    expect(updateStateActionMock).not.toHaveBeenCalled();
  });
});

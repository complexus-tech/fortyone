/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { post } from "@/lib/http";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import type { Workspace } from "@/types";
import { applyMayaWorkPlanTool, mayaWorkPlanTool } from "./maya";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/lib/http", () => ({
  post: jest.fn(),
}));

jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
}));

jest.mock("@/utils", () => ({
  getApiError: (error: unknown) =>
    error instanceof Error ? error.message : "Unknown API error",
}));

const authMock = jest.mocked(auth);
const getWorkspaceMock = jest.mocked(getWorkspace);
const postMock = jest.mocked(post);

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

const workspace = {
  id: "workspace-1",
  name: "Complexus",
  slug: "complexus",
  color: "#000000",
  avatarUrl: null,
  userRole: "member",
  trialEndsOn: null,
  deletedAt: null,
  isActive: true,
  createdAt: "2026-08-01T08:00:00.000Z",
  updatedAt: "2026-08-01T08:00:00.000Z",
} satisfies Workspace;

const workPlan = {
  actions: [],
  run: {
    id: "00000000-0000-4000-8000-000000000002",
    status: "completed",
    summary: "Prepared one scheduling action.",
  },
};

const toolOptions: ToolExecutionOptions = {
  toolCallId: "work-plan-call",
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

describe("Maya work-plan tools", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    getWorkspaceMock.mockResolvedValue(workspace);
    postMock.mockResolvedValue({ data: workPlan });
  });

  it("allows a workspace member to preview and apply their persisted plan", async () => {
    await expect(
      executeTool(mayaWorkPlanTool.execute, {
        storyId: "00000000-0000-4000-8000-000000000001",
      }),
    ).resolves.toMatchObject({ phase: "preview", success: true });

    await expect(
      executeTool(applyMayaWorkPlanTool.execute, {
        runId: "00000000-0000-4000-8000-000000000002",
      }),
    ).resolves.toMatchObject({ phase: "applied", success: true });

    expect(postMock).toHaveBeenNthCalledWith(
      1,
      "maya/work-plans",
      expect.objectContaining({ autoApply: false }),
      expect.objectContaining({ workspaceSlug: "complexus" }),
      { timeout: 30_000 },
    );
    expect(postMock).toHaveBeenNthCalledWith(
      2,
      "maya/work-plans/00000000-0000-4000-8000-000000000002/apply",
      {},
      expect.objectContaining({ workspaceSlug: "complexus" }),
      { timeout: 30_000 },
    );
  });

  it("keeps guests from previewing or applying work plans", async () => {
    getWorkspaceMock.mockResolvedValue({ ...workspace, userRole: "guest" });

    await expect(
      executeTool(mayaWorkPlanTool.execute, {
        storyId: "00000000-0000-4000-8000-000000000001",
      }),
    ).resolves.toEqual({
      error: "Only workspace admins and members can assign and schedule work.",
      success: false,
    });
    await expect(
      executeTool(applyMayaWorkPlanTool.execute, {
        runId: "00000000-0000-4000-8000-000000000002",
      }),
    ).resolves.toEqual({
      error: "Only workspace admins and members can assign and schedule work.",
      success: false,
    });
    expect(postMock).not.toHaveBeenCalled();
  });
});

/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { createGitHubInstallSessionAction } from "@/lib/actions/github/create-install-session";
import { resyncGitHubRepositoriesAction } from "@/lib/actions/github/resync-repositories";
import { updateGitHubWorkspaceSettingsAction } from "@/lib/actions/github/update-workspace-settings";
import { getGitHubIntegration } from "@/lib/queries/github/get-integration";
import {
  createGitHubInstallSessionTool,
  getGitHubIntegrationTool,
  resyncGitHubRepositoriesTool,
  updateGitHubWorkspaceSettingsTool,
} from "./github";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/lib/actions/github/create-install-session", () => ({
  createGitHubInstallSessionAction: jest.fn(),
}));
jest.mock("@/lib/actions/github/resync-repositories", () => ({
  resyncGitHubRepositoriesAction: jest.fn(),
}));
jest.mock("@/lib/actions/github/update-workspace-settings", () => ({
  updateGitHubWorkspaceSettingsAction: jest.fn(),
}));
jest.mock("@/lib/queries/github/get-integration", () => ({
  getGitHubIntegration: jest.fn(),
}));

const authMock = jest.mocked(auth);
const createInstallSessionMock = jest.mocked(createGitHubInstallSessionAction);
const getGitHubIntegrationMock = jest.mocked(getGitHubIntegration);
const resyncRepositoriesMock = jest.mocked(resyncGitHubRepositoriesAction);
const updateWorkspaceSettingsMock = jest.mocked(
  updateGitHubWorkspaceSettingsAction,
);

const session = {
  user: {
    email: "joseph@example.com",
    fullName: "Joseph Mukorivo",
    id: "user-1",
    image: null,
    isInternal: false,
    lastUsedWorkspaceId: "workspace-1",
    name: "Joseph Mukorivo",
    username: "joseph",
  },
};

const toolOptions: ToolExecutionOptions = {
  experimental_context: { workspaceSlug: "complexus" },
  messages: [],
  toolCallId: "tool-call-1",
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

describe("GitHub AI tools", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
  });

  it("does not query GitHub integration data without authentication", async () => {
    authMock.mockResolvedValue(null);

    const result = await executeTool(getGitHubIntegrationTool.execute, {});

    expect(result).toEqual({
      success: false,
      error: "Authentication required to access GitHub integration",
    });
    expect(getGitHubIntegrationMock).not.toHaveBeenCalled();
  });

  it("requires a workspace scope before querying GitHub data", async () => {
    const result = await executeTool(
      getGitHubIntegrationTool.execute,
      {},
      { ...toolOptions, experimental_context: {} },
    );

    expect(result).toEqual({
      success: false,
      error: "Workspace context is required to access GitHub integration",
    });
    expect(getGitHubIntegrationMock).not.toHaveBeenCalled();
  });

  it("returns the provider error when an authenticated integration query fails", async () => {
    getGitHubIntegrationMock.mockRejectedValue(
      new Error("GitHub is temporarily unavailable."),
    );

    const result = await executeTool(getGitHubIntegrationTool.execute, {});

    expect(result).toEqual({
      success: false,
      error: "GitHub is temporarily unavailable.",
    });
  });

  it("does not authenticate or mutate before an explicit resync confirmation", async () => {
    const result = await executeTool(resyncGitHubRepositoriesTool.execute, {
      confirmed: false,
    });

    expect(result).toEqual({
      success: false,
      needsConfirmation: true,
      message: "Please confirm before I resync GitHub repositories.",
    });
    expect(authMock).not.toHaveBeenCalled();
    expect(resyncRepositoriesMock).not.toHaveBeenCalled();
  });

  it("preserves API error messages from the install-session action", async () => {
    createInstallSessionMock.mockResolvedValue({
      error: { message: "GitHub installation was denied." },
    });

    const result = await executeTool(
      createGitHubInstallSessionTool.execute,
      {},
    );

    expect(createInstallSessionMock).toHaveBeenCalledWith("complexus");
    expect(result).toEqual({
      success: false,
      error: "GitHub installation was denied.",
    });
  });

  it("keeps false workspace settings when forwarding confirmed updates", async () => {
    updateWorkspaceSettingsMock.mockResolvedValue({
      data: {
        autoPopulatePrBody: true,
        branchFormat: "{story_ref}-{slug}",
        closeOnCommitKeywords: true,
        createdAt: "2026-08-01T00:00:00.000Z",
        linkCommitsByMagicWords: true,
        syncAssignees: true,
        syncLabels: false,
        updatedAt: "2026-08-01T00:00:00.000Z",
      },
    });

    const result = await executeTool(
      updateGitHubWorkspaceSettingsTool.execute,
      { confirmed: true, syncLabels: false },
    );

    expect(updateWorkspaceSettingsMock).toHaveBeenCalledWith(
      { syncLabels: false },
      "complexus",
    );
    expect(result).toMatchObject({
      success: true,
      message: "GitHub workspace settings updated.",
    });
  });
});

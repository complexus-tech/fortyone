/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { StoryDeletionOutcomeUncertainError } from "@/shared/story/deletion";
import { bulkDeleteAction } from "@/modules/stories/actions/bulk-delete-stories";
import type { Workspace } from "@/types";
import { bulkDeleteStories } from "./bulk-delete-stories";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
}));

jest.mock("@/modules/stories/actions/bulk-delete-stories", () => ({
  bulkDeleteAction: jest.fn(),
}));

const authMock = jest.mocked(auth);
const bulkDeleteActionMock = jest.mocked(bulkDeleteAction);
const getWorkspaceMock = jest.mocked(getWorkspace);

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

const workspace: Workspace = {
  id: "workspace-1",
  name: "Complexus",
  slug: "complexus",
  color: "#000000",
  avatarUrl: null,
  userRole: "admin",
  trialEndsOn: null,
  deletedAt: null,
  isActive: true,
  createdAt: "2026-08-01T08:00:00.000Z",
  updatedAt: "2026-08-01T08:00:00.000Z",
};

const storyIds = Array.from(
  { length: 50 },
  (_, index) =>
    `00000000-0000-4000-8000-${String(index + 1).padStart(12, "0")}`,
);
const storyTitles = storyIds.map((_, index) => `Story ${index + 1}`);

const toolOptions: ToolExecutionOptions = {
  toolCallId: "bulk-delete-call",
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

describe("bulkDeleteStories", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    getWorkspaceMock.mockResolvedValue(workspace);
  });

  it("deletes and returns all 50 requested story IDs", async () => {
    bulkDeleteActionMock.mockResolvedValue({
      data: { deletedCount: 50, storyIds },
    });

    const result = (await executeTool(bulkDeleteStories.execute, {
      confirmed: true,
      storyIds,
      storyTitles,
    })) as {
      deletedCount: number;
      missingStoryIds: string[];
      requestedCount: number;
      storyIds: string[];
      success: boolean;
    };

    expect(bulkDeleteActionMock).toHaveBeenCalledWith(
      { storyIds },
      "complexus",
    );
    expect(result).toEqual(
      expect.objectContaining({
        success: true,
        deletedCount: 50,
        requestedCount: 50,
        storyIds,
        missingStoryIds: [],
      }),
    );
  });

  it("reports exact missing IDs for a partial deletion", async () => {
    const deletedStoryIds = storyIds.slice(0, 47);
    bulkDeleteActionMock.mockResolvedValue({
      data: { deletedCount: 47, storyIds: deletedStoryIds },
    });

    const result = (await executeTool(bulkDeleteStories.execute, {
      confirmed: true,
      storyIds,
      storyTitles,
    })) as {
      deletedCount: number;
      missingStoryIds: string[];
      requestedCount: number;
      storyIds: string[];
      success: boolean;
    };

    expect(result).toMatchObject({
      success: false,
      deletedCount: 47,
      requestedCount: 50,
      storyIds: deletedStoryIds,
      missingStoryIds: storyIds.slice(47),
    });
  });

  it("rethrows an uncertain action outcome for mutation-ledger quarantine", async () => {
    const uncertainty = new StoryDeletionOutcomeUncertainError(
      new TypeError("response lost after commit"),
    );
    bulkDeleteActionMock.mockRejectedValue(uncertainty);

    await expect(
      executeTool(bulkDeleteStories.execute, {
        confirmed: true,
        storyIds,
        storyTitles,
      }),
    ).rejects.toBe(uncertainty);
  });
});

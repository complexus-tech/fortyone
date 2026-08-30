/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { deleteStoryAction } from "@/modules/story/actions/delete-story";
import { StoryDeletionOutcomeUncertainError } from "@/shared/story/deletion";
import { deleteStory } from "./delete-story";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/modules/story/actions/delete-story", () => ({
  deleteStoryAction: jest.fn(),
}));

const authMock = jest.mocked(auth);
const deleteStoryActionMock = jest.mocked(deleteStoryAction);

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

const story = {
  id: "00000000-0000-4000-8000-000000000001",
  title: "Prepare launch",
};

const toolOptions: ToolExecutionOptions = {
  toolCallId: "delete-story-call",
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

describe("deleteStory", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
  });

  it("sends an approved deletion directly to the authoritative idempotent endpoint", async () => {
    deleteStoryActionMock.mockResolvedValue({ data: null });

    await expect(
      executeTool(deleteStory.execute, {
        confirmed: true,
        storyId: story.id,
        storyTitle: story.title,
      }),
    ).resolves.toEqual({
      success: true,
      message: `Story "${story.title}" deleted successfully.`,
    });
    expect(deleteStoryActionMock).toHaveBeenCalledWith(story.id, "complexus");
  });

  it("returns a definite action failure", async () => {
    deleteStoryActionMock.mockResolvedValue({
      data: null,
      error: { message: "You cannot delete this story." },
    });

    await expect(
      executeTool(deleteStory.execute, {
        confirmed: true,
        storyId: story.id,
        storyTitle: story.title,
      }),
    ).resolves.toEqual({
      success: false,
      error: "You cannot delete this story.",
    });
  });

  it("rethrows an uncertain action outcome for mutation-ledger quarantine", async () => {
    const uncertainty = new StoryDeletionOutcomeUncertainError(
      new TypeError("response lost after commit"),
    );
    deleteStoryActionMock.mockRejectedValue(uncertainty);

    await expect(
      executeTool(deleteStory.execute, {
        confirmed: true,
        storyId: story.id,
        storyTitle: story.title,
      }),
    ).rejects.toBe(uncertainty);
  });
});

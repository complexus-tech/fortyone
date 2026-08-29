/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { StoryCreationOutcomeUncertainError } from "@/modules/story/actions/story-creation-error";
import type { Workspace } from "@/types";
import { createStory } from "./create-story";
import { createStoryInputSchema } from "./story-creation-schema";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
}));

jest.mock("@/modules/story/actions/create-story", () => ({
  createStoryAction: jest.fn(),
}));

jest.mock("./resolve-story-status", () => ({
  createStoryStatusResolver: () =>
    jest.fn(async () => "00000000-0000-4000-8000-000000000003"),
}));

jest.mock("./resolve-sprint-end-date", () => ({
  createSprintEndDateResolver: () =>
    jest.fn(
      async (_sprintId: string | undefined, requestedEndDate?: string | null) =>
        Promise.resolve(requestedEndDate),
    ),
}));

const authMock = jest.mocked(auth);
const createStoryActionMock = jest.mocked(createStoryAction);
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
const workspace = {
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
} satisfies Workspace;
const toolOptions: ToolExecutionOptions = {
  toolCallId: "create-call",
  messages: [],
  experimental_context: {
    chatId: "chat-123",
    workspaceSlug: "complexus",
  },
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

const input = createStoryInputSchema.parse({
  teamId: "00000000-0000-4000-8000-000000000001",
  title: "Prepare launch",
});

describe("createStory", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    getWorkspaceMock.mockResolvedValue(workspace);
  });

  it("returns a definite API rejection as an ordinary tool failure", async () => {
    createStoryActionMock.mockResolvedValue({
      data: null,
      error: { message: "The selected status is invalid." },
    });

    await expect(executeTool(createStory.execute, input)).resolves.toEqual({
      error: "The selected status is invalid.",
      success: false,
    });
  });

  it("propagates response loss so the approval ledger can quarantine the mutation", async () => {
    const uncertainty = new StoryCreationOutcomeUncertainError(
      new TypeError("response lost after commit"),
    );
    createStoryActionMock.mockRejectedValue(uncertainty);

    await expect(executeTool(createStory.execute, input)).rejects.toBe(
      uncertainty,
    );
    expect(createStoryActionMock).toHaveBeenCalledWith(
      expect.objectContaining({
        idempotencyKey: "maya:chat-123:create-call",
        title: "Prepare launch",
      }),
      "complexus",
    );
  });
});

/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { deleteMemoryAction } from "@/modules/ai-chats/actions/delete-memory";
import { updateMemoryAction } from "@/modules/ai-chats/actions/update-memory";
import { deleteMemory, updateMemory } from "./memory";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/modules/ai-chats/actions/create-memory", () => ({
  createMemoryAction: jest.fn(),
}));

jest.mock("@/modules/ai-chats/actions/delete-memory", () => ({
  deleteMemoryAction: jest.fn(),
}));

jest.mock("@/modules/ai-chats/actions/update-memory", () => ({
  updateMemoryAction: jest.fn(),
}));

jest.mock("@/modules/ai-chats/queries/get-memory", () => ({
  getMemories: jest.fn(),
}));

jest.mock("@/lib/queries/subscriptions/get-subscription", () => ({
  getSubscription: jest.fn(),
}));

jest.mock("@/lib/hooks/subscription-features", () => ({
  TIER_LIMITS: { free: { maxMemories: 10 } },
}));

const authMock = jest.mocked(auth);
const deleteMemoryActionMock = jest.mocked(deleteMemoryAction);
const updateMemoryActionMock = jest.mocked(updateMemoryAction);

const memoryID = "00000000-0000-4000-8000-000000000001";
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
  toolCallId: "memory-call",
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

describe("Maya memory mutation tools", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    updateMemoryActionMock.mockResolvedValue({ data: null });
    deleteMemoryActionMock.mockResolvedValue({ data: null });
  });

  it("updates and deletes only through the authenticated workspace actions", async () => {
    await expect(
      executeTool(updateMemory.execute, {
        id: memoryID,
        content: "Prefer concise weekly summaries.",
      }),
    ).resolves.toMatchObject({ success: true });
    await expect(
      executeTool(deleteMemory.execute, { id: memoryID }),
    ).resolves.toMatchObject({ success: true });

    expect(updateMemoryActionMock).toHaveBeenCalledWith(
      memoryID,
      { content: "Prefer concise weekly summaries." },
      "complexus",
    );
    expect(deleteMemoryActionMock).toHaveBeenCalledWith(memoryID, "complexus");
  });

  it("rejects mutation execution when authentication is missing", async () => {
    authMock.mockResolvedValue(null);

    await expect(
      executeTool(updateMemory.execute, {
        id: memoryID,
        content: "Do not save this.",
      }),
    ).resolves.toEqual({
      success: false,
      error: "Authentication required",
    });
    await expect(
      executeTool(deleteMemory.execute, { id: memoryID }),
    ).resolves.toEqual({
      success: false,
      error: "Authentication required",
    });

    expect(updateMemoryActionMock).not.toHaveBeenCalled();
    expect(deleteMemoryActionMock).not.toHaveBeenCalled();
  });
});

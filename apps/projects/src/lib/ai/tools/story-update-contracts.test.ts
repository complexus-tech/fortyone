/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { updateStoryAction } from "@/modules/story/actions/update-story";
import { bulkUpdateAction } from "@/modules/stories/actions/bulk-update-stories";
import { bulkUpdateStories } from "./stories/bulk-update-stories";
import { updateStory } from "./stories/update-story";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
}));
jest.mock("@/modules/story/actions/update-story", () => ({
  updateStoryAction: jest.fn(),
}));
jest.mock("@/modules/stories/actions/bulk-update-stories", () => ({
  bulkUpdateAction: jest.fn(),
}));

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
      toolCallId: "story-update-call",
    } as never,
  );
};

const parseToolInput = (definition: unknown, input: Record<string, unknown>) =>
  (
    definition as {
      inputSchema: { safeParse: (value: unknown) => { success: boolean } };
    }
  ).inputSchema.safeParse(input);

describe("story update tool contracts", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(auth).mockResolvedValue({ user: { id: "user-1" } } as never);
    jest.mocked(getWorkspace).mockResolvedValue({ userRole: "admin" } as never);
    jest.mocked(updateStoryAction).mockResolvedValue({} as never);
    jest.mocked(bulkUpdateAction).mockResolvedValue({
      data: {
        totalCount: 1,
        succeededCount: 1,
        failedCount: 0,
        partial: false,
        items: [{ storyId: "story-1", success: true }],
      },
    } as never);
  });

  it("does not advertise label replacement through story patch tools", () => {
    expect(
      parseToolInput(updateStory, {
        storyId: "story-1",
        labelIds: ["label-1"],
      }).success,
    ).toBe(false);
    expect(
      parseToolInput(bulkUpdateStories, {
        storyIds: ["18227a68-8e35-4aad-a8ee-fb6b1ef9feee"],
        updateData: { labelIds: ["label-1"] },
      }).success,
    ).toBe(false);
  });

  it("never forwards unsupported labelIds to the story patch endpoints", async () => {
    await execute(updateStory, {
      storyId: "story-1",
      confirmed: true,
      title: "Updated title",
      labelIds: ["label-1"],
    });
    await execute(bulkUpdateStories, {
      storyIds: ["story-1"],
      confirmed: true,
      updateData: { priority: "High", labelIds: ["label-1"] },
    });

    const singleUpdates = jest.mocked(updateStoryAction).mock.calls[0]?.[1];
    expect(singleUpdates).toMatchObject({ title: "Updated title" });
    expect(singleUpdates).not.toHaveProperty("labelIds");

    const bulkRequest = jest.mocked(bulkUpdateAction).mock.calls[0]?.[0];
    expect(bulkRequest.updates).toMatchObject({ priority: "High" });
    expect(bulkRequest.updates).not.toHaveProperty("labelIds");
  });
});

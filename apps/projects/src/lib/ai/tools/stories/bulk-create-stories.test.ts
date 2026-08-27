/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ToolExecutionOptions, ToolExecuteFunction } from "ai";
import { auth } from "@/auth";
import { getWorkspace } from "@/lib/queries/workspaces/get-workspace";
import { createStoryAction } from "@/modules/story/actions/create-story";
import { StoryCreationOutcomeUncertainError } from "@/modules/story/actions/story-creation-error";
import type { DetailedStory, NewStory } from "@/modules/story/types";
import type { Workspace } from "@/types";
import { bulkCreateStories } from "./bulk-create-stories";
import { bulkCreateStoriesInputSchema } from "./story-creation-schema";

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
    jest.fn(async (_teamId: string, requestedStatusId?: string | null) =>
      Promise.resolve(
        requestedStatusId ?? "00000000-0000-4000-8000-000000000003",
      ),
    ),
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

const toolOptions: ToolExecutionOptions = {
  toolCallId: "bulk-create-call",
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

const createdStory = (input: NewStory, index: number): DetailedStory => ({
  id: `00000000-0000-4000-8000-${String(index + 1).padStart(12, "0")}`,
  sequenceId: index + 1,
  title: input.title ?? `Story ${index + 1}`,
  estimateLabel: null,
  estimateValue: input.estimateValue ?? null,
  estimateScheme: "points",
  estimatedDurationMinutes: input.estimatedDurationMinutes ?? null,
  minimumFocusBlockMinutes: input.minimumFocusBlockMinutes ?? null,
  autoSchedulingEnabled: input.autoSchedulingEnabled ?? false,
  autoSchedulingLocked: false,
  autoSchedulingStatus: input.autoSchedulingEnabled ? "planning" : "off",
  autoSchedulingReason: null,
  autoSchedulingUpdatedAt: "2026-08-27T08:00:00.000Z",
  description: input.description ?? "",
  descriptionHTML: input.descriptionHTML ?? "",
  parentId: input.parentId ?? "",
  teamId: input.teamId ?? "",
  teamCode: "PROD",
  workspaceId: "workspace-1",
  objectiveId: input.objectiveId ?? null,
  keyResultId: input.keyResultId ?? null,
  statusId: input.statusId ?? "00000000-0000-4000-8000-000000000003",
  assigneeId: input.assigneeId ?? null,
  collaboratorIds: [],
  collaborators: [],
  collaboratorCount: 0,
  watcherCount: 0,
  watchers: [],
  isWatching: false,
  watchingReason: null,
  reporterId: "user-1",
  priority: input.priority ?? "No Priority",
  sprintId: input.sprintId ?? null,
  epicId: null,
  startDate: input.startDate ?? null,
  endDate: input.endDate ?? null,
  createdAt: "2026-08-27T08:00:00.000Z",
  updatedAt: "2026-08-27T08:00:00.000Z",
  deletedAt: null,
  completedAt: null,
  archivedAt: null,
  subStories: [],
  labels: input.labelIds ?? [],
  associations: [],
});

describe("bulkCreateStories", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
    getWorkspaceMock.mockResolvedValue(workspace);
    let creationIndex = 0;
    createStoryActionMock.mockImplementation(async (input) => {
      const story = createdStory(input, creationIndex);
      creationIndex += 1;
      return { data: story };
    });
  });

  it("creates exactly 50 stories with shared planning values and complete receipts", async () => {
    const parsedInput = bulkCreateStoriesInputSchema.safeParse({
      sharedValues: {
        teamId: "00000000-0000-4000-8000-000000000001",
        assigneeId: "00000000-0000-4000-8000-000000000002",
        priority: "Medium",
        estimatedDurationMinutes: 90,
        minimumFocusBlockMinutes: 30,
        autoSchedulingEnabled: true,
        endDate: "2026-09-30",
      },
      storiesData: Array.from({ length: 50 }, (_, index) => ({
        title: `Launch task ${index + 1}`,
        description: `Deliver launch task ${index + 1}.`,
        descriptionHTML: `<p>Deliver launch task ${index + 1}.</p>`,
      })),
    });

    expect(parsedInput.success).toBe(true);
    if (!parsedInput.success) throw parsedInput.error;

    const result = (await executeTool(
      bulkCreateStories.execute,
      parsedInput.data,
    )) as {
      calendarImpact: string;
      createdCount: number;
      errorCount: number;
      stories: { id: string; title: string }[];
      success: boolean;
    };

    expect(createStoryActionMock).toHaveBeenCalledTimes(50);
    createStoryActionMock.mock.calls.forEach(
      ([input, workspaceSlug], index) => {
        expect(input).toMatchObject({
          assigneeId: "00000000-0000-4000-8000-000000000002",
          autoSchedulingEnabled: true,
          endDate: "2026-09-30",
          estimatedDurationMinutes: 90,
          idempotencyKey: `maya:chat-123:bulk-create-call:${index}`,
          minimumFocusBlockMinutes: 30,
          priority: "Medium",
          teamId: "00000000-0000-4000-8000-000000000001",
          title: `Launch task ${index + 1}`,
        });
        expect(workspaceSlug).toBe("complexus");
      },
    );
    expect(result).toMatchObject({
      success: true,
      createdCount: 50,
      errorCount: 0,
    });
    expect(result.stories).toHaveLength(50);
    expect(result.stories.map(({ title }) => title)).toEqual(
      Array.from({ length: 50 }, (_, index) => `Launch task ${index + 1}`),
    );
    expect(result.calendarImpact).toContain("all 50 stories");
  });

  it("creates a 50-story batch unscheduled when planning details are omitted", async () => {
    const parsedInput = bulkCreateStoriesInputSchema.parse({
      sharedValues: {
        assigneeId: "00000000-0000-4000-8000-000000000002",
        teamId: "00000000-0000-4000-8000-000000000001",
      },
      storiesData: Array.from({ length: 50 }, (_, index) => ({
        title: `Manual planning task ${index + 1}`,
      })),
    });

    const result = (await executeTool(
      bulkCreateStories.execute,
      parsedInput,
    )) as {
      calendarImpact: string;
      createdCount: number;
      message: string;
      success: boolean;
    };

    expect(createStoryActionMock).toHaveBeenCalledTimes(50);
    for (const [input] of createStoryActionMock.mock.calls) {
      expect(input.estimatedDurationMinutes).toBeUndefined();
      expect(input.autoSchedulingEnabled).toBe(false);
    }
    expect(result).toMatchObject({
      createdCount: 50,
      success: true,
    });
    expect(result.calendarImpact).toBe(
      "Calendar scheduling is off for all 50 stories. Add each story's time needed and delivery details manually, or provide those details for Maya to schedule selected stories.",
    );
    expect(result.message).toContain(result.calendarImpact);
  });

  it("reports mixed calendar impact for an explicitly scheduled subset", async () => {
    const parsedInput = bulkCreateStoriesInputSchema.parse({
      sharedValues: {
        assigneeId: "00000000-0000-4000-8000-000000000002",
        teamId: "00000000-0000-4000-8000-000000000001",
      },
      storiesData: [
        {
          autoSchedulingEnabled: true,
          endDate: "2026-09-04",
          estimatedDurationMinutes: 30,
          title: "Scheduled task",
        },
        { title: "Manual task" },
      ],
    });

    const result = (await executeTool(
      bulkCreateStories.execute,
      parsedInput,
    )) as { calendarImpact: string; success: boolean };

    expect(createStoryActionMock).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({ autoSchedulingEnabled: true }),
      "complexus",
    );
    expect(createStoryActionMock).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({ autoSchedulingEnabled: false }),
      "complexus",
    );
    expect(result).toEqual(
      expect.objectContaining({
        calendarImpact:
          "Calendar scheduling is on for 1 story and off for 1 story. Maya is planning focus time for 1 story; no reservation is confirmed yet.",
        success: true,
      }),
    );
  });

  it("does not imply calendar work when every story fails", async () => {
    createStoryActionMock.mockResolvedValue({
      error: { message: "Story creation failed." },
    });
    const parsedInput = bulkCreateStoriesInputSchema.parse({
      sharedValues: {
        teamId: "00000000-0000-4000-8000-000000000001",
      },
      storiesData: [{ title: "A failed story" }],
    });

    const result = (await executeTool(
      bulkCreateStories.execute,
      parsedInput,
    )) as { calendarImpact: string; createdCount: number; success: boolean };

    expect(result).toMatchObject({
      success: false,
      createdCount: 0,
      calendarImpact:
        "No calendar time was reserved because no stories were created.",
    });
  });

  it("propagates one lost response from a mixed batch instead of completing with a misleading receipt", async () => {
    const uncertainty = new StoryCreationOutcomeUncertainError(
      new Error("Response failed after commit"),
    );
    createStoryActionMock
      .mockImplementationOnce(async (story) => ({
        data: createdStory(story, 0),
      }))
      .mockRejectedValueOnce(uncertainty)
      .mockImplementationOnce(async (story) => ({
        data: createdStory(story, 2),
      }));
    const parsedInput = bulkCreateStoriesInputSchema.parse({
      sharedValues: {
        teamId: "00000000-0000-4000-8000-000000000001",
      },
      storiesData: [
        { title: "Committed with a response" },
        { title: "Committed but response lost" },
        { title: "Another completed request" },
      ],
    });

    await expect(
      executeTool(bulkCreateStories.execute, parsedInput),
    ).rejects.toBe(uncertainty);
    expect(createStoryActionMock).toHaveBeenCalledTimes(3);
  });
});

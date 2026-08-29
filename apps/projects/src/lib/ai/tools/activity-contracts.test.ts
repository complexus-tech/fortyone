/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { getActivities } from "@/lib/queries/activities/get-activities";
import { getObjectiveActivities } from "@/modules/objectives/queries/get-objective-activities";
import { getKeyResultActivities } from "@/modules/objectives/queries/get-key-result-activities";
import { getStoryActivities } from "@/modules/story/queries/get-activities";
import { activitySummaryTool } from "./activity-summary";
import { getKeyResultActivitiesTool } from "./key-results/get-key-result-activities";
import { getObjectiveActivitiesTool } from "./objectives/get-objective-activities";
import { storyActivitiesTool } from "./story-activities";

jest.mock("ai", () => ({
  tool: (definition: unknown) => definition,
}));

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/queries/activities/get-activities", () => ({
  getActivities: jest.fn(),
}));
jest.mock("@/modules/story/queries/get-activities", () => ({
  getStoryActivities: jest.fn(),
}));
jest.mock("@/modules/objectives/queries/get-objective-activities", () => ({
  getObjectiveActivities: jest.fn(),
}));
jest.mock("@/modules/objectives/queries/get-key-result-activities", () => ({
  getKeyResultActivities: jest.fn(),
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
      toolCallId: "activity-call",
    } as never,
  ) as Promise<Record<string, unknown>>;
};

const activity = (index: number) => ({
  id: `activity-${index}`,
  storyId: "story-1",
  userId: "user-1",
  type: "update",
  field: "status",
  currentValue: "Started",
  createdAt: new Date(Date.UTC(2026, 7, 27, 12, index)).toISOString(),
});

describe("activity tool pagination contracts", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(auth).mockResolvedValue({ user: { id: "user-1" } } as never);
  });

  it("probes and reports truncation for the workspace activity window", async () => {
    jest
      .mocked(getActivities)
      .mockResolvedValue(
        Array.from({ length: 101 }, (_, index) => activity(index)) as never,
      );

    const result = await execute(activitySummaryTool, {});

    expect(getActivities).toHaveBeenCalledWith(
      expect.objectContaining({ workspaceSlug: "acme" }),
      expect.objectContaining({ limit: 101 }),
    );
    expect(result).toMatchObject({
      success: true,
      count: 20,
      source: {
        limit: 100,
        returnedCount: 100,
        truncated: true,
        completeForRequestedWindow: false,
      },
      pagination: {
        page: 1,
        pageSize: 20,
        totalCount: 100,
        hasMore: true,
        totalsScope: "loaded-source-window",
        completeForRequestedWindow: false,
      },
    });
    expect(result.message).toContain("older activity may exist");
  });

  it("uses backend story pagination instead of repaginating the first page", async () => {
    jest.mocked(getStoryActivities).mockResolvedValue({
      activities: [activity(1), activity(2)],
      pagination: { page: 2, pageSize: 5, hasMore: true, nextPage: 3 },
    } as never);

    const result = await execute(storyActivitiesTool, {
      action: "list-activities",
      storyId: "story-1",
      page: 2,
      pageSize: 5,
    });

    expect(getStoryActivities).toHaveBeenCalledWith(
      "story-1",
      expect.objectContaining({ workspaceSlug: "acme" }),
      2,
      5,
    );
    expect(result).toMatchObject({
      success: true,
      count: 2,
      pagination: {
        page: 2,
        pageSize: 5,
        returnedCount: 2,
        matchingCount: 2,
        hasMore: true,
        nextPage: 3,
        completeHistory: false,
        filtersAppliedTo: "current-page",
      },
    });
  });

  it.each([
    ["objective", getObjectiveActivitiesTool, getObjectiveActivities],
    ["key result", getKeyResultActivitiesTool, getKeyResultActivities],
  ] as const)(
    "reports hasMore for a paginated %s timeline",
    async (_, toolDefinition, query) => {
      jest.mocked(query).mockResolvedValue({
        activities: [activity(1)],
        pagination: { page: 1, pageSize: 1, hasMore: true },
      } as never);

      const input =
        toolDefinition === getObjectiveActivitiesTool
          ? { objectiveId: "objective-1", page: 1, pageSize: 1 }
          : { keyResultId: "key-result-1", page: 1, pageSize: 1 };
      const result = await execute(toolDefinition, input);

      expect(result).toMatchObject({
        success: true,
        pagination: {
          page: 1,
          pageSize: 1,
          returnedCount: 1,
          hasMore: true,
          nextPage: 2,
          completeHistory: false,
        },
      });
    },
  );
});

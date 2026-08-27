/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getStoryCreationIdempotencyKey } from "./story-creation-idempotency";
import {
  applyBulkStorySharedValues,
  bulkCreateStoriesInputSchema,
  createStoryInputSchema,
} from "./story-creation-schema";

const teamId = "00000000-0000-4000-8000-000000000001";
const assigneeId = "00000000-0000-4000-8000-000000000002";

describe("story creation planning contract", () => {
  it("requires complete planning inputs when calendar scheduling is on", () => {
    const result = createStoryInputSchema.safeParse({
      title: "Prepare launch",
      teamId,
      autoSchedulingEnabled: true,
    });

    expect(result.success).toBe(false);
    if (result.success) return;

    expect(result.error.issues.map(({ path }) => path.join("."))).toEqual([
      "assigneeId",
      "estimatedDurationMinutes",
      "endDate",
    ]);
  });

  it("accepts a sprint as the delivery-date source", () => {
    expect(
      createStoryInputSchema.safeParse({
        title: "Prepare launch",
        teamId,
        assigneeId,
        autoSchedulingEnabled: true,
        estimatedDurationMinutes: 120,
        sprintId: "00000000-0000-4000-8000-000000000003",
      }).success,
    ).toBe(true);
  });

  it("applies explicit shared values while preserving per-story overrides", () => {
    expect(
      applyBulkStorySharedValues(
        {
          teamId,
          assigneeId,
          priority: "Medium",
          estimatedDurationMinutes: 60,
          autoSchedulingEnabled: true,
          endDate: "2026-09-30",
        },
        {
          title: "Urgent launch fix",
          priority: "Urgent",
          estimatedDurationMinutes: 30,
        },
      ),
    ).toMatchObject({
      title: "Urgent launch fix",
      teamId,
      assigneeId,
      priority: "Urgent",
      estimatedDurationMinutes: 30,
      autoSchedulingEnabled: true,
      endDate: "2026-09-30",
    });
  });

  it("treats strict-schema null placeholders as shared-value inheritance", () => {
    expect(
      applyBulkStorySharedValues(
        {
          assigneeId,
          autoSchedulingEnabled: false,
          endDate: "2026-09-04",
          estimatedDurationMinutes: 60,
          teamId,
        },
        {
          assigneeId: null,
          autoSchedulingEnabled: false,
          endDate: null,
          estimatedDurationMinutes: null,
          teamId: null,
          title: "Generated strict-schema item",
        },
      ),
    ).toMatchObject({
      assigneeId,
      autoSchedulingEnabled: false,
      endDate: "2026-09-04",
      estimatedDurationMinutes: 60,
      teamId,
      title: "Generated strict-schema item",
    });
  });

  it("validates effective shared scheduling values for every story", () => {
    const result = bulkCreateStoriesInputSchema.safeParse({
      sharedValues: {
        teamId,
        autoSchedulingEnabled: true,
      },
      storiesData: [{ title: "First" }, { title: "Second" }],
    });

    expect(result.success).toBe(false);
    if (result.success) return;

    expect(result.error.issues.map(({ path }) => path.join("."))).toEqual([
      "storiesData.0.assigneeId",
      "storiesData.0.estimatedDurationMinutes",
      "storiesData.0.endDate",
      "storiesData.1.assigneeId",
      "storiesData.1.estimatedDurationMinutes",
      "storiesData.1.endDate",
    ]);
  });

  it("accepts a 50-story batch without a shared duration or calendar scheduling", () => {
    const storiesData = Array.from({ length: 50 }, (_, index) => ({
      title: `Unscheduled task ${index + 1}`,
    }));
    const result = bulkCreateStoriesInputSchema.safeParse({
      sharedValues: { assigneeId, teamId },
      storiesData,
    });

    expect(result.success).toBe(true);
    if (!result.success) throw result.error;

    for (const story of result.data.storiesData) {
      const resolvedStory = applyBulkStorySharedValues(
        result.data.sharedValues,
        story,
      );
      expect(resolvedStory.estimatedDurationMinutes).toBeUndefined();
      expect(resolvedStory.autoSchedulingEnabled).toBe(false);
    }
  });

  it("keeps supplied bulk dates and durations unscheduled without calendar consent", () => {
    const result = bulkCreateStoriesInputSchema.parse({
      sharedValues: {
        assigneeId,
        endDate: "2026-09-04",
        estimatedDurationMinutes: 30,
        teamId,
      },
      storiesData: [{ title: "First" }, { title: "Second" }],
    });

    for (const story of result.storiesData) {
      expect(
        applyBulkStorySharedValues(result.sharedValues, story),
      ).toMatchObject({
        autoSchedulingEnabled: false,
        endDate: "2026-09-04",
        estimatedDurationMinutes: 30,
      });
    }
  });

  it("supports an explicitly selected scheduled subset in a mixed batch", () => {
    const result = bulkCreateStoriesInputSchema.parse({
      sharedValues: { assigneeId, teamId },
      storiesData: [
        {
          autoSchedulingEnabled: true,
          endDate: "2026-09-04",
          estimatedDurationMinutes: 30,
          title: "Schedule this one",
        },
        { title: "Leave this one for manual planning" },
      ],
    });

    expect(
      applyBulkStorySharedValues(result.sharedValues, result.storiesData[0]),
    ).toMatchObject({
      autoSchedulingEnabled: true,
      estimatedDurationMinutes: 30,
    });
    expect(
      applyBulkStorySharedValues(result.sharedValues, result.storiesData[1]),
    ).toMatchObject({
      autoSchedulingEnabled: false,
    });
  });

  it("accepts strict provider placeholders across a full 50-story batch", () => {
    const strictStory = (index: number) => ({
      assigneeId: null,
      autoSchedulingEnabled: false,
      description: null,
      descriptionHTML: null,
      endDate: null,
      estimatedDurationMinutes: null,
      estimateValue: null,
      keyResultId: null,
      labelIds: null,
      minimumFocusBlockMinutes: null,
      objectiveId: null,
      parentId: null,
      priority: "No Priority" as const,
      sprintId: null,
      startDate: null,
      statusId: null,
      teamId: null,
      title: `Generated task ${index + 1}`,
    });

    expect(
      bulkCreateStoriesInputSchema.safeParse({
        sharedValues: {
          assigneeId,
          autoSchedulingEnabled: false,
          endDate: "2026-09-04",
          estimatedDurationMinutes: 60,
          estimateValue: null,
          keyResultId: null,
          labelIds: null,
          minimumFocusBlockMinutes: null,
          objectiveId: null,
          parentId: null,
          priority: "No Priority",
          sprintId: null,
          startDate: null,
          statusId: null,
          teamId,
        },
        storiesData: Array.from({ length: 50 }, (_, index) =>
          strictStory(index),
        ),
      }).success,
    ).toBe(true);
  });
});

describe("story creation idempotency keys", () => {
  it("binds a create to its chat and tool call", () => {
    expect(
      getStoryCreationIdempotencyKey({
        context: { chatId: "chat-123" },
        toolCallId: "call-456",
      }),
    ).toBe("maya:chat-123:call-456");
  });

  it("falls back to the tool call and bounds long keys", () => {
    expect(
      getStoryCreationIdempotencyKey({
        context: { workspaceSlug: "complexus" },
        toolCallId: "call-456",
      }),
    ).toBe("maya:call-456");

    const longKey = getStoryCreationIdempotencyKey({
      context: { chatId: "c".repeat(200) },
      toolCallId: "t".repeat(200),
      index: 49,
    });
    expect(longKey).toHaveLength(128);
    expect(longKey).toMatch(/:[0-9a-f]{16}$/);
  });
});

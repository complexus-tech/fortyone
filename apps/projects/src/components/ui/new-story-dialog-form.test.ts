/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { DetailedStory, NewStory } from "@/modules/story/types";
import {
  buildNewStoryDialogPayload,
  createInitialNewStoryDialogForm,
  getDeadlineForSprintSelection,
  getInitialDeadlineSource,
  runStoryCreatedFollowUp,
  storyFormReducer,
  toDateOnly,
} from "./new-story-dialog-form";

describe("new story dialog form", () => {
  it("creates a complete draft while preserving an explicit assignee choice", () => {
    expect(
      createInitialNewStoryDialogForm({
        assigneeId: "assignee-1",
        autoAssignedUserId: "current-user",
        autoSchedulingDefaultEnabled: true,
        currentTeamId: "team-1",
        endDate: "2026-06-20",
        objectiveId: "objective-1",
        priority: "High",
        sprintId: "sprint-1",
        statusId: "status-1",
      }),
    ).toMatchObject({
      assigneeId: "assignee-1",
      autoSchedulingEnabled: true,
      autoSchedulingLocked: false,
      endDate: "2026-06-20",
      estimatedDurationMinutes: 60,
      labelIds: [],
      objectiveId: "objective-1",
      priority: "High",
      sprintId: "sprint-1",
      statusId: "status-1",
      teamId: "team-1",
    });
  });

  it("falls back to the auto-assigned user only when no assignee was provided", () => {
    expect(
      createInitialNewStoryDialogForm({
        autoAssignedUserId: "current-user",
        autoSchedulingDefaultEnabled: false,
        endDate: null,
        priority: "No Priority",
      }).assigneeId,
    ).toBe("current-user");
  });

  it("updates individual fields without discarding the rest of the draft", () => {
    expect(
      storyFormReducer(
        { statusId: "status-1", title: "Preserve me" },
        { type: "SET_FIELD", field: "priority", value: "High" },
      ),
    ).toEqual({
      priority: "High",
      statusId: "status-1",
      title: "Preserve me",
    });
  });

  it("synchronizes team and status atomically", () => {
    expect(
      storyFormReducer(
        { priority: "Low", statusId: "old-status", teamId: "old-team" },
        {
          type: "SYNC_TEAM_STATUS",
          statusId: "new-status",
          teamId: "new-team",
        },
      ),
    ).toEqual({
      priority: "Low",
      statusId: "new-status",
      teamId: "new-team",
    });
  });

  it("removes the time component from sprint deadlines", () => {
    expect(toDateOnly("2026-06-20T23:59:59.000Z")).toBe("2026-06-20");
    expect(toDateOnly("2026-06-20")).toBe("2026-06-20");
    expect(toDateOnly(null)).toBeNull();
  });

  it("inherits the sprint end date when no deadline has been provided", () => {
    expect(
      getDeadlineForSprintSelection({
        currentEndDate: null,
        currentSource: getInitialDeadlineSource({ sprintId: null }),
        sprintEndDate: "2026-06-20",
      }),
    ).toEqual({
      endDate: "2026-06-20",
      source: "sprint",
    });
  });

  it("preserves a manually selected or cleared deadline when the sprint changes", () => {
    expect(
      getDeadlineForSprintSelection({
        currentEndDate: "2026-06-18",
        currentSource: "manual",
        sprintEndDate: "2026-06-20",
      }),
    ).toEqual({
      endDate: "2026-06-18",
      source: "manual",
    });

    expect(
      getDeadlineForSprintSelection({
        currentEndDate: null,
        currentSource: "cleared",
        sprintEndDate: "2026-06-20",
      }),
    ).toEqual({
      endDate: null,
      source: "cleared",
    });
  });

  it("preserves selected labels, complexity, and time in the create payload", () => {
    const storyForm: NewStory = {
      assigneeId: "user-1",
      autoSchedulingEnabled: true,
      endDate: "2026-06-20",
      estimateValue: 5,
      estimatedDurationMinutes: 120,
      minimumFocusBlockMinutes: 30,
      labelIds: ["label-1", "label-2"],
      objectiveId: "objective-1",
      keyResultId: "key-result-1",
      priority: "High",
      sprintId: "sprint-1",
      startDate: "2026-06-13",
      statusId: "status-1",
    };

    expect(
      buildNewStoryDialogPayload({
        currentTeamId: "team-1",
        description: "Plain text description",
        descriptionHTML: "<p>Plain text description</p>",
        storyForm,
        title: "Add reporting filters",
      }),
    ).toMatchObject({
      assigneeId: "user-1",
      autoSchedulingEnabled: true,
      description: "Plain text description",
      descriptionHTML: "<p>Plain text description</p>",
      estimateValue: 5,
      estimatedDurationMinutes: 120,
      minimumFocusBlockMinutes: 30,
      labelIds: ["label-1", "label-2"],
      objectiveId: "objective-1",
      keyResultId: "key-result-1",
      priority: "High",
      sprintId: "sprint-1",
      statusId: "status-1",
      teamId: "team-1",
      title: "Add reporting filters",
    });
  });

  it("serializes a timestamped deadline as an ISO date", () => {
    const payload = buildNewStoryDialogPayload({
      currentTeamId: "team-1",
      description: "",
      descriptionHTML: "",
      storyForm: {
        endDate: "2026-06-20T23:59:59.000Z",
      },
      title: "Use the sprint deadline",
    });

    expect(payload.endDate).toBe("2026-06-20");
  });

  it("defaults new stories to auto-scheduling when no choice is provided", () => {
    const payload = buildNewStoryDialogPayload({
      currentTeamId: "team-1",
      description: "",
      descriptionHTML: "",
      storyForm: {
        assigneeId: "user-1",
        estimatedDurationMinutes: 60,
      },
      title: "Keep the existing workflow",
    });

    expect(payload).toMatchObject({
      autoSchedulingEnabled: true,
    });
    expect(payload).not.toHaveProperty("autoSchedulingLocked");
  });

  it("always enables auto-scheduling when Maya is the assignee", () => {
    const payload = buildNewStoryDialogPayload({
      currentTeamId: "team-1",
      description: "",
      descriptionHTML: "",
      mayaAssigneeId: "maya-1",
      storyForm: {
        assigneeId: "maya-1",
        autoSchedulingEnabled: false,
      },
      title: "Let Maya schedule this work",
    });

    expect(payload.autoSchedulingEnabled).toBe(true);
  });

  it("rejects a focus block longer than the serialized duration", () => {
    const payload = buildNewStoryDialogPayload({
      currentTeamId: "team-1",
      description: "",
      descriptionHTML: "",
      storyForm: {
        estimatedDurationMinutes: 30,
        minimumFocusBlockMinutes: 60,
      },
      title: "Keep scheduling inputs valid",
    });

    expect(payload).toMatchObject({
      estimatedDurationMinutes: 30,
      minimumFocusBlockMinutes: null,
    });
  });

  it("does not send estimate scheme because the API derives it from team settings", () => {
    const payload = buildNewStoryDialogPayload({
      currentTeamId: "team-1",
      description: "",
      descriptionHTML: "",
      storyForm: {
        estimateValue: 2,
        priority: "Medium",
        statusId: "status-1",
      },
      title: "Add estimate input",
    });

    expect(payload).not.toHaveProperty("estimateScheme");
  });

  it("reports a follow-up failure without rejecting the committed creation", async () => {
    const story = { id: "story-1" } as DetailedStory;
    const callbackError = new Error("Could not create association");
    const onCreated = jest.fn().mockRejectedValue(callbackError);

    await expect(runStoryCreatedFollowUp(story, onCreated)).resolves.toBe(
      callbackError,
    );
    expect(onCreated).toHaveBeenCalledWith(story);
  });

  it("returns no follow-up error when the callback succeeds", async () => {
    const story = { id: "story-1" } as DetailedStory;
    const onCreated = jest.fn().mockResolvedValue(undefined);

    await expect(runStoryCreatedFollowUp(story, onCreated)).resolves.toBeNull();
  });
});

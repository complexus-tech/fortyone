/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getGitHubInstallSessionUrl,
  getMutationApproval,
  getMutationMessage,
  isRenderableToolPart,
  isSupportingToolType,
  type ToolMessagePart,
} from "./tool-output-policy";

describe("tool output policy", () => {
  it("keeps a signed GitHub install URL available only to the UI policy", () => {
    const installUrl =
      "https://github.com/apps/fortyone/installations/new?state=signed-session";

    expect(getGitHubInstallSessionUrl({ success: true, installUrl })).toBe(
      installUrl,
    );
    expect(
      getGitHubInstallSessionUrl({
        success: true,
        installUrl: "https://example.com/steal-session",
      }),
    ).toBeUndefined();
    expect(
      getGitHubInstallSessionUrl({
        success: true,
        installUrl: "https://github.com/apps/fortyone/installations/new",
      }),
    ).toBeUndefined();
    expect(
      getGitHubInstallSessionUrl({ success: false, installUrl }),
    ).toBeUndefined();
  });

  it("renders completed Maya work plans", () => {
    const part = {
      output: {
        kind: "maya-work-plan",
        plan: {
          actions: [],
          run: { id: "run-1", status: "applied", summary: "Scheduled." },
        },
        success: true,
      },
      state: "output-available",
      type: "tool-mayaWorkPlanTool",
    } as unknown as ToolMessagePart;

    expect(isRenderableToolPart(part)).toBe(true);
  });

  it("does not render an incomplete Maya work-plan tool part", () => {
    const part = {
      output: { kind: "maya-work-plan" },
      state: "input-streaming",
      type: "tool-mayaWorkPlanTool",
    } as unknown as ToolMessagePart;

    expect(isRenderableToolPart(part)).toBe(false);
  });

  it("renders the applied persisted Maya work plan", () => {
    const part = {
      output: {
        kind: "maya-work-plan",
        phase: "applied",
        plan: {
          actions: [{ id: "action-1", status: "applied" }],
          run: { id: "run-1", status: "succeeded", summary: "Scheduled." },
        },
        success: true,
      },
      state: "output-available",
      type: "tool-applyMayaWorkPlanTool",
    } as unknown as ToolMessagePart;

    expect(isRenderableToolPart(part)).toBe(true);
  });

  it.each([
    ["tool-resolveMember", { kind: "member-resolution", matches: [] }],
    ["tool-focusBrief", { kind: "focus-brief-data", candidates: [] }],
  ])("keeps %s supporting data out of generative UI", (type, output) => {
    const part = {
      input: {},
      output: { success: true, ...output },
      state: "output-available",
      type,
    } as unknown as ToolMessagePart;

    expect(isSupportingToolType(type)).toBe(true);
    expect(isRenderableToolPart(part)).toBe(false);
  });

  it("renders mutation results directly without another model response", () => {
    const part = {
      output: { message: "Team updated successfully.", success: true },
      state: "output-available",
      type: "tool-updateTeam",
    } as unknown as ToolMessagePart;

    expect(getMutationMessage(part.output)).toBe("Team updated successfully.");
    expect(isRenderableToolPart(part)).toBe(true);
  });

  it("renders a bounded approval model for prepared story creation", () => {
    const part = {
      approval: { id: "approval-1" },
      input: {
        storiesData: [
          { title: "Add onboarding checklist" },
          { title: "Track activation milestone" },
        ],
      },
      state: "approval-requested",
      toolCallId: "call-1",
      type: "tool-bulkCreateStories",
    } as unknown as ToolMessagePart;

    expect(getMutationApproval(part)).toEqual({
      approved: undefined,
      cancelledMessage: "Creation cancelled.",
      count: 2,
      description: "Maya will create the prepared stories exactly as shown.",
      details: [
        {
          label: "Story 1",
          value:
            "Add onboarding checklist — delivery: not specified · time needed: not specified · calendar scheduling: disabled (not specified)",
        },
        {
          label: "Story 2",
          value:
            "Track activation milestone — delivery: not specified · time needed: not specified · calendar scheduling: disabled (not specified)",
        },
      ],
      id: "approval-1",
      isStoryCreation: true,
      prompt: "Create 2 stories?",
      storyPreviews: [
        {
          id: "story-approval-1",
          priority: "No Priority",
          statusId: undefined,
          summary:
            "Add onboarding checklist — delivery: not specified · time needed: not specified · calendar scheduling: disabled (not specified)",
          title: "Add onboarding checklist",
        },
        {
          id: "story-approval-2",
          priority: "No Priority",
          statusId: undefined,
          summary:
            "Track activation milestone — delivery: not specified · time needed: not specified · calendar scheduling: disabled (not specified)",
          title: "Track activation milestone",
        },
      ],
      title: undefined,
    });
    expect(isRenderableToolPart(part)).toBe(true);
  });

  it("discloses delivery, effort, and calendar impact in story approval", () => {
    const part = {
      approval: { id: "approval-1" },
      input: {
        assigneeId: "user-1",
        autoSchedulingEnabled: true,
        description: "Ship the onboarding checklist.",
        endDate: "2026-09-04",
        estimatedDurationMinutes: 90,
        keyResultId: null,
        labelIds: null,
        parentId: null,
        startDate: null,
        teamId: "team-1",
        title: "Add onboarding checklist",
      },
      state: "approval-requested",
      toolCallId: "call-1",
      type: "tool-createStory",
    } as unknown as ToolMessagePart;

    expect(getMutationApproval(part)?.details).toEqual([
      {
        label: "Story 1",
        value:
          "Add onboarding checklist — description: Ship the onboarding checklist. · delivery: 2026-09-04 · time needed: 90 minutes · calendar scheduling: enabled",
      },
    ]);
  });

  it("applies shared planning values to every story shown for approval", () => {
    const part = {
      approval: { id: "approval-1" },
      input: {
        sharedValues: {
          autoSchedulingEnabled: false,
          endDate: "2026-09-04",
          estimatedDurationMinutes: 60,
          teamId: "team-1",
        },
        storiesData: [
          {
            autoSchedulingEnabled: null,
            endDate: null,
            estimatedDurationMinutes: null,
            teamId: null,
            title: "First task",
          },
        ],
      },
      state: "approval-requested",
      toolCallId: "call-1",
      type: "tool-bulkCreateStories",
    } as unknown as ToolMessagePart;

    expect(getMutationApproval(part)?.details).toEqual([
      {
        label: "Story 1",
        value:
          "First task — delivery: 2026-09-04 · time needed: 60 minutes · calendar scheduling: disabled",
      },
    ]);
    expect(getMutationApproval(part)?.storyPreviews).toEqual([
      expect.objectContaining({
        priority: "No Priority",
        title: "First task",
      }),
    ]);
  });

  it("keeps a denied story approval visible as a cancellation", () => {
    const part = {
      approval: { approved: false, id: "approval-1" },
      input: { title: "Add onboarding checklist" },
      state: "output-denied",
      toolCallId: "call-1",
      type: "tool-createStory",
    } as unknown as ToolMessagePart;

    expect(getMutationApproval(part)?.approved).toBe(false);
    expect(isRenderableToolPart(part)).toBe(true);
  });

  it("renders a concise generic approval for a prepared mutation", () => {
    const part = {
      approval: { id: "approval-1" },
      input: { storyId: "story-1", title: "Updated title" },
      state: "approval-requested",
      toolCallId: "call-1",
      type: "tool-updateStory",
    } as unknown as ToolMessagePart;

    expect(getMutationApproval(part)).toEqual(
      expect.objectContaining({
        description:
          "Maya will apply the prepared change exactly as requested.",
        details: [
          { label: "Story ID", value: "story-1" },
          { label: "Title", value: "Updated title" },
        ],
        isStoryCreation: false,
        prompt: "Update this story?",
      }),
    );
    expect(isRenderableToolPart(part)).toBe(true);
  });

  it("shows every destructive bulk target and the fields being changed", () => {
    const part = {
      approval: { id: "approval-1" },
      input: {
        storyIds: ["story-1", "story-2"],
        updateData: {
          autoSchedulingEnabled: true,
          endDate: "2026-09-04",
          estimatedDurationMinutes: 90,
        },
      },
      state: "approval-requested",
      toolCallId: "call-1",
      type: "tool-bulkUpdateStories",
    } as unknown as ToolMessagePart;

    expect(getMutationApproval(part)?.details).toEqual([
      { label: "Story 1", value: "Selected story 1" },
      { label: "Story 2", value: "Selected story 2" },
      {
        label: "Update Data · Auto Scheduling Enabled",
        value: "Enabled",
      },
      { label: "Update Data · End Date", value: "2026-09-04" },
      { label: "Update Data · Estimated Duration Minutes", value: "90" },
    ]);
  });

  it("shows human-readable titles for destructive bulk targets", () => {
    const part = {
      approval: { id: "approval-1" },
      input: {
        storyIds: ["story-1", "story-2"],
        storyTitles: ["Launch checklist", "Activation review"],
      },
      state: "approval-requested",
      toolCallId: "call-1",
      type: "tool-bulkDeleteStories",
    } as unknown as ToolMessagePart;

    expect(getMutationApproval(part)?.details).toEqual([
      { label: "Story 1", value: "Launch checklist" },
      { label: "Story 2", value: "Activation review" },
    ]);
  });

  it("shows the verified title instead of the internal ID for one deletion", () => {
    const part = {
      approval: { id: "approval-1" },
      input: {
        storyId: "6acd864b-5c13-4a04-8b64-ecbbfbbe5ea2",
        storyTitle: "MAYA-SCHEDULE-20260827-01",
      },
      state: "approval-requested",
      toolCallId: "call-1",
      type: "tool-deleteStory",
    } as unknown as ToolMessagePart;

    const approval = getMutationApproval(part);

    expect(approval?.prompt).toBe("Delete “MAYA-SCHEDULE-20260827-01”?");
    expect(approval?.details).toEqual([
      { label: "Story", value: "MAYA-SCHEDULE-20260827-01" },
    ]);
    expect(JSON.stringify(approval)).not.toContain(
      "6acd864b-5c13-4a04-8b64-ecbbfbbe5ea2",
    );
  });

  it("requests approval for mutating shared-tool actions only", () => {
    const mutationPart = {
      approval: { id: "approval-1" },
      input: { action: "create-label" },
      state: "approval-requested",
      toolCallId: "call-1",
      type: "tool-labels",
    } as unknown as ToolMessagePart;
    const readPart = {
      approval: { id: "approval-2" },
      input: { action: "list-labels" },
      state: "approval-requested",
      toolCallId: "call-2",
      type: "tool-labels",
    } as unknown as ToolMessagePart;

    expect(getMutationApproval(mutationPart)?.prompt).toBe("Create label?");
    expect(getMutationApproval(readPart)).toBeUndefined();
  });

  it("distinguishes mutating and read actions on a shared tool", () => {
    const mutationPart = {
      input: { action: "create-label" },
      output: { label: { name: "Customer" }, success: true },
      state: "output-available",
      type: "tool-labels",
    } as unknown as ToolMessagePart;
    const readPart = {
      input: { action: "list-labels" },
      output: { labels: [], success: true },
      state: "output-available",
      type: "tool-labels",
    } as unknown as ToolMessagePart;

    expect(getMutationMessage(mutationPart.output)).toBe(
      "Change completed successfully.",
    );
    expect(isRenderableToolPart(mutationPart)).toBe(true);
    expect(isRenderableToolPart(readPart)).toBe(true);
  });
});

/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getMutationMessage,
  getStoryCreationApproval,
  isRenderableToolPart,
  isSupportingToolType,
  type ToolMessagePart,
} from "./tool-output-policy";

describe("tool output policy", () => {
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

    expect(getStoryCreationApproval(part)).toEqual({
      approved: undefined,
      count: 2,
      id: "approval-1",
      title: undefined,
    });
    expect(isRenderableToolPart(part)).toBe(true);
  });

  it("keeps a denied story approval visible as a cancellation", () => {
    const part = {
      approval: { approved: false, id: "approval-1" },
      input: { title: "Add onboarding checklist" },
      state: "output-denied",
      toolCallId: "call-1",
      type: "tool-createStory",
    } as unknown as ToolMessagePart;

    expect(getStoryCreationApproval(part)?.approved).toBe(false);
    expect(isRenderableToolPart(part)).toBe(true);
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

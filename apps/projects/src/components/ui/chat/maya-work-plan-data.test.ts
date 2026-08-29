/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getMayaWorkPlanModel,
  isMayaWorkPlanOutput,
} from "./maya-work-plan-data";

const output = {
  kind: "maya-work-plan",
  message: "Maya created and applied the work plan.",
  plan: {
    actions: [
      {
        id: "action-1",
        payload: { assignStory: { assigneeId: "user-1" } },
        reason: "Frontend ownership best matches this work.",
        status: "applied",
        type: "assign_story",
      },
      {
        id: "action-2",
        payload: {
          risk: {
            code: "deadline_capacity",
            message: "Only half of the requested time fits before Friday.",
          },
        },
        reason: "Existing meetings leave a two-hour capacity gap.",
        status: "applied",
        type: "flag_schedule_risk",
      },
    ],
    run: {
      id: "run-1",
      status: "applied",
      summary: "Assigned the story and checked the delivery window.",
    },
  },
  success: true,
};

describe("Maya work plan output", () => {
  it("recognizes the dedicated result kind", () => {
    expect(isMayaWorkPlanOutput(output)).toBe(true);
    expect(isMayaWorkPlanOutput({ kind: "story-list" })).toBe(false);
  });

  it("preserves action reasons and risk copy for the renderer", () => {
    const model = getMayaWorkPlanModel(output);

    expect(model).toMatchObject({
      runStatus: "applied",
      summary: "Assigned the story and checked the delivery window.",
    });
    expect(model?.actions[0]).toMatchObject({
      assigneeId: "user-1",
      label: "Selected owner",
      reason: "Frontend ownership best matches this work.",
    });
    expect(model?.actions[1]).toMatchObject({
      label: "Schedule risk",
      reason: "Existing meetings leave a two-hour capacity gap.",
      riskCode: "deadline_capacity",
      riskMessage: "Only half of the requested time fits before Friday.",
    });
  });

  it("provides concise fallback copy for a partial successful response", () => {
    expect(
      getMayaWorkPlanModel({ kind: "maya-work-plan", success: true }),
    ).toMatchObject({
      actions: [],
      runStatus: "proposed",
      summary: "Maya prepared a work plan for this story.",
    });
  });

  it("derives preview and application phases independently from planner status", () => {
    expect(
      getMayaWorkPlanModel({
        kind: "maya-work-plan",
        phase: "preview",
        plan: { actions: [], run: { status: "succeeded" } },
      })?.runStatus,
    ).toBe("proposed");
    expect(
      getMayaWorkPlanModel({
        kind: "maya-work-plan",
        phase: "applied",
        plan: {
          actions: [
            { id: "action-1", status: "applied", type: "assign_story" },
          ],
          run: { status: "succeeded" },
        },
      })?.runStatus,
    ).toBe("applied");
  });
});

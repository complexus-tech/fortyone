/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  isRenderableToolPart,
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
});

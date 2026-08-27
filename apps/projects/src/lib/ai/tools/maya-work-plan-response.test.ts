/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { resolveMayaWorkPlanResponse } from "./maya-work-plan-response";

describe("resolveMayaWorkPlanResponse", () => {
  it("returns a valid work plan", () => {
    const plan = { actions: [], run: { id: "run-1" } };

    expect(
      resolveMayaWorkPlanResponse({ data: plan }, "Missing work plan."),
    ).toEqual({ data: plan });
  });

  it("fails closed when the API omits plan data", () => {
    expect(
      resolveMayaWorkPlanResponse({}, "Maya returned no work-plan preview."),
    ).toEqual({ error: "Maya returned no work-plan preview." });
    expect(
      resolveMayaWorkPlanResponse(
        { data: null },
        "Maya returned no applied work plan.",
      ),
    ).toEqual({ error: "Maya returned no applied work plan." });
  });

  it("preserves a concrete API error ahead of missing data", () => {
    expect(
      resolveMayaWorkPlanResponse(
        { error: { message: "The preview is stale." } },
        "Missing work plan.",
      ),
    ).toEqual({ error: "The preview is stale." });
  });
});

/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { normalizeOnboardingTourProgress } from "./normalize-progress";

const scope = {
  tourKey: "workspace-getting-started",
  tourVersion: "2.2.0",
};

describe("normalizeOnboardingTourProgress", () => {
  it("normalizes null response collections from mixed API deployments", () => {
    expect(
      normalizeOnboardingTourProgress(
        {
          ...scope,
          completedActionIds: null,
          completedStepIds: null,
          status: "active",
        },
        scope,
      ),
    ).toEqual({
      ...scope,
      completedActionIds: [],
      completedStepIds: [],
      status: "active",
    });
  });

  it("keeps only unique string identifiers and defaults malformed metadata", () => {
    expect(
      normalizeOnboardingTourProgress(
        {
          completedActionIds: ["story-created", null, "story-created"],
          completedStepIds: ["welcome", 42],
          status: "unknown",
        },
        scope,
      ),
    ).toEqual({
      ...scope,
      completedActionIds: ["story-created"],
      completedStepIds: ["welcome"],
      status: "active",
    });
  });

  it("rejects a response without a progress object", () => {
    expect(() => normalizeOnboardingTourProgress(null, scope)).toThrow(
      "Invalid onboarding tour progress response",
    );
  });
});

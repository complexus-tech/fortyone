/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { userKeys } from "./keys";

describe("userKeys", () => {
  it("scopes onboarding progress to the user tour rather than a workspace", () => {
    expect(
      userKeys.onboardingTourProgress(
        "user-1",
        "workspace-module-summary",
        "2.0.0",
      ),
    ).toEqual([
      "users",
      "onboarding-tour-progress",
      "user-1",
      "workspace-module-summary",
      "2.0.0",
    ]);
  });
});

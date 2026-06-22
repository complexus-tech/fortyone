/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { normalizeStoryInput } from "./normalize-story-input";

describe("normalizeStoryInput", () => {
  it("omits reporterId because the stories API derives the reporter from auth", () => {
    const payload = normalizeStoryInput({
      title: "Add onboarding checklist",
      teamId: "team-1",
      statusId: "status-1",
      reporterId: "user-1",
      priority: "Medium",
    });

    expect(payload).not.toHaveProperty("reporterId");
  });
});

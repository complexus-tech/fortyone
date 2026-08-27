/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  requiresMutationApproval,
  toApprovedMutationInput,
} from "./tool-policy";

describe("Maya tool policy", () => {
  it("injects confirmed only for legacy mutation executors", () => {
    expect(
      toApprovedMutationInput("updateStory", {
        confirmed: false,
        storyId: "story-1",
        title: "Updated",
      }),
    ).toEqual({
      confirmed: true,
      storyId: "story-1",
      title: "Updated",
    });
    expect(
      toApprovedMutationInput("createTeamTool", { name: "Product" }),
    ).toEqual({ name: "Product" });
  });

  it("keeps GitHub install setup outside native mutation approval", () => {
    expect(requiresMutationApproval("createGitHubInstallSessionTool", {})).toBe(
      false,
    );
    expect(requiresMutationApproval("resyncGitHubRepositoriesTool", {})).toBe(
      true,
    );
  });

  it("previews work plans without approval and approves only exact apply runs", () => {
    expect(requiresMutationApproval("mayaWorkPlanTool", {})).toBe(false);
    expect(
      requiresMutationApproval("applyMayaWorkPlanTool", { runId: "run-1" }),
    ).toBe(true);
  });
});

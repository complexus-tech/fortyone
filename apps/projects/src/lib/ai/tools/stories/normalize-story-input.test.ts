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

  it("omits estimateValue when the model uses 0 to mean unestimated", () => {
    const payload = normalizeStoryInput({
      title: "Move tracking links",
      teamId: "team-1",
      statusId: "status-1",
      priority: "Medium",
      estimateValue: 0,
    });

    expect(payload).not.toHaveProperty("estimateValue");
  });

  it("rejects estimate values that the API cannot store", () => {
    expect(() =>
      normalizeStoryInput({
        title: "Move tracking links",
        teamId: "team-1",
        statusId: "status-1",
        priority: "Medium",
        estimateValue: 4,
      }),
    ).toThrow("estimateValue must be one of 1, 2, 3, 5, or 8");
  });

  it("omits empty or placeholder optional IDs before calling the API", () => {
    const payload = normalizeStoryInput({
      title: "Move tracking links",
      teamId: "team-1",
      statusId: "status-1",
      assigneeId: "[your user ID or empty]",
      priority: "Medium",
      sprintId: "",
      objectiveId: "",
      parentId: "",
      startDate: "",
      endDate: "",
    });

    expect(payload).not.toHaveProperty("assigneeId");
    expect(payload).not.toHaveProperty("sprintId");
    expect(payload).not.toHaveProperty("objectiveId");
    expect(payload).not.toHaveProperty("parentId");
    expect(payload).not.toHaveProperty("startDate");
    expect(payload).not.toHaveProperty("endDate");
  });

  it("rejects placeholder required IDs with an actionable error", () => {
    expect(() =>
      normalizeStoryInput({
        title: "Move tracking links",
        teamId: "[Product team ID]",
        statusId: "[To Do status ID]",
        priority: "Medium",
      }),
    ).toThrow("teamId must be resolved to a real ID before creating a story");
  });
});

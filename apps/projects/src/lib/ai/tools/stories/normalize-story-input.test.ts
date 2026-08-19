/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { normalizeStoryInput } from "./normalize-story-input";

const TEAM_ID = "11111111-1111-4111-8111-111111111111";
const STATUS_ID = "22222222-2222-4222-8222-222222222222";
const ASSIGNEE_ID = "33333333-3333-4333-8333-333333333333";
const SPRINT_ID = "44444444-4444-4444-8444-444444444444";
const KEY_RESULT_ID = "66666666-6666-4666-8666-666666666666";
const LABEL_ID = "77777777-7777-4777-8777-777777777777";

describe("normalizeStoryInput", () => {
  it("omits reporterId because the stories API derives the reporter from auth", () => {
    const payload = normalizeStoryInput({
      title: "Add onboarding checklist",
      teamId: TEAM_ID,
      statusId: STATUS_ID,
      reporterId: "user-1",
      priority: "Medium",
    });

    expect(payload).not.toHaveProperty("reporterId");
  });

  it("omits estimateValue when the model uses 0 to mean unestimated", () => {
    const payload = normalizeStoryInput({
      title: "Move tracking links",
      teamId: TEAM_ID,
      statusId: STATUS_ID,
      priority: "Medium",
      estimateValue: 0,
    });

    expect(payload).not.toHaveProperty("estimateValue");
  });

  it("serializes scheduling time separately from relative complexity", () => {
    const payload = normalizeStoryInput({
      title: "Schedule focused work",
      teamId: TEAM_ID,
      statusId: STATUS_ID,
      priority: "High",
      estimateValue: 5,
      estimatedDurationMinutes: 120,
      minimumFocusBlockMinutes: 30,
      autoSchedulingEnabled: true,
    });

    expect(payload).toMatchObject({
      estimateValue: 5,
      estimatedDurationMinutes: 120,
      minimumFocusBlockMinutes: 30,
      autoSchedulingEnabled: true,
    });
  });

  it("rejects an invalid scheduling time contract", () => {
    expect(() =>
      normalizeStoryInput({
        title: "Impossible split",
        teamId: TEAM_ID,
        statusId: STATUS_ID,
        priority: "Medium",
        estimatedDurationMinutes: 30,
        minimumFocusBlockMinutes: 60,
      }),
    ).toThrow("cannot exceed estimatedDurationMinutes");
  });

  it("rejects estimate values that the API cannot store", () => {
    expect(() =>
      normalizeStoryInput({
        title: "Move tracking links",
        teamId: TEAM_ID,
        statusId: STATUS_ID,
        priority: "Medium",
        estimateValue: 4,
      }),
    ).toThrow("estimateValue must be one of 1, 2, 3, 5, or 8");
  });

  it("omits empty or placeholder optional IDs before calling the API", () => {
    const payload = normalizeStoryInput({
      title: "Move tracking links",
      teamId: TEAM_ID,
      statusId: STATUS_ID,
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
        statusId: STATUS_ID,
        priority: "Medium",
      }),
    ).toThrow("teamId must be resolved to a real ID before creating a story");
  });

  it("converts date-time values to calendar dates and preserves scheduling fields", () => {
    const payload = normalizeStoryInput({
      title: "  Plan launch  ",
      teamId: TEAM_ID,
      statusId: STATUS_ID,
      assigneeId: ASSIGNEE_ID,
      keyResultId: KEY_RESULT_ID,
      labelIds: [LABEL_ID, LABEL_ID, ""],
      priority: "High",
      startDate: "2026-08-19T00:00:00.000Z",
      endDate: "2026-08-29T23:59:59.000Z",
    });

    expect(payload).toMatchObject({
      assigneeId: ASSIGNEE_ID,
      endDate: "2026-08-29",
      keyResultId: KEY_RESULT_ID,
      labelIds: [LABEL_ID],
      startDate: "2026-08-19",
      title: "Plan launch",
    });
  });

  it("omits null label and association values", () => {
    const payload = normalizeStoryInput({
      title: "Move tracking links",
      teamId: TEAM_ID,
      statusId: STATUS_ID,
      labelIds: null,
      objectiveId: null,
      priority: "Medium",
      sprintId: SPRINT_ID,
    });

    expect(payload).not.toHaveProperty("labelIds");
    expect(payload).not.toHaveProperty("objectiveId");
    expect(payload).toHaveProperty("sprintId", SPRINT_ID);
  });

  it.each([
    ["invalid optional IDs", { assigneeId: "not-a-uuid" }, "assigneeId"],
    ["invalid label IDs", { labelIds: ["not-a-uuid"] }, "labelIds"],
    ["invalid dates", { endDate: "2026-02-30" }, "valid calendar date"],
    [
      "reversed dates",
      { startDate: "2026-09-01", endDate: "2026-08-31" },
      "endDate cannot be before startDate",
    ],
  ])("rejects %s", (_, values, message) => {
    expect(() =>
      normalizeStoryInput({
        title: "Move tracking links",
        teamId: TEAM_ID,
        statusId: STATUS_ID,
        priority: "Medium",
        ...values,
      }),
    ).toThrow(message);
  });
});

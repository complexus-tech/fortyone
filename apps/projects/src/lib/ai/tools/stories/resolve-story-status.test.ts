/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { selectStoryStatusId } from "./select-story-status";

const statuses = [
  {
    id: "backlog-status",
    category: "backlog",
    isDefault: false,
  },
  {
    id: "ready-status",
    category: "unstarted",
    isDefault: true,
  },
];

describe("selectStoryStatusId", () => {
  it("uses the requested status when it belongs to the target team", () => {
    expect(selectStoryStatusId(statuses, "backlog-status")).toBe(
      "backlog-status",
    );
  });

  it("uses the team's default status when Maya omits a status", () => {
    expect(selectStoryStatusId(statuses)).toBe("ready-status");
  });

  it("rejects a status from another team", () => {
    expect(() => selectStoryStatusId(statuses, "other-team-status")).toThrow(
      "The selected status is not available for the target team.",
    );
  });

  it("reports teams without workflow statuses", () => {
    expect(() => selectStoryStatusId([])).toThrow(
      "The target team has no workflow statuses.",
    );
  });
});

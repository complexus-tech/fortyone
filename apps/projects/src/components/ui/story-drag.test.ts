/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Story } from "@/modules/stories/types";
import { getStoryDropUpdate, getStoryGroupKey } from "./story-drag";

const story: Pick<Story, "assigneeId" | "priority" | "statusId"> = {
  assigneeId: null,
  priority: "High",
  statusId: "in-progress",
};

describe("getStoryGroupKey", () => {
  it("uses the API's null group key for unassigned stories", () => {
    expect(getStoryGroupKey(story, "assignee")).toBe("null");
  });
});

describe("getStoryDropUpdate", () => {
  it("treats a drop into the current group as a no-op", () => {
    expect(getStoryDropUpdate(story, "status", "in-progress")).toBeNull();
    expect(getStoryDropUpdate(story, "priority", "High")).toBeNull();
    expect(getStoryDropUpdate(story, "assignee", "null")).toBeNull();
  });

  it("builds the minimal update for a real group change", () => {
    expect(getStoryDropUpdate(story, "status", "done")).toEqual({
      statusId: "done",
    });
    expect(getStoryDropUpdate(story, "priority", "Urgent")).toEqual({
      priority: "Urgent",
    });
    expect(getStoryDropUpdate(story, "assignee", "user-2")).toEqual({
      assigneeId: "user-2",
    });
  });
});

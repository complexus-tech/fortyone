/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { DetailedStory } from "@/modules/story/types";
import { toStoryToolSummary } from "./story-tool-summary";

describe("toStoryToolSummary", () => {
  it("keeps only the fields needed by the model and confirmation UI", () => {
    const story = {
      id: "story-1",
      title: "Prepare launch",
      description: "A large description that should not be sent back to Maya.",
      activities: [{ id: "activity-1" }],
      labels: [{ id: "label-1", name: "Launch" }],
    } as unknown as DetailedStory;

    expect(toStoryToolSummary(story)).toEqual({
      id: "story-1",
      title: "Prepare launch",
    });
  });
});

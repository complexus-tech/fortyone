/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { NewStory } from "@/modules/story/types";
import { buildNewStoryDialogPayload } from "./new-story-dialog-form";

describe("new story dialog form", () => {
  it("preserves selected labels and estimate value in the create payload", () => {
    const storyForm: NewStory = {
      assigneeId: "user-1",
      endDate: "2026-06-20",
      estimateValue: 5,
      labelIds: ["label-1", "label-2"],
      objectiveId: "objective-1",
      priority: "High",
      sprintId: "sprint-1",
      startDate: "2026-06-13",
      statusId: "status-1",
    };

    expect(
      buildNewStoryDialogPayload({
        currentTeamId: "team-1",
        description: "Plain text description",
        descriptionHTML: "<p>Plain text description</p>",
        storyForm,
        title: "Add reporting filters",
      }),
    ).toMatchObject({
      assigneeId: "user-1",
      description: "Plain text description",
      descriptionHTML: "<p>Plain text description</p>",
      estimateValue: 5,
      labelIds: ["label-1", "label-2"],
      objectiveId: "objective-1",
      priority: "High",
      sprintId: "sprint-1",
      statusId: "status-1",
      teamId: "team-1",
      title: "Add reporting filters",
    });
  });

  it("does not send estimate scheme because the API derives it from team settings", () => {
    const payload = buildNewStoryDialogPayload({
      currentTeamId: "team-1",
      description: "",
      descriptionHTML: "",
      storyForm: {
        estimateValue: 2,
        priority: "Medium",
        statusId: "status-1",
      },
      title: "Add estimate input",
    });

    expect(payload).not.toHaveProperty("estimateScheme");
  });
});

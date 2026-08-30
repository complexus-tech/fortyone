/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getNewStoryDialogFieldSelections,
  getSelectedNewStoryLabels,
} from "./new-story-dialog-selections";

describe("new story dialog selections", () => {
  it("keeps only the labels included in the draft", () => {
    expect(
      getSelectedNewStoryLabels(
        [
          { id: "label-1", name: "Product" },
          { id: "label-2", name: "Engineering" },
        ],
        ["label-2"],
      ),
    ).toEqual([{ id: "label-2", name: "Engineering" }]);
  });

  it("derives the selected story metadata from the current form", () => {
    const mayaAssignee = { id: "maya-1", fullName: "Maya" };

    expect(
      getNewStoryDialogFieldSelections({
        currentTeamCode: "ENG",
        keyResults: [{ id: "key-result-1", name: "Ship it" }],
        mayaAssignee,
        members: [{ id: "member-1", fullName: "Ada" }],
        objectives: [
          { id: "objective-1", name: "Reliability", sequenceId: 12 },
        ],
        sprints: [{ id: "sprint-1", name: "June" }],
        storyForm: {
          assigneeId: "maya-1",
          keyResultId: "key-result-1",
          objectiveId: "objective-1",
          sprintId: "sprint-1",
        },
      }),
    ).toEqual({
      isMayaAssigned: true,
      member: mayaAssignee,
      sprint: { id: "sprint-1", name: "June" },
      strategyLinkLabel: "ENG-12 / Ship it",
    });
  });
});

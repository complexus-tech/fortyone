/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getActivityCopy,
  getActivityValueIds,
  getDisplayActivityReason,
} from "./activity-copy";

describe("activity copy", () => {
  it("uses natural copy for common story fields", () => {
    expect(
      getActivityCopy({
        currentValue: "In Progress",
        field: "status_id",
        fieldLabel: "Status",
        type: "update",
      }).text,
    ).toBe("moved the story to In Progress");

    expect(
      getActivityCopy({
        currentValue: "Launch notes",
        field: "title",
        fieldLabel: "Title",
        type: "update",
      }).text,
    ).toBe("renamed the story to Launch notes");
  });

  it("uses configured terminology in entity copy", () => {
    expect(
      getActivityCopy({
        currentValue: "Done",
        field: "status_id",
        fieldLabel: "Status",
        oldValue: "To Do",
        storyTerm: "issue",
        type: "update",
      }).text,
    ).toBe("moved the issue from To Do to Done");

    expect(
      getActivityCopy({
        currentValue: "Launch notes",
        field: "title",
        fieldLabel: "Title",
        storyTerm: "issue",
        type: "update",
      }).text,
    ).toBe("renamed the issue to Launch notes");
  });

  it("uses action-oriented copy for association changes", () => {
    expect(
      getActivityCopy({
        currentValue: "Testing",
        field: "duplicate_id",
        fieldLabel: "Duplicate of",
        reason: "association_added",
        type: "update",
      }).text,
    ).toBe("marked Testing as Duplicate of");

    expect(
      getActivityCopy({
        currentValue: "Testing",
        field: "duplicate_id",
        fieldLabel: "Duplicate of",
        oldValue: "Related to",
        reason: "association_updated",
        type: "update",
      }).text,
    ).toBe("changed Testing from Related to Duplicate of");

    expect(
      getActivityCopy({
        currentValue: "Testing",
        field: "duplicate_id",
        fieldLabel: "Duplicate of",
        reason: "association_removed",
        type: "update",
      }).text,
    ).toBe("removed the Duplicate of relationship with Testing");
  });

  it("uses natural copy for label changes", () => {
    expect(
      getActivityCopy({
        currentValue: "Test",
        field: "labels",
        fieldLabel: "Labels",
        type: "update",
      }).text,
    ).toBe("added Test label");

    expect(
      getActivityCopy({
        currentValue: "3 labels",
        field: "labels",
        fieldLabel: "Labels",
        type: "update",
      }).text,
    ).toBe("updated labels 3 labels");
  });

  it("uses natural copy for collaborator changes", () => {
    const copy = getActivityCopy({
      currentValue: "2 collaborators",
      field: "collaborator_ids",
      fieldLabel: "Collaborators",
      type: "update",
    });

    expect(copy.text).toBe("updated collaborators to 2 collaborators");
    expect(copy.segments).toContainEqual({
      type: "currentValue",
      value: "2 collaborators",
    });
  });

  it("extracts collaborator IDs from legacy activity values", () => {
    const collaboratorId = "8d852a9e-b36d-4790-9f92-8234407f1488";

    expect(getActivityValueIds(`[${collaboratorId}]`)).toEqual([
      collaboratorId,
    ]);
    expect(getActivityValueIds(`["${collaboratorId}"]`)).toEqual([
      collaboratorId,
    ]);
    expect(getActivityValueIds([collaboratorId])).toEqual([collaboratorId]);
  });

  it("does not expose internal association reasons as activity notes", () => {
    expect(getDisplayActivityReason("association_added")).toBe("");
    expect(getDisplayActivityReason("Moved because of planning")).toBe(
      "Moved because of planning",
    );
  });
});

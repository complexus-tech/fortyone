/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { ImportTask } from "../schema";
import { getImportTaskRelationshipPreview } from "./import-task-relationship-preview";

const task = {
  sourceId: "task-a",
  title: "Migration task",
  parentSourceId: "parent",
  keyResultSourceId: "kr-1",
  sprintSourceId: "sprint-1",
  labelSourceIds: ["label-1"],
  collaboratorPersonSourceIds: ["person-1"],
  links: [{ title: "Trello card", url: "https://trello.com/c/card" }],
  associations: [
    { type: "blocks", targetSourceId: "task-b" },
    { type: "blocked_by", targetSourceId: "task-c" },
    { type: "related", targetSourceId: "task-d" },
    { type: "duplicate", targetSourceId: "task-e" },
  ],
} as ImportTask;

describe("import task relationship preview", () => {
  it("makes hierarchy, planning links, people, URLs, and association direction reviewable", () => {
    expect(
      getImportTaskRelationshipPreview(
        task,
        {
          keyResults: new Map([["kr-1", "Increase activation"]]),
          labels: new Map([["label-1", "Migration"]]),
          people: new Map([["person-1", "Ada Lovelace"]]),
          sprints: new Map([["sprint-1", "September sprint"]]),
          tasks: new Map([
            ["parent", "Migration parent"],
            ["task-b", "Blocked work"],
            ["task-c", "Dependency"],
            ["task-d", "Related work"],
            ["task-e", "Duplicate work"],
          ]),
        },
        {
          keyResult: "Outcome",
          sprint: "Cycle",
        },
      ),
    ).toEqual([
      "Parent: Migration parent",
      "Outcome: Increase activation",
      "Cycle: September sprint",
      "Labels: Migration",
      "Collaborators: Ada Lovelace",
      "Link: Trello card",
      "Blocks: Blocked work",
      "Blocked by: Dependency",
      "Related to: Related work",
      "Duplicates: Duplicate work",
    ]);
  });
});

import assert from "node:assert/strict";
import test from "node:test";
import type { Status } from "@/types/statuses";
import { canCompleteStory, getCompletionStatus } from "./quick-actions";

const todo = { id: "todo", category: "unstarted", teamId: "team-a" } as Status;
const done = {
  id: "done",
  category: "completed",
  teamId: "team-a",
  orderIndex: 3,
} as Status;
const story = {
  teamId: "team-a",
  statusId: "todo",
  completedAt: null,
  archivedAt: null,
  deletedAt: null,
};

test("quick complete uses the team's ordered completed status without changing cached statuses", () => {
  const alternative = { ...done, id: "released", orderIndex: 4 };
  const otherTeam = {
    ...done,
    id: "other-done",
    teamId: "team-b",
    orderIndex: 0,
  };
  const statuses = [alternative, otherTeam, todo, done];
  assert.equal(getCompletionStatus(statuses, "team-a"), done);
  assert.equal(statuses[0], alternative);
  assert.equal(getCompletionStatus(statuses, "missing-team"), undefined);
});

test("quick complete is unavailable until the current and destination team statuses are known", () => {
  assert.equal(canCompleteStory(story, todo, done), true);
  assert.equal(canCompleteStory(story, undefined, done), false);
  assert.equal(canCompleteStory(story, todo), false);
  assert.equal(
    canCompleteStory(story, todo, { ...done, teamId: "team-b" }),
    false,
  );
  assert.equal(canCompleteStory(story, { ...todo, id: "stale" }, done), false);
  assert.equal(canCompleteStory(story, todo, todo), false);
});

test("quick complete does not change terminal, archived or deleted work", () => {
  for (const category of ["completed", "cancelled"] as const) {
    assert.equal(canCompleteStory(story, { ...todo, category }, done), false);
  }
  for (const field of ["completedAt", "archivedAt", "deletedAt"] as const) {
    assert.equal(
      canCompleteStory({ ...story, [field]: "2026-09-15" }, todo, done),
      false,
    );
  }
});

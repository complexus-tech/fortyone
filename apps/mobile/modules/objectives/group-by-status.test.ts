import assert from "node:assert/strict";
import test from "node:test";
import type { Objective, ObjectiveStatus } from "./types";
import { groupObjectivesByStatus } from "./group-by-status.ts";

const objective = (id: string, statusId: string): Objective => ({
  id,
  statusId,
  name: id,
  description: "",
  leadUser: "",
  teamId: "team",
  workspaceId: "workspace",
  startDate: "",
  endDate: "",
  isPrivate: false,
  createdAt: "",
  updatedAt: "",
  createdBy: "",
  health: null,
});
const status = (id: string, orderIndex: number): ObjectiveStatus => ({
  id,
  orderIndex,
  name: id,
  category: "unstarted",
  color: "#888888",
  isDefault: false,
  workspaceId: "workspace",
  createdAt: "",
  updatedAt: "",
});

test("groups objectives in configured status order without empty groups or input mutation", () => {
  const statuses = [
    status("done", 2),
    status("empty", 0),
    status("planned", 1),
  ];
  const items = [
    objective("one", "done"),
    objective("two", "planned"),
    objective("three", "done"),
  ];
  const groups = groupObjectivesByStatus(items, statuses);
  assert.deepEqual(
    groups.map((group) => [group.key, group.data.map((item) => item.id)]),
    [
      ["planned", ["two"]],
      ["done", ["one", "three"]],
    ],
  );
  assert.equal(groups[0].status, statuses[2]);
  assert.deepEqual(
    statuses.map((item) => item.id),
    ["done", "empty", "planned"],
  );
  assert.deepEqual(
    items.map((item) => item.id),
    ["one", "two", "three"],
  );
});

test("missing statuses never hide objectives and an empty list stays empty", () => {
  const items = [objective("one", "removed"), objective("two", "")];
  const groups = groupObjectivesByStatus(items, []);
  assert.deepEqual(
    groups.flatMap((group) => group.data),
    items,
  );
  assert.ok(
    groups.every((group) => group.title === "No status" && !group.status),
  );
  assert.deepEqual(groupObjectivesByStatus([], [status("planned", 1)]), []);
});

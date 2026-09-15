import assert from "node:assert/strict";
import test from "node:test";
import { filterSwitcherTeams } from "./team-switcher-utils";

const orderedTeams = Object.freeze([
  Object.freeze({ id: "3", name: "Zulu Engineering", code: "ZE" }),
  Object.freeze({ id: "1", name: "Design", code: "DES" }),
  Object.freeze({ id: "2", name: "Alpha Engineering", code: "AE" }),
]);

test("team search preserves saved ordering and the unfiltered cache", () => {
  assert.strictEqual(filterSwitcherTeams(orderedTeams, "  "), orderedTeams);
  assert.deepEqual(
    filterSwitcherTeams(orderedTeams, " ENGINEERING ").map((team) => team.id),
    ["3", "2"],
  );
  assert.deepEqual(
    orderedTeams.map((team) => team.id),
    ["3", "1", "2"],
  );
});

test("team search includes codes and returns no unrelated memberships", () => {
  assert.deepEqual(
    filterSwitcherTeams(orderedTeams, " des ").map((team) => team.id),
    ["1"],
  );
  assert.deepEqual(filterSwitcherTeams(orderedTeams, "missing"), []);
  assert.deepEqual(filterSwitcherTeams([], "design"), []);
});

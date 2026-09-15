import assert from "node:assert/strict";
import test from "node:test";
import {
  buildStoryFilterParams,
  createTeamStoryFilters,
} from "@/modules/teams/stories/team-story-filters";
import { changeStoryFilter } from "./story-filter-selection";

test("facets compose without replacing selections in other facets", () => {
  const original = createTeamStoryFilters("team");
  let filters = changeStoryFilter(original, "status", "status-one");
  filters = changeStoryFilter(filters, "priority", "High");
  filters = changeStoryFilter(filters, "sprint", "sprint-one");
  filters = changeStoryFilter(filters, "objective", "objective-one");
  filters = changeStoryFilter(filters, "assignee", "member-one");
  assert.deepEqual(buildStoryFilterParams(filters), {
    statusIds: ["status-one"],
    priorities: ["High"],
    sprintIds: ["sprint-one"],
    objectiveId: "objective-one",
    assigneeIds: ["member-one"],
  });
  assert.equal(filters.teamId, "team");
  assert.deepEqual(original, createTeamStoryFilters("team"));
});

test("multi-select facets toggle one value and Any clears only that facet", () => {
  for (const [facet, first, second, key] of [
    ["status", "status-one", "status-two", "statusIds"],
    ["priority", "No Priority", "High", "priorities"],
    ["sprint", "sprint-one", "sprint-two", "sprintIds"],
  ] as const) {
    const base = {
      ...createTeamStoryFilters("team"),
      objectiveId: "objective",
    };
    let filters = changeStoryFilter(base, facet, first);
    filters = changeStoryFilter(filters, facet, second);
    assert.deepEqual(filters[key], [first, second]);
    filters = changeStoryFilter(filters, facet, first);
    assert.deepEqual(filters[key], [second]);
    filters = changeStoryFilter(filters, facet, null);
    assert.deepEqual(filters[key], []);
    assert.equal(filters.objectiveId, "objective");
  }
});

test("Unassigned replaces selected members and a member replaces Unassigned", () => {
  let filters = changeStoryFilter(
    createTeamStoryFilters("team"),
    "assignee",
    "one",
  );
  filters = changeStoryFilter(filters, "assignee", "two");
  assert.deepEqual(filters.assignee, { kind: "members", ids: ["one", "two"] });
  filters = changeStoryFilter(filters, "assignee", "unassigned");
  assert.deepEqual(buildStoryFilterParams(filters), { hasNoAssignee: true });
  filters = changeStoryFilter(filters, "assignee", "three");
  assert.deepEqual(buildStoryFilterParams(filters), { assigneeIds: ["three"] });
  filters = changeStoryFilter(filters, "assignee", "three");
  assert.deepEqual(filters.assignee, { kind: "any" });
  assert.deepEqual(buildStoryFilterParams(filters), {});
});

test("Any assignee clears members or Unassigned without changing other filters", () => {
  for (const selected of ["member", "unassigned"]) {
    const base = { ...createTeamStoryFilters("team"), sprintIds: ["sprint"] };
    const filters = changeStoryFilter(
      changeStoryFilter(base, "assignee", selected),
      "assignee",
      null,
    );
    assert.deepEqual(filters.assignee, { kind: "any" });
    assert.deepEqual(buildStoryFilterParams(filters), {
      sprintIds: ["sprint"],
    });
  }
});

test("objective selection replaces the prior objective and clears explicitly", () => {
  const base = { ...createTeamStoryFilters("team"), statusIds: ["status"] };
  let filters = changeStoryFilter(base, "objective", "first");
  filters = changeStoryFilter(filters, "objective", "second");
  assert.equal(filters.objectiveId, "second");
  filters = changeStoryFilter(filters, "objective", "second");
  assert.equal(filters.objectiveId, null);
  filters = changeStoryFilter(filters, "objective", "first");
  filters = changeStoryFilter(filters, "objective", null);
  assert.deepEqual(buildStoryFilterParams(filters), { statusIds: ["status"] });
});

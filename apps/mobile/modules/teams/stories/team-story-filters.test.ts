import type { GroupStoryParams, StoryPriority } from "@/modules/stories/types";
import type { StoriesViewOptions } from "@/types/stories-view-options";
import type { TeamStoriesTab } from "../types";
import assert from "node:assert/strict";
import test from "node:test";
import { stringify } from "qs";
import {
  buildStoryFilterParams,
  buildTeamStoryQuery,
  createTeamStoryFilters,
  getTeamStoryFilterCount,
  getTeamStoryFilters,
} from "./team-story-filters";

const ids = {
  team: "11111111-1111-4111-8111-111111111111",
  otherTeam: "22222222-2222-4222-8222-222222222222",
  first: "33333333-3333-4333-8333-333333333333",
  second: "44444444-4444-4444-8444-444444444444",
};
const viewOptions: StoriesViewOptions = {
  groupBy: "priority",
  orderBy: "deadline",
  orderDirection: "asc",
  displayColumns: ["ID", "Assignee"],
};

test("cleared facets omit constraints; No Priority remains an explicit value", () => {
  const filters = createTeamStoryFilters(ids.team);
  assert.deepEqual(buildStoryFilterParams(filters), {});
  assert.equal(getTeamStoryFilterCount(filters), 0);
  assert.deepEqual(
    buildStoryFilterParams({ ...filters, priorities: ["No Priority"] }),
    { priorities: ["No Priority"] },
  );
  assert.deepEqual(
    buildStoryFilterParams({
      ...filters,
      assignee: { kind: "members", ids: [] },
    }),
    {},
  );
});

test("multiple selections produce stable keys without mutating user selections", () => {
  const filters = {
    ...createTeamStoryFilters(ids.team),
    statusIds: [ids.second, ids.first, ids.second],
    priorities: ["Low", "High", "Low"] as StoryPriority[],
    sprintIds: [ids.second, ids.first],
    objectiveId: ids.first,
    assignee: { kind: "members" as const, ids: [ids.second, ids.first] },
  };
  const before = structuredClone(filters);
  const query = buildTeamStoryQuery({
    teamId: ids.team,
    filters,
    tab: "all",
    viewOptions,
  });
  assert.deepEqual(query, {
    teamIds: [ids.team],
    groupBy: "priority",
    orderBy: "deadline",
    orderDirection: "asc",
    statusIds: [ids.first, ids.second],
    priorities: ["High", "Low"],
    sprintIds: [ids.first, ids.second],
    objectiveId: ids.first,
    assigneeIds: [ids.first, ids.second],
  });
  assert.equal(getTeamStoryFilterCount(filters), 5);
  assert.deepEqual(filters, before);
  assert.equal(
    JSON.stringify(query),
    JSON.stringify(
      buildTeamStoryQuery({
        teamId: ids.team,
        filters: {
          ...filters,
          statusIds: [ids.first, ids.second],
          priorities: ["High", "Low"],
          sprintIds: [ids.first, ids.second],
          assignee: { kind: "members", ids: [ids.first, ids.second] },
        },
        tab: "all",
        viewOptions,
      }),
    ),
  );
});

test("Unassigned and selected members are mutually exclusive API constraints", () => {
  const unassigned = {
    ...createTeamStoryFilters(ids.team),
    assignee: { kind: "unassigned" as const },
  };
  assert.deepEqual(buildStoryFilterParams(unassigned), { hasNoAssignee: true });
  assert.equal(getTeamStoryFilterCount(unassigned), 1);
  assert.deepEqual(
    buildStoryFilterParams({
      ...unassigned,
      assignee: { kind: "members", ids: [ids.first] },
    }),
    { assigneeIds: [ids.first] },
  );
  assert.deepEqual(
    buildStoryFilterParams({ ...unassigned, assignee: { kind: "any" } }),
    {},
  );
});

test("tab categories intersect explicit status filters", () => {
  const filters = {
    ...createTeamStoryFilters(ids.team),
    statusIds: [ids.first],
  };
  for (const [tab, categories] of [
    ["all", undefined],
    ["active", ["started"]],
    ["backlog", ["backlog"]],
  ] as const) {
    const query = buildTeamStoryQuery({
      teamId: ids.team,
      filters,
      tab,
      viewOptions,
    });
    assert.deepEqual(query.categories, categories);
    assert.deepEqual(query.statusIds, [ids.first]);
    assert.deepEqual(query.teamIds, [ids.team]);
  }
});

test("a team transition cannot reuse the previous team's filters", () => {
  const previous = {
    ...createTeamStoryFilters(ids.team),
    sprintIds: [ids.first],
    objectiveId: ids.second,
    assignee: { kind: "unassigned" as const },
  };
  const next = getTeamStoryFilters(ids.otherTeam, previous);
  assert.equal(next.teamId, ids.otherTeam);
  assert.equal(getTeamStoryFilterCount(next), 0);
  assert.deepEqual(
    buildTeamStoryQuery({
      teamId: ids.otherTeam,
      filters: previous,
      tab: "all",
      viewOptions,
    }),
    {
      teamIds: [ids.otherTeam],
      groupBy: "priority",
      orderBy: "deadline",
      orderDirection: "asc",
    },
  );
  assert.equal(getTeamStoryFilterCount(previous), 3);
  assert.strictEqual(getTeamStoryFilters(ids.team, previous), previous);
});

test("Reset clears only facets and keeps the tab and display preferences", () => {
  const tab: TeamStoriesTab = "backlog";
  const reset = createTeamStoryFilters(ids.team);
  const another = createTeamStoryFilters(ids.team);
  assert.notStrictEqual(reset.statusIds, another.statusIds);
  assert.notStrictEqual(reset.sprintIds, another.sprintIds);
  assert.notStrictEqual(reset.priorities, another.priorities);
  const query = buildTeamStoryQuery({
    teamId: ids.team,
    filters: reset,
    tab,
    viewOptions,
  });
  assert.deepEqual(query, {
    teamIds: [ids.team],
    groupBy: "priority",
    orderBy: "deadline",
    orderDirection: "asc",
    categories: ["backlog"],
  });
  assert.deepEqual(viewOptions.displayColumns, ["ID", "Assignee"]);
});

test("subsequent group pages retain every facet in their serialized query", () => {
  const query = buildTeamStoryQuery({
    teamId: ids.team,
    filters: {
      ...createTeamStoryFilters(ids.team),
      statusIds: [ids.first],
      priorities: ["No Priority", "High"],
      sprintIds: [ids.second],
      objectiveId: ids.first,
      assignee: { kind: "unassigned" },
    },
    tab: "active",
    viewOptions,
  });
  const page: GroupStoryParams = { ...query, groupKey: "High", page: 2 };
  const serialized = new URLSearchParams(
    stringify(page, {
      skipNulls: true,
      encodeValuesOnly: true,
      arrayFormat: "comma",
    }),
  );
  for (const [key, value] of Object.entries(query)) {
    assert.equal(
      serialized.get(key),
      Array.isArray(value) ? value.join(",") : String(value),
    );
  }
  assert.equal(serialized.get("page"), "2");
  assert.equal(serialized.get("groupKey"), "High");
  assert.equal(serialized.get("hasNoAssignee"), "true");
  assert.equal(serialized.has("assigneeIds"), false);
});

test("shared facets compose with My Work without emitting a fake team ID", () => {
  const filters = {
    ...createTeamStoryFilters("my-work:created"),
    objectiveId: ids.first,
    priorities: ["High"] as StoryPriority[],
  };
  const query = { createdByMe: true, ...buildStoryFilterParams(filters) };
  assert.deepEqual(query, {
    createdByMe: true,
    priorities: ["High"],
    objectiveId: ids.first,
  });
  assert.equal(
    getTeamStoryFilterCount(getTeamStoryFilters("my-work:assigned", filters)),
    0,
  );
});

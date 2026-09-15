import assert from "node:assert/strict";
import test from "node:test";
import {
  buildTeamStoryQuery,
  createTeamStoryFilters,
} from "./team-story-filters.ts";
import { parseTeamStoryRouteFilters } from "./team-story-route-filters.ts";

const SPRINT_ID = "ad54ecfb-199c-4169-adab-2c6912d19f64";
const OBJECTIVE_ID = "01995ff2-3130-7000-823e-9c77cd2b67d9";

test("an ordinary team link seeds fresh canonical filters without changing team identity", () => {
  const first = parseTeamStoryRouteFilters("team-route-id", {});
  const second = parseTeamStoryRouteFilters("team-route-id", {});
  assert.deepEqual(first, { filters: createTeamStoryFilters("team-route-id") });
  first.filters.sprintIds.push(SPRINT_ID);
  first.filters.statusIds.push("status");
  assert.deepEqual(second.filters, createTeamStoryFilters("team-route-id"));
});

test("sprint and objective IDs seed their corresponding filters with canonical casing", () => {
  const sprint = parseTeamStoryRouteFilters("team", {
    sprintId: SPRINT_ID.toUpperCase(),
  });
  const objective = parseTeamStoryRouteFilters("team", {
    objectiveId: OBJECTIVE_ID,
  });
  assert.deepEqual(sprint.filters.sprintIds, [SPRINT_ID]);
  assert.equal(sprint.filters.objectiveId, null);
  assert.equal(objective.filters.objectiveId, OBJECTIVE_ID);
  assert.deepEqual(objective.filters.sprintIds, []);
  assert.equal(sprint.error, undefined);
  assert.equal(objective.error, undefined);
});

test("combined entity filters remain an AND intersection in the actual team query", () => {
  const parsed = parseTeamStoryRouteFilters("team", {
    sprintId: SPRINT_ID,
    objectiveId: OBJECTIVE_ID,
  });
  assert.equal(parsed.error, undefined);
  const query = buildTeamStoryQuery({
    teamId: "team",
    filters: parsed.filters,
    tab: "all",
    viewOptions: {
      groupBy: "status",
      orderBy: "created",
      orderDirection: "desc",
    },
  });
  assert.deepEqual(query.sprintIds, [SPRINT_ID]);
  assert.equal(query.objectiveId, OBJECTIVE_ID);
  assert.deepEqual(query.teamIds, ["team"]);
});

test("collection links select a facet without inventing an entity filter", () => {
  for (const openFilter of ["sprint", "objective"] as const) {
    assert.deepEqual(parseTeamStoryRouteFilters("team", { openFilter }), {
      filters: createTeamStoryFilters("team"),
      initialFacet: openFilter,
    });
  }
  const parsed = parseTeamStoryRouteFilters("team", {
    sprintId: SPRINT_ID,
    objectiveId: OBJECTIVE_ID,
    openFilter: "objective",
  });
  assert.equal(parsed.initialFacet, "objective");
  assert.equal(parsed.error, undefined);
  assert.equal(parsed.filters.objectiveId, OBJECTIVE_ID);
  assert.deepEqual(parsed.filters.sprintIds, [SPRINT_ID]);
});

test("empty, malformed and nil UUID filters are explicit route errors", () => {
  const invalidIds = [
    "",
    " ",
    "undefined",
    "null",
    "nil",
    "not-a-uuid",
    "00000000-0000-0000-0000-000000000000",
    ` ${SPRINT_ID}`,
    `${SPRINT_ID} `,
    `${SPRINT_ID}\n`,
    `${SPRINT_ID}\r`,
    SPRINT_ID.replaceAll("-", ""),
    `{${SPRINT_ID}}`,
    `urn:uuid:${SPRINT_ID}`,
    `${SPRINT_ID}/tasks`,
    `${SPRINT_ID}?all=true`,
    SPRINT_ID.replace("a", "g"),
  ];
  for (const key of ["sprintId", "objectiveId"] as const) {
    for (const value of invalidIds) {
      const parsed = parseTeamStoryRouteFilters("team", { [key]: value });
      assert.match(parsed.error ?? "", /invalid/, `${key}: ${value}`);
      assert.equal(parsed.initialFacet, undefined);
    }
  }
});

test("repeated parameters are rejected even when they agree or contain only one item", () => {
  for (const key of ["sprintId", "objectiveId", "openFilter"] as const) {
    const value = key === "openFilter" ? "sprint" : SPRINT_ID;
    for (const repeated of [[], [value], [value, value]]) {
      const parsed = parseTeamStoryRouteFilters("team", { [key]: repeated });
      assert.match(parsed.error ?? "", /invalid/);
      assert.equal(parsed.initialFacet, undefined);
    }
  }
});

test("invalid facet names are errors rather than a silently ignored instruction", () => {
  for (const openFilter of ["", "Sprint", "objective ", "assignee", "all"]) {
    assert.match(
      parseTeamStoryRouteFilters("team", { openFilter }).error ?? "",
      /invalid filter selection/,
    );
  }
});

test("one invalid part rejects the whole link without retaining a misleading partial filter", () => {
  for (const params of [
    { sprintId: SPRINT_ID, objectiveId: "invalid" },
    { sprintId: "invalid", objectiveId: OBJECTIVE_ID },
    { sprintId: SPRINT_ID, objectiveId: OBJECTIVE_ID, openFilter: "invalid" },
  ]) {
    const parsed = parseTeamStoryRouteFilters("team", params);
    assert.match(parsed.error ?? "", /invalid/);
    assert.deepEqual(parsed.filters, createTeamStoryFilters("team"));
    assert.equal(parsed.initialFacet, undefined);
  }
});

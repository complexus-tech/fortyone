import assert from "node:assert/strict";
import test from "node:test";
import {
  legacyTeamStoriesHref,
  teamStoriesHref,
} from "./team-story-navigation.ts";

test("legacy entity links retain their team and entity filter without opening old pages", () => {
  assert.deepEqual(
    legacyTeamStoriesHref({ teamId: "team", sprintId: "sprint" }, "sprint"),
    {
      pathname: "/teams/[teamId]",
      params: { teamId: "team", sprintId: "sprint" },
    },
  );
  assert.deepEqual(
    legacyTeamStoriesHref(
      { teamId: "team", objectiveId: "objective" },
      "objective",
    ),
    {
      pathname: "/teams/[teamId]",
      params: { teamId: "team", objectiveId: "objective" },
    },
  );
});

test("legacy collection links request the corresponding filter picker", () => {
  for (const kind of ["sprint", "objective"] as const) {
    assert.deepEqual(legacyTeamStoriesHref({ teamId: "team" }, kind, true), {
      pathname: "/teams/[teamId]",
      params: { teamId: "team", openFilter: kind },
    });
  }
});

test("ambiguous or incomplete legacy links do not broaden into unfiltered team results", () => {
  assert.equal(legacyTeamStoriesHref({ teamId: "team" }, "sprint"), "/");
  assert.equal(
    legacyTeamStoriesHref({ teamId: "", sprintId: "sprint" }, "sprint"),
    "/",
  );
  assert.equal(
    legacyTeamStoriesHref(
      { teamId: ["first", "second"], objectiveId: "objective" },
      "objective",
    ),
    "/",
  );
  assert.equal(
    legacyTeamStoriesHref(
      { teamId: "team", objectiveId: ["first", "second"] },
      "objective",
    ),
    "/",
  );
});

test("direct search destinations use the same contract without opening a picker", () => {
  assert.deepEqual(
    teamStoriesHref({ teamId: "team", objectiveId: "objective" }),
    {
      pathname: "/teams/[teamId]",
      params: { teamId: "team", objectiveId: "objective" },
    },
  );
});

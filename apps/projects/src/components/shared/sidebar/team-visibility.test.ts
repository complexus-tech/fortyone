/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  MAX_VISIBLE_SIDEBAR_TEAMS,
  partitionSidebarTeams,
} from "./team-visibility";

const teams = Array.from({ length: 6 }, (_, index) => ({
  id: `team-${index + 1}`,
  name: `Team ${index + 1}`,
}));

describe("partitionSidebarTeams", () => {
  it("shows at most four ordered teams and puts the rest in overflow", () => {
    const result = partitionSidebarTeams(teams);

    expect(MAX_VISIBLE_SIDEBAR_TEAMS).toBe(4);
    expect(result.hasPromotedActiveTeam).toBe(false);
    expect(result.visibleTeams.map((team) => team.id)).toEqual([
      "team-1",
      "team-2",
      "team-3",
      "team-4",
    ]);
    expect(result.overflowTeams.map((team) => team.id)).toEqual([
      "team-5",
      "team-6",
    ]);
  });

  it("keeps an active overflow team visible without exceeding the limit", () => {
    const result = partitionSidebarTeams(teams, "team-6");

    expect(result.visibleTeams.map((team) => team.id)).toEqual([
      "team-1",
      "team-2",
      "team-3",
      "team-6",
    ]);
    expect(result.overflowTeams.map((team) => team.id)).toEqual([
      "team-4",
      "team-5",
    ]);
    expect(result.hasPromotedActiveTeam).toBe(true);
  });

  it("does not create overflow when the user belongs to four teams", () => {
    const result = partitionSidebarTeams(teams.slice(0, 4));

    expect(result.visibleTeams).toHaveLength(4);
    expect(result.overflowTeams).toHaveLength(0);
  });
});

/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { reorderCachedTeams } from "./team-order-cache";

const teams = Array.from({ length: 4 }, (_, index) => ({
  id: `team-${index + 1}`,
  name: `Team ${index + 1}`,
}));

describe("reorderCachedTeams", () => {
  it("orders known teams and preserves unlisted cached teams", () => {
    expect(
      reorderCachedTeams(teams, ["team-3", "team-1"])?.map((team) => team.id),
    ).toEqual(["team-3", "team-1", "team-2", "team-4"]);
  });

  it("ignores requested ids that are not in the cache", () => {
    expect(
      reorderCachedTeams(teams, ["missing", "team-2"])?.map((team) => team.id),
    ).toEqual(["team-2", "team-1", "team-3", "team-4"]);
  });
});

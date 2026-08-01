/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { alignObjectiveInStrategy } from "./strategy-cache";
import type { StrategyMap } from "./types";

const strategy: StrategyMap = {
  description: null,
  pillars: [
    {
      description: null,
      id: "pillar-a",
      name: "Pillar A",
      objectiveIds: ["objective-1", "objective-2"],
      orderIndex: 0,
    },
    {
      description: null,
      id: "pillar-b",
      name: "Pillar B",
      objectiveIds: ["objective-3"],
      orderIndex: 1,
    },
  ],
  ultimateGoal: "Build a durable business",
};

describe("alignObjectiveInStrategy", () => {
  it("moves an objective between pillars without duplicating it", () => {
    const result = alignObjectiveInStrategy(
      strategy,
      "objective-1",
      "pillar-b",
    );

    expect(result.pillars[0]?.objectiveIds).toEqual(["objective-2"]);
    expect(result.pillars[1]?.objectiveIds).toEqual([
      "objective-3",
      "objective-1",
    ]);
  });

  it("removes an objective alignment", () => {
    const result = alignObjectiveInStrategy(strategy, "objective-1", null);

    expect(result.pillars[0]?.objectiveIds).toEqual(["objective-2"]);
    expect(result.pillars[1]?.objectiveIds).toEqual(["objective-3"]);
  });
});

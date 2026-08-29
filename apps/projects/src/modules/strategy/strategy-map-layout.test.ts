/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { KeyResult, Objective } from "@/modules/objectives/types";
import {
  createStrategyMapLayout,
  getObjectiveNodeId,
  getPillarNodeId,
  getStrategyConnections,
  GOAL_NODE_ID,
  mergeStrategyNodePositions,
  parseStoredStrategyNodePositions,
} from "./strategy-map-layout";
import type { StrategyMap } from "./types";

const objective = (id: string) => ({ id }) as Objective;

const strategy: StrategyMap = {
  ultimateGoal: "Ship the strategy",
  description: null,
  pillars: [
    {
      id: "pillar-1",
      name: "Product",
      description: null,
      objectiveIds: ["objective-1"],
      orderIndex: 0,
    },
  ],
};

describe("strategy map layout", () => {
  it("creates positions for goal, pillar, aligned, and unaligned objectives", () => {
    const layout = createStrategyMapLayout(strategy, [
      objective("objective-1"),
      objective("objective-2"),
    ]);

    expect(layout.positions[GOAL_NODE_ID]).toBeDefined();
    expect(layout.positions[getPillarNodeId("pillar-1")]).toBeDefined();
    expect(layout.positions[getObjectiveNodeId("objective-1")]).toBeDefined();
    expect(layout.positions[getObjectiveNodeId("objective-2")]).toBeDefined();
  });

  it("only connects aligned objectives", () => {
    const connections = getStrategyConnections(
      strategy,
      new Set(["objective-1", "objective-2"]),
    );

    expect(connections).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          sourceId: GOAL_NODE_ID,
          targetId: getPillarNodeId("pillar-1"),
        }),
        expect.objectContaining({
          sourceId: getPillarNodeId("pillar-1"),
          targetId: getObjectiveNodeId("objective-1"),
        }),
      ]),
    );
    expect(
      connections.some(
        (connection) =>
          connection.targetId === getObjectiveNodeId("objective-2"),
      ),
    ).toBe(false);
  });

  it("validates and merges stored positions without retaining removed nodes", () => {
    const parsed = parseStoredStrategyNodePositions(
      JSON.stringify({
        goal: { x: 20, y: 30 },
        invalid: { x: "no", y: 10 },
      }),
    );
    const merged = mergeStrategyNodePositions(
      { goal: { x: 1, y: 2 }, pillar: { x: 3, y: 4 } },
      parsed,
    );

    expect(merged).toEqual({
      goal: { x: 20, y: 30 },
      pillar: { x: 3, y: 4 },
    });
  });

  it("omits positions and connections for collapsed key results", () => {
    const keyResult = {
      id: "key-result-1",
      objectiveId: "objective-1",
    } as KeyResult;
    const keyResultsByObjective = new Map([["objective-1", [keyResult]]]);
    const layout = createStrategyMapLayout(
      strategy,
      [objective("objective-1")],
      keyResultsByObjective,
      new Set(),
    );
    const connections = getStrategyConnections(
      strategy,
      new Set(["objective-1"]),
      keyResultsByObjective,
      new Set(),
    );

    expect(layout.positions["key-result:key-result-1"]).toBeUndefined();
    expect(
      connections.some(
        ({ targetId }) => targetId === "key-result:key-result-1",
      ),
    ).toBe(false);
  });
});

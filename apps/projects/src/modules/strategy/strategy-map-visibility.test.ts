/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Objective } from "@/modules/objectives/types";
import type {
  StrategyConnection,
  StrategyNodeDimensions,
  StrategyNodePositions,
} from "./strategy-map-layout";
import { getObjectiveNodeId } from "./strategy-map-layout";
import {
  getPersistedKeyResultFetchObjectiveIds,
  getStrategyKeyResultFetchObjectiveIds,
  getVisibleStrategyConnections,
  getVisibleStrategyNodeIds,
  type StrategyMapViewport,
} from "./strategy-map-visibility";

const dimensions: StrategyNodeDimensions = { height: 160, width: 340 };
const getDimensions = () => dimensions;
const viewport: StrategyMapViewport = {
  height: 600,
  left: 0,
  top: 0,
  width: 600,
};

const objective = (id: string, keyResultCount: number) =>
  ({ id, keyResultCount }) as Objective;

describe("strategy map visibility", () => {
  it("renders visible nodes and explicitly pinned nodes, not distant nodes", () => {
    const positions: StrategyNodePositions = {
      distant: { x: 3000, y: 3000 },
      pinned: { x: 3000, y: 3000 },
      visible: { x: 100, y: 100 },
    };

    const visibleNodeIds = getVisibleStrategyNodeIds({
      alwaysVisibleNodeIds: new Set(["pinned"]),
      getDimensions,
      nodeIds: ["visible", "distant", "pinned"],
      positions,
      viewport,
    });

    expect(visibleNodeIds).toEqual(new Set(["visible", "pinned"]));
  });

  it("keeps connectors that cross the viewport while culling distant paths", () => {
    const positions: StrategyNodePositions = {
      "cross-source": { x: 100, y: -700 },
      "cross-target": { x: 100, y: 1200 },
      "distant-source": { x: 3000, y: 0 },
      "distant-target": { x: 3000, y: 1200 },
    };
    const connections: StrategyConnection[] = [
      {
        id: "crosses-viewport",
        sourceId: "cross-source",
        targetId: "cross-target",
      },
      {
        id: "distant",
        sourceId: "distant-source",
        targetId: "distant-target",
      },
    ];

    expect(
      getVisibleStrategyConnections({
        connections,
        getDimensions,
        positions,
        viewport,
      }),
    ).toEqual([connections[0]]);
  });

  it("fetches expanded key results only for objective cells in the viewport or pinned for interaction", () => {
    const visibleObjective = objective("visible", 3);
    const distantObjective = objective("distant", 3);
    const selectedObjective = objective("selected", 3);
    const positions: StrategyNodePositions = {
      [getObjectiveNodeId(visibleObjective.id)]: { x: 100, y: 100 },
      [getObjectiveNodeId(distantObjective.id)]: { x: 100, y: 3000 },
      [getObjectiveNodeId(selectedObjective.id)]: { x: 3000, y: 3000 },
    };

    const objectiveIds = getStrategyKeyResultFetchObjectiveIds({
      alwaysIncludeObjectiveIds: new Set([selectedObjective.id]),
      expandedObjectiveIds: new Set([
        visibleObjective.id,
        distantObjective.id,
        selectedObjective.id,
      ]),
      objectives: [visibleObjective, distantObjective, selectedObjective],
      positions,
      viewport,
    });

    expect(objectiveIds).toEqual(
      new Set([visibleObjective.id, selectedObjective.id]),
    );
  });

  it("discovers a persisted offscreen key result through its sparse owner index", () => {
    const ownerObjective = objective("owner", 1);
    const nodeId = "key-result:positioned-away-from-parent";

    expect(
      getPersistedKeyResultFetchObjectiveIds({
        expandedObjectiveIds: new Set([ownerObjective.id]),
        keyResultOwnerIds: new Map([[nodeId, ownerObjective.id]]),
        positions: { [nodeId]: { x: 120, y: 120 } },
        viewport,
      }),
    ).toEqual(new Set([ownerObjective.id]));
  });
});

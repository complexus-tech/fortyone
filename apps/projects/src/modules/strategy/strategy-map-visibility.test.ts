/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { Objective } from "@/modules/objectives/types";
import type {
  StrategyConnection,
  StrategyNodeDimensions,
  StrategyNodePositions,
} from "./strategy-map-layout";
import { getObjectiveNodeId } from "./strategy-map-layout";
import {
  createStrategyMapViewport,
  getRetainedStrategyMapViewport,
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
  it("keeps the safety margin constant in screen pixels at every zoom", () => {
    for (const zoom of [0.5, 1, 1.6]) {
      const required = createStrategyMapViewport({
        clientHeight: 800,
        clientWidth: 1200,
        scrollLeft: 2000,
        scrollTop: 2000,
        zoom,
      });

      expect(2000 - required.left * zoom).toBeCloseTo(320);
      expect(2000 - required.top * zoom).toBeCloseTo(320);
      expect(required.width * zoom - 1200).toBeCloseTo(640);
      expect(required.height * zoom - 800).toBeCloseTo(640);
    }
  });

  it("keeps the same render window through small pans and zoom changes", () => {
    const initial = createStrategyMapViewport({
      clientHeight: 800,
      clientWidth: 1200,
      scrollLeft: 2000,
      scrollTop: 2000,
      zoom: 1,
    });
    const retained = getRetainedStrategyMapViewport(null, initial);

    for (let step = 1; step <= 60; step++) {
      const required = createStrategyMapViewport({
        clientHeight: 800,
        clientWidth: 1200,
        scrollLeft: 2000 + step * 3,
        scrollTop: 2000 + step * 2,
        zoom: 1 + step * 0.001,
      });
      expect(getRetainedStrategyMapViewport(retained, required)).toBe(retained);
    }
  });

  it("refills before the visible area reaches an edge without growing over long pans", () => {
    let retained: StrategyMapViewport | null = null;

    for (let step = 0; step < 100; step++) {
      const required = createStrategyMapViewport({
        clientHeight: 800,
        clientWidth: 1200,
        scrollLeft: step * 1500,
        scrollTop: step * 1000,
        zoom: 0.5,
      });
      retained = getRetainedStrategyMapViewport(retained, required);

      expect(retained.left).toBeLessThan(required.left);
      expect(retained.top).toBeLessThan(required.top);
      expect(retained.left + retained.width).toBeGreaterThan(
        required.left + required.width,
      );
      expect(retained.top + retained.height).toBeGreaterThan(
        required.top + required.height,
      );
      expect(retained.width).toBe(required.width * 2);
      expect(retained.height).toBe(required.height * 2);
    }
  });

  it("shrinks an oversized retained window after zooming in", () => {
    const wide = { height: 2000, left: 0, top: 0, width: 3000 };
    const retained = getRetainedStrategyMapViewport(null, wide);
    const zoomed = { height: 600, left: 1000, top: 1000, width: 900 };
    const next = getRetainedStrategyMapViewport(retained, zoomed);

    expect(next).not.toBe(retained);
    expect(next.width).toBe(zoomed.width * 2);
    expect(next.height).toBe(zoomed.height * 2);
  });

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

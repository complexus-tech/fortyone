"use client";

import { useMemo } from "react";
import {
  getKeyResultNodeId,
  getObjectiveNodeId,
  getPillarNodeId,
  GOAL_NODE_ID,
  type StrategyNodeDimensions,
  type StrategyNodePositions,
} from "./strategy-map-layout";
import { getDefaultNodeDimensions } from "./strategy-map-canvas-renderers";
import type { StrategyMap } from "./types";
import type { StrategyMapViewport } from "./strategy-map-visibility";
import { getVisibleStrategyNodeIds } from "./strategy-map-visibility";

type StrategyMapRenderObjective = {
  id: string;
};

type StrategyMapRenderKeyResult = {
  id: string;
};

type UseStrategyMapCanvasRenderDataOptions<
  TObjective extends StrategyMapRenderObjective,
  TKeyResult extends StrategyMapRenderKeyResult,
> = {
  dimensions: Record<string, StrategyNodeDimensions>;
  draggingNodeId: string | null;
  draggingNodeIds: readonly string[];
  dropTargetPillarId: string | null;
  expandedObjectiveIds: ReadonlySet<string>;
  keyResultsByObjective: ReadonlyMap<string, TKeyResult[]>;
  objectives: readonly TObjective[];
  positions: StrategyNodePositions;
  selectedNodeId: string | null | undefined;
  showUnaligned: boolean;
  strategy: StrategyMap;
  viewport: StrategyMapViewport;
};

export const useStrategyMapCanvasRenderData = <
  TObjective extends StrategyMapRenderObjective,
  TKeyResult extends StrategyMapRenderKeyResult,
>({
  dimensions,
  draggingNodeId,
  draggingNodeIds,
  dropTargetPillarId,
  expandedObjectiveIds,
  keyResultsByObjective,
  objectives,
  positions,
  selectedNodeId,
  showUnaligned,
  strategy,
  viewport,
}: UseStrategyMapCanvasRenderDataOptions<TObjective, TKeyResult>) => {
  const objectiveById = useMemo(
    () => new Map(objectives.map((objective) => [objective.id, objective])),
    [objectives],
  );
  const pillarByObjectiveId = useMemo(() => {
    const result = new Map<string, string>();
    strategy.pillars.forEach((pillar) => {
      pillar.objectiveIds.forEach((objectiveId) => {
        result.set(objectiveId, pillar.id);
      });
    });
    return result;
  }, [strategy.pillars]);
  const objectiveIds = useMemo(() => {
    const result = new Set<string>();
    objectives.forEach((objective) => {
      if (showUnaligned || pillarByObjectiveId.has(objective.id)) {
        result.add(objective.id);
      }
    });
    return result;
  }, [objectives, pillarByObjectiveId, showUnaligned]);
  const renderableNodeIds = useMemo(() => {
    const ids = [
      GOAL_NODE_ID,
      ...strategy.pillars.map((pillar) => getPillarNodeId(pillar.id)),
    ];
    objectives.forEach((objective) => {
      if (objectiveIds.has(objective.id)) {
        ids.push(getObjectiveNodeId(objective.id));
        if (expandedObjectiveIds.has(objective.id)) {
          keyResultsByObjective.get(objective.id)?.forEach((keyResult) => {
            ids.push(getKeyResultNodeId(keyResult.id));
          });
        }
      }
    });
    return ids;
  }, [
    expandedObjectiveIds,
    keyResultsByObjective,
    objectiveIds,
    objectives,
    strategy.pillars,
  ]);
  const alwaysRenderedNodeIds = useMemo(() => {
    const ids = new Set(draggingNodeIds);
    if (draggingNodeId?.startsWith("objective:")) {
      const objectiveId = draggingNodeId.slice("objective:".length);
      keyResultsByObjective.get(objectiveId)?.forEach((keyResult) => {
        ids.add(getKeyResultNodeId(keyResult.id));
      });
    }
    if (selectedNodeId) ids.add(selectedNodeId);
    if (dropTargetPillarId) ids.add(getPillarNodeId(dropTargetPillarId));
    return ids;
  }, [
    draggingNodeId,
    draggingNodeIds,
    dropTargetPillarId,
    keyResultsByObjective,
    selectedNodeId,
  ]);
  const visibleNodeIds = useMemo(
    () =>
      getVisibleStrategyNodeIds({
        alwaysVisibleNodeIds: alwaysRenderedNodeIds,
        getDimensions: (nodeId) =>
          dimensions[nodeId] ?? getDefaultNodeDimensions(nodeId),
        nodeIds: renderableNodeIds,
        positions,
        viewport,
      }),
    [alwaysRenderedNodeIds, dimensions, positions, renderableNodeIds, viewport],
  );
  const visiblePillars = useMemo(
    () =>
      strategy.pillars.filter((pillar) =>
        visibleNodeIds.has(getPillarNodeId(pillar.id)),
      ),
    [strategy.pillars, visibleNodeIds],
  );
  const visibleObjectives = useMemo(
    () =>
      objectives.filter(
        (objective) =>
          objectiveIds.has(objective.id) &&
          visibleNodeIds.has(getObjectiveNodeId(objective.id)),
      ),
    [objectiveIds, objectives, visibleNodeIds],
  );
  const visibleKeyResultNodes = useMemo(() => {
    const nodes: { keyResult: TKeyResult; objective: TObjective }[] = [];

    keyResultsByObjective.forEach((keyResults, objectiveId) => {
      const objective = objectiveById.get(objectiveId);
      if (
        !objective ||
        !objectiveIds.has(objectiveId) ||
        !expandedObjectiveIds.has(objectiveId)
      ) {
        return;
      }

      keyResults.forEach((keyResult) => {
        if (!visibleNodeIds.has(getKeyResultNodeId(keyResult.id))) return;
        nodes.push({ keyResult, objective });
      });
    });

    return nodes;
  }, [
    expandedObjectiveIds,
    keyResultsByObjective,
    objectiveById,
    objectiveIds,
    visibleNodeIds,
  ]);
  const keyResultObjectiveIdByNodeId = useMemo(() => {
    const result = new Map<string, string>();
    keyResultsByObjective.forEach((keyResults, objectiveId) => {
      keyResults.forEach((keyResult) => {
        result.set(getKeyResultNodeId(keyResult.id), objectiveId);
      });
    });
    return result;
  }, [keyResultsByObjective]);

  return {
    alwaysRenderedNodeIds,
    keyResultObjectiveIdByNodeId,
    objectiveById,
    objectiveIds,
    pillarByObjectiveId,
    visibleKeyResultNodes,
    visibleNodeIds,
    visibleObjectives,
    visiblePillars,
  };
};

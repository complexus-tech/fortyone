import type { Objective } from "@/modules/objectives/types";
import type { StrategyMap } from "./types";

export const STRATEGY_CANVAS_MIN_WIDTH = 2200;
export const STRATEGY_CANVAS_MIN_HEIGHT = 1400;
export const GOAL_NODE_WIDTH = 440;
export const PILLAR_NODE_WIDTH = 340;
export const OBJECTIVE_NODE_WIDTH = 340;

const CANVAS_GUTTER = 160;
const BRANCH_GAP = 120;
const OBJECTIVE_GAP = 64;
const OBJECTIVE_ROW_GAP = 320;
const MAX_OBJECTIVES_PER_ROW = 4;

export type StrategyNodePosition = {
  x: number;
  y: number;
};

export type StrategyNodeDimensions = {
  height: number;
  width: number;
};

export type StrategyNodePositions = Record<string, StrategyNodePosition>;

export type StrategyConnection = {
  id: string;
  sourceId: string;
  targetId: string;
};

export const GOAL_NODE_ID = "goal";
export const getPillarNodeId = (pillarId: string) => `pillar:${pillarId}`;
export const getObjectiveNodeId = (objectiveId: string) =>
  `objective:${objectiveId}`;

const getObjectiveRows = (count: number) =>
  Math.max(1, Math.ceil(count / MAX_OBJECTIVES_PER_ROW));

export const createStrategyMapLayout = (
  strategy: StrategyMap,
  objectives: Objective[],
) => {
  const objectiveById = new Map(
    objectives.map((objective) => [objective.id, objective]),
  );
  const alignedIds = new Set(
    strategy.pillars.flatMap((pillar) => pillar.objectiveIds),
  );
  const unalignedObjectives = objectives.filter(
    (objective) => !alignedIds.has(objective.id),
  );
  const branchWidths = strategy.pillars.map((pillar) => {
    const objectiveCount = pillar.objectiveIds.filter((id) =>
      objectiveById.has(id),
    ).length;
    const visibleInFirstRow = Math.min(
      Math.max(objectiveCount, 1),
      MAX_OBJECTIVES_PER_ROW,
    );

    return Math.max(
      PILLAR_NODE_WIDTH,
      visibleInFirstRow * OBJECTIVE_NODE_WIDTH +
        Math.max(0, visibleInFirstRow - 1) * OBJECTIVE_GAP,
    );
  });
  const branchesWidth = branchWidths.reduce(
    (total, width, index) => total + width + (index > 0 ? BRANCH_GAP : 0),
    0,
  );
  const unalignedWidth =
    Math.min(Math.max(unalignedObjectives.length, 1), MAX_OBJECTIVES_PER_ROW) *
      (OBJECTIVE_NODE_WIDTH + OBJECTIVE_GAP) -
    OBJECTIVE_GAP;
  const width = Math.max(
    STRATEGY_CANVAS_MIN_WIDTH,
    branchesWidth + CANVAS_GUTTER * 2,
    unalignedWidth + CANVAS_GUTTER * 2,
  );
  const maxAlignedRows = Math.max(
    1,
    ...strategy.pillars.map((pillar) =>
      getObjectiveRows(
        pillar.objectiveIds.filter((id) => objectiveById.has(id)).length,
      ),
    ),
  );
  const unalignedStartY = 620 + maxAlignedRows * OBJECTIVE_ROW_GAP + 180;
  const height = Math.max(
    STRATEGY_CANVAS_MIN_HEIGHT,
    unalignedStartY +
      getObjectiveRows(unalignedObjectives.length) * OBJECTIVE_ROW_GAP +
      260,
  );
  const positions: StrategyNodePositions = {
    [GOAL_NODE_ID]: {
      x: (width - GOAL_NODE_WIDTH) / 2,
      y: 72,
    },
  };

  let branchX = (width - branchesWidth) / 2;
  strategy.pillars.forEach((pillar, pillarIndex) => {
    const branchWidth = branchWidths[pillarIndex] ?? PILLAR_NODE_WIDTH;
    const alignedObjectives = pillar.objectiveIds
      .map((id) => objectiveById.get(id))
      .filter((objective): objective is Objective => Boolean(objective));

    positions[getPillarNodeId(pillar.id)] = {
      x: branchX + (branchWidth - PILLAR_NODE_WIDTH) / 2,
      y: 340,
    };

    alignedObjectives.forEach((objective, objectiveIndex) => {
      const row = Math.floor(objectiveIndex / MAX_OBJECTIVES_PER_ROW);
      const rowStart = row * MAX_OBJECTIVES_PER_ROW;
      const rowCount = Math.min(
        MAX_OBJECTIVES_PER_ROW,
        alignedObjectives.length - rowStart,
      );
      const rowWidth =
        rowCount * OBJECTIVE_NODE_WIDTH +
        Math.max(0, rowCount - 1) * OBJECTIVE_GAP;
      const column = objectiveIndex % MAX_OBJECTIVES_PER_ROW;

      positions[getObjectiveNodeId(objective.id)] = {
        x:
          branchX +
          (branchWidth - rowWidth) / 2 +
          column * (OBJECTIVE_NODE_WIDTH + OBJECTIVE_GAP),
        y: 620 + row * OBJECTIVE_ROW_GAP,
      };
    });

    branchX += branchWidth + BRANCH_GAP;
  });

  const unalignedColumns = Math.min(
    Math.max(unalignedObjectives.length, 1),
    MAX_OBJECTIVES_PER_ROW,
  );
  const unalignedRowWidth =
    unalignedColumns * OBJECTIVE_NODE_WIDTH +
    Math.max(0, unalignedColumns - 1) * OBJECTIVE_GAP;

  unalignedObjectives.forEach((objective, index) => {
    const row = Math.floor(index / MAX_OBJECTIVES_PER_ROW);
    const column = index % MAX_OBJECTIVES_PER_ROW;
    positions[getObjectiveNodeId(objective.id)] = {
      x:
        (width - unalignedRowWidth) / 2 +
        column * (OBJECTIVE_NODE_WIDTH + OBJECTIVE_GAP),
      y: unalignedStartY + row * OBJECTIVE_ROW_GAP,
    };
  });

  return { height, positions, width };
};

export const getStrategyConnections = (
  strategy: StrategyMap,
  objectiveIds: Set<string>,
): StrategyConnection[] => {
  const connections: StrategyConnection[] = [];

  strategy.pillars.forEach((pillar) => {
    const pillarNodeId = getPillarNodeId(pillar.id);
    connections.push({
      id: `${GOAL_NODE_ID}->${pillarNodeId}`,
      sourceId: GOAL_NODE_ID,
      targetId: pillarNodeId,
    });

    pillar.objectiveIds.forEach((objectiveId) => {
      if (!objectiveIds.has(objectiveId)) return;
      const objectiveNodeId = getObjectiveNodeId(objectiveId);
      connections.push({
        id: `${pillarNodeId}->${objectiveNodeId}`,
        sourceId: pillarNodeId,
        targetId: objectiveNodeId,
      });
    });
  });

  return connections;
};

export const createConnectionPath = (
  sourcePosition: StrategyNodePosition,
  sourceDimensions: StrategyNodeDimensions,
  targetPosition: StrategyNodePosition,
  targetDimensions: StrategyNodeDimensions,
) => {
  const sourceX = sourcePosition.x + sourceDimensions.width / 2;
  const sourceY = sourcePosition.y + sourceDimensions.height;
  const targetX = targetPosition.x + targetDimensions.width / 2;
  const targetY = targetPosition.y;
  const middleY = sourceY + (targetY - sourceY) / 2;

  return {
    path: `M ${sourceX} ${sourceY} V ${middleY} H ${targetX} V ${targetY}`,
    sourceX,
    sourceY,
    targetX,
    targetY,
  };
};

export const mergeStrategyNodePositions = (
  defaults: StrategyNodePositions,
  stored: StrategyNodePositions,
) =>
  Object.fromEntries(
    Object.entries(defaults).map(([id, position]) => [
      id,
      stored[id] ?? position,
    ]),
  );

export const parseStoredStrategyNodePositions = (value: string | null) => {
  if (!value) return {};

  try {
    const parsed = JSON.parse(value) as unknown;
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return {};
    }

    return Object.fromEntries(
      Object.entries(parsed).filter(
        (entry): entry is [string, StrategyNodePosition] => {
          const position = entry[1] as Partial<StrategyNodePosition> | null;
          return (
            Boolean(position) &&
            Number.isFinite(position?.x) &&
            Number.isFinite(position?.y)
          );
        },
      ),
    );
  } catch {
    return {};
  }
};

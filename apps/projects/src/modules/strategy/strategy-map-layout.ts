import type { KeyResult, Objective } from "@/modules/objectives/types";
import type { StrategyMap } from "./types";

export const STRATEGY_CANVAS_MIN_WIDTH = 2200;
export const STRATEGY_CANVAS_MIN_HEIGHT = 1400;
export const GOAL_NODE_WIDTH = 440;
export const PILLAR_NODE_WIDTH = 340;
export const OBJECTIVE_NODE_WIDTH = 340;
export const KEY_RESULT_NODE_WIDTH = 280;

const CANVAS_GUTTER = 160;
const BRANCH_GAP = 120;
const OBJECTIVE_GAP = 64;
const OBJECTIVE_ROW_MIN_HEIGHT = 320;
const OBJECTIVE_ROW_GAP = 80;
const OBJECTIVE_TO_KEY_RESULT_OFFSET = 250;
const KEY_RESULT_GAP = 28;
const KEY_RESULT_ROW_GAP = 184;
const MAX_OBJECTIVES_PER_ROW = 4;
const MAX_KEY_RESULTS_PER_ROW = 2;

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
export const getKeyResultNodeId = (keyResultId: string) =>
  `key-result:${keyResultId}`;

type ObjectiveCell = {
  objective: Objective;
  keyResults: KeyResult[];
  width: number;
};

type ObjectiveRow = {
  cells: ObjectiveCell[];
  height: number;
  width: number;
};

const getVisibleKeyResults = (
  objectiveId: string,
  keyResultsByObjective: ReadonlyMap<string, KeyResult[]>,
  expandedObjectiveIds: ReadonlySet<string>,
) =>
  expandedObjectiveIds.has(objectiveId)
    ? keyResultsByObjective.get(objectiveId) ?? []
    : [];

const getObjectiveCellWidth = (keyResultCount: number) => {
  const visibleColumns = Math.min(
    Math.max(keyResultCount, 1),
    MAX_KEY_RESULTS_PER_ROW,
  );

  return Math.max(
    OBJECTIVE_NODE_WIDTH,
    visibleColumns * KEY_RESULT_NODE_WIDTH +
      Math.max(0, visibleColumns - 1) * KEY_RESULT_GAP,
  );
};

const createObjectiveRows = (
  objectives: Objective[],
  keyResultsByObjective: ReadonlyMap<string, KeyResult[]>,
  expandedObjectiveIds: ReadonlySet<string>,
): ObjectiveRow[] => {
  if (objectives.length === 0) return [];

  const rows: ObjectiveRow[] = [];
  for (
    let index = 0;
    index < objectives.length;
    index += MAX_OBJECTIVES_PER_ROW
  ) {
    const cells = objectives
      .slice(index, index + MAX_OBJECTIVES_PER_ROW)
      .map((objective) => {
        const keyResults = getVisibleKeyResults(
          objective.id,
          keyResultsByObjective,
          expandedObjectiveIds,
        );
        return {
          keyResults,
          objective,
          width: getObjectiveCellWidth(keyResults.length),
        };
      });
    const keyResultRows = Math.max(
      0,
      ...cells.map(({ keyResults }) =>
        Math.ceil(keyResults.length / MAX_KEY_RESULTS_PER_ROW),
      ),
    );

    rows.push({
      cells,
      height: Math.max(
        OBJECTIVE_ROW_MIN_HEIGHT,
        OBJECTIVE_TO_KEY_RESULT_OFFSET + keyResultRows * KEY_RESULT_ROW_GAP,
      ),
      width:
        cells.reduce((total, cell) => total + cell.width, 0) +
        Math.max(0, cells.length - 1) * OBJECTIVE_GAP,
    });
  }

  return rows;
};

const getRowsHeight = (rows: ObjectiveRow[]) =>
  rows.reduce(
    (total, row, index) =>
      total + row.height + (index > 0 ? OBJECTIVE_ROW_GAP : 0),
    0,
  );

const positionObjectiveRows = ({
  groupWidth,
  groupX,
  rows,
  startY,
  positions,
}: {
  groupWidth: number;
  groupX: number;
  rows: ObjectiveRow[];
  startY: number;
  positions: StrategyNodePositions;
}) => {
  let rowY = startY;
  rows.forEach((row) => {
    let cellX = groupX + (groupWidth - row.width) / 2;

    row.cells.forEach(({ keyResults, objective, width }) => {
      positions[getObjectiveNodeId(objective.id)] = {
        x: cellX + (width - OBJECTIVE_NODE_WIDTH) / 2,
        y: rowY,
      };

      keyResults.forEach((keyResult, keyResultIndex) => {
        const keyResultRow = Math.floor(
          keyResultIndex / MAX_KEY_RESULTS_PER_ROW,
        );
        const keyResultRowStart = keyResultRow * MAX_KEY_RESULTS_PER_ROW;
        const keyResultRowCount = Math.min(
          MAX_KEY_RESULTS_PER_ROW,
          keyResults.length - keyResultRowStart,
        );
        const keyResultRowWidth =
          keyResultRowCount * KEY_RESULT_NODE_WIDTH +
          Math.max(0, keyResultRowCount - 1) * KEY_RESULT_GAP;
        const keyResultColumn = keyResultIndex % MAX_KEY_RESULTS_PER_ROW;

        positions[getKeyResultNodeId(keyResult.id)] = {
          x:
            cellX +
            (width - keyResultRowWidth) / 2 +
            keyResultColumn * (KEY_RESULT_NODE_WIDTH + KEY_RESULT_GAP),
          y:
            rowY +
            OBJECTIVE_TO_KEY_RESULT_OFFSET +
            keyResultRow * KEY_RESULT_ROW_GAP,
        };
      });

      cellX += width + OBJECTIVE_GAP;
    });

    rowY += row.height + OBJECTIVE_ROW_GAP;
  });
};

export const createStrategyMapLayout = (
  strategy: StrategyMap,
  objectives: Objective[],
  keyResultsByObjective: ReadonlyMap<string, KeyResult[]> = new Map(),
  expandedObjectiveIds: ReadonlySet<string> = new Set(),
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
  const branchRows = strategy.pillars.map((pillar) =>
    createObjectiveRows(
      pillar.objectiveIds
        .map((id) => objectiveById.get(id))
        .filter((objective): objective is Objective => Boolean(objective)),
      keyResultsByObjective,
      expandedObjectiveIds,
    ),
  );
  const branchWidths = branchRows.map((rows) =>
    Math.max(PILLAR_NODE_WIDTH, ...rows.map(({ width }) => width)),
  );
  const branchesWidth = branchWidths.reduce(
    (total, width, index) => total + width + (index > 0 ? BRANCH_GAP : 0),
    0,
  );
  const unalignedRows = createObjectiveRows(
    unalignedObjectives,
    keyResultsByObjective,
    expandedObjectiveIds,
  );
  const unalignedWidth = Math.max(
    OBJECTIVE_NODE_WIDTH,
    ...unalignedRows.map(({ width }) => width),
  );
  const width = Math.max(
    STRATEGY_CANVAS_MIN_WIDTH,
    branchesWidth + CANVAS_GUTTER * 2,
    unalignedWidth + CANVAS_GUTTER * 2,
  );
  const maxAlignedHeight = Math.max(
    OBJECTIVE_ROW_MIN_HEIGHT,
    ...branchRows.map(getRowsHeight),
  );
  const unalignedStartY = 620 + maxAlignedHeight + 180;
  const height = Math.max(
    STRATEGY_CANVAS_MIN_HEIGHT,
    unalignedStartY + getRowsHeight(unalignedRows) + 260,
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
    positions[getPillarNodeId(pillar.id)] = {
      x: branchX + (branchWidth - PILLAR_NODE_WIDTH) / 2,
      y: 340,
    };
    positionObjectiveRows({
      groupWidth: branchWidth,
      groupX: branchX,
      positions,
      rows: branchRows[pillarIndex] ?? [],
      startY: 620,
    });
    branchX += branchWidth + BRANCH_GAP;
  });

  positionObjectiveRows({
    groupWidth: unalignedWidth,
    groupX: (width - unalignedWidth) / 2,
    positions,
    rows: unalignedRows,
    startY: unalignedStartY,
  });

  return { height, positions, width };
};

export const getStrategyConnections = (
  strategy: StrategyMap,
  objectiveIds: ReadonlySet<string>,
  keyResultsByObjective: ReadonlyMap<string, KeyResult[]> = new Map(),
  expandedObjectiveIds: ReadonlySet<string> = new Set(),
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

  objectiveIds.forEach((objectiveId) => {
    getVisibleKeyResults(
      objectiveId,
      keyResultsByObjective,
      expandedObjectiveIds,
    ).forEach((keyResult) => {
      const objectiveNodeId = getObjectiveNodeId(objectiveId);
      const keyResultNodeId = getKeyResultNodeId(keyResult.id);
      connections.push({
        id: `${objectiveNodeId}->${keyResultNodeId}`,
        sourceId: objectiveNodeId,
        targetId: keyResultNodeId,
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

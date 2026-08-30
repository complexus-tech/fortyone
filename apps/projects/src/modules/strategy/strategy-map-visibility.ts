import {
  createConnectionPath,
  getObjectiveNodeId,
  getStrategyObjectiveCellDimensions,
  KEY_RESULT_NODE_HEIGHT,
  KEY_RESULT_NODE_WIDTH,
  OBJECTIVE_NODE_WIDTH,
  type StrategyConnection,
  type StrategyNodeDimensions,
  type StrategyNodePositions,
} from "./strategy-map-layout";

type StrategyObjectiveGeometry = {
  id: string;
  keyResultCount: number;
};

export type StrategyMapViewport = {
  height: number;
  left: number;
  top: number;
  width: number;
};

type StrategyMapBounds = StrategyMapViewport;

export const STRATEGY_MAP_VIEWPORT_OVERSCAN = 320;

const INITIAL_VIEWPORT_HEIGHT = 1000;
const INITIAL_VIEWPORT_WIDTH = 1440;

const getViewportRight = (viewport: StrategyMapViewport) =>
  viewport.left + viewport.width;
const getViewportBottom = (viewport: StrategyMapViewport) =>
  viewport.top + viewport.height;

export const createStrategyMapViewport = ({
  clientHeight,
  clientWidth,
  scrollLeft,
  scrollTop,
  zoom,
}: {
  clientHeight: number;
  clientWidth: number;
  scrollLeft: number;
  scrollTop: number;
  zoom: number;
}): StrategyMapViewport => {
  const scale = Math.max(zoom, Number.EPSILON);

  return {
    height: clientHeight / scale + STRATEGY_MAP_VIEWPORT_OVERSCAN * 2,
    left: scrollLeft / scale - STRATEGY_MAP_VIEWPORT_OVERSCAN,
    top: scrollTop / scale - STRATEGY_MAP_VIEWPORT_OVERSCAN,
    width: clientWidth / scale + STRATEGY_MAP_VIEWPORT_OVERSCAN * 2,
  };
};

export const getInitialStrategyMapViewport = (
  layoutWidth: number,
  zoom: number,
): StrategyMapViewport => {
  const scale = Math.max(zoom, Number.EPSILON);
  const width = INITIAL_VIEWPORT_WIDTH / scale;
  const height = INITIAL_VIEWPORT_HEIGHT / scale;

  return {
    height: height + STRATEGY_MAP_VIEWPORT_OVERSCAN * 2,
    left: (layoutWidth - width) / 2 - STRATEGY_MAP_VIEWPORT_OVERSCAN,
    top: -STRATEGY_MAP_VIEWPORT_OVERSCAN,
    width: width + STRATEGY_MAP_VIEWPORT_OVERSCAN * 2,
  };
};

export const isStrategyMapBoundsVisible = (
  bounds: StrategyMapBounds,
  viewport: StrategyMapViewport,
) =>
  bounds.left < getViewportRight(viewport) &&
  bounds.left + bounds.width > viewport.left &&
  bounds.top < getViewportBottom(viewport) &&
  bounds.top + bounds.height > viewport.top;

const getNodeBounds = (
  nodeId: string,
  positions: StrategyNodePositions,
  getDimensions: (nodeId: string) => StrategyNodeDimensions,
): StrategyMapBounds | null => {
  const position = Object.hasOwn(positions, nodeId)
    ? positions[nodeId]
    : undefined;
  if (!position) return null;

  const dimensions = getDimensions(nodeId);
  return {
    height: dimensions.height,
    left: position.x,
    top: position.y,
    width: dimensions.width,
  };
};

export const getVisibleStrategyNodeIds = ({
  alwaysVisibleNodeIds = new Set<string>(),
  getDimensions,
  nodeIds,
  positions,
  viewport,
}: {
  alwaysVisibleNodeIds?: ReadonlySet<string>;
  getDimensions: (nodeId: string) => StrategyNodeDimensions;
  nodeIds: readonly string[];
  positions: StrategyNodePositions;
  viewport: StrategyMapViewport;
}) => {
  const visibleNodeIds = new Set(alwaysVisibleNodeIds);

  nodeIds.forEach((nodeId) => {
    const bounds = getNodeBounds(nodeId, positions, getDimensions);
    if (bounds && isStrategyMapBoundsVisible(bounds, viewport)) {
      visibleNodeIds.add(nodeId);
    }
  });

  return visibleNodeIds;
};

const getObjectiveCellBounds = (
  objective: StrategyObjectiveGeometry,
  positions: StrategyNodePositions,
) => {
  const objectiveNodeId = getObjectiveNodeId(objective.id);
  const position = Object.hasOwn(positions, objectiveNodeId)
    ? positions[objectiveNodeId]
    : undefined;
  if (!position) return null;

  const dimensions = getStrategyObjectiveCellDimensions(
    objective.keyResultCount,
  );
  return {
    height: dimensions.height,
    left: position.x - (dimensions.width - OBJECTIVE_NODE_WIDTH) / 2,
    top: position.y,
    width: dimensions.width,
  };
};

export const getStrategyKeyResultFetchObjectiveIds = ({
  alwaysIncludeObjectiveIds = new Set<string>(),
  expandedObjectiveIds,
  objectives,
  positions,
  viewport,
}: {
  alwaysIncludeObjectiveIds?: ReadonlySet<string>;
  expandedObjectiveIds: ReadonlySet<string>;
  objectives: readonly StrategyObjectiveGeometry[];
  positions: StrategyNodePositions;
  viewport: StrategyMapViewport;
}) => {
  const objectiveIds = new Set<string>();

  objectives.forEach((objective) => {
    if (alwaysIncludeObjectiveIds.has(objective.id)) {
      objectiveIds.add(objective.id);
      return;
    }
    if (!expandedObjectiveIds.has(objective.id)) return;

    const bounds = getObjectiveCellBounds(objective, positions);
    if (bounds && isStrategyMapBoundsVisible(bounds, viewport)) {
      objectiveIds.add(objective.id);
    }
  });

  return objectiveIds;
};

export const getPersistedKeyResultFetchObjectiveIds = ({
  expandedObjectiveIds,
  keyResultOwnerIds,
  positions,
  viewport,
}: {
  expandedObjectiveIds: ReadonlySet<string>;
  keyResultOwnerIds: ReadonlyMap<string, string>;
  positions: StrategyNodePositions;
  viewport: StrategyMapViewport;
}) => {
  const objectiveIds = new Set<string>();

  keyResultOwnerIds.forEach((objectiveId, keyResultNodeId) => {
    if (!expandedObjectiveIds.has(objectiveId)) return;

    const position = Object.hasOwn(positions, keyResultNodeId)
      ? positions[keyResultNodeId]
      : undefined;
    if (!position) return;

    if (
      isStrategyMapBoundsVisible(
        {
          height: KEY_RESULT_NODE_HEIGHT,
          left: position.x,
          top: position.y,
          width: KEY_RESULT_NODE_WIDTH,
        },
        viewport,
      )
    ) {
      objectiveIds.add(objectiveId);
    }
  });

  return objectiveIds;
};

const isHorizontalSegmentVisible = (
  startX: number,
  endX: number,
  y: number,
  viewport: StrategyMapViewport,
) =>
  y >= viewport.top &&
  y <= getViewportBottom(viewport) &&
  Math.max(Math.min(startX, endX), viewport.left) <=
    Math.min(Math.max(startX, endX), getViewportRight(viewport));

const isVerticalSegmentVisible = (
  x: number,
  startY: number,
  endY: number,
  viewport: StrategyMapViewport,
) =>
  x >= viewport.left &&
  x <= getViewportRight(viewport) &&
  Math.max(Math.min(startY, endY), viewport.top) <=
    Math.min(Math.max(startY, endY), getViewportBottom(viewport));

export const isStrategyConnectionVisible = ({
  alwaysVisibleNodeIds = new Set<string>(),
  connection,
  getDimensions,
  positions,
  viewport,
}: {
  alwaysVisibleNodeIds?: ReadonlySet<string>;
  connection: StrategyConnection;
  getDimensions: (nodeId: string) => StrategyNodeDimensions;
  positions: StrategyNodePositions;
  viewport: StrategyMapViewport;
}) => {
  if (
    alwaysVisibleNodeIds.has(connection.sourceId) ||
    alwaysVisibleNodeIds.has(connection.targetId)
  ) {
    return true;
  }

  const sourcePosition = Object.hasOwn(positions, connection.sourceId)
    ? positions[connection.sourceId]
    : undefined;
  const targetPosition = Object.hasOwn(positions, connection.targetId)
    ? positions[connection.targetId]
    : undefined;
  if (!sourcePosition || !targetPosition) return false;

  const path = createConnectionPath(
    sourcePosition,
    getDimensions(connection.sourceId),
    targetPosition,
    getDimensions(connection.targetId),
  );

  return (
    isVerticalSegmentVisible(
      path.sourceX,
      path.sourceY,
      (path.sourceY + path.targetY) / 2,
      viewport,
    ) ||
    isHorizontalSegmentVisible(
      path.sourceX,
      path.targetX,
      (path.sourceY + path.targetY) / 2,
      viewport,
    ) ||
    isVerticalSegmentVisible(
      path.targetX,
      (path.sourceY + path.targetY) / 2,
      path.targetY,
      viewport,
    )
  );
};

export const getVisibleStrategyConnections = ({
  alwaysVisibleNodeIds,
  connections,
  getDimensions,
  positions,
  viewport,
}: {
  alwaysVisibleNodeIds?: ReadonlySet<string>;
  connections: readonly StrategyConnection[];
  getDimensions: (nodeId: string) => StrategyNodeDimensions;
  positions: StrategyNodePositions;
  viewport: StrategyMapViewport;
}) =>
  connections.filter((connection) =>
    isStrategyConnectionVisible({
      alwaysVisibleNodeIds,
      connection,
      getDimensions,
      positions,
      viewport,
    }),
  );

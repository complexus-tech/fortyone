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

export const getRetainedStrategyMapViewport = (
  current: StrategyMapViewport | null,
  required: StrategyMapViewport,
): StrategyMapViewport => {
  if (
    current &&
    required.left >= current.left &&
    required.top >= current.top &&
    getViewportRight(required) <= getViewportRight(current) &&
    getViewportBottom(required) <= getViewportBottom(current) &&
    current.width <= required.width * 3 &&
    current.height <= required.height * 3
  ) {
    return current;
  }

  // Keep a reserve beyond the already overscanned viewport. Ordinary camera
  // movement reuses the same mounted cards and paths; refill before they enter
  // view. Replace, rather than union, windows so long pans remain bounded.
  const horizontalReserve = required.width / 2;
  const verticalReserve = required.height / 2;

  return {
    height: required.height + verticalReserve * 2,
    left: required.left - horizontalReserve,
    top: required.top - verticalReserve,
    width: required.width + horizontalReserve * 2,
  };
};

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
  const overscan = STRATEGY_MAP_VIEWPORT_OVERSCAN / scale;

  return {
    height: clientHeight / scale + overscan * 2,
    left: scrollLeft / scale - overscan,
    top: scrollTop / scale - overscan,
    width: clientWidth / scale + overscan * 2,
  };
};

export const getInitialStrategyMapViewport = (
  layoutWidth: number,
  zoom: number,
): StrategyMapViewport => {
  const scale = Math.max(zoom, Number.EPSILON);
  const width = INITIAL_VIEWPORT_WIDTH / scale;
  const height = INITIAL_VIEWPORT_HEIGHT / scale;
  const overscan = STRATEGY_MAP_VIEWPORT_OVERSCAN / scale;

  return {
    height: height + overscan * 2,
    left: (layoutWidth - width) / 2 - overscan,
    top: -overscan,
    width: width + overscan * 2,
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

export const getStrategyConnectionPath = ({
  connection,
  getDimensions,
  positions,
}: {
  connection: StrategyConnection;
  getDimensions: (nodeId: string) => StrategyNodeDimensions;
  positions: StrategyNodePositions;
}) => {
  const sourcePosition = Object.hasOwn(positions, connection.sourceId)
    ? positions[connection.sourceId]
    : undefined;
  const targetPosition = Object.hasOwn(positions, connection.targetId)
    ? positions[connection.targetId]
    : undefined;
  if (!sourcePosition || !targetPosition) return null;

  return createConnectionPath(
    sourcePosition,
    getDimensions(connection.sourceId),
    targetPosition,
    getDimensions(connection.targetId),
  );
};

export const isStrategyConnectionPathVisible = ({
  alwaysVisibleNodeIds,
  connection,
  path,
  viewport,
}: {
  alwaysVisibleNodeIds?: ReadonlySet<string>;
  connection: StrategyConnection;
  path: ReturnType<typeof createConnectionPath> | null;
  viewport: StrategyMapViewport;
}) => {
  if (
    alwaysVisibleNodeIds?.has(connection.sourceId) ||
    alwaysVisibleNodeIds?.has(connection.targetId)
  ) {
    return true;
  }
  if (!path) return false;

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

export const isStrategyConnectionVisible = ({
  alwaysVisibleNodeIds,
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
}) =>
  isStrategyConnectionPathVisible({
    alwaysVisibleNodeIds,
    connection,
    path: getStrategyConnectionPath({ connection, getDimensions, positions }),
    viewport,
  });

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

"use client";

import type { PointerEvent as ReactPointerEvent, ReactNode } from "react";
import { memo, useMemo } from "react";
import { cn } from "lib";
import {
  getStrategyConnections,
  GOAL_NODE_ID,
  GOAL_NODE_WIDTH,
  KEY_RESULT_NODE_HEIGHT,
  KEY_RESULT_NODE_WIDTH,
  OBJECTIVE_NODE_WIDTH,
  PILLAR_NODE_WIDTH,
  type StrategyNodeDimensions,
  type StrategyNodePosition,
  type StrategyNodePositions,
  type StrategyKeyResultNode,
} from "./strategy-map-layout";
import type { StrategyMap } from "./types";
import {
  getStrategyConnectionPath,
  isStrategyConnectionPathVisible,
  type StrategyMapViewport,
} from "./strategy-map-visibility";

export const isInteractiveTarget = (target: EventTarget | null) => {
  if (!(target instanceof HTMLElement)) return false;
  if (target.closest("[data-card-select]")) return false;

  return Boolean(
    target.closest(
      "a, button, input, select, textarea, [role='menuitem'], [data-no-drag]",
    ),
  );
};

export const getDefaultNodeDimensions = (
  nodeId: string,
): StrategyNodeDimensions => {
  if (nodeId === GOAL_NODE_ID) {
    return { height: 196, width: GOAL_NODE_WIDTH };
  }
  if (nodeId.startsWith("pillar:")) {
    return { height: 150, width: PILLAR_NODE_WIDTH };
  }
  if (nodeId.startsWith("key-result:")) {
    return { height: KEY_RESULT_NODE_HEIGHT, width: KEY_RESULT_NODE_WIDTH };
  }
  return { height: 174, width: OBJECTIVE_NODE_WIDTH };
};

export const getNodePosition = (
  positions: StrategyNodePositions,
  nodeId: string,
): StrategyNodePosition | undefined =>
  Object.hasOwn(positions, nodeId) ? positions[nodeId] : undefined;

export const StrategyCanvasNode = ({
  children,
  id,
  isDragging,
  onPointerDown,
  onPointerMove,
  onPointerUp,
  position,
}: {
  children: ReactNode;
  id: string;
  isDragging: boolean;
  onPointerDown: (id: string, event: ReactPointerEvent<HTMLDivElement>) => void;
  onPointerMove: (event: ReactPointerEvent<HTMLDivElement>) => void;
  onPointerUp: (event: ReactPointerEvent<HTMLDivElement>) => void;
  position: StrategyNodePosition;
}) => (
  <div
    className={cn(
      "group/node absolute z-10 cursor-grab touch-none select-none active:cursor-grabbing",
      isDragging && "z-30 cursor-grabbing",
    )}
    data-dragging={isDragging}
    data-node-id={id}
    onPointerCancel={onPointerUp}
    onPointerDown={(event) => {
      onPointerDown(id, event);
    }}
    onPointerMove={onPointerMove}
    onPointerUp={onPointerUp}
    style={{ left: position.x, top: position.y }}
  >
    {children}
  </div>
);

export const CanvasConnections = memo(
  ({
    alwaysVisibleNodeIds,
    dimensions,
    expandedObjectiveIds,
    keyResultsByObjective,
    objectiveIds,
    positions,
    strategy,
    viewport,
  }: {
    alwaysVisibleNodeIds: ReadonlySet<string>;
    dimensions: Record<string, StrategyNodeDimensions>;
    expandedObjectiveIds: ReadonlySet<string>;
    keyResultsByObjective: ReadonlyMap<string, StrategyKeyResultNode[]>;
    objectiveIds: ReadonlySet<string>;
    positions: StrategyNodePositions;
    strategy: StrategyMap;
    viewport: StrategyMapViewport;
  }) => {
    const connections = useMemo(
      () =>
        getStrategyConnections(
          strategy,
          objectiveIds,
          keyResultsByObjective,
          expandedObjectiveIds,
        ),
      [expandedObjectiveIds, keyResultsByObjective, objectiveIds, strategy],
    );
    // Edge geometry depends on node positions and dimensions, not the camera.
    // Reuse each path for culling and drawing rather than rebuilding it per pan.
    const connectionPaths = useMemo(
      () =>
        connections.map((connection) => ({
          connection,
          path: getStrategyConnectionPath({
            connection,
            getDimensions: (nodeId) =>
              dimensions[nodeId] ?? getDefaultNodeDimensions(nodeId),
            positions,
          }),
        })),
      [connections, dimensions, positions],
    );
    const visibleConnections = useMemo(
      () =>
        connectionPaths.filter(({ connection, path }) =>
          isStrategyConnectionPathVisible({
            alwaysVisibleNodeIds,
            connection,
            path,
            viewport,
          }),
        ),
      [alwaysVisibleNodeIds, connectionPaths, viewport],
    );

    return (
      <svg
        aria-hidden
        className="pointer-events-none absolute inset-0 h-full w-full overflow-visible"
      >
        <g
          fill="none"
          stroke="var(--color-border-strong)"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2.25"
        >
          {visibleConnections.map(({ connection, path }) => {
            if (!path) return null;

            return (
              <g key={connection.id}>
                <path d={path.path} vectorEffect="non-scaling-stroke" />
                <circle
                  cx={path.targetX}
                  cy={path.targetY}
                  fill="var(--color-text-muted)"
                  r="3"
                  stroke="none"
                />
              </g>
            );
          })}
        </g>
      </svg>
    );
  },
);
CanvasConnections.displayName = "CanvasConnections";

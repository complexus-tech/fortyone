"use client";

import type { PointerEvent as ReactPointerEvent, ReactNode } from "react";
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { cn } from "lib";
import { Box, Flex, Text } from "ui";
import { useWorkspacePath } from "@/hooks";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useUpdateObjectiveMutation } from "@/modules/objectives/hooks";
import type { Objective, ObjectiveUpdate } from "@/modules/objectives/types";
import { useTeams } from "@/modules/teams/hooks/teams";
import {
  getObjectiveProgress,
  ObjectiveNodeCard,
  PillarNodeCard,
  UltimateGoalNodeCard,
} from "./strategy-map-cards";
import {
  createConnectionPath,
  createStrategyMapLayout,
  getObjectiveNodeId,
  getPillarNodeId,
  getStrategyConnections,
  GOAL_NODE_ID,
  GOAL_NODE_WIDTH,
  mergeStrategyNodePositions,
  OBJECTIVE_NODE_WIDTH,
  parseStoredStrategyNodePositions,
  PILLAR_NODE_WIDTH,
  type StrategyNodeDimensions,
  type StrategyNodePositions,
} from "./strategy-map-layout";
import type { StrategicPillar, StrategyMap } from "./types";

const LAYOUT_STORAGE_VERSION = 1;
const CANVAS_INSET = 28;
const CLICK_MOVEMENT_THRESHOLD = 4;
const subscribeToStaticStorage = () => () => undefined;

type ActiveNodeDrag = {
  id: string;
  pointerId: number;
  startClientX: number;
  startClientY: number;
  startPosition: { x: number; y: number };
};

type ActivePan = {
  pointerId: number;
  startClientX: number;
  startClientY: number;
  startScrollLeft: number;
  startScrollTop: number;
};

const isInteractiveTarget = (target: EventTarget | null) => {
  if (!(target instanceof HTMLElement)) return false;
  if (target.closest("[data-card-select]")) return false;

  return Boolean(
    target.closest(
      "a, button, input, select, textarea, [role='menuitem'], [data-no-drag]",
    ),
  );
};

const getDefaultNodeDimensions = (nodeId: string): StrategyNodeDimensions => {
  if (nodeId === GOAL_NODE_ID) {
    return { height: 196, width: GOAL_NODE_WIDTH };
  }
  if (nodeId.startsWith("pillar:")) {
    return { height: 150, width: PILLAR_NODE_WIDTH };
  }
  return { height: 174, width: OBJECTIVE_NODE_WIDTH };
};

const StrategyCanvasNode = ({
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
  position: { x: number; y: number };
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

const CanvasConnections = ({
  dimensions,
  objectiveIds,
  positions,
  strategy,
}: {
  dimensions: Record<string, StrategyNodeDimensions>;
  objectiveIds: Set<string>;
  positions: StrategyNodePositions;
  strategy: StrategyMap;
}) => {
  const connections = getStrategyConnections(strategy, objectiveIds);

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
        {connections.map((connection) => {
          const sourcePosition = positions[connection.sourceId];
          const targetPosition = positions[connection.targetId];

          const path = createConnectionPath(
            sourcePosition,
            dimensions[connection.sourceId] ??
              getDefaultNodeDimensions(connection.sourceId),
            targetPosition,
            dimensions[connection.targetId] ??
              getDefaultNodeDimensions(connection.targetId),
          );

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
};

export const StrategyMapCanvas = ({
  strategy,
  objectives,
  showUnaligned,
  onAlign,
  onAddPillar,
  onDeletePillar,
  onSelectGoal,
  onSelectObjective,
  onSelectPillar,
  canEdit,
  resetSignal = 0,
  zoom,
}: {
  strategy: StrategyMap;
  objectives: Objective[];
  showUnaligned: boolean;
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  onAddPillar: () => void;
  onDeletePillar: (pillarId: string) => void;
  onSelectGoal: () => void;
  onSelectObjective: (objective: Objective) => void;
  onSelectPillar: (pillar: StrategicPillar) => void;
  canEdit: boolean;
  resetSignal?: number;
  zoom: number;
}) => {
  const { workspaceSlug } = useWorkspacePath();
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: teams = [] } = useTeams();
  const updateObjective = useUpdateObjectiveMutation();
  const viewportRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLDivElement>(null);
  const activeDragRef = useRef<ActiveNodeDrag | null>(null);
  const activePanRef = useRef<ActivePan | null>(null);
  const positionsRef = useRef<StrategyNodePositions>({});
  const dimensionsRef = useRef<Record<string, StrategyNodeDimensions>>({});
  const hasPositionedViewportRef = useRef(false);
  const previousZoomRef = useRef(zoom);
  const previousResetSignalRef = useRef(resetSignal);
  const [draggingNodeId, setDraggingNodeId] = useState<string | null>(null);
  const [dropTargetPillarId, setDropTargetPillarId] = useState<string | null>(
    null,
  );
  const [isPanning, setIsPanning] = useState(false);
  const [dimensions, setDimensions] = useState<
    Record<string, StrategyNodeDimensions>
  >({});
  const layout = useMemo(
    () => createStrategyMapLayout(strategy, objectives),
    [objectives, strategy],
  );
  const storageKey = `strategy-map-layout:v${LAYOUT_STORAGE_VERSION}:${workspaceSlug}`;
  const getStoredLayoutSnapshot = useCallback(
    () => window.localStorage.getItem(storageKey),
    [storageKey],
  );
  const storedLayoutValue = useSyncExternalStore(
    subscribeToStaticStorage,
    getStoredLayoutSnapshot,
    () => null,
  );
  const storedPositions = useMemo(
    () => parseStoredStrategyNodePositions(storedLayoutValue),
    [storedLayoutValue],
  );
  const [transientLayout, setTransientLayout] = useState<{
    positions: StrategyNodePositions;
    storageKey: string;
  }>({ positions: {}, storageKey });
  const positions = useMemo(() => {
    const transientPositions =
      transientLayout.storageKey === storageKey
        ? transientLayout.positions
        : {};

    return mergeStrategyNodePositions(layout.positions, {
      ...storedPositions,
      ...transientPositions,
    });
  }, [layout.positions, storageKey, storedPositions, transientLayout]);
  const objectiveIds = useMemo(
    () => new Set(objectives.map((objective) => objective.id)),
    [objectives],
  );
  const objectiveById = useMemo(
    () => new Map(objectives.map((objective) => [objective.id, objective])),
    [objectives],
  );
  const statusById = useMemo(
    () => new Map(statuses.map((status) => [status.id, status])),
    [statuses],
  );
  const teamCodeById = useMemo(
    () => new Map(teams.map((team) => [team.id, team.code])),
    [teams],
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
  const averageProgress =
    objectives.length > 0
      ? Math.round(
          objectives.reduce(
            (total, objective) => total + getObjectiveProgress(objective),
            0,
          ) / objectives.length,
        )
      : 0;

  useLayoutEffect(() => {
    positionsRef.current = positions;
  }, [positions]);

  const renderedNodeIds = useMemo(() => {
    const ids = [
      GOAL_NODE_ID,
      ...strategy.pillars.map((pillar) => getPillarNodeId(pillar.id)),
    ];
    objectives.forEach((objective) => {
      if (showUnaligned || pillarByObjectiveId.has(objective.id)) {
        ids.push(getObjectiveNodeId(objective.id));
      }
    });
    return ids.join("|");
  }, [objectives, pillarByObjectiveId, showUnaligned, strategy.pillars]);

  useLayoutEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const measureNodes = () => {
      const nextDimensions: Record<string, StrategyNodeDimensions> = {};
      canvas.querySelectorAll<HTMLElement>("[data-node-id]").forEach((node) => {
        const nodeId = node.dataset.nodeId;
        if (!nodeId) return;
        nextDimensions[nodeId] = {
          height: node.offsetHeight,
          width: node.offsetWidth,
        };
      });
      dimensionsRef.current = nextDimensions;
      setDimensions(nextDimensions);
    };

    measureNodes();
    const observer = new ResizeObserver(measureNodes);
    canvas.querySelectorAll<HTMLElement>("[data-node-id]").forEach((node) => {
      observer.observe(node);
    });

    return () => {
      observer.disconnect();
    };
  }, [renderedNodeIds]);

  useLayoutEffect(() => {
    const viewport = viewportRef.current;
    if (!viewport) return;

    const shouldReset =
      !hasPositionedViewportRef.current ||
      previousResetSignalRef.current !== resetSignal;
    previousResetSignalRef.current = resetSignal;
    if (!shouldReset) return;

    viewport.scrollLeft = Math.max(
      0,
      (layout.width * zoom - viewport.clientWidth) / 2,
    );
    viewport.scrollTop = 0;
    hasPositionedViewportRef.current = true;
  }, [layout.width, resetSignal, zoom]);

  useLayoutEffect(() => {
    const viewport = viewportRef.current;
    const previousZoom = previousZoomRef.current;
    previousZoomRef.current = zoom;
    if (
      !viewport ||
      previousZoom === zoom ||
      !hasPositionedViewportRef.current
    ) {
      return;
    }

    const ratio = zoom / previousZoom;
    viewport.scrollLeft =
      (viewport.scrollLeft + viewport.clientWidth / 2) * ratio -
      viewport.clientWidth / 2;
    viewport.scrollTop =
      (viewport.scrollTop + viewport.clientHeight / 2) * ratio -
      viewport.clientHeight / 2;
  }, [zoom]);

  const persistPositions = useCallback(() => {
    try {
      window.localStorage.setItem(
        storageKey,
        JSON.stringify(positionsRef.current),
      );
    } catch {
      // The canvas remains usable if storage is unavailable or full.
    }
  }, [storageKey]);

  const getPointerWorldPosition = useCallback(
    (clientX: number, clientY: number) => {
      const viewport = viewportRef.current;
      if (!viewport) return null;
      const bounds = viewport.getBoundingClientRect();
      return {
        x: (clientX - bounds.left + viewport.scrollLeft) / zoom,
        y: (clientY - bounds.top + viewport.scrollTop) / zoom,
      };
    },
    [zoom],
  );

  const getPillarAtPointer = useCallback(
    (clientX: number, clientY: number) => {
      const pointer = getPointerWorldPosition(clientX, clientY);
      if (!pointer) return null;

      return (
        strategy.pillars.find((pillar) => {
          const nodeId = getPillarNodeId(pillar.id);
          const position = positionsRef.current[nodeId];
          const nodeDimensions =
            dimensionsRef.current[nodeId] ?? getDefaultNodeDimensions(nodeId);

          return (
            pointer.x >= position.x &&
            pointer.x <= position.x + nodeDimensions.width &&
            pointer.y >= position.y &&
            pointer.y <= position.y + nodeDimensions.height
          );
        })?.id ?? null
      );
    },
    [getPointerWorldPosition, strategy.pillars],
  );

  const handleNodePointerDown = useCallback(
    (nodeId: string, event: ReactPointerEvent<HTMLDivElement>) => {
      if (event.button !== 0 || isInteractiveTarget(event.target)) return;
      const startPosition = positionsRef.current[nodeId];

      event.preventDefault();
      event.stopPropagation();
      event.currentTarget.setPointerCapture(event.pointerId);
      activeDragRef.current = {
        id: nodeId,
        pointerId: event.pointerId,
        startClientX: event.clientX,
        startClientY: event.clientY,
        startPosition,
      };
      setDraggingNodeId(nodeId);
      document.body.style.cursor = "grabbing";
      document.body.style.userSelect = "none";
    },
    [],
  );

  const handleNodePointerMove = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>) => {
      const drag = activeDragRef.current;
      if (!drag || drag.pointerId !== event.pointerId) return;
      const nodeDimensions =
        dimensionsRef.current[drag.id] ?? getDefaultNodeDimensions(drag.id);
      const nextPosition = {
        x: Math.max(
          CANVAS_INSET,
          Math.min(
            layout.width - nodeDimensions.width - CANVAS_INSET,
            drag.startPosition.x + (event.clientX - drag.startClientX) / zoom,
          ),
        ),
        y: Math.max(
          CANVAS_INSET,
          Math.min(
            layout.height - nodeDimensions.height - CANVAS_INSET,
            drag.startPosition.y + (event.clientY - drag.startClientY) / zoom,
          ),
        ),
      };
      const nextPositions = {
        ...positionsRef.current,
        [drag.id]: nextPosition,
      };
      positionsRef.current = nextPositions;
      setTransientLayout({ positions: nextPositions, storageKey });

      if (drag.id.startsWith("objective:")) {
        const nextDropTarget = getPillarAtPointer(event.clientX, event.clientY);
        setDropTargetPillarId((current) =>
          current === nextDropTarget ? current : nextDropTarget,
        );
      }
    },
    [getPillarAtPointer, layout.height, layout.width, storageKey, zoom],
  );

  const finishNodeDrag = useCallback(
    (
      event: ReactPointerEvent<HTMLDivElement>,
      shouldCommit: boolean,
      onSelect: () => void,
    ) => {
      const drag = activeDragRef.current;
      if (!drag || drag.pointerId !== event.pointerId) return;
      const movement = Math.hypot(
        event.clientX - drag.startClientX,
        event.clientY - drag.startClientY,
      );
      const wasDragged = movement > CLICK_MOVEMENT_THRESHOLD;

      if (shouldCommit && wasDragged && drag.id.startsWith("objective:")) {
        const objectiveId = drag.id.slice("objective:".length);
        const targetPillarId = getPillarAtPointer(event.clientX, event.clientY);
        if (
          targetPillarId &&
          pillarByObjectiveId.get(objectiveId) !== targetPillarId
        ) {
          onAlign(objectiveId, targetPillarId);
        }
      }

      if (event.currentTarget.hasPointerCapture(event.pointerId)) {
        event.currentTarget.releasePointerCapture(event.pointerId);
      }
      activeDragRef.current = null;
      setDraggingNodeId(null);
      setDropTargetPillarId(null);
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      persistPositions();

      if (shouldCommit && !wasDragged) onSelect();
    },
    [getPillarAtPointer, onAlign, persistPositions, pillarByObjectiveId],
  );

  const handleObjectiveUpdate = useCallback(
    (objectiveId: string, data: ObjectiveUpdate) => {
      updateObjective.mutate({ objectiveId, data });
    },
    [updateObjective],
  );

  const handleViewportPointerDown = (
    event: ReactPointerEvent<HTMLDivElement>,
  ) => {
    if (
      event.button !== 0 ||
      isInteractiveTarget(event.target) ||
      (event.target instanceof HTMLElement &&
        event.target.closest("[data-node-id], [data-canvas-control]"))
    ) {
      return;
    }

    event.preventDefault();
    event.currentTarget.setPointerCapture(event.pointerId);
    activePanRef.current = {
      pointerId: event.pointerId,
      startClientX: event.clientX,
      startClientY: event.clientY,
      startScrollLeft: event.currentTarget.scrollLeft,
      startScrollTop: event.currentTarget.scrollTop,
    };
    setIsPanning(true);
  };

  const handleViewportPointerMove = (
    event: ReactPointerEvent<HTMLDivElement>,
  ) => {
    const pan = activePanRef.current;
    if (!pan || pan.pointerId !== event.pointerId) return;
    event.currentTarget.scrollLeft =
      pan.startScrollLeft - (event.clientX - pan.startClientX);
    event.currentTarget.scrollTop =
      pan.startScrollTop - (event.clientY - pan.startClientY);
  };

  const handleViewportPointerUp = (
    event: ReactPointerEvent<HTMLDivElement>,
  ) => {
    const pan = activePanRef.current;
    if (!pan || pan.pointerId !== event.pointerId) return;
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
    activePanRef.current = null;
    setIsPanning(false);
  };

  useEffect(
    () => () => {
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
    },
    [],
  );

  return (
    <Box className="relative h-full overflow-hidden">
      <div
        className={cn(
          "bg-surface-muted/20 dark:bg-surface-elevated/35 h-full overflow-auto overscroll-none",
          isPanning ? "cursor-grabbing" : "cursor-grab",
        )}
        onPointerCancel={handleViewportPointerUp}
        onPointerDown={handleViewportPointerDown}
        onPointerMove={handleViewportPointerMove}
        onPointerUp={handleViewportPointerUp}
        ref={viewportRef}
        style={{ touchAction: "none" }}
      >
        <div
          className="relative"
          style={{
            height: layout.height * zoom,
            width: layout.width * zoom,
          }}
        >
          <div
            className="bg-surface-muted/15 dark:bg-surface-elevated/25 absolute top-0 left-0 origin-top-left"
            ref={canvasRef}
            style={{
              backgroundImage:
                "radial-gradient(var(--color-border-strong) 1.15px, transparent 1.15px)",
              backgroundSize: "22px 22px",
              height: layout.height,
              transform: `scale(${zoom})`,
              width: layout.width,
            }}
          >
            <CanvasConnections
              dimensions={dimensions}
              objectiveIds={objectiveIds}
              positions={positions}
              strategy={strategy}
            />

            <StrategyCanvasNode
              id={GOAL_NODE_ID}
              isDragging={draggingNodeId === GOAL_NODE_ID}
              onPointerDown={handleNodePointerDown}
              onPointerMove={handleNodePointerMove}
              onPointerUp={(event) => {
                finishNodeDrag(
                  event,
                  event.type !== "pointercancel",
                  onSelectGoal,
                );
              }}
              position={
                positions[GOAL_NODE_ID] ?? layout.positions[GOAL_NODE_ID]
              }
            >
              <UltimateGoalNodeCard
                averageProgress={averageProgress}
                canEdit={canEdit}
                description={strategy.description}
                objectiveCount={objectives.length}
                onEdit={onSelectGoal}
                onOpenDetails={onSelectGoal}
                pillarCount={strategy.pillars.length}
                title={strategy.ultimateGoal}
              />
            </StrategyCanvasNode>

            {strategy.pillars.map((pillar) => {
              const nodeId = getPillarNodeId(pillar.id);
              const alignedObjectiveCount = pillar.objectiveIds.filter((id) =>
                objectiveById.has(id),
              ).length;
              return (
                <StrategyCanvasNode
                  id={nodeId}
                  isDragging={draggingNodeId === nodeId}
                  key={pillar.id}
                  onPointerDown={handleNodePointerDown}
                  onPointerMove={handleNodePointerMove}
                  onPointerUp={(event) => {
                    finishNodeDrag(
                      event,
                      event.type !== "pointercancel",
                      () => {
                        onSelectPillar(pillar);
                      },
                    );
                  }}
                  position={positions[nodeId] ?? layout.positions[nodeId]}
                >
                  <PillarNodeCard
                    canEdit={canEdit}
                    description={pillar.description}
                    isDropTarget={dropTargetPillarId === pillar.id}
                    name={pillar.name}
                    objectiveCount={alignedObjectiveCount}
                    onDelete={() => {
                      onDeletePillar(pillar.id);
                    }}
                    onEdit={() => {
                      onSelectPillar(pillar);
                    }}
                    onOpenDetails={() => {
                      onSelectPillar(pillar);
                    }}
                  />
                </StrategyCanvasNode>
              );
            })}

            {strategy.pillars.length === 0 ? (
              <button
                className="border-border text-text-muted hover:border-foreground/35 hover:bg-surface-elevated/35 hover:text-foreground disabled:hover:border-border disabled:hover:text-text-muted absolute rounded-xl border-2 border-dashed px-8 py-7 text-center transition-colors disabled:cursor-default disabled:hover:bg-transparent"
                data-no-drag
                disabled={!canEdit}
                onClick={onAddPillar}
                style={{
                  left: (layout.width - GOAL_NODE_WIDTH) / 2,
                  top: 340,
                  width: GOAL_NODE_WIDTH,
                }}
                type="button"
              >
                <Text fontWeight="medium">
                  Add a strategic pillar to start building the map.
                </Text>
              </button>
            ) : null}

            {objectives.map((objective) => {
              const currentPillarId =
                pillarByObjectiveId.get(objective.id) ?? null;
              if (!showUnaligned && !currentPillarId) return null;
              const nodeId = getObjectiveNodeId(objective.id);
              return (
                <StrategyCanvasNode
                  id={nodeId}
                  isDragging={draggingNodeId === nodeId}
                  key={objective.id}
                  onPointerDown={handleNodePointerDown}
                  onPointerMove={handleNodePointerMove}
                  onPointerUp={(event) => {
                    finishNodeDrag(
                      event,
                      event.type !== "pointercancel",
                      () => {
                        onSelectObjective(objective);
                      },
                    );
                  }}
                  position={positions[nodeId] ?? layout.positions[nodeId]}
                >
                  <ObjectiveNodeCard
                    canEdit={canEdit}
                    currentPillarId={currentPillarId}
                    objective={objective}
                    onAlign={onAlign}
                    onOpenDetails={() => {
                      onSelectObjective(objective);
                    }}
                    onUpdate={handleObjectiveUpdate}
                    pillars={strategy.pillars}
                    status={statusById.get(objective.statusId)}
                    statuses={statuses}
                    teamCode={teamCodeById.get(objective.teamId)}
                  />
                </StrategyCanvasNode>
              );
            })}
          </div>
        </div>
      </div>

      <Flex
        align="center"
        className="border-border bg-surface/90 text-text-muted pointer-events-none absolute bottom-4 left-1/2 z-40 -translate-x-1/2 gap-2 rounded-lg border px-3.5 py-2 text-sm shadow-lg backdrop-blur"
        data-canvas-control
      >
        <span>Drag cards to position</span>
        <span aria-hidden>·</span>
        <span>Drag the canvas to pan</span>
        <span aria-hidden>·</span>
        <span>Right-click for actions</span>
      </Flex>
    </Box>
  );
};

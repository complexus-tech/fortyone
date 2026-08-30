"use client";

import type {
  Dispatch,
  MutableRefObject,
  PointerEvent as ReactPointerEvent,
  RefObject,
  SetStateAction,
} from "react";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  getKeyResultNodeId,
  getPillarNodeId,
  type StrategyNodeDimensions,
  type StrategyNodePositions,
} from "./strategy-map-layout";
import {
  getDefaultNodeDimensions,
  getNodePosition,
  isInteractiveTarget,
} from "./strategy-map-canvas-renderers";
import type { StrategicPillar } from "./types";

const CANVAS_INSET = 28;
const CLICK_MOVEMENT_THRESHOLD = 4;
const EMPTY_DRAGGING_NODE_IDS: readonly string[] = [];

type ActiveNodeDrag = {
  deltaX: number;
  deltaY: number;
  id: string;
  pointerId: number;
  startClientX: number;
  startClientY: number;
  startPositions: StrategyNodePositions;
};

type ActivePan = {
  pointerId: number;
  startClientX: number;
  startClientY: number;
  startScrollLeft: number;
  startScrollTop: number;
};

type StrategyMapCanvasLayout = {
  height: number;
  positions: StrategyNodePositions;
  width: number;
};

type StrategyMapInteractionKeyResult = {
  id: string;
};

export type StrategyMapCanvasInteractionState = {
  draggingNodeId: string | null;
  draggingNodeIds: readonly string[];
  draggingObjectiveId: string | null;
  dropTargetPillarId: string | null;
  isPanning: boolean;
  setDraggingNodeId: Dispatch<SetStateAction<string | null>>;
  setDraggingNodeIds: Dispatch<SetStateAction<readonly string[]>>;
  setDraggingObjectiveId: Dispatch<SetStateAction<string | null>>;
  setDropTargetPillarId: Dispatch<SetStateAction<string | null>>;
  setIsPanning: Dispatch<SetStateAction<boolean>>;
};

export const useStrategyMapCanvasInteractionState =
  (): StrategyMapCanvasInteractionState => {
    const [draggingNodeId, setDraggingNodeId] = useState<string | null>(null);
    const [draggingNodeIds, setDraggingNodeIds] = useState<readonly string[]>(
      EMPTY_DRAGGING_NODE_IDS,
    );
    const [draggingObjectiveId, setDraggingObjectiveId] = useState<
      string | null
    >(null);
    const [dropTargetPillarId, setDropTargetPillarId] = useState<string | null>(
      null,
    );
    const [isPanning, setIsPanning] = useState(false);

    return {
      draggingNodeId,
      draggingNodeIds,
      draggingObjectiveId,
      dropTargetPillarId,
      isPanning,
      setDraggingNodeId,
      setDraggingNodeIds,
      setDraggingObjectiveId,
      setDropTargetPillarId,
      setIsPanning,
    };
  };

type UseStrategyMapCanvasInteractionsOptions = {
  dimensionsRef: MutableRefObject<Record<string, StrategyNodeDimensions>>;
  expandedObjectiveIds: ReadonlySet<string>;
  keyResultObjectiveIdByNodeId: ReadonlyMap<string, string>;
  keyResultsByObjective: ReadonlyMap<string, StrategyMapInteractionKeyResult[]>;
  layout: StrategyMapCanvasLayout;
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  persistPositions: () => void;
  pillarByObjectiveId: ReadonlyMap<string, string>;
  pillars: readonly StrategicPillar[];
  positionOverrides: StrategyNodePositions;
  positions: StrategyNodePositions;
  positionsRef: MutableRefObject<StrategyNodePositions>;
  setTransientPositions: (positions: StrategyNodePositions) => void;
  state: StrategyMapCanvasInteractionState;
  viewportRef: RefObject<HTMLDivElement | null>;
  zoom: number;
};

export const useStrategyMapCanvasInteractions = ({
  dimensionsRef,
  expandedObjectiveIds,
  keyResultObjectiveIdByNodeId,
  keyResultsByObjective,
  layout,
  onAlign,
  persistPositions,
  pillarByObjectiveId,
  pillars,
  positionOverrides,
  positions,
  positionsRef,
  setTransientPositions,
  state,
  viewportRef,
  zoom,
}: UseStrategyMapCanvasInteractionsOptions) => {
  const {
    setDraggingNodeId,
    setDraggingNodeIds,
    setDraggingObjectiveId,
    setDropTargetPillarId,
    setIsPanning,
  } = state;
  const activeDragRef = useRef<ActiveNodeDrag | null>(null);
  const activePanRef = useRef<ActivePan | null>(null);
  const dragAnimationFrameRef = useRef<number | null>(null);
  const pendingDragPositionsRef = useRef<StrategyNodePositions | null>(null);

  const flushPendingDragLayout = useCallback(() => {
    if (dragAnimationFrameRef.current !== null) {
      cancelAnimationFrame(dragAnimationFrameRef.current);
      dragAnimationFrameRef.current = null;
    }

    const pendingPositions = pendingDragPositionsRef.current;
    pendingDragPositionsRef.current = null;
    if (!pendingPositions) return;

    setTransientPositions(pendingPositions);
  }, [setTransientPositions]);
  const scheduleDragLayout = useCallback(
    (nextPositions: StrategyNodePositions) => {
      pendingDragPositionsRef.current = nextPositions;
      if (dragAnimationFrameRef.current !== null) return;

      dragAnimationFrameRef.current = requestAnimationFrame(() => {
        dragAnimationFrameRef.current = null;
        const pendingPositions = pendingDragPositionsRef.current;
        pendingDragPositionsRef.current = null;
        if (!pendingPositions) return;

        setTransientPositions(pendingPositions);
      });
    },
    [setTransientPositions],
  );

  useEffect(() => {
    const drag = activeDragRef.current;
    if (!drag?.id.startsWith("objective:")) return;

    const objectiveId = drag.id.slice("objective:".length);
    const keyResults = keyResultsByObjective.get(objectiveId);
    if (!keyResults?.length) return;

    const defaultObjectivePosition = getNodePosition(layout.positions, drag.id);
    const objectiveStartPosition = getNodePosition(
      drag.startPositions,
      drag.id,
    );
    if (!defaultObjectivePosition || !objectiveStartPosition) return;

    const parentOffset = {
      x: objectiveStartPosition.x - defaultObjectivePosition.x,
      y: objectiveStartPosition.y - defaultObjectivePosition.y,
    };
    const nextPositions = { ...positionsRef.current };
    const addedNodeIds: string[] = [];

    keyResults.forEach((keyResult) => {
      const nodeId = getKeyResultNodeId(keyResult.id);
      if (Object.hasOwn(drag.startPositions, nodeId)) return;

      const defaultKeyResultPosition = getNodePosition(
        layout.positions,
        nodeId,
      );
      const currentKeyResultPosition = getNodePosition(positions, nodeId);
      if (!defaultKeyResultPosition || !currentKeyResultPosition) return;

      const startPosition = Object.hasOwn(positionOverrides, nodeId)
        ? currentKeyResultPosition
        : {
            x: defaultKeyResultPosition.x + parentOffset.x,
            y: defaultKeyResultPosition.y + parentOffset.y,
          };
      drag.startPositions[nodeId] = startPosition;
      nextPositions[nodeId] = {
        x: startPosition.x + drag.deltaX,
        y: startPosition.y + drag.deltaY,
      };
      addedNodeIds.push(nodeId);
    });

    if (addedNodeIds.length === 0) return;
    positionsRef.current = nextPositions;
    scheduleDragLayout(nextPositions);
    setDraggingNodeIds((current) => [...current, ...addedNodeIds]);
  }, [
    keyResultsByObjective,
    layout.positions,
    positionOverrides,
    positions,
    positionsRef,
    scheduleDragLayout,
    setDraggingNodeIds,
  ]);

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
    [viewportRef, zoom],
  );
  const getPillarAtPointer = useCallback(
    (clientX: number, clientY: number) => {
      const pointer = getPointerWorldPosition(clientX, clientY);
      if (!pointer) return null;

      return (
        pillars.find((pillar) => {
          const nodeId = getPillarNodeId(pillar.id);
          const position = getNodePosition(positionsRef.current, nodeId);
          if (!position) return false;
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
    [dimensionsRef, getPointerWorldPosition, pillars, positionsRef],
  );

  const handleNodePointerDown = useCallback(
    (nodeId: string, event: ReactPointerEvent<HTMLDivElement>) => {
      if (event.button !== 0 || isInteractiveTarget(event.target)) return;
      const draggedNodeIds = [nodeId];
      if (nodeId.startsWith("objective:")) {
        const objectiveId = nodeId.slice("objective:".length);
        if (expandedObjectiveIds.has(objectiveId)) {
          keyResultsByObjective.get(objectiveId)?.forEach((keyResult) => {
            draggedNodeIds.push(getKeyResultNodeId(keyResult.id));
          });
        }
      }
      const startPositions = Object.fromEntries(
        draggedNodeIds.flatMap((id) => {
          const position = getNodePosition(positionsRef.current, id);
          return position ? [[id, position]] : [];
        }),
      );
      if (!Object.hasOwn(startPositions, nodeId)) return;

      event.preventDefault();
      event.stopPropagation();
      event.currentTarget.setPointerCapture(event.pointerId);
      activeDragRef.current = {
        deltaX: 0,
        deltaY: 0,
        id: nodeId,
        pointerId: event.pointerId,
        startClientX: event.clientX,
        startClientY: event.clientY,
        startPositions,
      };
      setDraggingNodeId(nodeId);
      setDraggingNodeIds(draggedNodeIds);
      setDraggingObjectiveId(
        nodeId.startsWith("objective:")
          ? nodeId.slice("objective:".length)
          : keyResultObjectiveIdByNodeId.get(nodeId) ?? null,
      );
      document.body.style.cursor = "grabbing";
      document.body.style.userSelect = "none";
    },
    [
      expandedObjectiveIds,
      keyResultObjectiveIdByNodeId,
      keyResultsByObjective,
      positionsRef,
      setDraggingNodeId,
      setDraggingNodeIds,
      setDraggingObjectiveId,
    ],
  );
  const handleNodePointerMove = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>) => {
      const drag = activeDragRef.current;
      if (!drag || drag.pointerId !== event.pointerId) return;
      const rawDeltaX = (event.clientX - drag.startClientX) / zoom;
      const rawDeltaY = (event.clientY - drag.startClientY) / zoom;
      let minimumDeltaX = Number.NEGATIVE_INFINITY;
      let maximumDeltaX = Number.POSITIVE_INFINITY;
      let minimumDeltaY = Number.NEGATIVE_INFINITY;
      let maximumDeltaY = Number.POSITIVE_INFINITY;

      Object.entries(drag.startPositions).forEach(([nodeId, position]) => {
        const nodeDimensions =
          dimensionsRef.current[nodeId] ?? getDefaultNodeDimensions(nodeId);
        minimumDeltaX = Math.max(minimumDeltaX, CANVAS_INSET - position.x);
        maximumDeltaX = Math.min(
          maximumDeltaX,
          layout.width - nodeDimensions.width - CANVAS_INSET - position.x,
        );
        minimumDeltaY = Math.max(minimumDeltaY, CANVAS_INSET - position.y);
        maximumDeltaY = Math.min(
          maximumDeltaY,
          layout.height - nodeDimensions.height - CANVAS_INSET - position.y,
        );
      });

      const deltaX = Math.max(
        minimumDeltaX,
        Math.min(maximumDeltaX, rawDeltaX),
      );
      const deltaY = Math.max(
        minimumDeltaY,
        Math.min(maximumDeltaY, rawDeltaY),
      );
      drag.deltaX = deltaX;
      drag.deltaY = deltaY;
      const nextPositions = { ...positionsRef.current };
      Object.entries(drag.startPositions).forEach(([nodeId, position]) => {
        nextPositions[nodeId] = {
          x: position.x + deltaX,
          y: position.y + deltaY,
        };
      });
      positionsRef.current = nextPositions;
      scheduleDragLayout(nextPositions);

      if (drag.id.startsWith("objective:")) {
        const nextDropTarget = getPillarAtPointer(event.clientX, event.clientY);
        setDropTargetPillarId((current) =>
          current === nextDropTarget ? current : nextDropTarget,
        );
      }
    },
    [
      dimensionsRef,
      getPillarAtPointer,
      layout.height,
      layout.width,
      positionsRef,
      scheduleDragLayout,
      setDropTargetPillarId,
      zoom,
    ],
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
      flushPendingDragLayout();
      activeDragRef.current = null;
      setDraggingNodeId(null);
      setDraggingNodeIds(EMPTY_DRAGGING_NODE_IDS);
      setDraggingObjectiveId(null);
      setDropTargetPillarId(null);
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      persistPositions();

      if (shouldCommit && !wasDragged) onSelect();
    },
    [
      flushPendingDragLayout,
      getPillarAtPointer,
      onAlign,
      persistPositions,
      pillarByObjectiveId,
      setDraggingNodeId,
      setDraggingNodeIds,
      setDraggingObjectiveId,
      setDropTargetPillarId,
    ],
  );
  const handleViewportPointerDown = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>) => {
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
    },
    [setIsPanning],
  );
  const handleViewportPointerMove = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>) => {
      const pan = activePanRef.current;
      if (!pan || pan.pointerId !== event.pointerId) return;
      event.currentTarget.scrollLeft =
        pan.startScrollLeft - (event.clientX - pan.startClientX);
      event.currentTarget.scrollTop =
        pan.startScrollTop - (event.clientY - pan.startClientY);
    },
    [],
  );
  const handleViewportPointerUp = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>) => {
      const pan = activePanRef.current;
      if (!pan || pan.pointerId !== event.pointerId) return;
      if (event.currentTarget.hasPointerCapture(event.pointerId)) {
        event.currentTarget.releasePointerCapture(event.pointerId);
      }
      activePanRef.current = null;
      setIsPanning(false);
    },
    [setIsPanning],
  );

  useEffect(
    () => () => {
      if (dragAnimationFrameRef.current !== null) {
        cancelAnimationFrame(dragAnimationFrameRef.current);
      }
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
    },
    [],
  );

  return {
    finishNodeDrag,
    handleNodePointerDown,
    handleNodePointerMove,
    handleViewportPointerDown,
    handleViewportPointerMove,
    handleViewportPointerUp,
  };
};

"use client";

import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
} from "react";
import { useWorkspacePath } from "@/hooks";
import { useMembers } from "@/lib/hooks/members";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { useUpdateObjectiveMutation } from "@/modules/objectives/hooks";
import type {
  KeyResult,
  Objective,
  ObjectiveUpdate,
} from "@/modules/objectives/types";
import { KeyResultContextMenu } from "@/modules/key-results/components/key-result-context-menu";
import { useTeams } from "@/modules/teams/public/client";
import { getNodePosition } from "./strategy-map-canvas-renderers";
import {
  persistStrategyMapCanvasLayout,
  persistStrategyMapKeyResultOwners,
  useStrategyMapCanvasStorage,
} from "./strategy-map-canvas-storage";
import {
  createStrategyMapLayout,
  getKeyResultNodeId,
  getObjectiveNodeId,
  GOAL_NODE_ID,
  mergeStrategyNodePositions,
  type StrategyNodeDimensions,
  type StrategyNodePositions,
} from "./strategy-map-layout";
import type { StrategicPillar, StrategyMap, StrategyMember } from "./types";
import {
  useStrategyMapCanvasInteractions,
  useStrategyMapCanvasInteractionState,
} from "./use-strategy-map-canvas-interactions";
import { useStrategyMapCanvasRenderData } from "./use-strategy-map-canvas-render-data";
import { StrategyMapCanvasSurface } from "./strategy-map-canvas-surface";
import { getCompleteStrategyMapAverageProgress } from "./strategy-map-progress";
import { useStrategyKeyResults } from "./use-strategy-key-results";
import {
  useStrategyMapCanvasNodeDimensions,
  useStrategyMapCanvasViewport,
} from "./use-strategy-map-canvas-viewport";
import {
  getPersistedKeyResultFetchObjectiveIds,
  getStrategyKeyResultFetchObjectiveIds,
} from "./strategy-map-visibility";

const EMPTY_KEY_RESULTS_BY_OBJECTIVE = new Map<string, KeyResult[]>();

type StrategyMapCanvasProps = {
  canEdit: boolean;
  objectives: Objective[];
  onAddPillar: () => void;
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  onDeletePillar: (pillarId: string) => void;
  onSelectGoal: () => void;
  onSelectKeyResult: (objective: Objective, keyResult: KeyResult) => void;
  onSelectObjective: (objective: Objective) => void;
  onSelectPillar: (pillar: StrategicPillar) => void;
  onZoomChange: (zoom: number) => void;
  resetSignal?: number;
  selectedNodeId?: string | null;
  selectedObjectiveId?: string | null;
  showUnaligned: boolean;
  strategy: StrategyMap;
  zoom: number;
};

export const StrategyMapCanvas = ({
  canEdit,
  objectives,
  onAddPillar,
  onAlign,
  onDeletePillar,
  onSelectGoal,
  onSelectKeyResult,
  onSelectObjective,
  onSelectPillar,
  onZoomChange,
  resetSignal = 0,
  selectedNodeId,
  selectedObjectiveId,
  showUnaligned,
  strategy,
  zoom,
}: StrategyMapCanvasProps) => {
  const { workspaceSlug } = useWorkspacePath();
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useObjectiveStatuses();
  const { data: teams = [] } = useTeams();
  const updateObjective = useUpdateObjectiveMutation();
  const viewportRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLDivElement>(null);
  const positionsRef = useRef<StrategyNodePositions>({});
  const dimensionsRef = useRef<Record<string, StrategyNodeDimensions>>({});
  const interactionState = useStrategyMapCanvasInteractionState();
  const {
    expandedObjectiveIds,
    keyResultOwnerStorageKey,
    setTransientPositions,
    storageKey,
    storedKeyResultOwnerIds,
    storedPositions,
    toggleObjectiveKeyResults,
    transientPositions,
  } = useStrategyMapCanvasStorage({ objectives, workspaceSlug });
  const {
    draggingNodeId,
    draggingNodeIds,
    draggingObjectiveId,
    dropTargetPillarId,
    isPanning,
  } = interactionState;
  const baseLayout = useMemo(
    () =>
      createStrategyMapLayout(
        strategy,
        objectives,
        EMPTY_KEY_RESULTS_BY_OBJECTIVE,
        expandedObjectiveIds,
      ),
    [expandedObjectiveIds, objectives, strategy],
  );
  const positionOverrides = useMemo(
    () => ({ ...storedPositions, ...transientPositions }),
    [storedPositions, transientPositions],
  );
  const basePositions = useMemo(
    () => mergeStrategyNodePositions(baseLayout.positions, positionOverrides),
    [baseLayout.positions, positionOverrides],
  );
  const viewport = useStrategyMapCanvasViewport({
    layoutWidth: baseLayout.width,
    onZoomChange,
    resetSignal,
    viewportRef,
    zoom,
  });
  const keyResultFetchObjectiveIds = useMemo(() => {
    const alwaysIncludeObjectiveIds = getPersistedKeyResultFetchObjectiveIds({
      expandedObjectiveIds,
      keyResultOwnerIds: storedKeyResultOwnerIds,
      positions: storedPositions,
      viewport,
    });
    if (selectedObjectiveId) alwaysIncludeObjectiveIds.add(selectedObjectiveId);
    if (draggingObjectiveId) alwaysIncludeObjectiveIds.add(draggingObjectiveId);

    return getStrategyKeyResultFetchObjectiveIds({
      alwaysIncludeObjectiveIds,
      expandedObjectiveIds,
      objectives,
      positions: basePositions,
      viewport,
    });
  }, [
    basePositions,
    draggingObjectiveId,
    expandedObjectiveIds,
    objectives,
    selectedObjectiveId,
    storedKeyResultOwnerIds,
    storedPositions,
    viewport,
  ]);
  const { keyResultsByObjective, loadedObjectiveIds } = useStrategyKeyResults(
    objectives,
    keyResultFetchObjectiveIds,
  );
  const layout = useMemo(
    () =>
      createStrategyMapLayout(
        strategy,
        objectives,
        keyResultsByObjective,
        expandedObjectiveIds,
      ),
    [expandedObjectiveIds, keyResultsByObjective, objectives, strategy],
  );
  const keyResultOwnerIds = useMemo(() => {
    const result = new Map(storedKeyResultOwnerIds);
    keyResultsByObjective.forEach((keyResults, objectiveId) => {
      keyResults.forEach((keyResult) => {
        result.set(getKeyResultNodeId(keyResult.id), objectiveId);
      });
    });
    return result;
  }, [keyResultsByObjective, storedKeyResultOwnerIds]);
  useEffect(() => {
    persistStrategyMapKeyResultOwners({
      keyResultOwnerIds,
      keyResultOwnerStorageKey,
    });
  }, [keyResultOwnerIds, keyResultOwnerStorageKey]);
  const positions = useMemo(() => {
    const merged = mergeStrategyNodePositions(
      layout.positions,
      positionOverrides,
    );

    objectives.forEach((objective) => {
      const objectiveNodeId = getObjectiveNodeId(objective.id);
      const defaultObjectivePosition = getNodePosition(
        layout.positions,
        objectiveNodeId,
      );
      const objectivePosition = getNodePosition(merged, objectiveNodeId);
      if (!defaultObjectivePosition || !objectivePosition) return;

      const deltaX = objectivePosition.x - defaultObjectivePosition.x;
      const deltaY = objectivePosition.y - defaultObjectivePosition.y;
      keyResultsByObjective.get(objective.id)?.forEach((keyResult) => {
        const keyResultNodeId = getKeyResultNodeId(keyResult.id);
        if (Object.hasOwn(positionOverrides, keyResultNodeId)) return;
        const defaultKeyResultPosition = getNodePosition(
          layout.positions,
          keyResultNodeId,
        );
        if (!defaultKeyResultPosition) return;
        merged[keyResultNodeId] = {
          x: defaultKeyResultPosition.x + deltaX,
          y: defaultKeyResultPosition.y + deltaY,
        };
      });
    });

    return merged;
  }, [keyResultsByObjective, layout.positions, objectives, positionOverrides]);
  const dimensions = useStrategyMapCanvasNodeDimensions({
    canvasRef,
    dimensionsRef,
  });
  const statusById = useMemo(
    () => new Map(statuses.map((status) => [status.id, status])),
    [statuses],
  );
  const teamCodeById = useMemo(
    () => new Map(teams.map((team) => [team.id, team.code])),
    [teams],
  );
  const memberById = useMemo<ReadonlyMap<string, StrategyMember>>(
    () => new Map(members.map((member) => [member.id, member])),
    [members],
  );
  const {
    alwaysRenderedNodeIds,
    keyResultObjectiveIdByNodeId,
    objectiveById,
    objectiveIds,
    pillarByObjectiveId,
    visibleKeyResultNodes,
    visibleNodeIds,
    visibleObjectives,
    visiblePillars,
  } = useStrategyMapCanvasRenderData({
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
  });
  const averageProgress = getCompleteStrategyMapAverageProgress(
    objectives,
    keyResultsByObjective,
    loadedObjectiveIds,
  );

  useLayoutEffect(() => {
    positionsRef.current = positions;
  }, [positions]);
  const persistPositions = useCallback(() => {
    persistStrategyMapCanvasLayout({
      keyResultOwnerIds,
      keyResultOwnerStorageKey,
      positions: positionsRef.current,
      storageKey,
    });
  }, [keyResultOwnerIds, keyResultOwnerStorageKey, storageKey]);
  const {
    finishNodeDrag,
    handleNodePointerDown,
    handleNodePointerMove,
    handleViewportPointerDown,
    handleViewportPointerMove,
    handleViewportPointerUp,
  } = useStrategyMapCanvasInteractions({
    dimensionsRef,
    expandedObjectiveIds,
    keyResultObjectiveIdByNodeId,
    keyResultsByObjective,
    layout,
    onAlign,
    persistPositions,
    pillarByObjectiveId,
    pillars: strategy.pillars,
    positionOverrides,
    positions,
    positionsRef,
    setTransientPositions,
    state: interactionState,
    viewportRef,
    zoom,
  });
  const handleObjectiveUpdate = useCallback(
    (objectiveId: string, data: ObjectiveUpdate) => {
      updateObjective.mutate({ objectiveId, data });
    },
    [updateObjective],
  );

  const canvasConnections = {
    alwaysVisibleNodeIds: alwaysRenderedNodeIds,
    dimensions,
    expandedObjectiveIds,
    keyResultsByObjective,
    objectiveIds,
    positions,
    strategy,
    viewport,
  };
  const canvasNodes = {
    averageProgress,
    canEdit,
    draggingNodeId,
    dropTargetPillarId,
    expandedObjectiveIds,
    keyResultContextMenu: KeyResultContextMenu,
    keyResultNodes: visibleKeyResultNodes,
    keyResultsByObjective,
    layoutPositions: layout.positions,
    layoutWidth: layout.width,
    memberById,
    objectiveById,
    objectiveCount: objectives.length,
    objectives: visibleObjectives,
    onAddPillar,
    onAlign,
    onDeletePillar,
    onFinishNodeDrag: finishNodeDrag,
    onNodePointerDown: handleNodePointerDown,
    onNodePointerMove: handleNodePointerMove,
    onSelectGoal,
    onSelectKeyResult,
    onSelectObjective,
    onSelectPillar,
    onToggleObjectiveKeyResults: toggleObjectiveKeyResults,
    onUpdateObjective: handleObjectiveUpdate,
    pillarByObjectiveId,
    pillars: visiblePillars,
    positions,
    showGoal: visibleNodeIds.has(GOAL_NODE_ID),
    statusById,
    statuses,
    strategy,
    teamCodeById,
  };

  return (
    <StrategyMapCanvasSurface
      canvasConnections={canvasConnections}
      canvasNodes={canvasNodes}
      canvasRef={canvasRef}
      isPanning={isPanning}
      layout={layout}
      onViewportPointerCancel={handleViewportPointerUp}
      onViewportPointerDown={handleViewportPointerDown}
      onViewportPointerMove={handleViewportPointerMove}
      onViewportPointerUp={handleViewportPointerUp}
      viewportRef={viewportRef}
      zoom={zoom}
    />
  );
};

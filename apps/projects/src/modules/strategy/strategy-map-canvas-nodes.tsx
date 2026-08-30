"use client";

import type {
  ComponentProps,
  ComponentType,
  PointerEvent as ReactPointerEvent,
  ReactNode,
} from "react";
import { memo } from "react";
import { Text } from "ui";
import {
  KeyResultNodeCard,
  ObjectiveNodeCard,
  PillarNodeCard,
  UltimateGoalNodeCard,
} from "./strategy-map-cards";
import { StrategyCanvasNode } from "./strategy-map-canvas-renderers";
import {
  getKeyResultNodeId,
  getObjectiveNodeId,
  getPillarNodeId,
  GOAL_NODE_ID,
  GOAL_NODE_WIDTH,
  type StrategyNodePosition,
  type StrategyNodePositions,
} from "./strategy-map-layout";
import type { StrategicPillar, StrategyMap, StrategyMember } from "./types";

type ObjectiveNodeCardProps = ComponentProps<typeof ObjectiveNodeCard>;
type KeyResultNodeCardProps = ComponentProps<typeof KeyResultNodeCard>;
type StrategyCanvasObjective = ObjectiveNodeCardProps["objective"];
type StrategyCanvasKeyResult = KeyResultNodeCardProps["keyResult"];
type StrategyCanvasObjectiveStatus = ObjectiveNodeCardProps["statuses"][number];
type StrategyCanvasObjectiveUpdate = Parameters<
  ObjectiveNodeCardProps["onUpdate"]
>[1];
type KeyResultContextMenuComponent = ComponentType<{
  children: ReactNode;
  keyResult: StrategyCanvasKeyResult;
  onOpenDetails: () => void;
}>;

type StrategyCanvasNodePointerHandler = (
  event: ReactPointerEvent<HTMLDivElement>,
) => void;

type FinishNodeDrag = (
  event: ReactPointerEvent<HTMLDivElement>,
  shouldCommit: boolean,
  onSelect: () => void,
) => void;

type StrategyCanvasKeyResultNode = {
  keyResult: StrategyCanvasKeyResult;
  objective: StrategyCanvasObjective;
};

type StrategyCanvasNodesProps = {
  averageProgress: number | null;
  canEdit: boolean;
  draggingNodeId: string | null;
  dropTargetPillarId: string | null;
  expandedObjectiveIds: ReadonlySet<string>;
  keyResultContextMenu: KeyResultContextMenuComponent;
  keyResultNodes: readonly StrategyCanvasKeyResultNode[];
  keyResultsByObjective: ReadonlyMap<string, StrategyCanvasKeyResult[]>;
  layoutPositions: StrategyNodePositions;
  layoutWidth: number;
  memberById: ReadonlyMap<string, StrategyMember>;
  objectiveById: ReadonlyMap<string, StrategyCanvasObjective>;
  objectiveCount: number;
  objectives: readonly StrategyCanvasObjective[];
  onAddPillar: () => void;
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  onDeletePillar: (pillarId: string) => void;
  onFinishNodeDrag: FinishNodeDrag;
  onNodePointerDown: (
    nodeId: string,
    event: ReactPointerEvent<HTMLDivElement>,
  ) => void;
  onNodePointerMove: StrategyCanvasNodePointerHandler;
  onSelectGoal: () => void;
  onSelectKeyResult: (
    objective: StrategyCanvasObjective,
    keyResult: StrategyCanvasKeyResult,
  ) => void;
  onSelectObjective: (objective: StrategyCanvasObjective) => void;
  onSelectPillar: (pillar: StrategicPillar) => void;
  onToggleObjectiveKeyResults: (objectiveId: string) => void;
  onUpdateObjective: (
    objectiveId: string,
    data: StrategyCanvasObjectiveUpdate,
  ) => void;
  pillarByObjectiveId: ReadonlyMap<string, string>;
  pillars: readonly StrategicPillar[];
  positions: StrategyNodePositions;
  showGoal: boolean;
  statusById: ReadonlyMap<string, StrategyCanvasObjectiveStatus>;
  statuses: ObjectiveNodeCardProps["statuses"];
  strategy: StrategyMap;
  teamCodeById: ReadonlyMap<string, string>;
};

const EMPTY_KEY_RESULTS: StrategyCanvasKeyResult[] = [];

const getNodePosition = (
  positions: StrategyNodePositions,
  fallbackPositions: StrategyNodePositions,
  nodeId: string,
): StrategyNodePosition => positions[nodeId] ?? fallbackPositions[nodeId];

const GoalCanvasNode = memo(
  ({
    averageProgress,
    canEdit,
    draggingNodeId,
    objectiveCount,
    onFinishNodeDrag,
    onNodePointerDown,
    onNodePointerMove,
    onSelectGoal,
    position,
    strategy,
  }: Pick<
    StrategyCanvasNodesProps,
    | "averageProgress"
    | "canEdit"
    | "draggingNodeId"
    | "objectiveCount"
    | "onFinishNodeDrag"
    | "onNodePointerDown"
    | "onNodePointerMove"
    | "onSelectGoal"
    | "strategy"
  > & { position: StrategyNodePosition }) => (
    <StrategyCanvasNode
      id={GOAL_NODE_ID}
      isDragging={draggingNodeId === GOAL_NODE_ID}
      onPointerDown={onNodePointerDown}
      onPointerMove={onNodePointerMove}
      onPointerUp={(event) => {
        onFinishNodeDrag(event, event.type !== "pointercancel", onSelectGoal);
      }}
      position={position}
    >
      <UltimateGoalNodeCard
        averageProgress={averageProgress}
        canEdit={canEdit}
        description={strategy.description}
        objectiveCount={objectiveCount}
        onEdit={onSelectGoal}
        onOpenDetails={onSelectGoal}
        pillarCount={strategy.pillars.length}
        title={strategy.ultimateGoal}
      />
    </StrategyCanvasNode>
  ),
);
GoalCanvasNode.displayName = "GoalCanvasNode";

const PillarCanvasNode = memo(
  ({
    canEdit,
    draggingNodeId,
    dropTargetPillarId,
    objectiveById,
    onDeletePillar,
    onFinishNodeDrag,
    onNodePointerDown,
    onNodePointerMove,
    onSelectPillar,
    pillar,
    position,
  }: Pick<
    StrategyCanvasNodesProps,
    | "canEdit"
    | "draggingNodeId"
    | "dropTargetPillarId"
    | "objectiveById"
    | "onDeletePillar"
    | "onFinishNodeDrag"
    | "onNodePointerDown"
    | "onNodePointerMove"
    | "onSelectPillar"
  > & { pillar: StrategicPillar; position: StrategyNodePosition }) => {
    const nodeId = getPillarNodeId(pillar.id);
    const alignedObjectiveCount = pillar.objectiveIds.filter((id) =>
      objectiveById.has(id),
    ).length;

    return (
      <StrategyCanvasNode
        id={nodeId}
        isDragging={draggingNodeId === nodeId}
        onPointerDown={onNodePointerDown}
        onPointerMove={onNodePointerMove}
        onPointerUp={(event) => {
          onFinishNodeDrag(event, event.type !== "pointercancel", () => {
            onSelectPillar(pillar);
          });
        }}
        position={position}
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
  },
);
PillarCanvasNode.displayName = "PillarCanvasNode";

const ObjectiveCanvasNode = memo(
  ({
    canEdit,
    draggingNodeId,
    expandedObjectiveIds,
    keyResultsByObjective,
    memberById,
    onAlign,
    onFinishNodeDrag,
    onNodePointerDown,
    onNodePointerMove,
    onSelectObjective,
    onToggleObjectiveKeyResults,
    onUpdateObjective,
    pillarByObjectiveId,
    statusById,
    statuses,
    strategy,
    teamCodeById,
    objective,
    position,
  }: Pick<
    StrategyCanvasNodesProps,
    | "canEdit"
    | "draggingNodeId"
    | "expandedObjectiveIds"
    | "keyResultsByObjective"
    | "memberById"
    | "onAlign"
    | "onFinishNodeDrag"
    | "onNodePointerDown"
    | "onNodePointerMove"
    | "onSelectObjective"
    | "onToggleObjectiveKeyResults"
    | "onUpdateObjective"
    | "pillarByObjectiveId"
    | "statusById"
    | "statuses"
    | "strategy"
    | "teamCodeById"
  > & {
    objective: StrategyCanvasObjective;
    position: StrategyNodePosition;
  }) => {
    const nodeId = getObjectiveNodeId(objective.id);
    const currentPillarId = pillarByObjectiveId.get(objective.id) ?? null;

    return (
      <StrategyCanvasNode
        id={nodeId}
        isDragging={draggingNodeId === nodeId}
        onPointerDown={onNodePointerDown}
        onPointerMove={onNodePointerMove}
        onPointerUp={(event) => {
          onFinishNodeDrag(event, event.type !== "pointercancel", () => {
            onSelectObjective(objective);
          });
        }}
        position={position}
      >
        <ObjectiveNodeCard
          canEdit={canEdit}
          currentPillarId={currentPillarId}
          isKeyResultsExpanded={expandedObjectiveIds.has(objective.id)}
          keyResultCount={objective.keyResultCount}
          keyResults={
            keyResultsByObjective.get(objective.id) ?? EMPTY_KEY_RESULTS
          }
          memberById={memberById}
          objective={objective}
          onAlign={onAlign}
          onOpenDetails={() => {
            onSelectObjective(objective);
          }}
          onToggleKeyResults={() => {
            onToggleObjectiveKeyResults(objective.id);
          }}
          onUpdate={onUpdateObjective}
          pillars={strategy.pillars}
          status={statusById.get(objective.statusId)}
          statuses={statuses}
          teamCode={teamCodeById.get(objective.teamId)}
        />
      </StrategyCanvasNode>
    );
  },
);
ObjectiveCanvasNode.displayName = "ObjectiveCanvasNode";

const KeyResultCanvasNode = memo(
  ({
    draggingNodeId,
    keyResultContextMenu: KeyResultContextMenu,
    memberById,
    onFinishNodeDrag,
    onNodePointerDown,
    onNodePointerMove,
    onSelectKeyResult,
    teamCodeById,
    keyResult,
    objective,
    position,
  }: Pick<
    StrategyCanvasNodesProps,
    | "draggingNodeId"
    | "keyResultContextMenu"
    | "memberById"
    | "onFinishNodeDrag"
    | "onNodePointerDown"
    | "onNodePointerMove"
    | "onSelectKeyResult"
    | "teamCodeById"
  > &
    StrategyCanvasKeyResultNode & { position: StrategyNodePosition }) => {
    const nodeId = getKeyResultNodeId(keyResult.id);
    const teamCode = teamCodeById.get(objective.teamId);

    return (
      <StrategyCanvasNode
        id={nodeId}
        isDragging={draggingNodeId === nodeId}
        onPointerDown={onNodePointerDown}
        onPointerMove={onNodePointerMove}
        onPointerUp={(event) => {
          onFinishNodeDrag(event, event.type !== "pointercancel", () => {
            onSelectKeyResult(objective, keyResult);
          });
        }}
        position={position}
      >
        <KeyResultContextMenu
          keyResult={keyResult}
          onOpenDetails={() => {
            onSelectKeyResult(objective, keyResult);
          }}
        >
          <KeyResultNodeCard
            code={
              teamCode
                ? `${teamCode}-${keyResult.sequenceId}`
                : String(keyResult.sequenceId)
            }
            keyResult={keyResult}
            memberById={memberById}
            onOpenDetails={() => {
              onSelectKeyResult(objective, keyResult);
            }}
          />
        </KeyResultContextMenu>
      </StrategyCanvasNode>
    );
  },
);
KeyResultCanvasNode.displayName = "KeyResultCanvasNode";

export const StrategyCanvasNodes = memo(
  ({
    averageProgress,
    canEdit,
    draggingNodeId,
    dropTargetPillarId,
    expandedObjectiveIds,
    keyResultContextMenu,
    keyResultNodes,
    keyResultsByObjective,
    layoutPositions,
    layoutWidth,
    memberById,
    objectiveById,
    objectiveCount,
    objectives,
    onAddPillar,
    onAlign,
    onDeletePillar,
    onFinishNodeDrag,
    onNodePointerDown,
    onNodePointerMove,
    onSelectGoal,
    onSelectKeyResult,
    onSelectObjective,
    onSelectPillar,
    onToggleObjectiveKeyResults,
    onUpdateObjective,
    pillarByObjectiveId,
    pillars,
    positions,
    showGoal,
    statusById,
    statuses,
    strategy,
    teamCodeById,
  }: StrategyCanvasNodesProps) => (
    <>
      {showGoal ? (
        <GoalCanvasNode
          averageProgress={averageProgress}
          canEdit={canEdit}
          draggingNodeId={draggingNodeId}
          objectiveCount={objectiveCount}
          onFinishNodeDrag={onFinishNodeDrag}
          onNodePointerDown={onNodePointerDown}
          onNodePointerMove={onNodePointerMove}
          onSelectGoal={onSelectGoal}
          position={getNodePosition(positions, layoutPositions, GOAL_NODE_ID)}
          strategy={strategy}
        />
      ) : null}

      {pillars.map((pillar) => (
        <PillarCanvasNode
          canEdit={canEdit}
          draggingNodeId={draggingNodeId}
          dropTargetPillarId={dropTargetPillarId}
          key={pillar.id}
          objectiveById={objectiveById}
          onDeletePillar={onDeletePillar}
          onFinishNodeDrag={onFinishNodeDrag}
          onNodePointerDown={onNodePointerDown}
          onNodePointerMove={onNodePointerMove}
          onSelectPillar={onSelectPillar}
          pillar={pillar}
          position={getNodePosition(
            positions,
            layoutPositions,
            getPillarNodeId(pillar.id),
          )}
        />
      ))}

      {strategy.pillars.length === 0 ? (
        <button
          className="border-border text-text-muted hover:border-foreground/35 hover:bg-surface-elevated/35 hover:text-foreground disabled:hover:border-border disabled:hover:text-text-muted absolute rounded-xl border-2 border-dashed px-8 py-7 text-center transition-colors disabled:cursor-default disabled:hover:bg-transparent"
          data-no-drag
          disabled={!canEdit}
          onClick={onAddPillar}
          style={{
            left: (layoutWidth - GOAL_NODE_WIDTH) / 2,
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

      {objectives.map((objective) => (
        <ObjectiveCanvasNode
          canEdit={canEdit}
          draggingNodeId={draggingNodeId}
          expandedObjectiveIds={expandedObjectiveIds}
          key={objective.id}
          keyResultsByObjective={keyResultsByObjective}
          memberById={memberById}
          objective={objective}
          onAlign={onAlign}
          onFinishNodeDrag={onFinishNodeDrag}
          onNodePointerDown={onNodePointerDown}
          onNodePointerMove={onNodePointerMove}
          onSelectObjective={onSelectObjective}
          onToggleObjectiveKeyResults={onToggleObjectiveKeyResults}
          onUpdateObjective={onUpdateObjective}
          pillarByObjectiveId={pillarByObjectiveId}
          position={getNodePosition(
            positions,
            layoutPositions,
            getObjectiveNodeId(objective.id),
          )}
          statusById={statusById}
          statuses={statuses}
          strategy={strategy}
          teamCodeById={teamCodeById}
        />
      ))}

      {keyResultNodes.map(({ keyResult, objective }) => (
        <KeyResultCanvasNode
          draggingNodeId={draggingNodeId}
          key={keyResult.id}
          keyResult={keyResult}
          keyResultContextMenu={keyResultContextMenu}
          memberById={memberById}
          objective={objective}
          onFinishNodeDrag={onFinishNodeDrag}
          onNodePointerDown={onNodePointerDown}
          onNodePointerMove={onNodePointerMove}
          onSelectKeyResult={onSelectKeyResult}
          position={getNodePosition(
            positions,
            layoutPositions,
            getKeyResultNodeId(keyResult.id),
          )}
          teamCodeById={teamCodeById}
        />
      ))}
    </>
  ),
);
StrategyCanvasNodes.displayName = "StrategyCanvasNodes";

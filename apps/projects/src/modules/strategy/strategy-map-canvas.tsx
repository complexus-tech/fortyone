"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import Link from "next/link";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useDraggable,
  useDroppable,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import {
  ArrowDownIcon,
  DeleteIcon,
  DragIcon,
  ObjectiveIcon,
  OKRIcon,
} from "icons";
import { cn } from "lib";
import { Avatar, Badge, Box, Button, Flex, Select, Text } from "ui";
import {
  AssigneesMenu,
  ObjectiveHealthIcon,
  PrioritiesMenu,
  PriorityIcon,
} from "@/components/ui";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { ObjectiveStatusesMenu } from "@/components/ui/objective-statuses-menu";
import { useWorkspacePath } from "@/hooks";
import { useMembers } from "@/lib/hooks/members";
import { useObjectiveStatuses } from "@/lib/hooks/objective-statuses";
import { ObjectiveHealthEditor } from "@/modules/objectives/components/objective-health-editor";
import {
  useCanUpdateObjective,
  useKeyResults,
  useUpdateObjectiveMutation,
} from "@/modules/objectives/hooks";
import type { Objective, ObjectiveUpdate } from "@/modules/objectives/types";
import { hexToRgba } from "@/utils";
import type { StrategicPillar, StrategyMap } from "./types";

const PILLAR_DROP_PREFIX = "pillar:";
const PILLAR_POSITION_PREFIX = "pillar-position:";
const UNALIGNED_DROP_ID = "unaligned";
const NUMBER_FORMAT = new Intl.NumberFormat();
const MAX_PILLAR_SPACING = 160;
const MAX_OBJECTIVE_SPACING = 160;

const getObjectiveProgress = (objective: Objective) => {
  const total = objective.stats?.total ?? 0;
  const completed = objective.stats?.completed ?? 0;
  return total > 0 ? Math.round((completed / total) * 100) : 0;
};

const getKeyResultProgress = (
  startValue: number,
  currentValue: number,
  targetValue: number,
) => {
  const range = targetValue - startValue;
  if (range === 0) return currentValue >= targetValue ? 100 : 0;
  return Math.max(
    0,
    Math.min(100, Math.round(((currentValue - startValue) / range) * 100)),
  );
};

const formatValue = (value: number, measurementType: string) => {
  if (measurementType === "percentage") return `${value}%`;
  if (measurementType === "boolean") {
    return value >= 1 ? "Complete" : "Incomplete";
  }
  return NUMBER_FORMAT.format(value);
};

const HierarchyBadge = ({ children }: { children: ReactNode }) => (
  <Badge color="tertiary" rounded="md" size="md">
    {children}
  </Badge>
);

const CollapseButton = ({
  isCollapsed,
  label,
  onToggle,
}: {
  isCollapsed: boolean;
  label: string;
  onToggle: () => void;
}) => (
  <Button
    aria-expanded={!isCollapsed}
    aria-label={label}
    asIcon
    color="tertiary"
    onClick={onToggle}
    rounded="md"
    size="sm"
    type="button"
    variant="naked"
  >
    <ArrowDownIcon
      className={cn(
        "text-text-muted h-4 w-auto transition-transform",
        isCollapsed && "-rotate-90",
      )}
      strokeWidth={1.8}
    />
  </Button>
);

const ProgressBar = ({ progress }: { progress: number }) => (
  <Box className="bg-surface-muted h-1.5 flex-1 overflow-hidden rounded-full">
    <Box
      className="bg-foreground h-full rounded-full transition-[width]"
      style={{ width: `${progress}%` }}
    />
  </Box>
);

type SiblingTreeItem = {
  id: string;
  node: ReactNode;
  spacingBefore?: number;
};

const SiblingTree = ({ items }: { items: SiblingTreeItem[] }) => {
  if (items.length === 0) return null;

  return (
    <div className="w-max min-w-full">
      <span aria-hidden className="bg-border mx-auto block h-8 w-px" />
      <div className="flex gap-12">
        {items.map((item, index) => {
          const isFirst = index === 0;
          const isLast = index === items.length - 1;
          const spacingBefore = item.spacingBefore ?? 0;

          return (
            <div
              className="relative shrink-0 pt-8"
              key={item.id}
              style={{ marginLeft: spacingBefore }}
            >
              <span
                aria-hidden
                className="bg-border absolute top-0 left-1/2 h-8 w-px"
              />
              {!isFirst ? (
                <span
                  aria-hidden
                  className="bg-border absolute top-0 right-1/2 h-px"
                  style={{ left: -(48 + spacingBefore) }}
                />
              ) : null}
              {!isLast ? (
                <span
                  aria-hidden
                  className="bg-border absolute top-0 right-[-3rem] left-1/2 h-px"
                />
              ) : null}
              {item.node}
            </div>
          );
        })}
      </div>
    </div>
  );
};

const KeyResultTree = ({ objectiveId }: { objectiveId: string }) => {
  const { data: keyResults = [], isPending } = useKeyResults(objectiveId);

  if (isPending) {
    return (
      <SiblingTree
        items={[
          {
            id: "loading",
            node: (
              <Box className="border-border bg-surface-muted/40 h-20 w-[20rem] animate-pulse rounded-xl border" />
            ),
          },
        ]}
      />
    );
  }

  if (keyResults.length === 0) {
    return (
      <SiblingTree
        items={[
          {
            id: "empty",
            node: (
              <Box className="border-border w-[20rem] rounded-xl border border-dashed px-4 py-3 text-center">
                <Text color="muted">No key results yet</Text>
              </Box>
            ),
          },
        ]}
      />
    );
  }

  return (
    <SiblingTree
      items={keyResults.map((keyResult) => {
        const progress = getKeyResultProgress(
          keyResult.startValue,
          keyResult.currentValue,
          keyResult.targetValue,
        );

        return {
          id: keyResult.id,
          node: (
            <Box className="border-border bg-background w-[20rem] rounded-xl border px-4 py-3.5 shadow-[0_1px_2px_rgba(0,0,0,0.04)]">
              <Flex align="start" className="gap-3">
                <Box className="bg-surface-muted mt-0.5 grid h-8 w-8 shrink-0 place-items-center rounded-lg">
                  <OKRIcon className="h-4 w-auto" strokeWidth={2} />
                </Box>
                <Box className="min-w-0 flex-1">
                  <HierarchyBadge>Key Result</HierarchyBadge>
                  <Text className="mt-2 line-clamp-2" fontWeight="medium">
                    {keyResult.name}
                  </Text>
                  <Flex align="center" className="mt-3 gap-3">
                    <ProgressBar progress={progress} />
                    <Text className="shrink-0" color="muted">
                      {formatValue(
                        keyResult.currentValue,
                        keyResult.measurementType,
                      )}
                      {" / "}
                      {formatValue(
                        keyResult.targetValue,
                        keyResult.measurementType,
                      )}
                    </Text>
                  </Flex>
                </Box>
              </Flex>
            </Box>
          ),
        };
      })}
    />
  );
};

const ObjectiveProperties = ({
  canEdit,
  objective,
}: {
  canEdit: boolean;
  objective: Objective;
}) => {
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useObjectiveStatuses();
  const canUpdateObjective = useCanUpdateObjective();
  const updateMutation = useUpdateObjectiveMutation();
  const lead = members.find(({ id }) => id === objective.leadUser);
  const status = statuses.find(({ id }) => id === objective.statusId);
  const canUpdate = canEdit && canUpdateObjective;

  const handleUpdate = (data: ObjectiveUpdate) => {
    if (!canUpdate) return;
    updateMutation.mutate({ objectiveId: objective.id, data });
  };

  return (
    <Flex align="center" className="gap-2" wrap>
      <AssigneesMenu>
        <AssigneesMenu.Trigger>
          <Button
            aria-label={lead ? `Lead: ${lead.username}` : "Add lead"}
            className="gap-2 px-2"
            color="tertiary"
            disabled={!canUpdate}
            rounded="md"
            size="sm"
            type="button"
            variant="outline"
          >
            <Avatar
              name={lead?.fullName || lead?.username}
              rounded="md"
              size="xs"
              src={lead?.avatarUrl}
            />
            <span className="max-w-28 truncate">
              {lead?.username ?? "Lead"}
            </span>
          </Button>
        </AssigneesMenu.Trigger>
        <AssigneesMenu.Items
          assigneeId={objective.leadUser}
          onAssigneeSelected={(leadUser) => {
            handleUpdate({ leadUser });
          }}
          teamId={objective.teamId}
        />
      </AssigneesMenu>

      <ObjectiveStatusesMenu>
        <ObjectiveStatusesMenu.Trigger>
          <Button
            className="gap-2 pr-2.5"
            disabled={!canUpdate}
            rounded="md"
            size="sm"
            style={{
              backgroundColor: hexToRgba(status?.color, 0.1),
              borderColor: hexToRgba(status?.color, 0.22),
            }}
            type="button"
            variant="outline"
          >
            <ObjectiveStatusIcon statusId={objective.statusId} />
            {status?.name ?? "Status"}
          </Button>
        </ObjectiveStatusesMenu.Trigger>
        <ObjectiveStatusesMenu.Items
          setStatusId={(statusId) => {
            handleUpdate({ statusId });
          }}
          statusId={objective.statusId}
        />
      </ObjectiveStatusesMenu>

      <PrioritiesMenu>
        <PrioritiesMenu.Trigger>
          <Button
            className="gap-2 pr-2.5"
            color="tertiary"
            disabled={!canUpdate}
            rounded="md"
            size="sm"
            type="button"
            variant="outline"
          >
            <PriorityIcon
              className="h-[1.15rem]"
              priority={objective.priority}
            />
            {objective.priority ?? "No Priority"}
          </Button>
        </PrioritiesMenu.Trigger>
        <PrioritiesMenu.Items
          priority={objective.priority}
          setPriority={(priority) => {
            handleUpdate({ priority });
          }}
        />
      </PrioritiesMenu>

      <ObjectiveHealthEditor
        health={objective.health}
        objectiveId={objective.id}
      >
        <Button
          className="gap-2 pr-2.5"
          color="tertiary"
          disabled={!canUpdate}
          rounded="md"
          size="sm"
          type="button"
          variant="outline"
        >
          <ObjectiveHealthIcon health={objective.health} />
          {objective.health ?? "No Health"}
        </Button>
      </ObjectiveHealthEditor>
    </Flex>
  );
};

const ObjectiveBranch = ({
  canEdit,
  objective,
  onAlign,
  pillars,
}: {
  canEdit: boolean;
  objective: Objective;
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  pillars: StrategicPillar[];
}) => {
  const [isCollapsed, setIsCollapsed] = useState(true);
  const { withWorkspace } = useWorkspacePath();
  const { data: statuses = [] } = useObjectiveStatuses();
  const status = statuses.find(({ id }) => id === objective.statusId);
  const progress = getObjectiveProgress(objective);
  const { attributes, isDragging, listeners, setNodeRef } = useDraggable({
    id: objective.id,
    disabled: !canEdit,
  });

  return (
    <div className="w-max min-w-[23rem]" ref={setNodeRef}>
      <Box
        className={cn(
          "border-border bg-surface mx-auto w-[23rem] rounded-xl border p-4 shadow-[0_1px_3px_rgba(0,0,0,0.05)] transition-[opacity,background-color,border-color]",
          isDragging && "opacity-40",
        )}
        style={{
          backgroundColor: hexToRgba(status?.color, 0.045),
          borderColor: hexToRgba(status?.color, 0.2),
          boxShadow: `inset 3px 0 0 ${status?.color ?? "transparent"}, 0 1px 3px rgba(0,0,0,0.05)`,
        }}
      >
        <Flex align="start" className="gap-3" justify="between">
          <Box className="min-w-0 flex-1">
            <Flex align="center" className="gap-2">
              <HierarchyBadge>Objective</HierarchyBadge>
              <CollapseButton
                isCollapsed={isCollapsed}
                label={`${isCollapsed ? "Show" : "Hide"} key results for ${objective.name}`}
                onToggle={() => {
                  setIsCollapsed((current) => !current);
                }}
              />
            </Flex>
            <Link
              className="mt-2 block hover:opacity-75"
              href={withWorkspace(
                `/teams/${objective.teamId}/objectives/${objective.id}`,
              )}
            >
              <Text
                className="line-clamp-2 text-[1.1rem]"
                fontWeight="semibold"
              >
                {objective.name}
              </Text>
            </Link>
          </Box>
          <Flex align="center" className="shrink-0 gap-1.5">
            <button
              aria-label={`Drag ${objective.name}`}
              className="text-text-secondary hover:bg-state-hover hover:text-text-primary flex h-8 w-8 cursor-grab items-center justify-center rounded-lg active:cursor-grabbing disabled:cursor-default disabled:opacity-40"
              disabled={!canEdit}
              type="button"
              {...attributes}
              {...listeners}
            >
              <DragIcon className="h-4 w-4" />
            </button>
            <Select
              disabled={!canEdit}
              onValueChange={(value) => {
                onAlign(objective.id, value === "__unaligned" ? null : value);
              }}
              value={
                pillars.find((pillar) =>
                  pillar.objectiveIds.includes(objective.id),
                )?.id ?? "__unaligned"
              }
            >
              <Select.Trigger
                aria-label="Move objective"
                className="h-8 w-auto bg-transparent px-2.5 text-[1rem]"
                title="Move objective"
              >
                Move
              </Select.Trigger>
              <Select.Content align="end">
                <Select.Option className="text-[1rem]" value="__unaligned">
                  Unaligned
                </Select.Option>
                {pillars.map((pillar) => (
                  <Select.Option
                    className="text-[1rem]"
                    key={pillar.id}
                    value={pillar.id}
                  >
                    {pillar.name}
                  </Select.Option>
                ))}
              </Select.Content>
            </Select>
          </Flex>
        </Flex>

        <Flex align="center" className="mt-4 gap-3">
          <ProgressBar progress={progress} />
          <Text color="muted">{progress}%</Text>
        </Flex>
        <Box className="mt-4">
          <ObjectiveProperties canEdit={canEdit} objective={objective} />
        </Box>
      </Box>

      {!isCollapsed ? <KeyResultTree objectiveId={objective.id} /> : null}
    </div>
  );
};

const PillarBranch = ({
  canAdjustSpacing,
  canEdit,
  objectiveSpacing,
  objectives,
  onAlign,
  onDelete,
  onEdit,
  pillar,
  pillars,
}: {
  canAdjustSpacing: boolean;
  canEdit: boolean;
  objectiveSpacing: Record<string, number>;
  objectives: Objective[];
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  onDelete: (pillarId: string) => void;
  onEdit: (pillar: StrategicPillar) => void;
  pillar: StrategicPillar;
  pillars: StrategicPillar[];
}) => {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const { isOver, setNodeRef } = useDroppable({
    id: `${PILLAR_DROP_PREFIX}${pillar.id}`,
    disabled: !canEdit,
  });
  const {
    attributes,
    isDragging,
    listeners,
    setNodeRef: setPositionNodeRef,
    transform,
  } = useDraggable({
    id: `${PILLAR_POSITION_PREFIX}${pillar.id}`,
    disabled: !canEdit || !canAdjustSpacing,
  });
  let objectiveTree: ReactNode = null;

  if (!isCollapsed) {
    objectiveTree = (
      <SiblingTree
        items={
          objectives.length > 0
            ? objectives.map((objective) => ({
                id: objective.id,
                node: (
                  <ObjectiveBranch
                    canEdit={canEdit}
                    objective={objective}
                    onAlign={onAlign}
                    pillars={pillars}
                  />
                ),
                spacingBefore: objectiveSpacing[objective.id] ?? 0,
              }))
            : [
                {
                  id: "empty",
                  node: (
                    <Box className="border-border w-[23rem] rounded-xl border border-dashed px-5 py-7 text-center">
                      <Text color="muted">
                        Drag an objective here, or use Move on a card.
                      </Text>
                    </Box>
                  ),
                },
              ]
        }
      />
    );
  }

  return (
    <div
      className={cn(
        "relative w-max min-w-[25rem] shrink-0 transition-colors",
        isOver && "bg-state-hover rounded-xl",
      )}
      ref={setNodeRef}
    >
      <div
        className={cn(
          "mx-auto w-[23rem] transition-opacity",
          isDragging && "opacity-60",
        )}
        ref={setPositionNodeRef}
        style={{
          transform: transform
            ? `translate3d(${transform.x}px, ${transform.y}px, 0)`
            : undefined,
        }}
      >
        <Box className="border-border bg-surface-muted/45 rounded-xl border p-4 shadow-[0_1px_2px_rgba(0,0,0,0.04)]">
          <Flex align="start" className="gap-3" justify="between">
            <Box className="min-w-0 flex-1">
              <Flex align="center" className="gap-2">
                <HierarchyBadge>Strategic Pillar</HierarchyBadge>
                <CollapseButton
                  isCollapsed={isCollapsed}
                  label={`${isCollapsed ? "Show" : "Hide"} objectives for ${pillar.name}`}
                  onToggle={() => {
                    setIsCollapsed((current) => !current);
                  }}
                />
              </Flex>
              <button
                className="mt-2 block max-w-full text-left hover:opacity-75"
                disabled={!canEdit}
                onClick={() => {
                  onEdit(pillar);
                }}
                type="button"
              >
                <Text
                  className="line-clamp-2 text-[1.1rem]"
                  fontWeight="semibold"
                >
                  {pillar.name}
                </Text>
              </button>
            </Box>
            <Flex align="center" className="shrink-0 gap-1">
              {canAdjustSpacing ? (
                <button
                  aria-label={`Adjust spacing before ${pillar.name}`}
                  className="text-text-secondary hover:bg-state-hover hover:text-text-primary flex h-8 w-8 cursor-ew-resize items-center justify-center rounded-lg active:cursor-grabbing disabled:cursor-default disabled:opacity-40"
                  disabled={!canEdit}
                  title="Drag horizontally to adjust branch spacing"
                  type="button"
                  {...attributes}
                  {...listeners}
                >
                  <DragIcon className="h-4 w-4" />
                </button>
              ) : null}
              <Button
                aria-label={`Delete ${pillar.name}`}
                asIcon
                color="tertiary"
                disabled={!canEdit}
                onClick={() => {
                  onDelete(pillar.id);
                }}
                rounded="md"
                size="sm"
                title="Delete pillar"
                type="button"
                variant="naked"
              >
                <DeleteIcon className="h-4 w-auto" />
              </Button>
            </Flex>
          </Flex>
          {pillar.description ? (
            <Text className="mt-3 line-clamp-3" color="muted">
              {pillar.description}
            </Text>
          ) : null}
          <Flex align="center" className="mt-4 gap-2">
            <ObjectiveIcon className="text-text-muted h-4 w-auto" />
            <Text color="muted">
              {objectives.length} objective{objectives.length === 1 ? "" : "s"}
            </Text>
          </Flex>
        </Box>
      </div>

      {objectiveTree}
    </div>
  );
};

const UnalignedObjectives = ({
  canEdit,
  objectives,
  onAlign,
  pillars,
}: {
  canEdit: boolean;
  objectives: Objective[];
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  pillars: StrategicPillar[];
}) => {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const { isOver, setNodeRef } = useDroppable({
    id: UNALIGNED_DROP_ID,
    disabled: !canEdit,
  });
  let objectiveContent: ReactNode = null;

  if (!isCollapsed) {
    objectiveContent =
      objectives.length > 0 ? (
        <Box className="grid grid-cols-3 gap-5">
          {objectives.map((objective) => (
            <ObjectiveBranch
              canEdit={canEdit}
              key={objective.id}
              objective={objective}
              onAlign={onAlign}
              pillars={pillars}
            />
          ))}
        </Box>
      ) : (
        <Box className="border-border rounded-xl border border-dashed px-5 py-8 text-center">
          <Text color="muted">There are no unaligned objectives.</Text>
        </Box>
      );
  }

  return (
    <div
      className={cn(
        "mt-12 w-full max-w-[78rem] rounded-xl p-3 transition-colors",
        isOver && "bg-state-hover",
      )}
      ref={setNodeRef}
    >
      <Flex align="center" className="mb-4 gap-2">
        <CollapseButton
          isCollapsed={isCollapsed}
          label={`${isCollapsed ? "Show" : "Hide"} unaligned objectives`}
          onToggle={() => {
            setIsCollapsed((current) => !current);
          }}
        />
        <Text fontWeight="semibold">Unaligned objectives</Text>
        <Text color="muted">
          Drag an objective here to remove its pillar alignment.
        </Text>
      </Flex>
      {objectiveContent}
    </div>
  );
};

export const StrategyMapCanvas = ({
  strategy,
  objectives,
  showUnaligned,
  onEditGoal,
  onAlign,
  onDeletePillar,
  onEditPillar,
  canEdit,
  zoom,
}: {
  strategy: StrategyMap;
  objectives: Objective[];
  showUnaligned: boolean;
  onEditGoal: () => void;
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  onDeletePillar: (pillarId: string) => void;
  onEditPillar: (pillar: StrategicPillar) => void;
  canEdit: boolean;
  zoom: number;
}) => {
  const [activeObjectiveId, setActiveObjectiveId] = useState<string | null>(
    null,
  );
  const [isGoalCollapsed, setIsGoalCollapsed] = useState(false);
  const [pillarSpacing, setPillarSpacing] = useState<Record<string, number>>(
    {},
  );
  const [objectiveSpacing, setObjectiveSpacing] = useState<
    Record<string, number>
  >({});
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
  );
  const alignedIds = new Set(
    strategy.pillars.flatMap((pillar) => pillar.objectiveIds),
  );
  const unaligned = objectives.filter(
    (objective) => !alignedIds.has(objective.id),
  );
  const activeObjective = objectives.find(
    (objective) => objective.id === activeObjectiveId,
  );

  const handleDragStart = ({ active }: DragStartEvent) => {
    const activeId = active.id.toString();
    if (!activeId.startsWith(PILLAR_POSITION_PREFIX)) {
      setActiveObjectiveId(activeId);
    }
  };

  const handleDragEnd = ({ active, delta, over }: DragEndEvent) => {
    setActiveObjectiveId(null);
    if (!canEdit) return;

    const objectiveId = active.id.toString();
    if (objectiveId.startsWith(PILLAR_POSITION_PREFIX)) {
      const pillarId = objectiveId.slice(PILLAR_POSITION_PREFIX.length);
      setPillarSpacing((current) => ({
        ...current,
        [pillarId]: Math.max(
          0,
          Math.min(MAX_PILLAR_SPACING, (current[pillarId] ?? 0) + delta.x),
        ),
      }));
      return;
    }

    if (!over) return;

    const destination = over.id.toString();
    const currentPillarId =
      strategy.pillars.find((pillar) =>
        pillar.objectiveIds.includes(objectiveId),
      )?.id ?? null;
    let nextPillarId: string | null | undefined;

    if (destination.startsWith(PILLAR_DROP_PREFIX)) {
      nextPillarId = destination.slice(PILLAR_DROP_PREFIX.length);
    } else if (destination === UNALIGNED_DROP_ID) {
      nextPillarId = null;
    }

    if (nextPillarId === currentPillarId) {
      if (currentPillarId && Math.abs(delta.x) > 12) {
        setObjectiveSpacing((current) => ({
          ...current,
          [objectiveId]: Math.max(
            0,
            Math.min(
              MAX_OBJECTIVE_SPACING,
              (current[objectiveId] ?? 0) + delta.x,
            ),
          ),
        }));
      }
      return;
    }

    if (nextPillarId === undefined) return;
    setObjectiveSpacing((current) => ({ ...current, [objectiveId]: 0 }));
    onAlign(objectiveId, nextPillarId);
  };

  return (
    <DndContext
      onDragCancel={() => {
        setActiveObjectiveId(null);
      }}
      onDragEnd={handleDragEnd}
      onDragStart={handleDragStart}
      sensors={sensors}
    >
      <Box className="bg-surface-muted/20 relative h-full overflow-auto">
        <Box
          className="min-w-max origin-top-left p-10 transition-transform"
          style={{ transform: `scale(${zoom})`, width: `${100 / zoom}%` }}
        >
          <Flex align="center" direction="column">
            <Box className="border-border bg-background w-[34rem] max-w-[calc(100vw-5rem)] rounded-xl border px-6 py-5 shadow-[0_1px_3px_rgba(0,0,0,0.05)]">
              <Flex align="start" className="gap-4" justify="between">
                <button
                  className="min-w-0 flex-1 text-left hover:opacity-75"
                  disabled={!canEdit}
                  onClick={onEditGoal}
                  type="button"
                >
                  <HierarchyBadge>Ultimate Goal</HierarchyBadge>
                  <Text className="mt-3 text-xl" fontWeight="semibold">
                    {strategy.ultimateGoal || "Define your ultimate goal"}
                  </Text>
                  <Text className="mt-2" color="muted">
                    {strategy.description ||
                      "Describe the long-term outcome that every pillar should support."}
                  </Text>
                </button>
                <CollapseButton
                  isCollapsed={isGoalCollapsed}
                  label={`${isGoalCollapsed ? "Show" : "Hide"} strategic pillars`}
                  onToggle={() => {
                    setIsGoalCollapsed((current) => !current);
                  }}
                />
              </Flex>
            </Box>

            {!isGoalCollapsed ? (
              <>
                {strategy.pillars.length > 0 ? (
                  <SiblingTree
                    items={strategy.pillars.map((pillar, index) => ({
                      id: pillar.id,
                      node: (
                        <PillarBranch
                          canAdjustSpacing={index > 0}
                          canEdit={canEdit}
                          objectiveSpacing={objectiveSpacing}
                          objectives={objectives.filter((objective) =>
                            pillar.objectiveIds.includes(objective.id),
                          )}
                          onAlign={onAlign}
                          onDelete={onDeletePillar}
                          onEdit={onEditPillar}
                          pillar={pillar}
                          pillars={strategy.pillars}
                        />
                      ),
                      spacingBefore: pillarSpacing[pillar.id] ?? 0,
                    }))}
                  />
                ) : (
                  <>
                    <span aria-hidden className="bg-border h-8 w-px shrink-0" />
                    <Box className="border-border w-[34rem] rounded-xl border border-dashed px-8 py-10 text-center">
                      <Text fontWeight="medium">
                        Add the strategic pillars that make the goal achievable.
                      </Text>
                    </Box>
                  </>
                )}

                {showUnaligned ? (
                  <UnalignedObjectives
                    canEdit={canEdit}
                    objectives={unaligned}
                    onAlign={onAlign}
                    pillars={strategy.pillars}
                  />
                ) : null}
              </>
            ) : null}
          </Flex>
        </Box>
      </Box>
      <DragOverlay>
        {activeObjective ? (
          <Box className="border-border bg-surface w-[23rem] rounded-xl border px-4 py-3 shadow-xl">
            <HierarchyBadge>Objective</HierarchyBadge>
            <Text className="mt-2 line-clamp-2" fontWeight="semibold">
              {activeObjective.name}
            </Text>
          </Box>
        ) : null}
      </DragOverlay>
    </DndContext>
  );
};

"use client";

import type { ReactNode } from "react";
import { useMemo, useState } from "react";
import Link from "next/link";
import { Box, Flex, Select, Text } from "ui";
import { cn } from "lib";
import { MinusIcon, MoreHorizontalIcon, OKRIcon, PlusIcon } from "icons";
import { useWorkspacePath } from "@/hooks";
import { useKeyResults } from "@/modules/objectives/hooks";
import type { Objective } from "@/modules/objectives/types";
import type { StrategicPillar, StrategyMap } from "./types";

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
  if (measurementType === "boolean")
    return value >= 1 ? "Complete" : "Incomplete";
  return new Intl.NumberFormat().format(value);
};

const NodeLabel = ({ children }: { children: ReactNode }) => (
  <Text
    className="text-[0.95rem] tracking-[0.08em] uppercase"
    color="muted"
    fontWeight="medium"
  >
    {children}
  </Text>
);

const MonochromeProgress = ({ progress }: { progress: number }) => (
  <Box className="bg-surface-muted h-1.5 flex-1 overflow-hidden rounded-full">
    <Box
      className="bg-foreground h-full rounded-full transition-[width]"
      style={{ width: `${progress}%` }}
    />
  </Box>
);

const KeyResults = ({ objectiveId }: { objectiveId: string }) => {
  const { data: keyResults = [], isPending } = useKeyResults(objectiveId);

  if (isPending) {
    return (
      <Box className="border-border bg-surface-muted/30 mt-3 h-16 animate-pulse rounded-lg border" />
    );
  }

  if (keyResults.length === 0) {
    return (
      <Box className="border-border mt-3 rounded-lg border border-dashed px-4 py-3">
        <Text className="text-[0.95rem]" color="muted">
          No key results yet
        </Text>
      </Box>
    );
  }

  return (
    <Flex className="mt-3 gap-2.5" direction="column">
      {keyResults.map((keyResult) => {
        const progress = getKeyResultProgress(
          keyResult.startValue,
          keyResult.currentValue,
          keyResult.targetValue,
        );
        return (
          <Box
            className="border-border bg-background rounded-lg border p-3.5"
            key={keyResult.id}
          >
            <Flex align="start" className="gap-3">
              <OKRIcon className="mt-0.5 h-[1.05rem] w-auto shrink-0" />
              <Box className="min-w-0 flex-1">
                <NodeLabel>Key Result</NodeLabel>
                <Text className="mt-1 line-clamp-2" fontWeight="medium">
                  {keyResult.name}
                </Text>
                <Flex align="center" className="mt-3 gap-3">
                  <MonochromeProgress progress={progress} />
                  <Text className="text-[0.95rem]" color="muted">
                    {formatValue(
                      keyResult.currentValue,
                      keyResult.measurementType,
                    )}
                    /
                    {formatValue(
                      keyResult.targetValue,
                      keyResult.measurementType,
                    )}
                  </Text>
                </Flex>
              </Box>
            </Flex>
          </Box>
        );
      })}
    </Flex>
  );
};

const ObjectiveCard = ({
  objective,
  pillars,
  onAlign,
}: {
  objective: Objective;
  pillars: StrategicPillar[];
  onAlign: (objectiveId: string, pillarId: string | null) => void;
}) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const { withWorkspace } = useWorkspacePath();
  const progress = getObjectiveProgress(objective);

  return (
    <Box className="border-border bg-surface rounded-xl border p-4 shadow-[0_1px_2px_rgba(0,0,0,0.04)]">
      <Flex align="start" justify="between">
        <NodeLabel>Objective</NodeLabel>
        <Select
          onValueChange={(value) => {
            onAlign(objective.id, value === "__unaligned" ? null : value);
          }}
          value={
            pillars.find((pillar) => pillar.objectiveIds.includes(objective.id))
              ?.id ?? "__unaligned"
          }
        >
          <Select.Trigger
            aria-label="Move objective"
            className="h-7 w-7 border-0 bg-transparent px-0 [&>svg]:hidden"
            title="Move objective"
          >
            <span className="sr-only">
              <Select.Input />
            </span>
            <MoreHorizontalIcon className="h-4 w-4" />
          </Select.Trigger>
          <Select.Content align="end">
            <Select.Option value="__unaligned">Unaligned</Select.Option>
            {pillars.map((pillar) => (
              <Select.Option key={pillar.id} value={pillar.id}>
                {pillar.name}
              </Select.Option>
            ))}
          </Select.Content>
        </Select>
      </Flex>
      <Link
        className="mt-1 block hover:opacity-75"
        href={withWorkspace(
          `/teams/${objective.teamId}/objectives/${objective.id}`,
        )}
      >
        <Text className="line-clamp-2 text-[1.02rem]" fontWeight="semibold">
          {objective.name}
        </Text>
      </Link>
      <Flex align="center" className="mt-4 gap-3">
        <MonochromeProgress progress={progress} />
        <Text className="text-[0.95rem]" color="muted">
          {progress}%
        </Text>
      </Flex>
      <Flex align="center" className="mt-3" justify="between">
        <Text className="text-[0.95rem]" color="muted">
          {objective.health ?? "No health update"}
        </Text>
        <button
          aria-expanded={isExpanded}
          className="text-text-secondary hover:text-text-primary flex items-center gap-1.5 font-medium"
          onClick={() => {
            setIsExpanded((current) => !current);
          }}
          type="button"
        >
          {isExpanded ? "Hide" : "Key results"}
          <span
            className={cn("transition-transform", isExpanded && "rotate-180")}
          >
            ⌄
          </span>
        </button>
      </Flex>
      {isExpanded ? <KeyResults objectiveId={objective.id} /> : null}
    </Box>
  );
};

const PillarColumn = ({
  pillar,
  objectives,
  pillars,
  onAlign,
  onDelete,
  onEdit,
}: {
  pillar: StrategicPillar;
  objectives: Objective[];
  pillars: StrategicPillar[];
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  onDelete: (pillarId: string) => void;
  onEdit: (pillar: StrategicPillar) => void;
}) => (
  <Box className="relative w-[22rem] shrink-0 pt-8">
    <span className="border-border absolute top-0 left-1/2 h-8 border-l" />
    <Box className="border-border bg-surface-muted/30 min-h-28 rounded-xl border p-4">
      <Flex align="start" justify="between">
        <NodeLabel>Strategic Pillar</NodeLabel>
        <button
          aria-label={`Delete ${pillar.name}`}
          className="text-text-secondary hover:text-text-primary"
          onClick={() => {
            onDelete(pillar.id);
          }}
          title="Delete pillar"
          type="button"
        >
          ×
        </button>
      </Flex>
      <button
        className="mt-1 text-left hover:opacity-75"
        onClick={() => {
          onEdit(pillar);
        }}
        type="button"
      >
        <Text className="text-[1.08rem]" fontWeight="semibold">
          {pillar.name}
        </Text>
      </button>
      {pillar.description ? (
        <Text className="mt-2 line-clamp-2 text-[0.95rem]" color="muted">
          {pillar.description}
        </Text>
      ) : null}
      <Text className="mt-3 text-[0.95rem]" color="muted">
        {objectives.length} objective{objectives.length === 1 ? "" : "s"}
      </Text>
    </Box>

    <Flex className="mt-4 gap-3" direction="column">
      {objectives.map((objective) => (
        <ObjectiveCard
          key={objective.id}
          objective={objective}
          onAlign={onAlign}
          pillars={pillars}
        />
      ))}
      {objectives.length === 0 ? (
        <Box className="border-border rounded-xl border border-dashed px-5 py-8 text-center">
          <Text className="text-[0.95rem]" color="muted">
            Move an objective here to connect execution to this pillar.
          </Text>
        </Box>
      ) : null}
    </Flex>
  </Box>
);

export const StrategyMapCanvas = ({
  strategy,
  objectives,
  showUnaligned,
  onEditGoal,
  onAlign,
  onDeletePillar,
  onEditPillar,
}: {
  strategy: StrategyMap;
  objectives: Objective[];
  showUnaligned: boolean;
  onEditGoal: () => void;
  onAlign: (objectiveId: string, pillarId: string | null) => void;
  onDeletePillar: (pillarId: string) => void;
  onEditPillar: (pillar: StrategicPillar) => void;
}) => {
  const [zoom, setZoom] = useState(1);
  const alignedIds = useMemo(
    () => new Set(strategy.pillars.flatMap((pillar) => pillar.objectiveIds)),
    [strategy.pillars],
  );
  const unaligned = objectives.filter(
    (objective) => !alignedIds.has(objective.id),
  );

  return (
    <Box className="bg-surface-muted/20 relative h-full overflow-auto">
      <Flex
        align="center"
        className="border-border bg-background sticky top-0 left-0 z-30 h-12 border-b px-5"
        justify="between"
      >
        <Flex align="center" className="gap-5">
          <Text className="text-[0.95rem]" color="muted">
            {strategy.pillars.length} pillars · {alignedIds.size} aligned
            objectives
          </Text>
          {showUnaligned ? (
            <Text className="text-[0.95rem]" color="muted">
              {unaligned.length} unaligned
            </Text>
          ) : null}
        </Flex>
        <Flex
          align="center"
          className="border-border overflow-hidden rounded-lg border"
        >
          <button
            aria-label="Zoom out"
            className="hover:bg-state-hover grid h-8 w-9 place-items-center"
            onClick={() => {
              setZoom((value) => Math.max(0.75, value - 0.1));
            }}
            type="button"
          >
            <MinusIcon className="h-4 w-4" />
          </button>
          <Text className="border-border min-w-14 border-x text-center text-[0.95rem]">
            {Math.round(zoom * 100)}%
          </Text>
          <button
            aria-label="Zoom in"
            className="hover:bg-state-hover grid h-8 w-9 place-items-center"
            onClick={() => {
              setZoom((value) => Math.min(1.25, value + 0.1));
            }}
            type="button"
          >
            <PlusIcon className="h-4 w-4" />
          </button>
          <button
            className="border-border hover:bg-state-hover h-8 border-l px-3 font-medium"
            onClick={() => {
              setZoom(1);
            }}
            type="button"
          >
            Reset
          </button>
        </Flex>
      </Flex>

      <Box
        className="origin-top-left p-10 transition-transform"
        style={{ transform: `scale(${zoom})`, width: `${100 / zoom}%` }}
      >
        <Flex align="center" direction="column">
          <button
            className="border-border bg-background hover:bg-state-hover w-[32rem] max-w-[calc(100vw-5rem)] rounded-xl border px-7 py-5 text-left shadow-[0_1px_3px_rgba(0,0,0,0.05)]"
            onClick={onEditGoal}
            type="button"
          >
            <NodeLabel>Ultimate Goal</NodeLabel>
            <Text className="mt-1.5 text-xl" fontWeight="semibold">
              {strategy.ultimateGoal || "Define your ultimate goal"}
            </Text>
            <Text className="mt-2 text-[0.95rem]" color="muted">
              {strategy.description ||
                "Describe the long-term outcome that every pillar should support."}
            </Text>
          </button>
          <span className="border-border h-10 border-l" />

          {strategy.pillars.length > 0 ? (
            <Box className="relative">
              <span className="border-border absolute top-0 right-[11rem] left-[11rem] border-t" />
              <Flex className="gap-7">
                {strategy.pillars.map((pillar) => (
                  <PillarColumn
                    key={pillar.id}
                    objectives={objectives.filter((objective) =>
                      pillar.objectiveIds.includes(objective.id),
                    )}
                    onAlign={onAlign}
                    onDelete={onDeletePillar}
                    onEdit={onEditPillar}
                    pillar={pillar}
                    pillars={strategy.pillars}
                  />
                ))}
              </Flex>
            </Box>
          ) : (
            <Box className="border-border w-[32rem] rounded-xl border border-dashed px-8 py-10 text-center">
              <Text fontWeight="medium">
                Add the strategic pillars that make the goal achievable.
              </Text>
            </Box>
          )}

          {showUnaligned && unaligned.length > 0 ? (
            <Box className="mt-10 w-full max-w-[72rem]">
              <Flex align="center" className="mb-4 gap-3">
                <Text fontWeight="semibold">Unaligned objectives</Text>
                <Text className="text-[0.95rem]" color="muted">
                  Connect these to a pillar using the menu on each card.
                </Text>
              </Flex>
              <Box className="grid grid-cols-3 gap-4">
                {unaligned.map((objective) => (
                  <ObjectiveCard
                    key={objective.id}
                    objective={objective}
                    onAlign={onAlign}
                    pillars={strategy.pillars}
                  />
                ))}
              </Box>
            </Box>
          ) : null}
        </Flex>
      </Box>
    </Box>
  );
};

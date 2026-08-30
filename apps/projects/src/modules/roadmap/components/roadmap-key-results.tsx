"use client";

import { useState } from "react";
import { CalendarIcon, OKRIcon } from "icons";
import { format } from "date-fns";
import {
  Avatar,
  Box,
  Button,
  CircleProgressBar,
  Flex,
  Popover,
  Text,
  Tooltip,
} from "ui";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { useTerminology } from "@/hooks";
import { useMembers } from "@/lib/hooks/members";
import { KeyResultContextMenu } from "@/modules/key-results/components/key-result-context-menu";
import {
  formatKeyResultValue,
  getKeyResultProgress,
  getKeyResultReference,
} from "@/modules/key-results/utils";
import { ObjectiveCard } from "@/modules/objectives/components/card";
import { useKeyResults } from "@/modules/objectives/hooks/use-key-results";
import type { KeyResult, Objective } from "@/modules/objectives/types";

const KeyResultPreview = ({
  interactive = false,
  keyResults,
  onSelect,
}: {
  interactive?: boolean;
  keyResults: KeyResult[];
  onSelect: (keyResult: KeyResult) => void;
}) => (
  <Box className="max-w-80 min-w-64">
    {keyResults.map((keyResult) => {
      const content = (
        <Flex align="center" className="min-w-0 gap-2.5" justify="between">
          <Text className="line-clamp-1 min-w-0">{keyResult.name}</Text>
          <Text className="shrink-0 text-[0.9rem] tabular-nums" color="muted">
            {getKeyResultProgress(keyResult)}%
          </Text>
        </Flex>
      );

      return interactive ? (
        <button
          className="border-border hover:bg-state-hover w-full border-t px-3 py-2.5 text-left first:border-t-0"
          key={keyResult.id}
          onClick={() => {
            onSelect(keyResult);
          }}
          type="button"
        >
          {content}
        </button>
      ) : (
        <Box
          className="border-border border-t py-2.5 first:border-t-0"
          key={keyResult.id}
        >
          {content}
        </Box>
      );
    })}
  </Box>
);

export const RoadmapKeyResultSummary = ({
  objective,
  onSelect,
}: {
  objective: Objective;
  onSelect: (keyResult: KeyResult) => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const { data: keyResults = [] } = useKeyResults(
    objective.id,
    objective.keyResultCount > 0,
  );

  if (objective.keyResultCount === 0) return null;

  return (
    <Popover>
      <Tooltip
        className="border-border-strong dark:border-border-strong dark:bg-surface-elevated p-3"
        delayDuration={250}
        title={
          keyResults.length > 0 ? (
            <KeyResultPreview keyResults={keyResults} onSelect={onSelect} />
          ) : (
            <Text color="muted">Loading key results…</Text>
          )
        }
      >
        <span>
          <Popover.Trigger asChild>
            <Button
              className="gap-1.5 px-2"
              color="tertiary"
              leftIcon={<OKRIcon className="h-4 w-4" strokeWidth={2} />}
              size="xs"
              type="button"
              variant="outline"
            >
              {objective.keyResultCount}{" "}
              {getTermDisplay("keyResultTerm", {
                variant: objective.keyResultCount === 1 ? "singular" : "plural",
              })}
            </Button>
          </Popover.Trigger>
        </span>
      </Tooltip>
      <Popover.Content align="start" className="w-80 p-1.5">
        {keyResults.length > 0 ? (
          <KeyResultPreview
            interactive
            keyResults={keyResults}
            onSelect={onSelect}
          />
        ) : (
          <Text className="px-3 py-2.5" color="muted">
            Loading key results…
          </Text>
        )}
      </Popover.Content>
    </Popover>
  );
};

const RoadmapKeyResultRow = ({
  keyResult,
  onSelect,
  teamCode,
}: {
  keyResult: KeyResult;
  onSelect: (keyResult: KeyResult) => void;
  teamCode?: string;
}) => {
  const { data: members = [] } = useMembers();
  const lead = keyResult.lead
    ? members.find(({ id }) => id === keyResult.lead)
    : undefined;
  const progress = getKeyResultProgress(keyResult);
  const reference = getKeyResultReference(teamCode, keyResult.sequenceId);
  const currentValue = formatKeyResultValue(
    keyResult.currentValue,
    keyResult.measurementType,
  );
  const measurementValue =
    keyResult.measurementType === "boolean"
      ? currentValue
      : `${currentValue} / ${formatKeyResultValue(
          keyResult.targetValue,
          keyResult.measurementType,
        )}`;

  return (
    <KeyResultContextMenu
      keyResult={keyResult}
      onOpenDetails={() => {
        onSelect(keyResult);
      }}
    >
      <RowWrapper className="px-5 py-2.5 md:pr-12 md:pl-18">
        <Box className="relative flex min-w-10 flex-1 items-center gap-2 @sm:min-w-20">
          <button
            className="focus-visible:ring-primary flex min-w-0 flex-1 items-center gap-2 rounded-sm text-left outline-none hover:opacity-90 focus-visible:ring-1"
            onClick={() => {
              onSelect(keyResult);
            }}
            type="button"
          >
            <OKRIcon
              className="text-text-muted h-[1.1rem] w-[1.1rem] shrink-0"
              strokeWidth={2}
            />
            {reference ? (
              <Text className="text-text-muted shrink-0 text-[0.875rem] uppercase">
                {reference}
              </Text>
            ) : null}
            <Text className="min-w-0 flex-1 truncate pr-4">
              {keyResult.name}
            </Text>
          </button>
        </Box>
        <Flex align="center" className="ml-4 shrink-0 gap-4">
          <Text
            className="hidden shrink-0 text-right tabular-nums md:block"
            color="muted"
          >
            {measurementValue}
          </Text>
          <Flex align="center" className="hidden w-[60px] gap-1.5 sm:flex">
            <CircleProgressBar progress={progress} size={16} strokeWidth={2} />
            <Text className="tabular-nums">{progress}%</Text>
          </Flex>
          <Flex align="center" className="hidden w-[100px] gap-1.5 lg:flex">
            <CalendarIcon className="text-text-muted h-4 shrink-0" />
            <Text className="truncate" color="muted">
              {format(new Date(keyResult.endDate), "MMM d, yy")}
            </Text>
          </Flex>
          <Tooltip title={lead?.fullName || lead?.username || "No lead"}>
            <span className="flex w-8 shrink-0 justify-end">
              <Avatar
                name={lead?.fullName || lead?.username}
                size="sm"
                src={lead?.avatarUrl}
              />
            </span>
          </Tooltip>
        </Flex>
      </RowWrapper>
    </KeyResultContextMenu>
  );
};

export const RoadmapObjectiveListItem = ({
  objective,
  onKeyResultSelect,
  onObjectiveSelect,
  onSelectionChange,
  expanded,
  onExpandedChange,
  selected,
  teamCode,
}: {
  objective: Objective;
  onKeyResultSelect: (keyResult: KeyResult) => void;
  onObjectiveSelect: () => void;
  onSelectionChange: (checked: boolean) => void;
  expanded?: boolean;
  onExpandedChange?: (expanded: boolean) => void;
  selected: boolean;
  teamCode?: string;
}) => {
  const [internallyExpanded, setInternallyExpanded] = useState(false);
  const isExpanded = expanded ?? internallyExpanded;
  const { data: keyResults = [] } = useKeyResults(
    objective.id,
    objective.keyResultCount > 0,
  );
  const keyResultProgress =
    keyResults.length > 0
      ? Math.round(
          keyResults.reduce(
            (total, keyResult) => total + getKeyResultProgress(keyResult),
            0,
          ) / keyResults.length,
        )
      : undefined;

  return (
    <Box>
      <ObjectiveCard
        {...objective}
        childCount={objective.keyResultCount}
        isExpanded={isExpanded}
        onSelect={onObjectiveSelect}
        onSelectionChange={onSelectionChange}
        onToggleExpanded={() => {
          const nextExpanded = !isExpanded;
          setInternallyExpanded(nextExpanded);
          onExpandedChange?.(nextExpanded);
        }}
        progress={keyResultProgress}
        selected={selected}
      />
      {isExpanded
        ? keyResults.map((keyResult) => (
            <RoadmapKeyResultRow
              key={keyResult.id}
              keyResult={keyResult}
              onSelect={onKeyResultSelect}
              teamCode={teamCode}
            />
          ))
        : null}
    </Box>
  );
};

"use client";

import { useDroppable } from "@dnd-kit/core";
import { cn } from "lib";
import { type ReactNode, useMemo, useRef, useState } from "react";
import { Box, Flex, Text } from "ui";
import {
  getHiddenObjectiveGroupKeys,
  hideObjectiveGroup,
  showObjectiveGroup,
  type ObjectiveGroup,
  type ObjectiveViewOptions,
} from "../objective-board-utils";
import {
  HiddenObjectiveGroups,
  ObjectiveGroupHeader,
} from "./objective-group-sections";
import {
  ObjectiveBoardCard,
  type ObjectiveBoardCardProps,
} from "./objective-board-card";
import { VirtualizedObjectiveItems } from "./virtualized-objective-items";

export const OBJECTIVE_GROUP_DND_ID_PREFIX = "objective-group:";

const KANBAN_COLUMN_WIDTH = 340;
const KANBAN_COLUMN_GAP = 24;
const KANBAN_COLUMN_STRIDE = KANBAN_COLUMN_WIDTH + KANBAN_COLUMN_GAP;
const ESTIMATED_KANBAN_CARD_HEIGHT = 176;

type Objective = ObjectiveBoardCardProps["objective"];

type ObjectivesKanbanProps = {
  activeObjectiveId?: string;
  canDrag: boolean;
  selectedObjectiveId?: string;
  groups: ObjectiveGroup[];
  viewOptions: ObjectiveViewOptions;
  setViewOptions: (viewOptions: ObjectiveViewOptions) => void;
  teamCodeById: ReadonlyMap<string, string>;
  onObjectiveSelect: (objective: Objective) => void;
  onCreateObjective: () => void;
  renderCardControls: (objective: Objective) => ReactNode;
};

type ObjectiveKanbanColumnProps = Pick<
  ObjectivesKanbanProps,
  | "activeObjectiveId"
  | "canDrag"
  | "selectedObjectiveId"
  | "onObjectiveSelect"
  | "teamCodeById"
  | "renderCardControls"
> & {
  group: ObjectiveGroup;
};

const ObjectiveKanbanColumn = ({
  activeObjectiveId,
  canDrag,
  selectedObjectiveId,
  group,
  onObjectiveSelect,
  teamCodeById,
  renderCardControls,
}: ObjectiveKanbanColumnProps) => {
  const scrollElementRef = useRef<HTMLDivElement>(null);
  const [focusedObjectiveId, setFocusedObjectiveId] = useState<string>();
  const { setNodeRef, isOver } = useDroppable({
    id: `${OBJECTIVE_GROUP_DND_ID_PREFIX}${group.key}`,
  });
  const pinnedObjectiveIds = [
    activeObjectiveId,
    focusedObjectiveId,
    selectedObjectiveId,
  ].filter((objectiveId): objectiveId is string => Boolean(objectiveId));

  return (
    <div
      aria-label={`Objectives in ${group.key}`}
      className={cn(
        "flex h-full w-[340px] flex-col gap-3 overflow-y-auto rounded-md pb-6 transition-colors",
        { "bg-surface-muted": isOver },
      )}
      ref={(element) => {
        scrollElementRef.current = element;
        setNodeRef(element);
      }}
      role="region"
    >
      <VirtualizedObjectiveItems
        className="w-full shrink-0"
        estimatedSize={ESTIMATED_KANBAN_CARD_HEIGHT}
        getItemKey={(objective) => objective.id}
        itemClassName="pb-3"
        items={group.objectives}
        onItemFocus={(objective) => {
          setFocusedObjectiveId(objective.id);
        }}
        overscan={3}
        pinnedKeys={pinnedObjectiveIds}
        renderItem={(objective) => (
          <ObjectiveBoardCard
            canDrag={canDrag}
            objective={objective}
            onSelect={onObjectiveSelect}
            teamCode={teamCodeById.get(objective.teamId)}
          >
            {renderCardControls(objective)}
          </ObjectiveBoardCard>
        )}
        scrollElementRef={scrollElementRef}
      />
    </div>
  );
};

type KanbanColumn =
  | {
      key: string;
      group: ObjectiveGroup;
      type: "group";
    }
  | {
      key: "hidden-groups";
      groups: ObjectiveGroup[];
      type: "hidden";
    };

export const ObjectivesKanban = ({
  activeObjectiveId,
  canDrag,
  selectedObjectiveId,
  groups,
  viewOptions,
  setViewOptions,
  teamCodeById,
  onObjectiveSelect,
  onCreateObjective,
  renderCardControls,
}: ObjectivesKanbanProps) => {
  const scrollElementRef = useRef<HTMLDivElement>(null);
  const [focusedColumnKey, setFocusedColumnKey] = useState<string>();
  const { columns, hiddenGroups, objectiveColumnKeyById } = useMemo(() => {
    const hiddenGroupKeySet = new Set(getHiddenObjectiveGroupKeys(viewOptions));
    const visibleGroups: ObjectiveGroup[] = [];
    const nextHiddenGroups: ObjectiveGroup[] = [];
    const nextObjectiveColumnKeyById = new Map<string, string>();

    for (const group of groups) {
      if (hiddenGroupKeySet.has(group.key)) {
        nextHiddenGroups.push(group);
      } else {
        visibleGroups.push(group);
      }
    }

    const nextColumns: KanbanColumn[] = visibleGroups.map((group) => {
      const key = `group:${group.key}`;

      for (const objective of group.objectives) {
        nextObjectiveColumnKeyById.set(objective.id, key);
      }

      return { group, key, type: "group" };
    });

    if (nextHiddenGroups.length > 0) {
      nextColumns.push({
        groups: nextHiddenGroups,
        key: "hidden-groups",
        type: "hidden",
      });
    }

    return {
      columns: nextColumns,
      hiddenGroups: nextHiddenGroups,
      objectiveColumnKeyById: nextObjectiveColumnKeyById,
    };
  }, [groups, viewOptions]);
  const activeColumnKey = activeObjectiveId
    ? objectiveColumnKeyById.get(activeObjectiveId)
    : undefined;
  const selectedColumnKey = selectedObjectiveId
    ? objectiveColumnKeyById.get(selectedObjectiveId)
    : undefined;
  const pinnedColumnKeys = [
    activeColumnKey,
    focusedColumnKey,
    selectedColumnKey,
  ].filter((columnKey): columnKey is string => Boolean(columnKey));

  return (
    <Box
      className="bg-surface-muted/40 dark:bg-surface-muted/20 h-full min-h-0 overflow-x-auto overflow-y-auto"
      data-body-container
      ref={scrollElementRef}
    >
      <Box className="sticky top-0 z-1 h-14 w-max px-6 backdrop-blur">
        <VirtualizedObjectiveItems
          axis="horizontal"
          className="h-full"
          estimatedSize={KANBAN_COLUMN_STRIDE}
          getItemKey={(column) => column.key}
          items={columns}
          onItemFocus={(column) => {
            setFocusedColumnKey(column.key);
          }}
          overscan={1}
          pinnedKeys={pinnedColumnKeys}
          renderItem={(column) =>
            column.type === "group" ? (
              <Box className="w-[340px] pl-1">
                <ObjectiveGroupHeader
                  group={column.group}
                  groupBy={viewOptions.groupBy}
                  onCreateObjective={onCreateObjective}
                  onHide={() => {
                    setViewOptions(
                      hideObjectiveGroup(viewOptions, column.group.key),
                    );
                  }}
                />
              </Box>
            ) : (
              <Flex align="center" className="w-[340px] pl-1" gap={2}>
                <Text color="muted" fontWeight="medium">
                  Hidden columns
                </Text>
                <Text color="muted">{hiddenGroups.length}</Text>
              </Flex>
            )
          }
          scrollElementRef={scrollElementRef}
        />
      </Box>
      <Box className="h-[calc(100%-3.5rem)] w-max px-7">
        <VirtualizedObjectiveItems
          axis="horizontal"
          className="h-full"
          estimatedSize={KANBAN_COLUMN_STRIDE}
          getItemKey={(column) => column.key}
          items={columns}
          onItemFocus={(column) => {
            setFocusedColumnKey(column.key);
          }}
          overscan={1}
          pinnedKeys={pinnedColumnKeys}
          renderItem={(column) =>
            column.type === "group" ? (
              <ObjectiveKanbanColumn
                activeObjectiveId={activeObjectiveId}
                canDrag={canDrag}
                group={column.group}
                onObjectiveSelect={onObjectiveSelect}
                renderCardControls={renderCardControls}
                selectedObjectiveId={selectedObjectiveId}
                teamCodeById={teamCodeById}
              />
            ) : (
              <HiddenObjectiveGroups
                groupBy={viewOptions.groupBy}
                groups={column.groups}
                onShow={(groupKey) => {
                  setViewOptions(showObjectiveGroup(viewOptions, groupKey));
                }}
              />
            )
          }
          scrollElementRef={scrollElementRef}
        />
      </Box>
    </Box>
  );
};

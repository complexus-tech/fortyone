"use client";

import {
  type ComponentProps,
  useCallback,
  useMemo,
  useRef,
  useState,
} from "react";
import { Box } from "ui";
import { getCategoryHeaderStyle } from "@/components/ui/category-header-style";
import type {
  ObjectiveGroup,
  ObjectiveViewOptions,
} from "../objective-board-utils";
import { ObjectivesToolbar } from "./objectives-toolbar";
import { ObjectiveGroupHeader } from "./objective-group-sections";
import { RoadmapObjectiveListItem } from "./roadmap-key-results";
import { VirtualizedObjectiveItems } from "./virtualized-objective-items";

const ESTIMATED_LIST_ROW_HEIGHT = 56;

type Objective = ComponentProps<typeof RoadmapObjectiveListItem>["objective"];
type KeyResult = Parameters<
  ComponentProps<typeof RoadmapObjectiveListItem>["onKeyResultSelect"]
>[0];

type ObjectivesGroupedListProps = {
  groups: ObjectiveGroup[];
  teamCodeById: ReadonlyMap<string, string>;
  viewOptions: ObjectiveViewOptions;
  onObjectiveSelect: (objective: Objective) => void;
  onKeyResultSelect: (objective: Objective, keyResult: KeyResult) => void;
  onCreateObjective: () => void;
  selectedObjectiveId?: string;
};

type ObjectiveListEntry =
  | {
      group: ObjectiveGroup;
      key: string;
      type: "group";
    }
  | {
      group: ObjectiveGroup;
      key: string;
      objective: Objective;
      type: "objective";
    };

const getListEntries = (
  groups: ObjectiveGroup[],
  collapsedGroups: ReadonlySet<string>,
) => {
  const entries: ObjectiveListEntry[] = [];

  for (const group of groups) {
    entries.push({
      group,
      key: `group:${group.key}`,
      type: "group",
    });

    if (collapsedGroups.has(group.key)) continue;

    for (const objective of group.objectives) {
      entries.push({
        group,
        key: `objective:${objective.id}`,
        objective,
        type: "objective",
      });
    }
  }

  return entries;
};

export const ObjectivesGroupedList = ({
  groups,
  teamCodeById,
  onKeyResultSelect,
  onObjectiveSelect,
  onCreateObjective,
  selectedObjectiveId,
  viewOptions,
}: ObjectivesGroupedListProps) => {
  const scrollElementRef = useRef<HTMLDivElement>(null);
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(
    () => new Set(),
  );
  const [selectedObjectives, setSelectedObjectives] = useState<string[]>([]);
  const [expandedObjectiveIds, setExpandedObjectiveIds] = useState<Set<string>>(
    () => new Set(),
  );
  const [focusedEntryKey, setFocusedEntryKey] = useState<string>();
  const selectedObjectiveIdSet = useMemo(
    () => new Set(selectedObjectives),
    [selectedObjectives],
  );
  const objectives = useMemo(
    () => groups.flatMap((group) => group.objectives),
    [groups],
  );
  const entries = useMemo(
    () => getListEntries(groups, collapsedGroups),
    [collapsedGroups, groups],
  );
  const toggleGroup = useCallback((groupKey: string) => {
    setCollapsedGroups((current) => {
      const next = new Set(current);
      if (next.has(groupKey)) {
        next.delete(groupKey);
      } else {
        next.add(groupKey);
      }
      return next;
    });
  }, []);
  const setObjectiveExpanded = useCallback(
    (objectiveId: string, expanded: boolean) => {
      setExpandedObjectiveIds((current) => {
        const next = new Set(current);
        if (expanded) {
          next.add(objectiveId);
        } else {
          next.delete(objectiveId);
        }
        return next;
      });
    },
    [],
  );
  const setObjectiveSelected = useCallback(
    (objectiveId: string, checked: boolean) => {
      setSelectedObjectives((current) =>
        checked
          ? Array.from(new Set([...current, objectiveId]))
          : current.filter((id) => id !== objectiveId),
      );
    },
    [],
  );

  return (
    <Box
      className="h-full min-h-0 overflow-x-auto overflow-y-auto pb-6"
      data-body-container
      ref={scrollElementRef}
    >
      <VirtualizedObjectiveItems
        className="min-w-6xl"
        estimatedSize={ESTIMATED_LIST_ROW_HEIGHT}
        getItemKey={(entry) => entry.key}
        items={entries}
        onItemFocus={(entry) => {
          setFocusedEntryKey(entry.key);
        }}
        pinnedKeys={[
          focusedEntryKey,
          selectedObjectiveId ? `objective:${selectedObjectiveId}` : undefined,
        ].filter((entryKey): entryKey is string => Boolean(entryKey))}
        renderItem={(entry) => {
          const { group } = entry;

          if (entry.type === "group") {
            const isCollapsed = collapsedGroups.has(group.key);
            const headerStyle = getCategoryHeaderStyle({
              statusColor: group.status?.color,
              priority: group.priority,
            });

            return (
              <Box
                className="border-border bg-surface-muted/85 dark:border-border/70 border-b-[0.5px] px-12 py-[0.4rem] backdrop-blur"
                style={headerStyle}
              >
                <ObjectiveGroupHeader
                  collapsible
                  group={group}
                  groupBy={viewOptions.groupBy}
                  isCollapsed={isCollapsed}
                  onCreateObjective={onCreateObjective}
                  onToggle={() => {
                    toggleGroup(group.key);
                  }}
                  selectedObjectives={selectedObjectives}
                  setSelectedObjectives={setSelectedObjectives}
                />
              </Box>
            );
          }

          const { objective } = entry;
          return (
            <RoadmapObjectiveListItem
              expanded={expandedObjectiveIds.has(objective.id)}
              objective={objective}
              onExpandedChange={(expanded) => {
                setObjectiveExpanded(objective.id, expanded);
              }}
              onKeyResultSelect={(keyResult) => {
                onKeyResultSelect(objective, keyResult);
              }}
              onObjectiveSelect={() => {
                onObjectiveSelect(objective);
              }}
              onSelectionChange={(checked) => {
                setObjectiveSelected(objective.id, checked);
              }}
              selected={selectedObjectiveIdSet.has(objective.id)}
              teamCode={teamCodeById.get(objective.teamId)}
            />
          );
        }}
        scrollElementRef={scrollElementRef}
      />
      {selectedObjectives.length > 0 ? (
        <ObjectivesToolbar
          objectives={objectives}
          onClear={() => {
            setSelectedObjectives([]);
          }}
          selectedObjectiveIds={selectedObjectives}
        />
      ) : null}
    </Box>
  );
};

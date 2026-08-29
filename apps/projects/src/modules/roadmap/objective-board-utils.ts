import type { StoryPriority } from "@/modules/stories/types";
import type { Objective, ObjectiveStatus } from "@/modules/objectives/types";
import type { Member } from "@/types";

export type ObjectiveGroupBy = "status" | "lead" | "priority";
export type ObjectiveOrderBy = "priority" | "target" | "created" | "updated";
export type ObjectiveOrderDirection = "asc" | "desc";

export type ObjectiveViewOptions = {
  groupBy: ObjectiveGroupBy;
  orderBy: ObjectiveOrderBy;
  orderDirection: ObjectiveOrderDirection;
  showEmptyGroups: boolean;
  hiddenKanbanGroups?: Partial<Record<ObjectiveGroupBy, string[]>>;
};

export type ObjectiveGroup = {
  key: string;
  objectives: Objective[];
  status?: ObjectiveStatus;
  member?: Member;
  priority?: StoryPriority;
};

export const DEFAULT_OBJECTIVE_VIEW_OPTIONS: ObjectiveViewOptions = {
  groupBy: "status",
  orderBy: "priority",
  orderDirection: "desc",
  showEmptyGroups: true,
};

export const UNASSIGNED_GROUP_KEY = "unassigned";

const PRIORITIES: StoryPriority[] = [
  "Urgent",
  "High",
  "Medium",
  "Low",
  "No Priority",
];

const PRIORITY_RANK = new Map(
  PRIORITIES.map((priority, index) => [priority, PRIORITIES.length - index]),
);

const getObjectiveGroupKey = (
  objective: Objective,
  groupBy: ObjectiveGroupBy,
) => {
  switch (groupBy) {
    case "status":
      return objective.statusId;
    case "lead":
      return objective.leadUser || UNASSIGNED_GROUP_KEY;
    case "priority":
      return objective.priority ?? "No Priority";
  }
};

const getSortValue = (objective: Objective, orderBy: ObjectiveOrderBy) => {
  switch (orderBy) {
    case "priority":
      return PRIORITY_RANK.get(objective.priority ?? "No Priority") ?? 0;
    case "target":
      return objective.endDate
        ? new Date(objective.endDate).getTime()
        : Number.MAX_SAFE_INTEGER;
    case "created":
      return new Date(objective.createdAt).getTime();
    case "updated":
      return new Date(objective.updatedAt).getTime();
  }
};

const sortObjectives = (
  objectives: Objective[],
  orderBy: ObjectiveOrderBy,
  orderDirection: ObjectiveOrderDirection,
) => {
  const direction = orderDirection === "asc" ? 1 : -1;

  return [...objectives].sort((left, right) => {
    const difference =
      (getSortValue(left, orderBy) - getSortValue(right, orderBy)) * direction;

    return difference || left.name.localeCompare(right.name);
  });
};

export const groupObjectives = ({
  objectives,
  statuses,
  members,
  viewOptions,
}: {
  objectives: Objective[];
  statuses: ObjectiveStatus[];
  members: Member[];
  viewOptions: ObjectiveViewOptions;
}): ObjectiveGroup[] => {
  const groupedObjectives = new Map<string, Objective[]>();

  for (const objective of objectives) {
    const key = getObjectiveGroupKey(objective, viewOptions.groupBy);
    const group = groupedObjectives.get(key) ?? [];
    group.push(objective);
    groupedObjectives.set(key, group);
  }

  const createGroup = (
    key: string,
    metadata: Omit<ObjectiveGroup, "key" | "objectives">,
  ): ObjectiveGroup => ({
    key,
    objectives: sortObjectives(
      groupedObjectives.get(key) ?? [],
      viewOptions.orderBy,
      viewOptions.orderDirection,
    ),
    ...metadata,
  });

  let groups: ObjectiveGroup[];
  switch (viewOptions.groupBy) {
    case "status":
      groups = [...statuses]
        .sort((left, right) => left.orderIndex - right.orderIndex)
        .map((status) => createGroup(status.id, { status }));
      break;
    case "lead":
      groups = [
        ...members.map((member) => createGroup(member.id, { member })),
        createGroup(UNASSIGNED_GROUP_KEY, {}),
      ];
      break;
    case "priority":
      groups = PRIORITIES.map((priority) =>
        createGroup(priority, { priority }),
      );
      break;
  }

  const knownKeys = new Set(groups.map(({ key }) => key));
  for (const key of Array.from(groupedObjectives.keys())) {
    if (!knownKeys.has(key)) {
      groups.push(createGroup(key, {}));
    }
  }

  return viewOptions.showEmptyGroups
    ? groups
    : groups.filter(({ objectives }) => objectives.length > 0);
};

export const getObjectiveGroupUpdate = (
  groupBy: ObjectiveGroupBy,
  groupKey: string,
) => {
  switch (groupBy) {
    case "status":
      return { statusId: groupKey };
    case "lead":
      return {
        leadUser: groupKey === UNASSIGNED_GROUP_KEY ? null : groupKey,
      };
    case "priority":
      return { priority: groupKey as StoryPriority };
  }
};

export const getHiddenObjectiveGroupKeys = (
  viewOptions: ObjectiveViewOptions,
): string[] => viewOptions.hiddenKanbanGroups?.[viewOptions.groupBy] ?? [];

export const hideObjectiveGroup = (
  viewOptions: ObjectiveViewOptions,
  groupKey: string,
): ObjectiveViewOptions => {
  const hiddenKeys = getHiddenObjectiveGroupKeys(viewOptions);

  return {
    ...viewOptions,
    hiddenKanbanGroups: {
      ...viewOptions.hiddenKanbanGroups,
      [viewOptions.groupBy]: hiddenKeys.includes(groupKey)
        ? hiddenKeys
        : [...hiddenKeys, groupKey],
    },
  };
};

export const showObjectiveGroup = (
  viewOptions: ObjectiveViewOptions,
  groupKey: string,
): ObjectiveViewOptions => ({
  ...viewOptions,
  hiddenKanbanGroups: {
    ...viewOptions.hiddenKanbanGroups,
    [viewOptions.groupBy]: getHiddenObjectiveGroupKeys(viewOptions).filter(
      (hiddenKey) => hiddenKey !== groupKey,
    ),
  },
});

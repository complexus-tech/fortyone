import type {
  StoriesViewOptions,
  ViewOptionsGroupBy,
} from "./stories-view-options-button";

type HideableGroupBy = Exclude<ViewOptionsGroupBy, "none">;

const isHideableGroupBy = (
  groupBy: ViewOptionsGroupBy,
): groupBy is HideableGroupBy => groupBy !== "none";

export const getHiddenKanbanGroupKeys = (
  viewOptions: StoriesViewOptions,
): string[] => {
  if (!isHideableGroupBy(viewOptions.groupBy)) return [];

  return viewOptions.hiddenKanbanGroups?.[viewOptions.groupBy] ?? [];
};

export const hideKanbanGroup = (
  viewOptions: StoriesViewOptions,
  groupKey: string,
): StoriesViewOptions => {
  if (!isHideableGroupBy(viewOptions.groupBy)) return viewOptions;

  const hiddenKeys = getHiddenKanbanGroupKeys(viewOptions);
  const nextHiddenKeys = hiddenKeys.includes(groupKey)
    ? hiddenKeys
    : [...hiddenKeys, groupKey];

  return {
    ...viewOptions,
    hiddenKanbanGroups: {
      ...viewOptions.hiddenKanbanGroups,
      [viewOptions.groupBy]: nextHiddenKeys,
    },
  };
};

export const showKanbanGroup = (
  viewOptions: StoriesViewOptions,
  groupKey: string,
): StoriesViewOptions => {
  if (!isHideableGroupBy(viewOptions.groupBy)) return viewOptions;

  return {
    ...viewOptions,
    hiddenKanbanGroups: {
      ...viewOptions.hiddenKanbanGroups,
      [viewOptions.groupBy]: getHiddenKanbanGroupKeys(viewOptions).filter(
        (hiddenKey) => hiddenKey !== groupKey,
      ),
    },
  };
};

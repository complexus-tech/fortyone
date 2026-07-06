import type { StoriesFilterField } from "./stories-filter-bar";

export const STORIES_FILTER_BUTTON_FIELDS = [
  "assignedToMe",
  "createdByMe",
  "hasNoAssignee",
  "statusIds",
  "assigneeIds",
  "reporterIds",
  "priorities",
  "teamIds",
] as const satisfies readonly StoriesFilterField[];

export type StoriesFilterButtonField =
  (typeof STORIES_FILTER_BUTTON_FIELDS)[number];

export const getVisibleStoriesFilterButtonFields = ({
  hasRouteTeam,
  hiddenFields,
}: {
  hasRouteTeam: boolean;
  hiddenFields: readonly StoriesFilterField[];
}): StoriesFilterButtonField[] => {
  const hiddenFieldSet = new Set(hiddenFields);

  return STORIES_FILTER_BUTTON_FIELDS.filter((field) => {
    if (hiddenFieldSet.has(field)) {
      return false;
    }

    if (field === "teamIds" && hasRouteTeam) {
      return false;
    }

    return true;
  });
};

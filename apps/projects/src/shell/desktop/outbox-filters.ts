import type { QueryClient } from "@tanstack/react-query";
import type { DetailedStory } from "@/modules/story/types";
import { statusKeys } from "@/constants/keys";
import type { DesktopCacheScope } from "./cache-types";
import { isRecord } from "./outbox-records";

export function acceptsStory(
  client: QueryClient,
  scope: DesktopCacheScope,
  story: DetailedStory,
  filters: Record<string, unknown>,
) {
  if (
    story.parentId &&
    !filters.showSubStories &&
    filters.parentId !== story.parentId
  )
    return false;
  const mine = [
    filters.assignedToMe === true && story.assigneeId === scope.accountId,
    filters.createdByMe === true && story.reporterId === scope.accountId,
    filters.collaboratingWithMe === true &&
      (story.collaboratorIds.includes(scope.accountId) ||
        story.collaborators.some((user) => user.id === scope.accountId)),
  ];
  if (
    (filters.assignedToMe ||
      filters.createdByMe ||
      filters.collaboratingWithMe) &&
    !mine.some(Boolean)
  )
    return false;
  if (story.archivedAt && !filters.includeArchived) return false;
  if (story.deletedAt && !filters.includeDeleted) return false;
  const ignored = new Set([
    "groupBy",
    "groupKey",
    "orderBy",
    "orderDirection",
    "page",
    "pageSize",
    "storiesPerGroup",
    "showSubStories",
    "includeArchived",
    "includeDeleted",
    "assignedToMe",
    "createdByMe",
    "collaboratingWithMe",
  ]);
  for (const [key, value] of Object.entries(filters)) {
    if (
      ignored.has(key) ||
      value === undefined ||
      value === null ||
      value === false ||
      (Array.isArray(value) && value.length === 0)
    )
      continue;
    if (key === "hasNoAssignee") {
      if (story.assigneeId) return false;
      continue;
    }
    if (key === "hasAssignee") {
      if (!story.assigneeId) return false;
      continue;
    }
    if (key === "categories" && Array.isArray(value)) {
      const statuses = client.getQueryData<{ id: string; category: string }[]>(
        statusKeys.lists(scope.workspaceSlug),
      );
      if (
        !value.includes(
          statuses?.find((status) => status.id === story.statusId)?.category,
        )
      )
        return false;
      continue;
    }
    if (key === "titleContains" || key === "titleNotContains") {
      const contains = story.title
        .toLowerCase()
        .includes(String(value).toLowerCase());
      if (contains === (key === "titleNotContains")) return false;
      continue;
    }
    const listFields: Partial<Record<string, keyof DetailedStory>> = {
      statusIds: "statusId",
      assigneeIds: "assigneeId",
      reporterIds: "reporterId",
      priorities: "priority",
      teamIds: "teamId",
      sprintIds: "sprintId",
      estimateValues: "estimateValue",
      labelIds: "labels",
    };
    const excluded = key.startsWith("excluded");
    const normalized = excluded
      ? key.slice(8, 9).toLowerCase() + key.slice(9)
      : key;
    const field = listFields[normalized];
    if (field && Array.isArray(value)) {
      const current = story[field];
      const matches = Array.isArray(current)
        ? current.some((item) => value.includes(item))
        : value.includes(current);
      if (matches === excluded) return false;
      continue;
    }
    if (["parentId", "objectiveId", "keyResultId", "epicId"].includes(key)) {
      if (story[key as "parentId"] !== value) return false;
      continue;
    }
    // Unknown advanced filters are left to the server. The item still appears
    // in the durable pending-work list rather than an unrelated filtered view.
    return false;
  }
  return true;
}

export function queryFilters(
  key: readonly unknown[],
  fallback: Record<string, unknown> = {},
) {
  const last = key.at(-1);
  const filters = { ...(isRecord(last) ? last : {}), ...fallback };
  if (key[2] === "list" && key[3] === "team") filters.teamIds = [key[4]];
  if (key[2] === "objectives") filters.objectiveId = key[3];
  if (key[2] === "sprints") filters.sprintIds = [key[3]];
  if (
    key[2] === "list" &&
    key[3] === "mine" &&
    !filters.assignedToMe &&
    !filters.createdByMe &&
    !filters.collaboratingWithMe
  )
    filters.assignedToMe = true;
  return filters;
}
